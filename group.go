package akita

import (
	"path"
)

type (
	// Group is a set of sub-routes for a specified route. It can be used for inner
	// routes that share a common middleware or functionality that should be separate
	// from the parent akita instance while still inheriting from it.
	Group struct {
		prefix     string
		middleware []MiddlewareFunc
		akita      *Akita
	}
)

// Use implements `Akita#Use()` for sub-routes within the Group.
func (g *Group) Use(middleware ...MiddlewareFunc) {
	g.middleware = append(g.middleware, middleware...)
	// Allow all requests to reach the group as they might get dropped if router
	// doesn't find a match, making none of the group middleware process.
	g.akita.Any(path.Clean(g.prefix+"/*"), func(c Context) error {
		return NotFoundHandler(c)
	}, g.middleware...)
}

// CONNECT implements `Akita#CONNECT()` for sub-routes within the Group.
func (g *Group) CONNECT(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return g.Add(CONNECT, path, h, m...)
}

// DELETE implements `Akita#DELETE()` for sub-routes within the Group.
func (g *Group) DELETE(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return g.Add(DELETE, path, h, m...)
}

// GET implements `Akita#GET()` for sub-routes within the Group.
func (g *Group) GET(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return g.Add(GET, path, h, m...)
}

// HEAD implements `Akita#HEAD()` for sub-routes within the Group.
func (g *Group) HEAD(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return g.Add(HEAD, path, h, m...)
}

// OPTIONS implements `Akita#OPTIONS()` for sub-routes within the Group.
func (g *Group) OPTIONS(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return g.Add(OPTIONS, path, h, m...)
}

// PATCH implements `Akita#PATCH()` for sub-routes within the Group.
func (g *Group) PATCH(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return g.Add(PATCH, path, h, m...)
}

// POST implements `Akita#POST()` for sub-routes within the Group.
func (g *Group) POST(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return g.Add(POST, path, h, m...)
}

// PUT implements `Akita#PUT()` for sub-routes within the Group.
func (g *Group) PUT(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return g.Add(PUT, path, h, m...)
}

// TRACE implements `Akita#TRACE()` for sub-routes within the Group.
func (g *Group) TRACE(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return g.Add(TRACE, path, h, m...)
}

// Any implements `Akita#Any()` for sub-routes within the Group.
func (g *Group) Any(path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	for _, m := range methods {
		g.Add(m, path, handler, middleware...)
	}
}

// Match implements `Akita#Match()` for sub-routes within the Group.
func (g *Group) Match(methods []string, path string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	for _, m := range methods {
		g.Add(m, path, handler, middleware...)
	}
}

// Group creates a new sub-group with prefix and optional sub-group-level middleware.
func (g *Group) Group(prefix string, middleware ...MiddlewareFunc) *Group {
	m := []MiddlewareFunc{}
	m = append(m, g.middleware...)
	m = append(m, middleware...)
	return g.akita.Group(g.prefix+prefix, m...)
}

// Static implements `Akita#Static()` for sub-routes within the Group.
func (g *Group) Static(prefix, root string) {
	static(g, prefix, root)
}

// File implements `Akita#File()` for sub-routes within the Group.
func (g *Group) File(path, file string) {
	g.akita.File(g.prefix+path, file)
}

// Add implements `Akita#Add()` for sub-routes within the Group.
func (g *Group) Add(method, path string, handler HandlerFunc, middleware ...MiddlewareFunc) *Route {
	// Combine into a new slice to avoid accidentally passing the same slice for
	// multiple routes, which would lead to later add() calls overwriting the
	// middleware from earlier calls.
	m := []MiddlewareFunc{}
	m = append(m, g.middleware...)
	m = append(m, middleware...)
	return g.akita.Add(method, g.prefix+path, handler, m...)
}
