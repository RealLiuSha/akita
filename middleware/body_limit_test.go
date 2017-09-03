package middleware

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/itchenyi/akita"
	"github.com/stretchr/testify/assert"
)

func TestBodyLimit(t *testing.T) {
	a := akita.New()
	hw := []byte("Hello, World!")
	req := httptest.NewRequest(akita.POST, "/", bytes.NewReader(hw))
	rec := httptest.NewRecorder()
	ctx := a.NewContext(req, rec)
	h := func(ctx akita.Context) error {
		body, err := ioutil.ReadAll(ctx.Request().Body)
		if err != nil {
			return err
		}
		return ctx.String(http.StatusOK, string(body))
	}

	// Based on content length (within limit)
	if assert.NoError(t, BodyLimit("2M")(h)(ctx)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, hw, rec.Body.Bytes())
	}

	// Based on content read (overlimit)
	he := BodyLimit("2B")(h)(ctx).(*akita.HTTPError)
	assert.Equal(t, http.StatusRequestEntityTooLarge, he.Code)

	// Based on content read (within limit)
	req = httptest.NewRequest(akita.POST, "/", bytes.NewReader(hw))
	rec = httptest.NewRecorder()
	ctx = a.NewContext(req, rec)
	if assert.NoError(t, BodyLimit("2M")(h)(ctx)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "Hello, World!", rec.Body.String())
	}

	// Based on content read (overlimit)
	req = httptest.NewRequest(akita.POST, "/", bytes.NewReader(hw))
	rec = httptest.NewRecorder()
	ctx = a.NewContext(req, rec)
	he = BodyLimit("2B")(h)(ctx).(*akita.HTTPError)
	assert.Equal(t, http.StatusRequestEntityTooLarge, he.Code)
}
