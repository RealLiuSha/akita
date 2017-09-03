package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/itchenyi/akita"
	"github.com/stretchr/testify/assert"
)

func TestKeyAuth(t *testing.T) {
	a := akita.New()
	req := httptest.NewRequest(akita.GET, "/", nil)
	res := httptest.NewRecorder()
	ctx := a.NewContext(req, res)
	config := KeyAuthConfig{
		Validator: func(key string, ctx akita.Context) (bool, error) {
			return key == "valid-key", nil
		},
	}
	h := KeyAuthWithConfig(config)(func(ctx akita.Context) error {
		return ctx.String(http.StatusOK, "test")
	})

	// Valid key
	auth := DefaultKeyAuthConfig.AuthScheme + " " + "valid-key"
	req.Header.Set(akita.HeaderAuthorization, auth)
	assert.NoError(t, h(ctx))

	// Invalid key
	auth = DefaultKeyAuthConfig.AuthScheme + " " + "invalid-key"
	req.Header.Set(akita.HeaderAuthorization, auth)
	he := h(ctx).(*akita.HTTPError)
	assert.Equal(t, http.StatusUnauthorized, he.Code)

	// Missing Authorization header
	req.Header.Del(akita.HeaderAuthorization)
	he = h(ctx).(*akita.HTTPError)
	assert.Equal(t, http.StatusBadRequest, he.Code)

	// Key from custom header
	config.KeyLookup = "header:API-Key"
	h = KeyAuthWithConfig(config)(func(ctx akita.Context) error {
		return ctx.String(http.StatusOK, "test")
	})
	req.Header.Set("API-Key", "valid-key")
	assert.NoError(t, h(ctx))

	// Key from query string
	config.KeyLookup = "query:key"
	h = KeyAuthWithConfig(config)(func(ctx akita.Context) error {
		return ctx.String(http.StatusOK, "test")
	})
	q := req.URL.Query()
	q.Add("key", "valid-key")
	req.URL.RawQuery = q.Encode()
	assert.NoError(t, h(ctx))
}
