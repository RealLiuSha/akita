package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/itchenyi/akita"
	"github.com/stretchr/testify/assert"
)

func TestStatic(t *testing.T) {
	a := akita.New()
	req := httptest.NewRequest(akita.GET, "/", nil)
	rec := httptest.NewRecorder()
	ctx := a.NewContext(req, rec)
	config := StaticConfig{
		Root: "../_fixture",
	}

	// Directory
	h := StaticWithConfig(config)(akita.NotFoundHandler)
	if assert.NoError(t, h(ctx)) {
		assert.Contains(t, rec.Body.String(), "Akita")
	}

	// File found
	req = httptest.NewRequest(akita.GET, "/images/akita.png", nil)
	rec = httptest.NewRecorder()
	ctx = a.NewContext(req, rec)
	if assert.NoError(t, h(ctx)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, rec.Header().Get(akita.HeaderContentLength), "219885")
	}

	// File not found
	req = httptest.NewRequest(akita.GET, "/none", nil)
	rec = httptest.NewRecorder()
	ctx = a.NewContext(req, rec)
	he := h(ctx).(*akita.HTTPError)
	assert.Equal(t, http.StatusNotFound, he.Code)

	// HTML5
	req = httptest.NewRequest(akita.GET, "/random", nil)
	rec = httptest.NewRecorder()
	ctx = a.NewContext(req, rec)
	config.HTML5 = true
	static := StaticWithConfig(config)
	h = static(akita.NotFoundHandler)
	if assert.NoError(t, h(ctx)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "Akita")
	}

	// Browse
	req = httptest.NewRequest(akita.GET, "/", nil)
	rec = httptest.NewRecorder()
	ctx = a.NewContext(req, rec)
	config.Root = "../_fixture/certs"
	config.Browse = true
	static = StaticWithConfig(config)
	h = static(akita.NotFoundHandler)
	if assert.NoError(t, h(ctx)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "cert.pem")
	}
}
