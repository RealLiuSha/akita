package akita

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResponse(t *testing.T) {
	a := New()
	req := httptest.NewRequest(GET, "/", nil)
	rec := httptest.NewRecorder()
	ctx := a.NewContext(req, rec)
	res := &Response{akita: a, Writer: rec}

	// Before
	res.Before(func() {
		ctx.Response().Header().Set(HeaderServer, "akita")
	})
	res.Write([]byte("test"))
	assert.Equal(t, "akita", rec.Header().Get(HeaderServer))
}
