package middleware

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/itchenyi/akita"
	"github.com/itchenyi/common/random"
)

type (
	// CSRFConfig defines the config for CSRF middleware.
	CSRFConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		// TokenLength is the length of the generated token.
		TokenLength uint8 `json:"token_length"`
		// Optional. Default value 32.

		// TokenLookup is a string in the form of "<source>:<key>" that is used
		// to extract token from the request.
		// Optional. Default value "header:X-CSRF-Token".
		// Possible values:
		// - "header:<name>"
		// - "form:<name>"
		// - "query:<name>"
		TokenLookup string `json:"token_lookup"`

		// Context key to store generated CSRF token into context.
		// Optional. Default value "csrf".
		ContextKey string `json:"context_key"`

		// Name of the CSRF cookie. This cookie will store CSRF token.
		// Optional. Default value "csrf".
		CookieName string `json:"cookie_name"`

		// Domain of the CSRF cookie.
		// Optional. Default value none.
		CookieDomain string `json:"cookie_domain"`

		// Path of the CSRF cookie.
		// Optional. Default value none.
		CookiePath string `json:"cookie_path"`

		// Max age (in seconds) of the CSRF cookie.
		// Optional. Default value 86400 (24hr).
		CookieMaxAge int `json:"cookie_max_age"`

		// Indicates if CSRF cookie is secure.
		// Optional. Default value false.
		CookieSecure bool `json:"cookie_secure"`

		// Indicates if CSRF cookie is HTTP only.
		// Optional. Default value false.
		CookieHTTPOnly bool `json:"cookie_http_only"`
	}

	// csrfTokenExtractor defines a function that takes `akita.Context` and returns
	// either a token or an error.
	csrfTokenExtractor func(akita.Context) (string, error)
)

var (
	// DefaultCSRFConfig is the default CSRF middleware config.
	DefaultCSRFConfig = CSRFConfig{
		Skipper:      DefaultSkipper,
		TokenLength:  32,
		TokenLookup:  "header:" + akita.HeaderXCSRFToken,
		ContextKey:   "csrf",
		CookieName:   "_csrf",
		CookieMaxAge: 86400,
	}
)

// CSRF returns a Cross-Site Request Forgery (CSRF) middleware.
// See: https://en.wikipedia.org/wiki/Cross-site_request_forgery
func CSRF() akita.MiddlewareFunc {
	ctx := DefaultCSRFConfig
	return CSRFWithConfig(ctx)
}

// CSRFWithConfig returns a CSRF middleware with config.
// See `CSRF()`.
func CSRFWithConfig(config CSRFConfig) akita.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultCSRFConfig.Skipper
	}
	if config.TokenLength == 0 {
		config.TokenLength = DefaultCSRFConfig.TokenLength
	}
	if config.TokenLookup == "" {
		config.TokenLookup = DefaultCSRFConfig.TokenLookup
	}
	if config.ContextKey == "" {
		config.ContextKey = DefaultCSRFConfig.ContextKey
	}
	if config.CookieName == "" {
		config.CookieName = DefaultCSRFConfig.CookieName
	}
	if config.CookieMaxAge == 0 {
		config.CookieMaxAge = DefaultCSRFConfig.CookieMaxAge
	}

	// Initialize
	parts := strings.Split(config.TokenLookup, ":")
	extractor := csrfTokenFromHeader(parts[1])
	switch parts[0] {
	case "form":
		extractor = csrfTokenFromForm(parts[1])
	case "query":
		extractor = csrfTokenFromQuery(parts[1])
	}

	return func(next akita.HandlerFunc) akita.HandlerFunc {
		return func(ctx akita.Context) error {
			if config.Skipper(ctx) {
				return next(ctx)
			}

			req := ctx.Request()
			k, err := ctx.Cookie(config.CookieName)
			token := ""

			if err != nil {
				// Generate token
				token = random.String(config.TokenLength)
			} else {
				// Reuse token
				token = k.Value
			}

			switch req.Method {
			case akita.GET, akita.HEAD, akita.OPTIONS, akita.TRACE:
			default:
				// Validate token only for requests which are not defined as 'safe' by RFC7231
				clientToken, err := extractor(ctx)
				if err != nil {
					return akita.NewHTTPError(http.StatusBadRequest, err.Error())
				}
				if !validateCSRFToken(token, clientToken) {
					return akita.NewHTTPError(http.StatusForbidden, "Invalid csrf token")
				}
			}

			// Set CSRF cookie
			cookie := new(http.Cookie)
			cookie.Name = config.CookieName
			cookie.Value = token
			if config.CookiePath != "" {
				cookie.Path = config.CookiePath
			}
			if config.CookieDomain != "" {
				cookie.Domain = config.CookieDomain
			}
			cookie.Expires = time.Now().Add(time.Duration(config.CookieMaxAge) * time.Second)
			cookie.Secure = config.CookieSecure
			cookie.HttpOnly = config.CookieHTTPOnly
			ctx.SetCookie(cookie)

			// Store token in the context
			ctx.Set(config.ContextKey, token)

			// Protect clients from caching the response
			ctx.Response().Header().Add(akita.HeaderVary, akita.HeaderCookie)

			return next(ctx)
		}
	}
}

// csrfTokenFromForm returns a `csrfTokenExtractor` that extracts token from the
// provided request header.
func csrfTokenFromHeader(header string) csrfTokenExtractor {
	return func(ctx akita.Context) (string, error) {
		return ctx.Request().Header.Get(header), nil
	}
}

// csrfTokenFromForm returns a `csrfTokenExtractor` that extracts token from the
// provided form parameter.
func csrfTokenFromForm(param string) csrfTokenExtractor {
	return func(ctx akita.Context) (string, error) {
		token := ctx.FormValue(param)
		if token == "" {
			return "", errors.New("Missing csrf token in the form parameter")
		}
		return token, nil
	}
}

// csrfTokenFromQuery returns a `csrfTokenExtractor` that extracts token from the
// provided query parameter.
func csrfTokenFromQuery(param string) csrfTokenExtractor {
	return func(ctx akita.Context) (string, error) {
		token := ctx.QueryParam(param)
		if token == "" {
			return "", errors.New("Missing csrf token in the query string")
		}
		return token, nil
	}
}

func validateCSRFToken(token, clientToken string) bool {
	return subtle.ConstantTimeCompare([]byte(token), []byte(clientToken)) == 1
}
