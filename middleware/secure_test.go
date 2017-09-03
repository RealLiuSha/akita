package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/itchenyi/akita"
	"github.com/stretchr/testify/assert"
)

func TestSecure(t *testing.T) {
	a := akita.New()
	req := httptest.NewRequest(akita.GET, "/", nil)
	rec := httptest.NewRecorder()
	ctx := a.NewContext(req, rec)
	h := func(ctx akita.Context) error {
		return ctx.String(http.StatusOK, "test")
	}

	// Default
	Secure()(h)(ctx)
	assert.Equal(t, "1; mode=block", rec.Header().Get(akita.HeaderXXSSProtection))
	assert.Equal(t, "nosniff", rec.Header().Get(akita.HeaderXContentTypeOptions))
	assert.Equal(t, "SAMEORIGIN", rec.Header().Get(akita.HeaderXFrameOptions))
	assert.Equal(t, "", rec.Header().Get(akita.HeaderStrictTransportSecurity))
	assert.Equal(t, "", rec.Header().Get(akita.HeaderContentSecurityPolicy))

	// Custom
	req.Header.Set(akita.HeaderXForwardedProto, "https")
	rec = httptest.NewRecorder()
	ctx = a.NewContext(req, rec)
	SecureWithConfig(SecureConfig{
		XSSProtection:         "",
		ContentTypeNosniff:    "",
		XFrameOptions:         "",
		HSTSMaxAge:            3600,
		ContentSecurityPolicy: "default-src 'self'",
	})(h)(ctx)
	assert.Equal(t, "", rec.Header().Get(akita.HeaderXXSSProtection))
	assert.Equal(t, "", rec.Header().Get(akita.HeaderXContentTypeOptions))
	assert.Equal(t, "", rec.Header().Get(akita.HeaderXFrameOptions))
	assert.Equal(t, "max-age=3600; includeSubdomains", rec.Header().Get(akita.HeaderStrictTransportSecurity))
	assert.Equal(t, "default-src 'self'", rec.Header().Get(akita.HeaderContentSecurityPolicy))
}
