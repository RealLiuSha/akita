package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/itchenyi/akita"
	"github.com/stretchr/testify/assert"
)

func TestRedirectHTTPSRedirect(t *testing.T) {
	a := akita.New()
	next := func(ctx akita.Context) (err error) {
		return ctx.NoContent(http.StatusOK)
	}
	req := httptest.NewRequest(akita.GET, "/", nil)
	req.Host = "liusha.me"
	res := httptest.NewRecorder()
	ctx := a.NewContext(req, res)
	HTTPSRedirect()(next)(ctx)
	assert.Equal(t, http.StatusMovedPermanently, res.Code)
	assert.Equal(t, "https://liusha.me/", res.Header().Get(akita.HeaderLocation))
}

func TestRedirectHTTPSWWWRedirect(t *testing.T) {
	a := akita.New()
	next := func(ctx akita.Context) (err error) {
		return ctx.NoContent(http.StatusOK)
	}
	req := httptest.NewRequest(akita.GET, "/", nil)
	req.Host = "liusha.me"
	res := httptest.NewRecorder()
	ctx := a.NewContext(req, res)
	HTTPSWWWRedirect()(next)(ctx)
	assert.Equal(t, http.StatusMovedPermanently, res.Code)
	assert.Equal(t, "https://www.liusha.me/", res.Header().Get(akita.HeaderLocation))
}

func TestRedirectHTTPSNonWWWRedirect(t *testing.T) {
	a := akita.New()
	next := func(ctx akita.Context) (err error) {
		return ctx.NoContent(http.StatusOK)
	}
	req := httptest.NewRequest(akita.GET, "/", nil)
	req.Host = "www.liusha.me"
	res := httptest.NewRecorder()
	ctx := a.NewContext(req, res)
	HTTPSNonWWWRedirect()(next)(ctx)
	assert.Equal(t, http.StatusMovedPermanently, res.Code)
	assert.Equal(t, "https://liusha.me/", res.Header().Get(akita.HeaderLocation))
}

func TestRedirectWWWRedirect(t *testing.T) {
	a := akita.New()
	next := func(ctx akita.Context) (err error) {
		return ctx.NoContent(http.StatusOK)
	}
	req := httptest.NewRequest(akita.GET, "/", nil)
	req.Host = "liusha.me"
	res := httptest.NewRecorder()
	ctx := a.NewContext(req, res)
	WWWRedirect()(next)(ctx)
	assert.Equal(t, http.StatusMovedPermanently, res.Code)
	assert.Equal(t, "http://www.liusha.me/", res.Header().Get(akita.HeaderLocation))
}

func TestRedirectNonWWWRedirect(t *testing.T) {
	a := akita.New()
	next := func(ctx akita.Context) (err error) {
		return ctx.NoContent(http.StatusOK)
	}
	req := httptest.NewRequest(akita.GET, "/", nil)
	req.Host = "www.liusha.me"
	res := httptest.NewRecorder()
	ctx := a.NewContext(req, res)
	NonWWWRedirect()(next)(ctx)
	assert.Equal(t, http.StatusMovedPermanently, res.Code)
	assert.Equal(t, "http://liusha.me/", res.Header().Get(akita.HeaderLocation))
}
