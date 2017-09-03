package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/itchenyi/akita"
	"github.com/stretchr/testify/assert"
)

func TestRecover(t *testing.T) {
	a := akita.New()
	buf := new(bytes.Buffer)
	a.Logger.SetOutput(buf)
	req := httptest.NewRequest(akita.GET, "/", nil)
	rec := httptest.NewRecorder()
	ctx := a.NewContext(req, rec)
	h := Recover()(akita.HandlerFunc(func(ctx akita.Context) error {
		panic("test")
	}))
	h(ctx)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, buf.String(), "PANIC RECOVER")
}
