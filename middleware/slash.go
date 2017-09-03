package middleware

import (
	"github.com/itchenyi/akita"
)

type (
	// TrailingSlashConfig defines the config for TrailingSlash middleware.
	TrailingSlashConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		// Status code to be used when redirecting the request.
		// Optional, but when provided the request is redirected using this code.
		RedirectCode int `json:"redirect_code"`
	}
)

var (
	// DefaultTrailingSlashConfig is the default TrailingSlash middleware config.
	DefaultTrailingSlashConfig = TrailingSlashConfig{
		Skipper: DefaultSkipper,
	}
)

// AddTrailingSlash returns a root level (before router) middleware which adds a
// trailing slash to the request `URL#Path`.
//
// Usage `Akita#Pre(AddTrailingSlash())`
func AddTrailingSlash() akita.MiddlewareFunc {
	return AddTrailingSlashWithConfig(DefaultTrailingSlashConfig)
}

// AddTrailingSlashWithConfig returns a AddTrailingSlash middleware with config.
// See `AddTrailingSlash()`.
func AddTrailingSlashWithConfig(config TrailingSlashConfig) akita.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultTrailingSlashConfig.Skipper
	}

	return func(next akita.HandlerFunc) akita.HandlerFunc {
		return func(ctx akita.Context) error {
			if config.Skipper(ctx) {
				return next(ctx)
			}

			req := ctx.Request()
			url := req.URL
			path := url.Path
			qs := ctx.QueryString()
			if path != "/" && path[len(path)-1] != '/' {
				path += "/"
				uri := path
				if qs != "" {
					uri += "?" + qs
				}

				// Redirect
				if config.RedirectCode != 0 {
					return ctx.Redirect(config.RedirectCode, uri)
				}

				// Forward
				req.RequestURI = uri
				url.Path = path
			}
			return next(ctx)
		}
	}
}

// RemoveTrailingSlash returns a root level (before router) middleware which removes
// a trailing slash from the request URI.
//
// Usage `Akita#Pre(RemoveTrailingSlash())`
func RemoveTrailingSlash() akita.MiddlewareFunc {
	return RemoveTrailingSlashWithConfig(TrailingSlashConfig{})
}

// RemoveTrailingSlashWithConfig returns a RemoveTrailingSlash middleware with config.
// See `RemoveTrailingSlash()`.
func RemoveTrailingSlashWithConfig(config TrailingSlashConfig) akita.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultTrailingSlashConfig.Skipper
	}

	return func(next akita.HandlerFunc) akita.HandlerFunc {
		return func(ctx akita.Context) error {
			if config.Skipper(ctx) {
				return next(ctx)
			}

			req := ctx.Request()
			url := req.URL
			path := url.Path
			qs := ctx.QueryString()
			l := len(path) - 1
			if l >= 0 && path != "/" && path[l] == '/' {
				path = path[:l]
				uri := path
				if qs != "" {
					uri += "?" + qs
				}

				// Redirect
				if config.RedirectCode != 0 {
					return ctx.Redirect(config.RedirectCode, uri)
				}

				// Forward
				req.RequestURI = uri
				url.Path = path
			}
			return next(ctx)
		}
	}
}
