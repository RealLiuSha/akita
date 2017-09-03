package akita

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"reflect"
	"strings"

	"errors"

	"time"

	"github.com/stretchr/testify/assert"
)

type (
	user struct {
		ID   int    `json:"id" xml:"id" form:"id" query:"id"`
		Name string `json:"name" xml:"name" form:"name" query:"name"`
	}
)

const (
	userJSON       = `{"id":1,"name":"Jon Snow"}`
	userXML        = `<user><id>1</id><name>Jon Snow</name></user>`
	userForm       = `id=1&name=Jon Snow`
	invalidContent = "invalid content"
)

const userJSONPretty = `{
  "id": 1,
  "name": "Jon Snow"
}`

const userXMLPretty = `<user>
  <id>1</id>
  <name>Jon Snow</name>
</user>`

func TestAkita(t *testing.T) {
	a := New()
	req := httptest.NewRequest(GET, "/", nil)
	rec := httptest.NewRecorder()
	ctx := a.NewContext(req, rec)

	// Router
	assert.NotNil(t, a.Router())

	// DefaultHTTPErrorHandler
	a.DefaultHTTPErrorHandler(errors.New("error"), ctx)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestAkitaStatic(t *testing.T) {
	a := New()

	// OK
	a.Static("/images", "_fixture/images")
	c, b := request(GET, "/images/akita.png", a)
	assert.Equal(t, http.StatusOK, c)
	assert.NotEmpty(t, b)

	// No file
	a.Static("/images", "_fixture/scripts")
	c, _ = request(GET, "/images/bolt.png", a)
	assert.Equal(t, http.StatusNotFound, c)

	// Directory
	a.Static("/images", "_fixture/images")
	c, _ = request(GET, "/images", a)
	assert.Equal(t, http.StatusNotFound, c)

	// Directory with index.html
	a.Static("/", "_fixture")
	c, r := request(GET, "/", a)
	assert.Equal(t, http.StatusOK, c)
	assert.Equal(t, true, strings.HasPrefix(r, "<!doctype html>"))

	// Sub-directory with index.html
	c, r = request(GET, "/folder", a)
	assert.Equal(t, http.StatusOK, c)
	assert.Equal(t, true, strings.HasPrefix(r, "<!doctype html>"))
}

func TestAkitaFile(t *testing.T) {
	a := New()
	a.File("/akita", "_fixture/images/akita.png")
	c, b := request(GET, "/akita", a)
	assert.Equal(t, http.StatusOK, c)
	assert.NotEmpty(t, b)
}

func TestAkitaMiddleware(t *testing.T) {
	a := New()
	buf := new(bytes.Buffer)

	a.Pre(func(next HandlerFunc) HandlerFunc {
		return func(ctx Context) error {
			assert.Empty(t, ctx.Path())
			buf.WriteString("-1")
			return next(ctx)
		}
	})

	a.Use(func(next HandlerFunc) HandlerFunc {
		return func(ctx Context) error {
			buf.WriteString("1")
			return next(ctx)
		}
	})

	a.Use(func(next HandlerFunc) HandlerFunc {
		return func(ctx Context) error {
			buf.WriteString("2")
			return next(ctx)
		}
	})

	a.Use(func(next HandlerFunc) HandlerFunc {
		return func(ctx Context) error {
			buf.WriteString("3")
			return next(ctx)
		}
	})

	// Route
	a.GET("/", func(ctx Context) error {
		return ctx.String(http.StatusOK, "OK")
	})

	c, b := request(GET, "/", a)
	assert.Equal(t, "-1123", buf.String())
	assert.Equal(t, http.StatusOK, c)
	assert.Equal(t, "OK", b)
}

func TestAkitaMiddlewareError(t *testing.T) {
	a := New()
	a.Use(func(next HandlerFunc) HandlerFunc {
		return func(ctx Context) error {
			return errors.New("error")
		}
	})
	a.GET("/", NotFoundHandler)
	c, _ := request(GET, "/", a)
	assert.Equal(t, http.StatusInternalServerError, c)
}

func TestAkitaHandler(t *testing.T) {
	a := New()

	// HandlerFunc
	a.GET("/ok", func(c Context) error {
		return c.String(http.StatusOK, "OK")
	})

	c, b := request(GET, "/ok", a)
	assert.Equal(t, http.StatusOK, c)
	assert.Equal(t, "OK", b)
}

func TestAkitaWrapHandler(t *testing.T) {
	a := New()
	req := httptest.NewRequest(GET, "/", nil)
	rec := httptest.NewRecorder()
	ctx := a.NewContext(req, rec)
	h := WrapHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	}))
	if assert.NoError(t, h(ctx)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "test", rec.Body.String())
	}
}

func TestAkitaWrapMiddleware(t *testing.T) {
	a := New()
	req := httptest.NewRequest(GET, "/", nil)
	rec := httptest.NewRecorder()
	ctx := a.NewContext(req, rec)
	buf := new(bytes.Buffer)
	mw := WrapMiddleware(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			buf.Write([]byte("mw"))
			h.ServeHTTP(w, r)
		})
	})
	h := mw(func(c Context) error {
		return c.String(http.StatusOK, "OK")
	})
	if assert.NoError(t, h(ctx)) {
		assert.Equal(t, "mw", buf.String())
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "OK", rec.Body.String())
	}
}

func TestAkitaConnect(t *testing.T) {
	a := New()
	testMethod(t, CONNECT, "/", a)
}

func TestAkitaDelete(t *testing.T) {
	a := New()
	testMethod(t, DELETE, "/", a)
}

func TestAkitaGet(t *testing.T) {
	a := New()
	testMethod(t, GET, "/", a)
}

func TestAkitaHead(t *testing.T) {
	a := New()
	testMethod(t, HEAD, "/", a)
}

func TestAkitaOptions(t *testing.T) {
	a := New()
	testMethod(t, OPTIONS, "/", a)
}

func TestAkitaPatch(t *testing.T) {
	a := New()
	testMethod(t, PATCH, "/", a)
}

func TestAkitaPost(t *testing.T) {
	a := New()
	testMethod(t, POST, "/", a)
}

func TestAkitaPut(t *testing.T) {
	a := New()
	testMethod(t, PUT, "/", a)
}

func TestAkitaTrace(t *testing.T) {
	a := New()
	testMethod(t, TRACE, "/", a)
}

func TestAkitaAny(t *testing.T) { // JFC
	a := New()
	a.Any("/", func(ctx Context) error {
		return ctx.String(http.StatusOK, "Any")
	})
}

func TestAkitaMatch(t *testing.T) { // JFC
	a := New()
	a.Match([]string{GET, POST}, "/", func(ctx Context) error {
		return ctx.String(http.StatusOK, "Match")
	})
}

func TestAkitaURL(t *testing.T) {
	a := New()
	static := func(Context) error { return nil }
	getUser := func(Context) error { return nil }
	getFile := func(Context) error { return nil }

	a.GET("/static/file", static)
	a.GET("/users/:id", getUser)
	g := a.Group("/group")
	g.GET("/users/:uid/files/:fid", getFile)

	assert.Equal(t, "/static/file", a.URL(static))
	assert.Equal(t, "/users/:id", a.URL(getUser))
	assert.Equal(t, "/users/1", a.URL(getUser, "1"))
	assert.Equal(t, "/group/users/1/files/:fid", a.URL(getFile, "1"))
	assert.Equal(t, "/group/users/1/files/1", a.URL(getFile, "1", "1"))
}

func TestAkitaRoutes(t *testing.T) {
	a := New()
	routes := []*Route{
		{GET, "/users/:user/events", ""},
		{GET, "/users/:user/events/public", ""},
		{POST, "/repos/:owner/:repo/git/refs", ""},
		{POST, "/repos/:owner/:repo/git/tags", ""},
	}
	for _, r := range routes {
		a.Add(r.Method, r.Path, func(c Context) error {
			return c.String(http.StatusOK, "OK")
		})
	}

	if assert.Equal(t, len(routes), len(a.Routes())) {
		for _, r := range a.Routes() {
			found := false
			for _, rr := range routes {
				if r.Method == rr.Method && r.Path == rr.Path {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Route %s %s not found", r.Method, r.Path)
			}
		}
	}
}

func TestAkitaEncodedPath(t *testing.T) {
	a := New()
	a.GET("/:id", func(ctx Context) error {
		return ctx.NoContent(http.StatusOK)
	})
	req := httptest.NewRequest(GET, "/with%2Fslash", nil)
	rec := httptest.NewRecorder()
	a.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAkitaGroup(t *testing.T) {
	a := New()
	buf := new(bytes.Buffer)
	a.Use(MiddlewareFunc(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			buf.WriteString("0")
			return next(c)
		}
	}))
	h := func(c Context) error {
		return c.NoContent(http.StatusOK)
	}

	//--------
	// Routes
	//--------

	a.GET("/users", h)

	// Group
	g1 := a.Group("/group1")
	g1.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			buf.WriteString("1")
			return next(c)
		}
	})
	g1.GET("", h)

	// Nested groups with middleware
	g2 := a.Group("/group2")
	g2.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			buf.WriteString("2")
			return next(c)
		}
	})
	g3 := g2.Group("/group3")
	g3.Use(func(next HandlerFunc) HandlerFunc {
		return func(c Context) error {
			buf.WriteString("3")
			return next(c)
		}
	})
	g3.GET("", h)

	request(GET, "/users", a)
	assert.Equal(t, "0", buf.String())

	buf.Reset()
	request(GET, "/group1", a)
	assert.Equal(t, "01", buf.String())

	buf.Reset()
	request(GET, "/group2/group3", a)
	assert.Equal(t, "023", buf.String())
}

func TestAkitaNotFound(t *testing.T) {
	a := New()
	req := httptest.NewRequest(GET, "/files", nil)
	rec := httptest.NewRecorder()
	a.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAkitaMethodNotAllowed(t *testing.T) {
	a := New()
	a.GET("/", func(c Context) error {
		return c.String(http.StatusOK, "Akita!")
	})
	req := httptest.NewRequest(POST, "/", nil)
	rec := httptest.NewRecorder()
	a.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestAkitaContext(t *testing.T) {
	a := New()
	c := a.AcquireContext()
	assert.IsType(t, new(context), c)
	a.ReleaseContext(c)
}

func TestAkitaStart(t *testing.T) {
	a := New()
	go func() {
		assert.NoError(t, a.Start(":0"))
	}()
	time.Sleep(200 * time.Millisecond)
}

func TestAkitaStartTLS(t *testing.T) {
	a := New()
	go func() {
		assert.NoError(t, a.StartTLS(":0", "_fixture/certs/cert.pem", "_fixture/certs/key.pem"))
	}()
	time.Sleep(200 * time.Millisecond)
}

func testMethod(t *testing.T, method, path string, a *Akita) {
	p := reflect.ValueOf(path)
	h := reflect.ValueOf(func(c Context) error {
		return c.String(http.StatusOK, method)
	})
	i := interface{}(a)
	reflect.ValueOf(i).MethodByName(method).Call([]reflect.Value{p, h})
	_, body := request(method, path, a)
	assert.Equal(t, method, body)
}

func request(method, path string, a *Akita) (int, string) {
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	a.ServeHTTP(rec, req)
	return rec.Code, rec.Body.String()
}

func TestHTTPError(t *testing.T) {
	err := NewHTTPError(400, map[string]interface{}{
		"code": 12,
	})
	assert.Equal(t, "code=400, message=map[code:12]", err.Error())
}
