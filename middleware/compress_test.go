package middleware

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/itchenyi/akita"
	"github.com/stretchr/testify/assert"
)

func TestGzip(t *testing.T) {
	a := akita.New()
	req := httptest.NewRequest(akita.GET, "/", nil)
	rec := httptest.NewRecorder()
	ctx := a.NewContext(req, rec)

	// Skip if no Accept-Encoding header
	h := Gzip()(func(ctx akita.Context) error {
		ctx.Response().Write([]byte("test")) // For Content-Type sniffing
		return nil
	})
	h(ctx)
	assert.Equal(t, "test", rec.Body.String())

	// Gzip
	req = httptest.NewRequest(akita.GET, "/", nil)
	req.Header.Set(akita.HeaderAcceptEncoding, gzipScheme)
	rec = httptest.NewRecorder()
	ctx = a.NewContext(req, rec)
	h(ctx)
	assert.Equal(t, gzipScheme, rec.Header().Get(akita.HeaderContentEncoding))
	assert.Contains(t, rec.Header().Get(akita.HeaderContentType), akita.MIMETextPlain)
	r, err := gzip.NewReader(rec.Body)
	defer r.Close()
	if assert.NoError(t, err) {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r)
		assert.Equal(t, "test", buf.String())
	}
}

func TestGzipNoContent(t *testing.T) {
	a := akita.New()
	req := httptest.NewRequest(akita.GET, "/", nil)
	req.Header.Set(akita.HeaderAcceptEncoding, gzipScheme)
	rec := httptest.NewRecorder()
	ctx := a.NewContext(req, rec)
	h := Gzip()(func(ctx akita.Context) error {
		return ctx.NoContent(http.StatusNoContent)
	})
	if assert.NoError(t, h(ctx)) {
		assert.Empty(t, rec.Header().Get(akita.HeaderContentEncoding))
		assert.Empty(t, rec.Header().Get(akita.HeaderContentType))
		assert.Equal(t, 0, len(rec.Body.Bytes()))
	}
}

func TestGzipErrorReturned(t *testing.T) {
	a := akita.New()
	a.Use(Gzip())
	a.GET("/", func(ctx akita.Context) error {
		return akita.ErrNotFound
	})
	req := httptest.NewRequest(akita.GET, "/", nil)
	req.Header.Set(akita.HeaderAcceptEncoding, gzipScheme)
	rec := httptest.NewRecorder()
	a.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Empty(t, rec.Header().Get(akita.HeaderContentEncoding))
}

// Issue #806
func TestGzipWithStatic(t *testing.T) {
	a := akita.New()
	a.Use(Gzip())
	a.Static("/test", "../_fixture/images")
	req := httptest.NewRequest(akita.GET, "/test/akita.png", nil)
	req.Header.Set(akita.HeaderAcceptEncoding, gzipScheme)
	rec := httptest.NewRecorder()
	a.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	// Data is written out in chunks when Content-Length == "", so only
	// validate the content length if it's not set.
	if cl := rec.Header().Get("Content-Length"); cl != "" {
		assert.Equal(t, cl, rec.Body.Len())
	}
	r, err := gzip.NewReader(rec.Body)
	assert.NoError(t, err)
	defer r.Close()
	want, err := ioutil.ReadFile("../_fixture/images/akita.png")
	if assert.NoError(t, err) {
		var buf bytes.Buffer
		buf.ReadFrom(r)
		assert.Equal(t, want, buf.Bytes())
	}
}
