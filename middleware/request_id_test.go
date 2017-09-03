package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/itchenyi/akita"
	"github.com/stretchr/testify/assert"
)

func TestRequestID(t *testing.T) {
	a := akita.New()
	req := httptest.NewRequest(akita.GET, "/", nil)
	rec := httptest.NewRecorder()
	ctx := a.NewContext(req, rec)
	handler := func(ctx akita.Context) error {
		return ctx.String(http.StatusOK, "test")
	}

	rid := RequestIDWithConfig(RequestIDConfig{})
	h := rid(handler)
	h(ctx)
	assert.Len(t, rec.Header().Get(akita.HeaderXRequestID), 32)

	// Custom generator
	rid = RequestIDWithConfig(RequestIDConfig{
		Generator: func() string { return "customGenerator" },
	})
	h = rid(handler)
	h(ctx)
	assert.Equal(t, rec.Header().Get(akita.HeaderXRequestID), "customGenerator")
}
