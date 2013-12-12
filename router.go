package flog 

import (
	//"fmt"
	//"github.com/codegangsta/inject"
	"net/http"
	//"reflect"
	//"regexp"
	//"strconv"
)

// Params is a map of name/value pairs for named routes. An instance of flag.Params is available to be injected into any route handler.
type Params map[string]string

// Router is Martini's de-facto routing interface. Supports HTTP verbs, stacked handlers, and dependency injection.
type Router interface {
	// Get adds a route for a HTTP GET request to the specified matching pattern.
	Get(string, string, ...Handler) Route
	// Patch adds a route for a HTTP PATCH request to the specified matching pattern.
	Patch(string, string, ...Handler) Route
	// Post adds a route for a HTTP POST request to the specified matching pattern.
	Post(string, string, ...Handler) Route
	// Put adds a route for a HTTP PUT request to the specified matching pattern.
	Put(string, string, ...Handler) Route
	// Delete adds a route for a HTTP DELETE request to the specified matching pattern.
	Delete(string, string, ...Handler) Route
	// Options adds a route for a HTTP OPTIONS request to the specified matching pattern.
	Options(string, string, ...Handler) Route
    // Head adds a route for a HTTP HEAD request to the specified matching pattern.
    Head(string, string, ...Handler) Route
	// Any adds a route for any HTTP method request to the specified matching pattern.
	Any(string, string, ...Handler) Route

	// NotFound sets the handlers that are called when a no route matches a request. Throws a basic 404 by default.
	NotFound(...Handler)

	// Handle is the entry point for routing. This is used as a martini.Handler
	Handle(http.ResponseWriter, *http.Request, Context)
 
    UrlFor(string, ...interface{}) string
}

type router struct {
    routes map[string]*route
	notFounds []Handler
}

// NewRouter creates a new Router instance.
func NewRouter() Router {
    return &router{routes: make(map[string]*route),
                   notFounds: []Handler{http.NotFound}}
}

func (r *router) Get(endpoint string, pattern string, h ...Handler) Route {
	return r.addRoute(endpoint, "GET", pattern, h)
}

func (r *router) Patch(endpoint string, pattern string, h ...Handler) Route {
	return r.addRoute(endpoint, "PATCH", pattern, h)
}

func (r *router) Post(endpoint string, pattern string, h ...Handler) Route {
	return r.addRoute(endpoint, "POST", pattern, h)
}

func (r *router) Put(endpoint string, pattern string, h ...Handler) Route {
	return r.addRoute(endpoint, "PUT", pattern, h)
}

func (r *router) Delete(endpoint string, pattern string, h ...Handler) Route {
	return r.addRoute(endpoint, "DELETE", pattern, h)
}

func (r *router) Options(endpoint string, pattern string, h ...Handler) Route {
	return r.addRoute(endpoint, "OPTIONS", pattern, h)
}

func (r *router) Head(endpoint string, pattern string, h ...Handler) Route {
    return r.addRoute(endpoint, "HEAD", pattern, h)
}

func (r *router) Any(endpoint string, pattern string, h ...Handler) Route {
	return r.addRoute(endpoint, "*", pattern, h)
}

func (r *router) Handle(res http.ResponseWriter, req *http.Request, ctx Context) {
	for _, route := range r.routes {
		ok, vals := route.Match(req.Method, req.URL.Path)
		if ok {
			params := Params(vals)
			ctx.Map(params)
			//r := routes{}
			//ctx.MapTo(r, (*Routes)(nil))
			_, err := ctx.Invoke(route.Handle)
			if err != nil {
				panic(err)
			}
			return
		}
	}

	// no routes exist, 404
	c := &routeContext{ctx, 0, r.notFounds}
	ctx.MapTo(c, (*Context)(nil))
	c.run()
}

func (r *router) NotFound(handler ...Handler) {
	r.notFounds = handler
}

func (r *router) addRoute(endpoint string, method string, pattern string, handlers []Handler) *route {
	route := newRoute(endpoint, method, pattern, handlers)
	route.Validate()
    r.routes[endpoint] = route
	return route
}

func (r *router) UrlFor(endpoint string, params ...interface{}) string {
    rte := r.routes[endpoint]
    return rte.urlFor(params)
}
