package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/itchenyi/akita"
	"github.com/stretchr/testify/assert"
)

func TestMethodOverride(t *testing.T) {
	a := akita.New()
	m := MethodOverride()
	h := func(ctx akita.Context) error {
		return ctx.String(http.StatusOK, "test")
	}

	// Override with http header
	req := httptest.NewRequest(akita.POST, "/", nil)
	rec := httptest.NewRecorder()
	req.Header.Set(akita.HeaderXHTTPMethodOverride, akita.DELETE)
	ctx := a.NewContext(req, rec)
	m(h)(ctx)
	assert.Equal(t, akita.DELETE, req.Method)

	// Override with form parameter
	m = MethodOverrideWithConfig(MethodOverrideConfig{Getter: MethodFromForm("_method")})
	req = httptest.NewRequest(akita.POST, "/", bytes.NewReader([]byte("_method="+akita.DELETE)))
	rec = httptest.NewRecorder()
	req.Header.Set(akita.HeaderContentType, akita.MIMEApplicationForm)
	ctx = a.NewContext(req, rec)
	m(h)(ctx)
	assert.Equal(t, akita.DELETE, req.Method)

	// Override with query parameter
	m = MethodOverrideWithConfig(MethodOverrideConfig{Getter: MethodFromQuery("_method")})
	req = httptest.NewRequest(akita.POST, "/?_method="+akita.DELETE, nil)
	rec = httptest.NewRecorder()
	ctx = a.NewContext(req, rec)
	m(h)(ctx)
	assert.Equal(t, akita.DELETE, req.Method)

	// Ignore `GET`
	req = httptest.NewRequest(akita.GET, "/", nil)
	req.Header.Set(akita.HeaderXHTTPMethodOverride, akita.DELETE)
	assert.Equal(t, akita.GET, req.Method)
}
