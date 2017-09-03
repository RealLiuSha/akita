package middleware

import (
	"encoding/base64"
	"strconv"

	"github.com/itchenyi/akita"
)

type (
	// BasicAuthConfig defines the config for BasicAuth middleware.
	BasicAuthConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		// Validator is a function to validate BasicAuth credentials.
		// Required.
		Validator BasicAuthValidator

		// Realm is a string to define realm attribute of BasicAuth.
		// Default value "Restricted".
		Realm string
	}

	// BasicAuthValidator defines a function to validate BasicAuth credentials.
	BasicAuthValidator func(string, string, akita.Context) (bool, error)
)

const (
	basic        = "Basic"
	defaultRealm = "Restricted"
)

var (
	// DefaultBasicAuthConfig is the default BasicAuth middleware config.
	DefaultBasicAuthConfig = BasicAuthConfig{
		Skipper: DefaultSkipper,
		Realm:   defaultRealm,
	}
)

// BasicAuth returns an BasicAuth middleware.
//
// For valid credentials it calls the next handler.
// For missing or invalid credentials, it sends "401 - Unauthorized" response.
func BasicAuth(fn BasicAuthValidator) akita.MiddlewareFunc {
	c := DefaultBasicAuthConfig
	c.Validator = fn
	return BasicAuthWithConfig(c)
}

// BasicAuthWithConfig returns an BasicAuth middleware with config.
// See `BasicAuth()`.
func BasicAuthWithConfig(config BasicAuthConfig) akita.MiddlewareFunc {
	// Defaults
	if config.Validator == nil {
		panic("akita: basic-auth middleware requires a validator function")
	}
	if config.Skipper == nil {
		config.Skipper = DefaultBasicAuthConfig.Skipper
	}
	if config.Realm == "" {
		config.Realm = defaultRealm
	}

	return func(next akita.HandlerFunc) akita.HandlerFunc {
		return func(ctx akita.Context) error {
			if config.Skipper(ctx) {
				return next(ctx)
			}

			auth := ctx.Request().Header.Get(akita.HeaderAuthorization)
			l := len(basic)

			if len(auth) > l+1 && auth[:l] == basic {
				b, err := base64.StdEncoding.DecodeString(auth[l+1:])
				if err != nil {
					return err
				}
				cred := string(b)
				for i := 0; i < len(cred); i++ {
					if cred[i] == ':' {
						// Verify credentials
						valid, err := config.Validator(cred[:i], cred[i+1:], ctx)
						if err != nil {
							return err
						} else if valid {
							return next(ctx)
						}
					}
				}
			}

			realm := ""
			if config.Realm == defaultRealm {
				realm = defaultRealm
			} else {
				realm = strconv.Quote(config.Realm)
			}

			// Need to return `401` for browsers to pop-up login box.
			ctx.Response().Header().Set(akita.HeaderWWWAuthenticate, basic+" realm="+realm)
			return akita.ErrUnauthorized
		}
	}
}
