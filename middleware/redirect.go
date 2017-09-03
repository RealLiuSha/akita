package middleware

import (
	"net/http"

	"github.com/itchenyi/akita"
)

type (
	// RedirectConfig defines the config for Redirect middleware.
	RedirectConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		// Status code to be used when redirecting the request.
		// Optional. Default value http.StatusMovedPermanently.
		Code int `json:"code"`
	}
)

const (
	www = "www"
)

var (
	// DefaultRedirectConfig is the default Redirect middleware config.
	DefaultRedirectConfig = RedirectConfig{
		Skipper: DefaultSkipper,
		Code:    http.StatusMovedPermanently,
	}
)

// HTTPSRedirect redirects http requests to https.
// For example, http://liusha.me will be redirect to https://liusha.me.
//
// Usage `Akita#Pre(HTTPSRedirect())`
func HTTPSRedirect() akita.MiddlewareFunc {
	return HTTPSRedirectWithConfig(DefaultRedirectConfig)
}

// HTTPSRedirectWithConfig returns an HTTPSRedirect middleware with config.
// See `HTTPSRedirect()`.
func HTTPSRedirectWithConfig(config RedirectConfig) akita.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultTrailingSlashConfig.Skipper
	}
	if config.Code == 0 {
		config.Code = DefaultRedirectConfig.Code
	}

	return func(next akita.HandlerFunc) akita.HandlerFunc {
		return func(c akita.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			req := c.Request()
			host := req.Host
			uri := req.RequestURI
			if !c.IsTLS() {
				return c.Redirect(config.Code, "https://"+host+uri)
			}
			return next(c)
		}
	}
}

// HTTPSWWWRedirect redirects http requests to https www.
// For example, http://liusha.me.com will be redirect to https://www.liusha.me.
//
// Usage `Akita#Pre(HTTPSWWWRedirect())`
func HTTPSWWWRedirect() akita.MiddlewareFunc {
	return HTTPSWWWRedirectWithConfig(DefaultRedirectConfig)
}

// HTTPSWWWRedirectWithConfig returns an HTTPSRedirect middleware with config.
// See `HTTPSWWWRedirect()`.
func HTTPSWWWRedirectWithConfig(config RedirectConfig) akita.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultTrailingSlashConfig.Skipper
	}
	if config.Code == 0 {
		config.Code = DefaultRedirectConfig.Code
	}

	return func(next akita.HandlerFunc) akita.HandlerFunc {
		return func(ctx akita.Context) error {
			if config.Skipper(ctx) {
				return next(ctx)
			}

			req := ctx.Request()
			host := req.Host
			uri := req.RequestURI
			if !ctx.IsTLS() && host[:3] != www {
				return ctx.Redirect(config.Code, "https://www."+host+uri)
			}
			return next(ctx)
		}
	}
}

// HTTPSNonWWWRedirect redirects http requests to https non www.
// For example, http://www.liusha.me will be redirect to https://liusha.me.
//
// Usage `Akita#Pre(HTTPSNonWWWRedirect())`
func HTTPSNonWWWRedirect() akita.MiddlewareFunc {
	return HTTPSNonWWWRedirectWithConfig(DefaultRedirectConfig)
}

// HTTPSNonWWWRedirectWithConfig returns an HTTPSRedirect middleware with config.
// See `HTTPSNonWWWRedirect()`.
func HTTPSNonWWWRedirectWithConfig(config RedirectConfig) akita.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultTrailingSlashConfig.Skipper
	}
	if config.Code == 0 {
		config.Code = DefaultRedirectConfig.Code
	}

	return func(next akita.HandlerFunc) akita.HandlerFunc {
		return func(ctx akita.Context) error {
			if config.Skipper(ctx) {
				return next(ctx)
			}

			req := ctx.Request()
			host := req.Host
			uri := req.RequestURI
			if !ctx.IsTLS() {
				if host[:3] == www {
					return ctx.Redirect(config.Code, "https://"+host[4:]+uri)
				}
				return ctx.Redirect(config.Code, "https://"+host+uri)
			}
			return next(ctx)
		}
	}
}

// WWWRedirect redirects non www requests to www.
// For example, http://liusha.me will be redirect to http://www.liusha.me.
//
// Usage `Akita#Pre(WWWRedirect())`
func WWWRedirect() akita.MiddlewareFunc {
	return WWWRedirectWithConfig(DefaultRedirectConfig)
}

// WWWRedirectWithConfig returns an HTTPSRedirect middleware with config.
// See `WWWRedirect()`.
func WWWRedirectWithConfig(config RedirectConfig) akita.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultTrailingSlashConfig.Skipper
	}
	if config.Code == 0 {
		config.Code = DefaultRedirectConfig.Code
	}

	return func(next akita.HandlerFunc) akita.HandlerFunc {
		return func(c akita.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			req := c.Request()
			scheme := c.Scheme()
			host := req.Host
			if host[:3] != www {
				uri := req.RequestURI
				return c.Redirect(config.Code, scheme+"://www."+host+uri)
			}
			return next(c)
		}
	}
}

// NonWWWRedirect redirects www requests to non www.
// For example, http://www.liusha.me will be redirect to http://liusha.me.
//
// Usage `Akita#Pre(NonWWWRedirect())`
func NonWWWRedirect() akita.MiddlewareFunc {
	return NonWWWRedirectWithConfig(DefaultRedirectConfig)
}

// NonWWWRedirectWithConfig returns an HTTPSRedirect middleware with config.
// See `NonWWWRedirect()`.
func NonWWWRedirectWithConfig(config RedirectConfig) akita.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = DefaultTrailingSlashConfig.Skipper
	}
	if config.Code == 0 {
		config.Code = DefaultRedirectConfig.Code
	}

	return func(next akita.HandlerFunc) akita.HandlerFunc {
		return func(ctx akita.Context) error {
			if config.Skipper(ctx) {
				return next(ctx)
			}

			req := ctx.Request()
			scheme := ctx.Scheme()
			host := req.Host
			if host[:3] == www {
				uri := req.RequestURI
				return ctx.Redirect(config.Code, scheme+"://"+host[4:]+uri)
			}
			return next(ctx)
		}
	}
}
