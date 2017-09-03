package akita

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type (
	// Context represents the context of the current HTTP request. It holds request and
	// response objects, path, path parameters, data and registered handler.
	Context interface {
		// Request returns `*http.Request`.
		Request() *http.Request

		// SetRequest sets `*http.Request`.
		SetRequest(r *http.Request)

		// Response returns `*Response`.
		Response() *Response

		// IsTLS returns true if HTTP connection is TLS otherwise false.
		IsTLS() bool

		// IsWebSocket returns true if HTTP connection is WebSocket otherwise false.
		IsWebSocket() bool

		// Scheme returns the HTTP protocol scheme, `http` or `https`.
		Scheme() string

		// RealIP returns the client's network address based on `X-Forwarded-For`
		// or `X-Real-IP` request header.
		RealIP() string

		// Path returns the registered path for the handler.
		Path() string

		// SetPath sets the registered path for the handler.
		SetPath(p string)

		// Param returns path parameter by name.
		Param(name string) string

		// ParamNames returns path parameter names.
		ParamNames() []string

		// SetParamNames sets path parameter names.
		SetParamNames(names ...string)

		// ParamValues returns path parameter values.
		ParamValues() []string

		// SetParamValues sets path parameter values.
		SetParamValues(values ...string)

		// QueryParam returns the query param for the provided name.
		QueryParam(name string) string

		// QueryParams returns the query parameters as `url.Values`.
		QueryParams() url.Values

		// QueryString returns the URL query string.
		QueryString() string

		// FormValue returns the form field value for the provided name.
		FormValue(name string) string

		// FormParams returns the form parameters as `url.Values`.
		FormParams() (url.Values, error)

		// FormFile returns the multipart form file for the provided name.
		FormFile(name string) (*multipart.FileHeader, error)

		// MultipartForm returns the multipart form.
		MultipartForm() (*multipart.Form, error)

		// Cookie returns the named cookie provided in the request.
		Cookie(name string) (*http.Cookie, error)

		// SetCookie adds a `Set-Cookie` header in HTTP response.
		SetCookie(cookie *http.Cookie)

		// Cookies returns the HTTP cookies sent with the request.
		Cookies() []*http.Cookie

		// Get retrieves data from the context.
		Get(key string) interface{}

		// Set saves data in the context.
		Set(key string, val interface{})

		// Bind binds the request body into provided type `i`. The default binder
		// does it based on Content-Type header.
		Bind(i interface{}) error

		// Validate validates provided `i`. It is usually called after `Context#Bind()`.
		// Validator must be registered using `Akita#Validator`.
		Validate(i interface{}) error

		// Render renders a template with data and sends a text/html response with status
		// code. Renderer must be registered using `Akita.Renderer`.
		Render(code int, name string, data interface{}) error

		// HTML sends an HTTP response with status code.
		HTML(code int, html string) error

		// HTMLBlob sends an HTTP blob response with status code.
		HTMLBlob(code int, b []byte) error

		// String sends a string response with status code.
		String(code int, s string) error

		// JSON sends a JSON response with status code.
		JSON(code int, i interface{}) error

		// JSONPretty sends a pretty-print JSON with status code.
		JSONPretty(code int, i interface{}, indent string) error

		// JSONBlob sends a JSON blob response with status code.
		JSONBlob(code int, b []byte) error

		// JSONP sends a JSONP response with status code. It uses `callback` to construct
		// the JSONP payload.
		JSONP(code int, callback string, i interface{}) error

		// JSONPBlob sends a JSONP blob response with status code. It uses `callback`
		// to construct the JSONP payload.
		JSONPBlob(code int, callback string, b []byte) error

		// XML sends an XML response with status code.
		XML(code int, i interface{}) error

		// XMLPretty sends a pretty-print XML with status code.
		XMLPretty(code int, i interface{}, indent string) error

		// XMLBlob sends an XML blob response with status code.
		XMLBlob(code int, b []byte) error

		// Blob sends a blob response with status code and content type.
		Blob(code int, contentType string, b []byte) error

		// Stream sends a streaming response with status code and content type.
		Stream(code int, contentType string, r io.Reader) error

		// File sends a response with the content of the file.
		File(file string) error

		// Attachment sends a response as attachment, prompting client to save the
		// file.
		Attachment(file string, name string) error

		// Inline sends a response as inline, opening the file in the browser.
		Inline(file string, name string) error

		// NoContent sends a response with no body and a status code.
		NoContent(code int) error

		// Redirect redirects the request to a provided URL with status code.
		Redirect(code int, url string) error

		// Error invokes the registered HTTP error handler. Generally used by middleware.
		Error(err error)

		// Handler returns the matched handler by router.
		Handler() HandlerFunc

		// SetHandler sets the matched handler by router.
		SetHandler(h HandlerFunc)

		// Logger returns the `Logger` instance.
		Logger() Logger

		// Akita returns the `Akita` instance.
		Akita() *Akita

		// Reset resets the context after request completes. It must be called along
		// with `Akita#AcquireContext()` and `Akita#ReleaseContext()`.
		// See `Akita#ServeHTTP()`
		Reset(r *http.Request, w http.ResponseWriter)
	}

	context struct {
		request  *http.Request
		response *Response
		path     string
		pnames   []string
		pvalues  []string
		query    url.Values
		handler  HandlerFunc
		store    Map
		akita    *Akita
	}
)

const (
	defaultMemory = 32 << 20 // 32 MB
	indexPage     = "index.html"
)

func (ctx *context) Request() *http.Request {
	return ctx.request
}

func (ctx *context) SetRequest(r *http.Request) {
	ctx.request = r
}

func (ctx *context) Response() *Response {
	return ctx.response
}

func (ctx *context) IsTLS() bool {
	return ctx.request.TLS != nil
}

func (ctx *context) IsWebSocket() bool {
	upgrade := ctx.request.Header.Get(HeaderUpgrade)
	return upgrade == "websocket" || upgrade == "Websocket"
}

func (ctx *context) Scheme() string {
	// Can't use `r.Request.URL.Scheme`
	// See: https://groups.google.com/forum/#!topic/golang-nuts/pMUkBlQBDF0
	if ctx.IsTLS() {
		return "https"
	}
	if scheme := ctx.request.Header.Get(HeaderXForwardedProto); scheme != "" {
		return scheme
	}
	if scheme := ctx.request.Header.Get(HeaderXForwardedProtocol); scheme != "" {
		return scheme
	}
	if ssl := ctx.request.Header.Get(HeaderXForwardedSsl); ssl == "on" {
		return "https"
	}
	if scheme := ctx.request.Header.Get(HeaderXUrlScheme); scheme != "" {
		return scheme
	}
	return "http"
}

func (ctx *context) RealIP() string {
	ra := ctx.request.RemoteAddr
	if ip := ctx.request.Header.Get(HeaderXForwardedFor); ip != "" {
		ra = strings.Split(ip, ", ")[0]
	} else if ip := ctx.request.Header.Get(HeaderXRealIP); ip != "" {
		ra = ip
	} else {
		ra, _, _ = net.SplitHostPort(ra)
	}
	return ra
}

func (ctx *context) Path() string {
	return ctx.path
}

func (ctx *context) SetPath(p string) {
	ctx.path = p
}

func (ctx *context) Param(name string) string {
	for i, n := range ctx.pnames {
		if i < len(ctx.pvalues) {
			if n == name {
				return ctx.pvalues[i]
			}

			// Param name with aliases
			for _, p := range strings.Split(n, ",") {
				if p == name {
					return ctx.pvalues[i]
				}
			}
		}
	}
	return ""
}

func (ctx *context) ParamNames() []string {
	return ctx.pnames
}

func (ctx *context) SetParamNames(names ...string) {
	ctx.pnames = names
}

func (ctx *context) ParamValues() []string {
	return ctx.pvalues[:len(ctx.pnames)]
}

func (ctx *context) SetParamValues(values ...string) {
	ctx.pvalues = values
}

func (ctx *context) QueryParam(name string) string {
	if ctx.query == nil {
		ctx.query = ctx.request.URL.Query()
	}
	return ctx.query.Get(name)
}

func (ctx *context) QueryParams() url.Values {
	if ctx.query == nil {
		ctx.query = ctx.request.URL.Query()
	}
	return ctx.query
}

func (ctx *context) QueryString() string {
	return ctx.request.URL.RawQuery
}

func (ctx *context) FormValue(name string) string {
	return ctx.request.FormValue(name)
}

func (ctx *context) FormParams() (url.Values, error) {
	if strings.HasPrefix(ctx.request.Header.Get(HeaderContentType), MIMEMultipartForm) {
		if err := ctx.request.ParseMultipartForm(defaultMemory); err != nil {
			return nil, err
		}
	} else {
		if err := ctx.request.ParseForm(); err != nil {
			return nil, err
		}
	}
	return ctx.request.Form, nil
}

func (ctx *context) FormFile(name string) (*multipart.FileHeader, error) {
	_, fh, err := ctx.request.FormFile(name)
	return fh, err
}

func (ctx *context) MultipartForm() (*multipart.Form, error) {
	err := ctx.request.ParseMultipartForm(defaultMemory)
	return ctx.request.MultipartForm, err
}

func (ctx *context) Cookie(name string) (*http.Cookie, error) {
	return ctx.request.Cookie(name)
}

func (ctx *context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(ctx.Response(), cookie)
}

func (ctx *context) Cookies() []*http.Cookie {
	return ctx.request.Cookies()
}

func (ctx *context) Get(key string) interface{} {
	return ctx.store[key]
}

func (ctx *context) Set(key string, val interface{}) {
	if ctx.store == nil {
		ctx.store = make(Map)
	}
	ctx.store[key] = val
}

func (ctx *context) Bind(i interface{}) error {
	return ctx.akita.Binder.Bind(i, ctx)
}

func (ctx *context) Validate(i interface{}) error {
	if ctx.akita.Validator == nil {
		return ErrValidatorNotRegistered
	}
	return ctx.akita.Validator.Validate(i)
}

func (ctx *context) Render(code int, name string, data interface{}) (err error) {
	if ctx.akita.Renderer == nil {
		return ErrRendererNotRegistered
	}
	buf := new(bytes.Buffer)
	if err = ctx.akita.Renderer.Render(buf, name, data, ctx); err != nil {
		return
	}
	return ctx.HTMLBlob(code, buf.Bytes())
}

func (ctx *context) HTML(code int, html string) (err error) {
	return ctx.HTMLBlob(code, []byte(html))
}

func (ctx *context) HTMLBlob(code int, b []byte) (err error) {
	return ctx.Blob(code, MIMETextHTMLCharsetUTF8, b)
}

func (ctx *context) String(code int, s string) (err error) {
	return ctx.Blob(code, MIMETextPlainCharsetUTF8, []byte(s))
}

func (ctx *context) JSON(code int, i interface{}) (err error) {
	_, pretty := ctx.QueryParams()["pretty"]
	if ctx.akita.Debug || pretty {
		return ctx.JSONPretty(code, i, "  ")
	}
	b, err := json.Marshal(i)
	if err != nil {
		return
	}
	return ctx.JSONBlob(code, b)
}

func (ctx *context) JSONPretty(code int, i interface{}, indent string) (err error) {
	b, err := json.MarshalIndent(i, "", indent)
	if err != nil {
		return
	}
	return ctx.JSONBlob(code, b)
}

func (ctx *context) JSONBlob(code int, b []byte) (err error) {
	return ctx.Blob(code, MIMEApplicationJSONCharsetUTF8, b)
}

func (ctx *context) JSONP(code int, callback string, i interface{}) (err error) {
	b, err := json.Marshal(i)
	if err != nil {
		return
	}
	return ctx.JSONPBlob(code, callback, b)
}

func (ctx *context) JSONPBlob(code int, callback string, b []byte) (err error) {
	ctx.response.Header().Set(HeaderContentType, MIMEApplicationJavaScriptCharsetUTF8)
	ctx.response.WriteHeader(code)
	if _, err = ctx.response.Write([]byte(callback + "(")); err != nil {
		return
	}
	if _, err = ctx.response.Write(b); err != nil {
		return
	}
	_, err = ctx.response.Write([]byte(");"))
	return
}

func (ctx *context) XML(code int, i interface{}) (err error) {
	_, pretty := ctx.QueryParams()["pretty"]
	if ctx.akita.Debug || pretty {
		return ctx.XMLPretty(code, i, "  ")
	}
	b, err := xml.Marshal(i)
	if err != nil {
		return
	}
	return ctx.XMLBlob(code, b)
}

func (ctx *context) XMLPretty(code int, i interface{}, indent string) (err error) {
	b, err := xml.MarshalIndent(i, "", indent)
	if err != nil {
		return
	}
	return ctx.XMLBlob(code, b)
}

func (ctx *context) XMLBlob(code int, b []byte) (err error) {
	ctx.response.Header().Set(HeaderContentType, MIMEApplicationXMLCharsetUTF8)
	ctx.response.WriteHeader(code)
	if _, err = ctx.response.Write([]byte(xml.Header)); err != nil {
		return
	}
	_, err = ctx.response.Write(b)
	return
}

func (ctx *context) Blob(code int, contentType string, b []byte) (err error) {
	ctx.response.Header().Set(HeaderContentType, contentType)
	ctx.response.WriteHeader(code)
	_, err = ctx.response.Write(b)
	return
}

func (ctx *context) Stream(code int, contentType string, r io.Reader) (err error) {
	ctx.response.Header().Set(HeaderContentType, contentType)
	ctx.response.WriteHeader(code)
	_, err = io.Copy(ctx.response, r)
	return
}

func (ctx *context) File(file string) (err error) {
	f, err := os.Open(file)
	if err != nil {
		return NotFoundHandler(ctx)
	}
	defer f.Close()

	fi, _ := f.Stat()
	if fi.IsDir() {
		file = filepath.Join(file, indexPage)
		f, err = os.Open(file)
		if err != nil {
			return NotFoundHandler(ctx)
		}
		defer f.Close()
		if fi, err = f.Stat(); err != nil {
			return
		}
	}
	http.ServeContent(ctx.Response(), ctx.Request(), fi.Name(), fi.ModTime(), f)
	return
}

func (ctx *context) Attachment(file, name string) (err error) {
	return ctx.contentDisposition(file, name, "attachment")
}

func (ctx *context) Inline(file, name string) (err error) {
	return ctx.contentDisposition(file, name, "inline")
}

func (ctx *context) contentDisposition(file, name, dispositionType string) (err error) {
	ctx.response.Header().Set(HeaderContentDisposition, fmt.Sprintf("%s; filename=%q", dispositionType, name))
	ctx.File(file)
	return
}

func (ctx *context) NoContent(code int) error {
	ctx.response.WriteHeader(code)
	return nil
}

func (ctx *context) Redirect(code int, url string) error {
	if code < 300 || code > 308 {
		return ErrInvalidRedirectCode
	}
	ctx.response.Header().Set(HeaderLocation, url)
	ctx.response.WriteHeader(code)
	return nil
}

func (ctx *context) Error(err error) {
	ctx.akita.HTTPErrorHandler(err, ctx)
}

func (ctx *context) Akita() *Akita {
	return ctx.akita
}

func (ctx *context) Handler() HandlerFunc {
	return ctx.handler
}

func (ctx *context) SetHandler(h HandlerFunc) {
	ctx.handler = h
}

func (ctx *context) Logger() Logger {
	return ctx.akita.Logger
}

func (ctx *context) Reset(r *http.Request, w http.ResponseWriter) {
	ctx.request = r
	ctx.response.reset(w)
	ctx.query = nil
	ctx.handler = NotFoundHandler
	ctx.store = nil
	ctx.path = ""
	ctx.pnames = nil
	// NOTE: Don't reset because it has to have length ctx.akita.maxParam at all times
	// ctx.pvalues = nil
}
