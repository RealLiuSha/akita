package middleware

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/itchenyi/akita"
	"github.com/stretchr/testify/assert"
)

func TestBodyDump(t *testing.T) {
	a := akita.New()
	hw := "Hello, World!"
	req := httptest.NewRequest(akita.POST, "/", strings.NewReader(hw))
	rec := httptest.NewRecorder()
	ctx := a.NewContext(req, rec)
	h := func(ctx akita.Context) error {
		body, err := ioutil.ReadAll(ctx.Request().Body)
		if err != nil {
			return err
		}
		return ctx.String(http.StatusOK, string(body))
	}

	requestBody := ""
	responseBody := ""
	mw := BodyDump(func(c akita.Context, reqBody, resBody []byte) {
		requestBody = string(reqBody)
		responseBody = string(resBody)
	})
	if assert.NoError(t, mw(h)(ctx)) {
		assert.Equal(t, requestBody, hw)
		assert.Equal(t, responseBody, hw)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, hw, rec.Body.String())
	}
}
