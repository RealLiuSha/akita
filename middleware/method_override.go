package middleware

import "github.com/itchenyi/akita"

type (
	// MethodOverrideConfig defines the config for MethodOverride middleware.
	MethodOverrideConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		// Getter is a function that gets overridden method from the request.
		// Optional. Default values MethodFromHeader(akita.HeaderXHTTPMethodOverride).
		Getter MethodOverrideGetter
	}

	// MethodOverrideGetter is a function that gets overridden method from the request
	MethodOverrideGetter func(akita.Context) string
)

var (
	// DefaultMethodOverrideConfig is the default MethodOverride middleware config.
	DefaultMethodOverrideConfig = MethodOverrideConfig{
		Skipper: DefaultSkipper,
		Getter:  MethodFromHeader(akita.HeaderXHTTPMethodOverride),
	}
)

// MethodOverride returns a MethodOverride middleware.
// MethodOverride  middleware checks for the overridden method from the request and
// uses it instead of the original method.
//
// For security reasons, only `POST` method can be overridden.
func MethodOverride() akita.MiddlewareFunc {
	return MethodOverrideWithConfig(DefaultMethodOverrideConfig)
}

// MethodOverrideWithConfig returns a MethodOverride middleware with config.
// See: `MethodOverride()`.
func MethodOverrideWithConfig(config MethodOverrideConfig) akita.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultMethodOverrideConfig.Skipper
	}
	if config.Getter == nil {
		config.Getter = DefaultMethodOverrideConfig.Getter
	}

	return func(next akita.HandlerFunc) akita.HandlerFunc {
		return func(ctx akita.Context) error {
			if config.Skipper(ctx) {
				return next(ctx)
			}

			req := ctx.Request()
			if req.Method == akita.POST {
				m := config.Getter(ctx)
				if m != "" {
					req.Method = m
				}
			}
			return next(ctx)
		}
	}
}

// MethodFromHeader is a `MethodOverrideGetter` that gets overridden method from
// the request header.
func MethodFromHeader(header string) MethodOverrideGetter {
	return func(ctx akita.Context) string {
		return ctx.Request().Header.Get(header)
	}
}

// MethodFromForm is a `MethodOverrideGetter` that gets overridden method from the
// form parameter.
func MethodFromForm(param string) MethodOverrideGetter {
	return func(ctx akita.Context) string {
		return ctx.FormValue(param)
	}
}

// MethodFromQuery is a `MethodOverrideGetter` that gets overridden method from
// the query parameter.
func MethodFromQuery(param string) MethodOverrideGetter {
	return func(ctx akita.Context) string {
		return ctx.QueryParam(param)
	}
}
