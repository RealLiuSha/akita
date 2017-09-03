package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/itchenyi/akita"
)

type (
	// CORSConfig defines the config for CORS middleware.
	CORSConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		// AllowOrigin defines a list of origins that may access the resource.
		// Optional. Default value []string{"*"}.
		AllowOrigins []string `json:"allow_origins"`

		// AllowMethods defines a list methods allowed when accessing the resource.
		// This is used in response to a preflight request.
		// Optional. Default value DefaultCORSConfig.AllowMethods.
		AllowMethods []string `json:"allow_methods"`

		// AllowHeaders defines a list of request headers that can be used when
		// making the actual request. This in response to a preflight request.
		// Optional. Default value []string{}.
		AllowHeaders []string `json:"allow_headers"`

		// AllowCredentials indicates whether or not the response to the request
		// can be exposed when the credentials flag is true. When used as part of
		// a response to a preflight request, this indicates whether or not the
		// actual request can be made using credentials.
		// Optional. Default value false.
		AllowCredentials bool `json:"allow_credentials"`

		// ExposeHeaders defines a whitelist headers that clients are allowed to
		// access.
		// Optional. Default value []string{}.
		ExposeHeaders []string `json:"expose_headers"`

		// MaxAge indicates how long (in seconds) the results of a preflight request
		// can be cached.
		// Optional. Default value 0.
		MaxAge int `json:"max_age"`
	}
)

var (
	// DefaultCORSConfig is the default CORS middleware config.
	DefaultCORSConfig = CORSConfig{
		Skipper:      DefaultSkipper,
		AllowOrigins: []string{"*"},
		AllowMethods: []string{akita.GET, akita.HEAD, akita.PUT, akita.PATCH, akita.POST, akita.DELETE},
	}
)

// CORS returns a Cross-Origin Resource Sharing (CORS) middleware.
// See: https://developer.mozilla.org/en/docs/Web/HTTP/Access_control_CORS
func CORS() akita.MiddlewareFunc {
	return CORSWithConfig(DefaultCORSConfig)
}

// CORSWithConfig returns a CORS middleware with config.
// See: `CORS()`.
func CORSWithConfig(config CORSConfig) akita.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultCORSConfig.Skipper
	}
	if len(config.AllowOrigins) == 0 {
		config.AllowOrigins = DefaultCORSConfig.AllowOrigins
	}
	if len(config.AllowMethods) == 0 {
		config.AllowMethods = DefaultCORSConfig.AllowMethods
	}

	allowMethods := strings.Join(config.AllowMethods, ",")
	allowHeaders := strings.Join(config.AllowHeaders, ",")
	exposeHeaders := strings.Join(config.ExposeHeaders, ",")
	maxAge := strconv.Itoa(config.MaxAge)

	return func(next akita.HandlerFunc) akita.HandlerFunc {
		return func(ctx akita.Context) error {
			if config.Skipper(ctx) {
				return next(ctx)
			}

			req := ctx.Request()
			res := ctx.Response()
			origin := req.Header.Get(akita.HeaderOrigin)
			allowOrigin := ""

			// Check allowed origins
			for _, o := range config.AllowOrigins {
				if o == "*" || o == origin {
					allowOrigin = o
					break
				}
			}

			// Simple request
			if req.Method != akita.OPTIONS {
				res.Header().Add(akita.HeaderVary, akita.HeaderOrigin)
				res.Header().Set(akita.HeaderAccessControlAllowOrigin, allowOrigin)
				if config.AllowCredentials {
					res.Header().Set(akita.HeaderAccessControlAllowCredentials, "true")
				}
				if exposeHeaders != "" {
					res.Header().Set(akita.HeaderAccessControlExposeHeaders, exposeHeaders)
				}
				return next(ctx)
			}

			// Preflight request
			res.Header().Add(akita.HeaderVary, akita.HeaderOrigin)
			res.Header().Add(akita.HeaderVary, akita.HeaderAccessControlRequestMethod)
			res.Header().Add(akita.HeaderVary, akita.HeaderAccessControlRequestHeaders)
			res.Header().Set(akita.HeaderAccessControlAllowOrigin, allowOrigin)
			res.Header().Set(akita.HeaderAccessControlAllowMethods, allowMethods)
			if config.AllowCredentials {
				res.Header().Set(akita.HeaderAccessControlAllowCredentials, "true")
			}
			if allowHeaders != "" {
				res.Header().Set(akita.HeaderAccessControlAllowHeaders, allowHeaders)
			} else {
				h := req.Header.Get(akita.HeaderAccessControlRequestHeaders)
				if h != "" {
					res.Header().Set(akita.HeaderAccessControlAllowHeaders, h)
				}
			}
			if config.MaxAge > 0 {
				res.Header().Set(akita.HeaderAccessControlMaxAge, maxAge)
			}
			return ctx.NoContent(http.StatusNoContent)
		}
	}
}
