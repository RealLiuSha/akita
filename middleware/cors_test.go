package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/itchenyi/akita"
	"github.com/stretchr/testify/assert"
)

func TestCORS(t *testing.T) {
	a := akita.New()

	// Wildcard origin
	req := httptest.NewRequest(akita.GET, "/", nil)
	rec := httptest.NewRecorder()
	ctx := a.NewContext(req, rec)
	h := CORS()(akita.NotFoundHandler)
	h(ctx)
	assert.Equal(t, "*", rec.Header().Get(akita.HeaderAccessControlAllowOrigin))

	// Allow origins
	req = httptest.NewRequest(akita.GET, "/", nil)
	rec = httptest.NewRecorder()
	ctx = a.NewContext(req, rec)
	h = CORSWithConfig(CORSConfig{
		AllowOrigins: []string{"localhost"},
	})(akita.NotFoundHandler)
	req.Header.Set(akita.HeaderOrigin, "localhost")
	h(ctx)
	assert.Equal(t, "localhost", rec.Header().Get(akita.HeaderAccessControlAllowOrigin))

	// Preflight request
	req = httptest.NewRequest(akita.OPTIONS, "/", nil)
	rec = httptest.NewRecorder()
	ctx = a.NewContext(req, rec)
	req.Header.Set(akita.HeaderOrigin, "localhost")
	req.Header.Set(akita.HeaderContentType, akita.MIMEApplicationJSON)
	cors := CORSWithConfig(CORSConfig{
		AllowOrigins:     []string{"localhost"},
		AllowCredentials: true,
		MaxAge:           3600,
	})
	h = cors(akita.NotFoundHandler)
	h(ctx)
	assert.Equal(t, "localhost", rec.Header().Get(akita.HeaderAccessControlAllowOrigin))
	assert.NotEmpty(t, rec.Header().Get(akita.HeaderAccessControlAllowMethods))
	assert.Equal(t, "true", rec.Header().Get(akita.HeaderAccessControlAllowCredentials))
	assert.Equal(t, "3600", rec.Header().Get(akita.HeaderAccessControlMaxAge))
}
