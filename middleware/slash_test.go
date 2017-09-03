package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/itchenyi/akita"
	"github.com/stretchr/testify/assert"
)

func TestAddTrailingSlash(t *testing.T) {
	a := akita.New()
	req := httptest.NewRequest(akita.GET, "/add-slash", nil)
	rec := httptest.NewRecorder()
	ctx := a.NewContext(req, rec)
	h := AddTrailingSlash()(func(ctx akita.Context) error {
		return nil
	})
	h(ctx)
	assert.Equal(t, "/add-slash/", req.URL.Path)
	assert.Equal(t, "/add-slash/", req.RequestURI)

	// With config
	req = httptest.NewRequest(akita.GET, "/add-slash?key=value", nil)
	rec = httptest.NewRecorder()
	ctx = a.NewContext(req, rec)
	h = AddTrailingSlashWithConfig(TrailingSlashConfig{
		RedirectCode: http.StatusMovedPermanently,
	})(func(ctx akita.Context) error {
		return nil
	})
	h(ctx)
	assert.Equal(t, http.StatusMovedPermanently, rec.Code)
	assert.Equal(t, "/add-slash/?key=value", rec.Header().Get(akita.HeaderLocation))
}

func TestRemoveTrailingSlash(t *testing.T) {
	a := akita.New()
	req := httptest.NewRequest(akita.GET, "/remove-slash/", nil)
	rec := httptest.NewRecorder()
	ctx := a.NewContext(req, rec)
	h := RemoveTrailingSlash()(func(ctx akita.Context) error {
		return nil
	})
	h(ctx)
	assert.Equal(t, "/remove-slash", req.URL.Path)
	assert.Equal(t, "/remove-slash", req.RequestURI)

	// With config
	req = httptest.NewRequest(akita.GET, "/remove-slash/?key=value", nil)
	rec = httptest.NewRecorder()
	ctx = a.NewContext(req, rec)
	h = RemoveTrailingSlashWithConfig(TrailingSlashConfig{
		RedirectCode: http.StatusMovedPermanently,
	})(func(ctx akita.Context) error {
		return nil
	})
	h(ctx)
	assert.Equal(t, http.StatusMovedPermanently, rec.Code)
	assert.Equal(t, "/remove-slash?key=value", rec.Header().Get(akita.HeaderLocation))

	// With bare URL
	req = httptest.NewRequest(akita.GET, "http://localhost", nil)
	rec = httptest.NewRecorder()
	ctx = a.NewContext(req, rec)
	h = RemoveTrailingSlash()(func(ctx akita.Context) error {
		return nil
	})
	h(ctx)
	assert.Equal(t, "", req.URL.Path)
}
