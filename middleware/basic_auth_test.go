package middleware

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/itchenyi/akita"
	"github.com/stretchr/testify/assert"
)

func TestBasicAuth(t *testing.T) {
	a := akita.New()
	req := httptest.NewRequest(akita.GET, "/", nil)
	res := httptest.NewRecorder()
	ctx := a.NewContext(req, res)
	f := func(u, p string, ctx akita.Context) (bool, error) {
		if u == "joe" && p == "secret" {
			return true, nil
		}
		return false, nil
	}
	h := BasicAuth(f)(func(ctx akita.Context) error {
		return ctx.String(http.StatusOK, "test")
	})

	// Valid credentials
	auth := basic + " " + base64.StdEncoding.EncodeToString([]byte("joe:secret"))
	req.Header.Set(akita.HeaderAuthorization, auth)
	assert.NoError(t, h(ctx))

	// Invalid credentials
	auth = basic + " " + base64.StdEncoding.EncodeToString([]byte("joe:invalid-password"))
	req.Header.Set(akita.HeaderAuthorization, auth)
	he := h(ctx).(*akita.HTTPError)
	assert.Equal(t, http.StatusUnauthorized, he.Code)
	assert.Equal(t, basic+" realm=Restricted", res.Header().Get(akita.HeaderWWWAuthenticate))

	// Missing Authorization header
	req.Header.Del(akita.HeaderAuthorization)
	he = h(ctx).(*akita.HTTPError)
	assert.Equal(t, http.StatusUnauthorized, he.Code)

	// Invalid Authorization header
	auth = base64.StdEncoding.EncodeToString([]byte("invalid"))
	req.Header.Set(akita.HeaderAuthorization, auth)
	he = h(ctx).(*akita.HTTPError)
	assert.Equal(t, http.StatusUnauthorized, he.Code)
}
