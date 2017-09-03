package middleware

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"net"
	"net/http"

	"io"

	"github.com/itchenyi/akita"
)

type (
	// BodyDumpConfig defines the config for BodyDump middleware.
	BodyDumpConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper

		// Handler receives request and response payload.
		// Required.
		Handler BodyDumpHandler
	}

	// BodyDumpHandler receives the request and response payload.
	BodyDumpHandler func(akita.Context, []byte, []byte)

	bodyDumpResponseWriter struct {
		io.Writer
		http.ResponseWriter
	}
)

var (
	// DefaultBodyDumpConfig is the default Gzip middleware config.
	DefaultBodyDumpConfig = BodyDumpConfig{
		Skipper: DefaultSkipper,
	}
)

// BodyDump returns a BodyDump middleware.
//
// BodyLimit middleware captures the request and response payload and calls the
// registered handler.
func BodyDump(handler BodyDumpHandler) akita.MiddlewareFunc {
	c := DefaultBodyDumpConfig
	c.Handler = handler
	return BodyDumpWithConfig(c)
}

// BodyDumpWithConfig returns a BodyDump middleware with config.
// See: `BodyDump()`.
func BodyDumpWithConfig(config BodyDumpConfig) akita.MiddlewareFunc {
	// Defaults
	if config.Handler == nil {
		panic("akita: body-dump middleware requires a handler function")
	}
	if config.Skipper == nil {
		config.Skipper = DefaultBodyDumpConfig.Skipper
	}

	return func(next akita.HandlerFunc) akita.HandlerFunc {
		return func(ctx akita.Context) (err error) {
			if config.Skipper(ctx) {
				return next(ctx)
			}

			// Request
			reqBody := []byte{}
			if ctx.Request().Body != nil { // Read
				reqBody, _ = ioutil.ReadAll(ctx.Request().Body)
			}
			ctx.Request().Body = ioutil.NopCloser(bytes.NewBuffer(reqBody)) // Reset

			// Response
			resBody := new(bytes.Buffer)
			mw := io.MultiWriter(ctx.Response().Writer, resBody)
			writer := &bodyDumpResponseWriter{Writer: mw, ResponseWriter: ctx.Response().Writer}
			ctx.Response().Writer = writer

			if err = next(ctx); err != nil {
				ctx.Error(err)
			}

			// Callback
			config.Handler(ctx, reqBody, resBody.Bytes())

			return
		}
	}
}

func (w *bodyDumpResponseWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
}

func (w *bodyDumpResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w *bodyDumpResponseWriter) Flush() {
	w.ResponseWriter.(http.Flusher).Flush()
}

func (w *bodyDumpResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

func (w *bodyDumpResponseWriter) CloseNotify() <-chan bool {
	return w.ResponseWriter.(http.CloseNotifier).CloseNotify()
}
