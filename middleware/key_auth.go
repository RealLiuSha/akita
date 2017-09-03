package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/itchenyi/akita"
)

type (
	// KeyAuthConfig defines the config for KeyAuth middleware.
	KeyAuthConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		// KeyLookup is a string in the form of "<source>:<name>" that is used
		// to extract key from the request.
		// Optional. Default value "header:Authorization".
		// Possible values:
		// - "header:<name>"
		// - "query:<name>"
		KeyLookup string `json:"key_lookup"`

		// AuthScheme to be used in the Authorization header.
		// Optional. Default value "Bearer".
		AuthScheme string

		// Validator is a function to validate key.
		// Required.
		Validator KeyAuthValidator
	}

	// KeyAuthValidator defines a function to validate KeyAuth credentials.
	KeyAuthValidator func(string, akita.Context) (bool, error)

	keyExtractor func(akita.Context) (string, error)
)

var (
	// DefaultKeyAuthConfig is the default KeyAuth middleware config.
	DefaultKeyAuthConfig = KeyAuthConfig{
		Skipper:    DefaultSkipper,
		KeyLookup:  "header:" + akita.HeaderAuthorization,
		AuthScheme: "Bearer",
	}
)

// KeyAuth returns an KeyAuth middleware.
//
// For valid key it calls the next handler.
// For invalid key, it sends "401 - Unauthorized" response.
// For missing key, it sends "400 - Bad Request" response.
func KeyAuth(fn KeyAuthValidator) akita.MiddlewareFunc {
	c := DefaultKeyAuthConfig
	c.Validator = fn
	return KeyAuthWithConfig(c)
}

// KeyAuthWithConfig returns an KeyAuth middleware with config.
// See `KeyAuth()`.
func KeyAuthWithConfig(config KeyAuthConfig) akita.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultKeyAuthConfig.Skipper
	}
	// Defaults
	if config.AuthScheme == "" {
		config.AuthScheme = DefaultKeyAuthConfig.AuthScheme
	}
	if config.KeyLookup == "" {
		config.KeyLookup = DefaultKeyAuthConfig.KeyLookup
	}
	if config.Validator == nil {
		panic("akita: key-auth middleware requires a validator function")
	}

	// Initialize
	parts := strings.Split(config.KeyLookup, ":")
	extractor := keyFromHeader(parts[1], config.AuthScheme)
	switch parts[0] {
	case "query":
		extractor = keyFromQuery(parts[1])
	}

	return func(next akita.HandlerFunc) akita.HandlerFunc {
		return func(ctx akita.Context) error {
			if config.Skipper(ctx) {
				return next(ctx)
			}

			// Extract and verify key
			key, err := extractor(ctx)
			if err != nil {
				return akita.NewHTTPError(http.StatusBadRequest, err.Error())
			}
			valid, err := config.Validator(key, ctx)
			if err != nil {
				return err
			} else if valid {
				return next(ctx)
			}

			return akita.ErrUnauthorized
		}
	}
}

// keyFromHeader returns a `keyExtractor` that extracts key from the request header.
func keyFromHeader(header string, authScheme string) keyExtractor {
	return func(ctx akita.Context) (string, error) {
		auth := ctx.Request().Header.Get(header)
		if auth == "" {
			return "", errors.New("Missing key in request header")
		}
		if header == akita.HeaderAuthorization {
			l := len(authScheme)
			if len(auth) > l+1 && auth[:l] == authScheme {
				return auth[l+1:], nil
			}
			return "", errors.New("Invalid key in the request header")
		}
		return auth, nil
	}
}

// keyFromQuery returns a `keyExtractor` that extracts key from the query string.
func keyFromQuery(param string) keyExtractor {
	return func(ctx akita.Context) (string, error) {
		key := ctx.QueryParam(param)
		if key == "" {
			return "", errors.New("Missing key in the query string")
		}
		return key, nil
	}
}
