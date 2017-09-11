/*
Package akita implements high performance, minimalist Go web framework.

Example:

  package main

  import (
    "net/http"

    "github.com/itchenyi/akita"
    "github.com/itchenyi/akita/middleware"
  )

  // Handler
  func hello(ctx akita.Context) error {
    return ctx.String(http.StatusOK, "Hello, World!")
  }

  func main() {
    // Akita instance
    a := akita.New()

    // Middleware
    a.Use(middleware.Logger())
    a.Use(middleware.Recover())

    // Routes
    e.GET("/", hello)

    // Start server
    e.Logger.Fatal(e.Start(":1323"))
  }

Learn more at https://liusha.me/tags/akita
*/
package akita

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	stdLog "log"
	"net"
	"net/http"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/itchenyi/common/color"
	"github.com/itchenyi/common/log"
	"golang.org/x/crypto/acme/autocert"
)

type (
	// Akita is the top-level framework instance.
	Akita struct {
		stdLogger        *stdLog.Logger
		colorer          *color.Color
		premiddleware    []MiddlewareFunc
		middleware       []MiddlewareFunc
		maxParam         *int
		router           *Router
		notFoundHandler  HandlerFunc
		pool             sync.Pool
		Server           *http.Server
		TLSServer        *http.Server
		Listener         net.Listener
		TLSListener      net.Listener
		AutoTLSManager   autocert.Manager
		DisableHTTP2     bool
		Debug            bool
		HideBanner       bool
		HTTPErrorHandler HTTPErrorHandler
		Binder           Binder
		Validator        Validator
		Renderer         Renderer
		// Mutex            sync.RWMutex
		Logger Logger
	}

	// Route contains a handler and information for matching against requests.
	Route struct {
		Method string `json:"method"`
		Path   string `json:"path"`
		Name   string `json:"name"`
	}

	// HTTPError represents an error that occurred while handling a request.
	HTTPError struct {
		Code    int
		Message interface{}
		Inner   error // Stores the error returned by an external dependency
	}

	// MiddlewareFunc defines a function to process middleware.
	MiddlewareFunc func(HandlerFunc) HandlerFunc

	// HandlerFunc defines a function to server HTTP requests.
	HandlerFunc func(Context) error

	// HTTPErrorHandler is a centralized HTTP error handler.
	HTTPErrorHandler func(error, Context)

	// Validator is the interface that wraps the Validate function.
	Validator interface {
		Validate(i interface{}) error
	}

	// Renderer is the interface that wraps the Render function.
	Renderer interface {
		Render(io.Writer, string, interface{}, Context) error
	}

	// Map defines a generic map of type `map[string]interface{}`.
	Map map[string]interface{}

	// i is the interface for Akita and Group.
	i interface {
		GET(string, HandlerFunc, ...MiddlewareFunc) *Route
	}
)

// HTTP methods
const (
	CONNECT = "CONNECT"
	DELETE  = "DELETE"
	GET     = "GET"
	HEAD    = "HEAD"
	OPTIONS = "OPTIONS"
	PATCH   = "PATCH"
	POST    = "POST"
	PUT     = "PUT"
	TRACE   = "TRACE"
)

// MIME types
const (
	MIMEApplicationJSON                  = "application/json"
	MIMEApplicationJSONCharsetUTF8       = MIMEApplicationJSON + "; " + charsetUTF8
	MIMEApplicationJavaScript            = "application/javascript"
	MIMEApplicationJavaScriptCharsetUTF8 = MIMEApplicationJavaScript + "; " + charsetUTF8
	MIMEApplicationXML                   = "application/xml"
	MIMEApplicationXMLCharsetUTF8        = MIMEApplicationXML + "; " + charsetUTF8
	MIMETextXML                          = "text/xml"
	MIMETextXMLCharsetUTF8               = MIMETextXML + "; " + charsetUTF8
	MIMEApplicationForm                  = "application/x-www-form-urlencoded"
	MIMEApplicationProtobuf              = "application/protobuf"
	MIMEApplicationMsgpack               = "application/msgpack"
	MIMETextHTML                         = "text/html"
	MIMETextHTMLCharsetUTF8              = MIMETextHTML + "; " + charsetUTF8
	MIMETextPlain                        = "text/plain"
	MIMETextPlainCharsetUTF8             = MIMETextPlain + "; " + charsetUTF8
	MIMEMultipartForm                    = "multipart/form-data"
	MIMEOctetStream                      = "application/octet-stream"
)

const (
	charsetUTF8 = "charset=UTF-8"
)

// Headers
const (
	HeaderAccept              = "Accept"
	HeaderAcceptEncoding      = "Accept-Encoding"
	HeaderAllow               = "Allow"
	HeaderAuthorization       = "Authorization"
	HeaderContentDisposition  = "Content-Disposition"
	HeaderContentEncoding     = "Content-Encoding"
	HeaderContentLength       = "Content-Length"
	HeaderContentType         = "Content-Type"
	HeaderCookie              = "Cookie"
	HeaderSetCookie           = "Set-Cookie"
	HeaderIfModifiedSince     = "If-Modified-Since"
	HeaderLastModified        = "Last-Modified"
	HeaderLocation            = "Location"
	HeaderUpgrade             = "Upgrade"
	HeaderVary                = "Vary"
	HeaderWWWAuthenticate     = "WWW-Authenticate"
	HeaderXForwardedFor       = "X-Forwarded-For"
	HeaderXForwardedProto     = "X-Forwarded-Proto"
	HeaderXForwardedProtocol  = "X-Forwarded-Protocol"
	HeaderXForwardedSsl       = "X-Forwarded-Ssl"
	HeaderXUrlScheme          = "X-Url-Scheme"
	HeaderXHTTPMethodOverride = "X-HTTP-Method-Override"
	HeaderXRealIP             = "X-Real-IP"
	HeaderXRequestID          = "X-Request-ID"
	HeaderServer              = "Server"
	HeaderOrigin              = "Origin"

	// Access control
	HeaderAccessControlRequestMethod    = "Access-Control-Request-Method"
	HeaderAccessControlRequestHeaders   = "Access-Control-Request-Headers"
	HeaderAccessControlAllowOrigin      = "Access-Control-Allow-Origin"
	HeaderAccessControlAllowMethods     = "Access-Control-Allow-Methods"
	HeaderAccessControlAllowHeaders     = "Access-Control-Allow-Headers"
	HeaderAccessControlAllowCredentials = "Access-Control-Allow-Credentials"
	HeaderAccessControlExposeHeaders    = "Access-Control-Expose-Headers"
	HeaderAccessControlMaxAge           = "Access-Control-Max-Age"

	// Security
	HeaderStrictTransportSecurity = "Strict-Transport-Security"
	HeaderXContentTypeOptions     = "X-Content-Type-Options"
	HeaderXXSSProtection          = "X-XSS-Protection"
	HeaderXFrameOptions           = "X-Frame-Options"
	HeaderContentSecurityPolicy   = "Content-Security-Policy"
	HeaderXCSRFToken              = "X-CSRF-Token"
)

const (
	version = "1.0.1"
	website = "https://liusha.me"
	// http://patorjk.com/software/taag/#p=display&f=Small%20Slant&t=Akita
	banner = `
   ___   __    _ __
  / _ | / /__ (_) /____ _
 / __ |/  '_// / __/ _ '/
/_/ |_/_/\_\/_/\__/\_,_/%s
High performance, minimalist Go web framework
%s
____________________________________O/_______
                                    O\
`
)

var (
	methods = [...]string{
		CONNECT,
		DELETE,
		GET,
		HEAD,
		OPTIONS,
		PATCH,
		POST,
		PUT,
		TRACE,
	}
)

// Errors
var (
	ErrUnsupportedMediaType        = NewHTTPError(http.StatusUnsupportedMediaType)
	ErrNotFound                    = NewHTTPError(http.StatusNotFound)
	ErrUnauthorized                = NewHTTPError(http.StatusUnauthorized)
	ErrForbidden                   = NewHTTPError(http.StatusForbidden)
	ErrMethodNotAllowed            = NewHTTPError(http.StatusMethodNotAllowed)
	ErrStatusRequestEntityTooLarge = NewHTTPError(http.StatusRequestEntityTooLarge)
	ErrValidatorNotRegistered      = errors.New("Validator not registered")
	ErrRendererNotRegistered       = errors.New("Renderer not registered")
	ErrInvalidRedirectCode         = errors.New("Invalid redirect status code")
	ErrCookieNotFound              = errors.New("Cookie not found")
)

// Error handlers
var (
	NotFoundHandler = func(c Context) error {
		return ErrNotFound
	}

	MethodNotAllowedHandler = func(c Context) error {
		return ErrMethodNotAllowed
	}
)

// New creates an instance of Akita.
func New() (a *Akita) {
	a = &Akita{
		Server:    new(http.Server),
		TLSServer: new(http.Server),
		AutoTLSManager: autocert.Manager{
			Prompt: autocert.AcceptTOS,
		},
		Logger:   log.New("akita"),
		colorer:  color.New(),
		maxParam: new(int),
	}
	a.Server.Handler = a
	a.TLSServer.Handler = a
	a.HTTPErrorHandler = a.DefaultHTTPErrorHandler
	a.Binder = &DefaultBinder{}
	a.Logger.SetLevel(log.ERROR)
	a.stdLogger = stdLog.New(a.Logger.Output(), a.Logger.Prefix()+": ", 0)
	a.pool.New = func() interface{} {
		return a.NewContext(nil, nil)
	}
	a.router = NewRouter(a)
	return
}

// NewContext returns a Context instance.
func (a *Akita) NewContext(r *http.Request, w http.ResponseWriter) Context {
	return &context{
		request:  r,
		response: NewResponse(w, a),
		store:    make(Map),
		akita:    a,
		pvalues:  make([]string, *a.maxParam),
		handler:  NotFoundHandler,
	}
}

// Router returns router.
func (a *Akita) Router() *Router {
	return a.router
}

// DefaultHTTPErrorHandler is the default HTTP error handler. It sends a JSON response
// with status code.
func (a *Akita) DefaultHTTPErrorHandler(err error, ctx Context) {
	var (
		code = http.StatusInternalServerError
		msg  interface{}
	)

	if he, ok := err.(*HTTPError); ok {
		code = he.Code
		msg = he.Message
	} else if a.Debug {
		msg = err.Error()
		if he.Inner != nil {
			msg = fmt.Sprintf("%v, %v", err, he.Inner)
		}
	} else {
		msg = http.StatusText(code)
	}
	if _, ok := msg.(string); ok {
		msg = Map{"message": msg}
	}

	a.Logger.Error(err)

	// Send response
	if !ctx.Response().Committed {
		if ctx.Request().Method == HEAD { // Issue #608
			err = ctx.NoContent(code)
		} else {
			err = ctx.JSON(code, msg)
		}
		if err != nil {
			a.Logger.Error(err)
		}
	}
}

// Pre adds middleware to the chain which is run before router.
func (a *Akita) Pre(middleware ...MiddlewareFunc) {
	a.premiddleware = append(a.premiddleware, middleware...)
}

// Use adds middleware to the chain which is run after router.
func (a *Akita) Use(middleware ...MiddlewareFunc) {
	a.middleware = append(a.middleware, middleware...)
}

// CONNECT registers a new CONNECT route for a path with matching handler in the
// router with optional route-level middleware.
func (a *Akita) CONNECT(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return a.Add(CONNECT, path, h, m...)
}

// DELETE registers a new DELETE route for a path with matching handler in the router
// with optional route-level middleware.
func (a *Akita) DELETE(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return a.Add(DELETE, path, h, m...)
}

// GET registers a new GET route for a path with matching handler in the router
// with optional route-level middleware.
func (a *Akita) GET(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return a.Add(GET, path, h, m...)
}

// HEAD registers a new HEAD route for a path with matching handler in the
// router with optional route-level middleware.
func (a *Akita) HEAD(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return a.Add(HEAD, path, h, m...)
}

// OPTIONS registers a new OPTIONS route for a path with matching handler in the
// router with optional route-level middleware.
func (a *Akita) OPTIONS(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return a.Add(OPTIONS, path, h, m...)
}

// PATCH registers a new PATCH route for a path with matching handler in the
// router with optional route-level middleware.
func (a *Akita) PATCH(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return a.Add(PATCH, path, h, m...)
}

// POST registers a new POST route for a path with matching handler in the
// router with optional route-level middleware.
func (a *Akita) POST(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return a.Add(POST, path, h, m...)
}

// PUT registers a new PUT route for a path with matching handler in the
// router with optional route-level middleware.
func (a *Akita) PUT(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return a.Add(PUT, path, h, m...)
}

// TRACE registers a new TRACE route for a path with matching handler in the
// router with optional route-level middleware.
func (a *Akita) TRACE(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return a.Add(TRACE, path, h, m...)
}

// Any registers a new route for all HTTP methods and path with matching handler
// in the router with optional route-level middleware.
func (a *Akita) Any(path string, handler HandlerFunc, middleware ...MiddlewareFunc) []*Route {
	routes := make([]*Route, 0)
	for _, m := range methods {
		routes = append(routes, a.Add(m, path, handler, middleware...))
	}
	return routes
}

// Match registers a new route for multiple HTTP methods and path with matching
// handler in the router with optional route-level middleware.
func (a *Akita) Match(methods []string, path string, handler HandlerFunc, middleware ...MiddlewareFunc) []*Route {
	routes := make([]*Route, 0)
	for _, m := range methods {
		routes = append(routes, a.Add(m, path, handler, middleware...))
	}
	return routes
}

// Static registers a new route with path prefix to serve static files from the
// provided root directory.
func (a *Akita) Static(prefix, root string) *Route {
	if root == "" {
		root = "." // For security we want to restrict to CWD.
	}
	return static(a, prefix, root)
}

func static(i i, prefix, root string) *Route {
	h := func(c Context) error {
		p, err := PathUnescape(c.Param("*"))
		if err != nil {
			return err
		}
		name := filepath.Join(root, path.Clean("/"+p)) // "/"+ for security
		return c.File(name)
	}
	i.GET(prefix, h)
	if prefix == "/" {
		return i.GET(prefix+"*", h)
	}

	return i.GET(prefix+"/*", h)
}

// File registers a new route with path to serve a static file.
func (a *Akita) File(path, file string) *Route {
	return a.GET(path, func(ctx Context) error {
		return ctx.File(file)
	})
}

// Add registers a new route for an HTTP method and path with matching handler
// in the router with optional route-level middleware.
func (a *Akita) Add(method, path string, handler HandlerFunc, middleware ...MiddlewareFunc) *Route {
	name := handlerName(handler)
	a.router.Add(method, path, func(ctx Context) error {
		h := handler
		// Chain middleware
		for i := len(middleware) - 1; i >= 0; i-- {
			h = middleware[i](h)
		}
		return h(ctx)
	})
	r := &Route{
		Method: method,
		Path:   path,
		Name:   name,
	}
	a.router.routes[method+path] = r
	return r
}

// Group creates a new router group with prefix and optional group-level middleware.
func (a *Akita) Group(prefix string, m ...MiddlewareFunc) (g *Group) {
	g = &Group{prefix: prefix, akita: a}
	g.Use(m...)
	return
}

// URI generates a URI from handler.
func (a *Akita) URI(handler HandlerFunc, params ...interface{}) string {
	name := handlerName(handler)
	return a.Reverse(name, params...)
}

// URL is an alias for `URI` function.
func (a *Akita) URL(h HandlerFunc, params ...interface{}) string {
	return a.URI(h, params...)
}

// Reverse generates an URL from route name and provided parameters.
func (a *Akita) Reverse(name string, params ...interface{}) string {
	uri := new(bytes.Buffer)
	ln := len(params)
	n := 0
	for _, r := range a.router.routes {
		if r.Name == name {
			for i, l := 0, len(r.Path); i < l; i++ {
				if r.Path[i] == ':' && n < ln {
					for ; i < l && r.Path[i] != '/'; i++ {
					}
					uri.WriteString(fmt.Sprintf("%v", params[n]))
					n++
				}
				if i < l {
					uri.WriteByte(r.Path[i])
				}
			}
			break
		}
	}
	return uri.String()
}

// Routes returns the registered routes.
func (a *Akita) Routes() []*Route {
	routes := []*Route{}
	for _, v := range a.router.routes {
		routes = append(routes, v)
	}
	return routes
}

// AcquireContext returns an empty `Context` instance from the pool.
// You must return the context by calling `ReleaseContext()`.
func (a *Akita) AcquireContext() Context {
	return a.pool.Get().(Context)
}

// ReleaseContext returns the `Context` instance back to the pool.
// You must call it after `AcquireContext()`.
func (a *Akita) ReleaseContext(ctx Context) {
	a.pool.Put(ctx)
}

// ServeHTTP implements `http.Handler` interface, which serves HTTP requests.
func (a *Akita) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Acquire lock
	// e.Mutex.RLock()
	// defer e.Mutex.RUnlock()

	// Acquire context
	ctx := a.pool.Get().(*context)
	defer a.pool.Put(ctx)
	ctx.Reset(r, w)

	// Middleware
	h := func(ctx Context) error {
		method := r.Method
		urlPath := r.URL.RawPath
		if urlPath == "" {
			urlPath = r.URL.Path
		}
		a.router.Find(method, urlPath, ctx)
		h := ctx.Handler()
		for i := len(a.middleware) - 1; i >= 0; i-- {
			h = a.middleware[i](h)
		}
		return h(ctx)
	}

	// Premiddleware
	for i := len(a.premiddleware) - 1; i >= 0; i-- {
		h = a.premiddleware[i](h)
	}

	// Execute chain
	if err := h(ctx); err != nil {
		a.HTTPErrorHandler(err, ctx)
	}
}

// Start starts an HTTP server.
func (a *Akita) Start(address string) error {
	a.Server.Addr = address
	return a.StartServer(a.Server)
}

// StartTLS starts an HTTPS server.
func (a *Akita) StartTLS(address string, certFile, keyFile string) (err error) {
	if certFile == "" || keyFile == "" {
		return errors.New("invalid tls configuration")
	}
	s := a.TLSServer
	s.TLSConfig = new(tls.Config)
	s.TLSConfig.Certificates = make([]tls.Certificate, 1)
	s.TLSConfig.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return
	}
	return a.startTLS(address)
}

// StartAutoTLS starts an HTTPS server using certificates automatically installed from https://letsencrypt.org.
func (a *Akita) StartAutoTLS(address string) error {
	s := a.TLSServer
	s.TLSConfig = new(tls.Config)
	s.TLSConfig.GetCertificate = a.AutoTLSManager.GetCertificate
	return a.startTLS(address)
}

func (a *Akita) startTLS(address string) error {
	s := a.TLSServer
	s.Addr = address
	if !a.DisableHTTP2 {
		s.TLSConfig.NextProtos = append(s.TLSConfig.NextProtos, "h2")
	}
	return a.StartServer(a.TLSServer)
}

// StartServer starts a custom http server.
func (a *Akita) StartServer(s *http.Server) (err error) {
	// Setup
	a.colorer.SetOutput(a.Logger.Output())
	s.ErrorLog = a.stdLogger
	s.Handler = a
	if a.Debug {
		a.Logger.SetLevel(log.DEBUG)
	}

	if !a.HideBanner {
		a.colorer.Printf(banner, a.colorer.Red("v"+version), a.colorer.Blue(website))
	}

	if s.TLSConfig == nil {
		if a.Listener == nil {
			a.Listener, err = newListener(s.Addr)
			if err != nil {
				return err
			}
		}
		if !a.HideBanner {
			a.colorer.Printf("⇨ http server started on %s\n", a.colorer.Green(a.Listener.Addr()))
		}
		return s.Serve(a.Listener)
	}
	if a.TLSListener == nil {
		l, err := newListener(s.Addr)
		if err != nil {
			return err
		}
		a.TLSListener = tls.NewListener(l, s.TLSConfig)
	}
	if !a.HideBanner {
		a.colorer.Printf("⇨ https server started on %s\n", a.colorer.Green(a.TLSListener.Addr()))
	}
	return s.Serve(a.TLSListener)
}

// NewHTTPError creates a new HTTPError instance.
func NewHTTPError(code int, message ...interface{}) *HTTPError {
	he := &HTTPError{Code: code, Message: http.StatusText(code)}
	if len(message) > 0 {
		he.Message = message[0]
	}
	return he
}

// Error makes it compatible with `error` interface.
func (he *HTTPError) Error() string {
	return fmt.Sprintf("code=%d, message=%v", he.Code, he.Message)
}

// WrapHandler wraps `http.Handler` into `akita.HandlerFunc`.
func WrapHandler(h http.Handler) HandlerFunc {
	return func(ctx Context) error {
		h.ServeHTTP(ctx.Response(), ctx.Request())
		return nil
	}
}

// WrapMiddleware wraps `func(http.Handler) http.Handler` into `akita.MiddlewareFunc`
func WrapMiddleware(m func(http.Handler) http.Handler) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx Context) (err error) {
			m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx.SetRequest(r)
				err = next(ctx)
			})).ServeHTTP(ctx.Response(), ctx.Request())
			return
		}
	}
}

func handlerName(h HandlerFunc) string {
	t := reflect.ValueOf(h).Type()
	if t.Kind() == reflect.Func {
		return runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
	}
	return t.String()
}

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}

func newListener(address string) (*tcpKeepAliveListener, error) {
	l, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	return &tcpKeepAliveListener{l.(*net.TCPListener)}, nil
}
