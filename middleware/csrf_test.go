package middleware

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/itchenyi/akita"
	"github.com/itchenyi/common/random"
	"github.com/stretchr/testify/assert"
)

func TestCSRF(t *testing.T) {
	a := akita.New()
	req := httptest.NewRequest(akita.GET, "/", nil)
	rec := httptest.NewRecorder()
	ctx := a.NewContext(req, rec)
	csrf := CSRFWithConfig(CSRFConfig{
		TokenLength: 16,
	})
	h := csrf(func(ctx akita.Context) error {
		return ctx.String(http.StatusOK, "test")
	})

	// Generate CSRF token
	h(ctx)
	assert.Contains(t, rec.Header().Get(akita.HeaderSetCookie), "_csrf")

	// Without CSRF cookie
	req = httptest.NewRequest(akita.POST, "/", nil)
	rec = httptest.NewRecorder()
	ctx = a.NewContext(req, rec)
	assert.Error(t, h(ctx))

	// Empty/invalid CSRF token
	req = httptest.NewRequest(akita.POST, "/", nil)
	rec = httptest.NewRecorder()
	ctx = a.NewContext(req, rec)
	req.Header.Set(akita.HeaderXCSRFToken, "")
	assert.Error(t, h(ctx))

	// Valid CSRF token
	token := random.String(16)
	req.Header.Set(akita.HeaderCookie, "_csrf="+token)
	req.Header.Set(akita.HeaderXCSRFToken, token)
	if assert.NoError(t, h(ctx)) {
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}

func TestCSRFTokenFromForm(t *testing.T) {
	f := make(url.Values)
	f.Set("csrf", "token")
	a := akita.New()
	req := httptest.NewRequest(akita.POST, "/", strings.NewReader(f.Encode()))
	req.Header.Add(akita.HeaderContentType, akita.MIMEApplicationForm)
	c := a.NewContext(req, nil)
	token, err := csrfTokenFromForm("csrf")(c)
	if assert.NoError(t, err) {
		assert.Equal(t, "token", token)
	}
	_, err = csrfTokenFromForm("invalid")(c)
	assert.Error(t, err)
}

func TestCSRFTokenFromQuery(t *testing.T) {
	q := make(url.Values)
	q.Set("csrf", "token")
	e := akita.New()
	req := httptest.NewRequest(akita.GET, "/?"+q.Encode(), nil)
	req.Header.Add(akita.HeaderContentType, akita.MIMEApplicationForm)
	c := e.NewContext(req, nil)
	token, err := csrfTokenFromQuery("csrf")(c)
	if assert.NoError(t, err) {
		assert.Equal(t, "token", token)
	}
	_, err = csrfTokenFromQuery("invalid")(c)
	assert.Error(t, err)
	csrfTokenFromQuery("csrf")
}
