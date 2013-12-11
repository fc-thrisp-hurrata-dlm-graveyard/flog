package flog

import (
	"github.com/codegangsta/inject"
	"log"
	"net/http"
	"os"
	"reflect"
)

const VERSION = "0.0.1"

// App represents the top level web application.
// inject.Injector methods can be invoked to map services on a global level.
type App struct {
	inject.Injector
	handlers []Handler
	action   Handler
	logger   *log.Logger
}

// New creates a bare bones Flog instance.
// Use this method if you want to have full control over the middleware that is used.
func New() *App {
	a := &App{inject.New(), []Handler{}, func() {}, log.New(os.Stdout, "[flog app] ", 0)}
	a.Map(a.logger)
	return a
}

// Handler can be any callable function.
type Handler interface{}

func validateHandler(handler Handler) {
	if reflect.TypeOf(handler).Kind() != reflect.Func {
		panic("app handler must be a callable func")
	}
}

// Use adds a middleware Service to the stack.
// Will panic if the service is not a callable func.
// Middleware services are invoked in the order that they are added.
func (a *App) Use(handler Handler) {
	validateHandler(handler)
	a.handlers = append(a.handlers, handler)
}

// Handlers sets the entire middleware stack with the given Handlers. This will clear any current middleware handlers.
// Will panic if any of the handlers is not a callable function
func (a *App) Handlers(handlers ...Handler) {
	a.handlers = make([]Handler, 0)
	for _, handler := range handlers {
		a.Use(handler)
	}
}

// ServeHTTP is the HTTP Entry point for a Flog instance. Useful if you want to control your own HTTP server.
func (a *App) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	a.createContext(res, req).run()
}

// Action sets the handler that will be called after all the middleware has been invoked.
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

// represents an App with reasonable defaults. Embeds the router functions for convenience.
type FlogApplication struct {
	*App
	Router
}

// Flog with some basic default middleware
func Flog() *FlogApplication {
	r := NewRouter()
	a := New()
    a.Handlers(Logger(), Recovery(), Static("static"), Renderer("templates")) 
	a.Action(r.Handle)
	return &FlogApplication{a, r}
}

func (a *App) createContext(res http.ResponseWriter, req *http.Request) *context {
	c := &context{inject.New(), append(a.handlers, a.action), NewResponseWriter(res), 0}
	c.SetParent(a)
	c.MapTo(c, (*Context)(nil))
	c.MapTo(c.rw, (*http.ResponseWriter)(nil))
	c.Map(req)
	return c
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
