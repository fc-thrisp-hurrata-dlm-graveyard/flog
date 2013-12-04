package flog

import (
	"github.com/codegangsta/inject"
	"log"
	"net/http"
	"os"
	"reflect"
)

// App represents the top level web application. inject.Injector methods can be invoked to map services on a global level.
type App struct {
	inject.Injector
	handlers []Handler
	action   Handler
	logger   *log.Logger
}

// New creates a bare bones Flog instance. Use this method if you want to have full control over the middleware that is used.
func New() *App {
	a := &App{inject.New(), []Handler{}, func() {}, log.New(os.Stdout, "[flog app] ", 0)}
	a.Map(a.logger)
	return a
}

// Use adds a middleware Handler to the stack. Will panic if the handler is not a callable func. Middleware Handlers are invoked in the order that they are added.
func (a *App) Use(handler Handler) {
	validateHandler(handler)

	a.handlers = append(a.handlers, handler)
}

// ServeHTTP is the HTTP Entry point for a Flog instance. Useful if you want to control your own HTTP server.
func (a *App) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	a.createContext(res, req).run()
}

// Action sets the handler that will be called after all the middleware has been invoked. This is set to flog.Router in a flog.Classic().
func (a *App) Action(handler Handler) {
	validateHandler(handler)
	a.action = handler
}

// Run the http server. Listening on os.GetEnv("PORT") or 3000 by default.
func (a *App) Run() {
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "3000"
	}

	a.logger.Println("listening on port " + port)
	a.logger.Fatalln(http.ListenAndServe(":"+port, a))
}

// Handlers sets the entire middleware stack with the given Handlers. This will clear any current middleware handlers.
// Will panic if any of the handlers is not a callable function
func (a *App) Handlers(handlers ...Handler) {
	a.handlers = make([]Handler, 0)
	for _, handler := range handlers {
		a.Use(handler)
	}
}

func (a *App) createContext(res http.ResponseWriter, req *http.Request) *context {
	c := &context{inject.New(), append(a.handlers, a.action), NewResponseWriter(res), 0}
	c.SetParent(a)
	c.MapTo(c, (*Context)(nil))
	c.MapTo(c.rw, (*http.ResponseWriter)(nil))
	c.Map(req)
	return c
}

// represents an App with reasonable defaults. Embeds the router functions for convenience.
type FlogApplication struct {
	*App
	Router
}

// Classic creates a classic Flog with some basic default middleware - flog.Logger, flog.Recovery, flog.Static, flog.Renderer.
func Flog() *FlogApplication {
	r := NewRouter()
	fa := New()
    fa.Handlers(Logger(), Recovery(), Static("static"), Renderer("templates")) 
	fa.Action(r.Handle)
	return &FlogApplication{fa, r}
}

// Handler can be any callable function. Flog attempts to inject services into the handler's argument list.
// Flog will panic if an argument could not be fullfilled via dependency injection.
type Handler interface{}

func validateHandler(handler Handler) {
	if reflect.TypeOf(handler).Kind() != reflect.Func {
		panic("app handler must be a callable func")
	}
}

// Context represents a request context. Services can be mapped on the request level from this interface.
type Context interface {
	inject.Injector
	// Next is an optional function that Middleware Handlers can call to yield the until after
	// the other Handlers have been executed. This works really well for any operations that must
	// happen after an http request
	Next()
	written() bool
}

type context struct {
	inject.Injector
	handlers []Handler
	rw       ResponseWriter
	index    int
}

func (c *context) Next() {
	c.index += 1
	c.run()
}

func (c *context) written() bool {
	return c.rw.Written()
}

func (c *context) run() {
	for c.index < len(c.handlers) {
		_, err := c.Invoke(c.handlers[c.index])
		if err != nil {
			panic(err)
		}
		c.index += 1

		if c.rw.Written() {
			return
		}
	}
}
