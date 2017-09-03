package middleware

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/itchenyi/akita"
	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	// Note: Just for the test coverage, not a real test.
	a := akita.New()
	req := httptest.NewRequest(akita.GET, "/", nil)
	rec := httptest.NewRecorder()
	ctx := a.NewContext(req, rec)
	h := Logger()(func(ctx akita.Context) error {
		return ctx.String(http.StatusOK, "test")
	})

	// Status 2xx
	h(ctx)

	// Status 3xx
	rec = httptest.NewRecorder()
	ctx = a.NewContext(req, rec)
	h = Logger()(func(ctx akita.Context) error {
		return ctx.String(http.StatusTemporaryRedirect, "test")
	})
	h(ctx)

	// Status 4xx
	rec = httptest.NewRecorder()
	ctx = a.NewContext(req, rec)
	h = Logger()(func(ctx akita.Context) error {
		return ctx.String(http.StatusNotFound, "test")
	})
	h(ctx)

	// Status 5xx with empty path
	req = httptest.NewRequest(akita.GET, "/", nil)
	rec = httptest.NewRecorder()
	ctx = a.NewContext(req, rec)
	h = Logger()(func(ctx akita.Context) error {
		return errors.New("error")
	})
	h(ctx)
}

func TestLoggerIPAddress(t *testing.T) {
	a := akita.New()
	req := httptest.NewRequest(akita.GET, "/", nil)
	rec := httptest.NewRecorder()
	ctx := a.NewContext(req, rec)
	buf := new(bytes.Buffer)
	a.Logger.SetOutput(buf)
	ip := "127.0.0.1"
	h := Logger()(func(ctx akita.Context) error {
		return ctx.String(http.StatusOK, "test")
	})

	// With X-Real-IP
	req.Header.Add(akita.HeaderXRealIP, ip)
	h(ctx)
	assert.Contains(t, ip, buf.String())

	// With X-Forwarded-For
	buf.Reset()
	req.Header.Del(akita.HeaderXRealIP)
	req.Header.Add(akita.HeaderXForwardedFor, ip)
	h(ctx)
	assert.Contains(t, ip, buf.String())

	buf.Reset()
	h(ctx)
	assert.Contains(t, ip, buf.String())
}

func TestLoggerTemplate(t *testing.T) {
	buf := new(bytes.Buffer)

	e := akita.New()
	e.Use(LoggerWithConfig(LoggerConfig{
		Format: `{"time":"${time_rfc3339_nano}","id":"${id}","remote_ip":"${remote_ip}","host":"${host}","user_agent":"${user_agent}",` +
			`"method":"${method}","uri":"${uri}","status":${status}, "latency":${latency},` +
			`"latency_human":"${latency_human}","bytes_in":${bytes_in}, "path":"${path}", "referer":"${referer}",` +
			`"bytes_out":${bytes_out},"ch":"${header:X-Custom-Header}",` +
			`"us":"${query:username}", "cf":"${form:username}", "session":"${cookie:session}"}` + "\n",
		Output: buf,
	}))

	e.GET("/", func(ctx akita.Context) error {
		return ctx.String(http.StatusOK, "Header Logged")
	})

	req := httptest.NewRequest(akita.GET, "/?username=apagano-param&password=secret", nil)
	req.RequestURI = "/"
	req.Header.Add(akita.HeaderXRealIP, "127.0.0.1")
	req.Header.Add("Referer", "google.com")
	req.Header.Add("User-Agent", "akita-tests-agent")
	req.Header.Add("X-Custom-Header", "AAA-CUSTOM-VALUE")
	req.Header.Add("X-Request-ID", "6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	req.Header.Add("Cookie", "_ga=GA1.2.000000000.0000000000; session=ac08034cd216a647fc2eb62f2bcf7b810")
	req.Form = url.Values{
		"username": []string{"apagano-form"},
		"password": []string{"secret-form"},
	}

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	cases := map[string]bool{
		"apagano-param":                        true,
		"apagano-form":                         true,
		"AAA-CUSTOM-VALUE":                     true,
		"BBB-CUSTOM-VALUE":                     false,
		"secret-form":                          false,
		"hexvalue":                             false,
		"GET":                                  true,
		"127.0.0.1":                            true,
		"\"path\":\"/\"":                       true,
		"\"uri\":\"/\"":                        true,
		"\"status\":200":                       true,
		"\"bytes_in\":0":                       true,
		"google.com":                           true,
		"akita-tests-agent":                    true,
		"6ba7b810-9dad-11d1-80b4-00c04fd430c8": true,
		"ac08034cd216a647fc2eb62f2bcf7b810":    true,
	}

	for token, present := range cases {
		assert.True(t, strings.Contains(buf.String(), token) == present, "Case: "+token)
	}
}
