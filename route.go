package flog

import (
	"fmt"
	//"github.com/codegangsta/inject"
    "../inject"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
)

// Route is an interface representing a Route in Flog's routing layer.
type Route interface {
	// URLWith returns a rendering of the Route's url with the given string params.
	urlFor(params ...interface{}) string
}

type route struct {
    endpoint string
	method   string
	regex    *regexp.Regexp
	handlers []Handler
	pattern  string
}

func newRoute(endpoint string, method string, pattern string, handlers []Handler) *route {
	route := route{endpoint, method, nil, handlers, pattern}
	r := regexp.MustCompile(`:[^/#?()\.\\]+`)
	pattern = r.ReplaceAllStringFunc(pattern, func(m string) string {
		return fmt.Sprintf(`(?P<%s>[^/#?]+)`, m[1:])
	})
	r2 := regexp.MustCompile(`\*\*`)
	var index int
	pattern = r2.ReplaceAllStringFunc(pattern, func(m string) string {
		index++
		return fmt.Sprintf(`(?P<_%d>[^#?]*)`, index)
	})
	pattern += `\/?`
	route.regex = regexp.MustCompile(pattern)
	return &route
}

func (r route) Match(method string, path string) (bool, map[string]string) {
	// add Any method matching support
	if r.method != "*" && method != r.method {
		return false, nil
	}

	matches := r.regex.FindStringSubmatch(path)
	if len(matches) > 0 && matches[0] == path {
		params := make(map[string]string)
		for i, name := range r.regex.SubexpNames() {
			if len(name) > 0 {
				params[name] = matches[i]
			}
		}
		return true, params
	}
	return false, nil
}

func (r *route) Validate() {
	for _, handler := range r.handlers {
		validateHandler(handler)
	}
}

func (r *route) Handle(c Context, res http.ResponseWriter) {
	ctx := &routeContext{c, 0, r.handlers}
	c.MapTo(ctx, (*Context)(nil))
	ctx.run()
}


// URLFor returns the url for the given route name.
func (r *route) urlFor(params ...interface{}) string {
    args := urlArgs(params)
    if len(args) > 0 {
		reg := regexp.MustCompile(`:[^/#?()\.\\]+`)
		argCount := len(args)
		i := 0
		url := reg.ReplaceAllStringFunc(r.pattern, func(m string) string {
			var val interface{}
			if i < argCount {
				val = args[i]
			} else {
				val = m
			}
			i += 1
			return fmt.Sprintf(`%v`, val)
		})

		return url
	}
	return r.pattern
}

// args for URLFor
func urlArgs(params ...interface{}) []string {
    var args []string
	for _, param := range params {
		switch v := param.(type) {
		case int:
			args = append(args, strconv.FormatInt(int64(v), 10))
		case string:
			args = append(args, v)
		default:
			if v != nil {
				panic("Arguments passed to UrlFor must be integers or strings")
			}
		}
	}
    return args
}

type routeContext struct {
	Context
	index    int
	handlers []Handler
}

func (r *routeContext) Next() {
	r.index += 1
	r.run()
}

func (r *routeContext) run() {
	for r.index < len(r.handlers) {
		handler := r.handlers[r.index]
		vals, err := r.Invoke(handler)
		if err != nil {
			panic(err)
		}
		r.index += 1

		// if the handler returned something, write it to
		// the http response
		rv := r.Get(inject.InterfaceOf((*http.ResponseWriter)(nil)))
		res := rv.Interface().(http.ResponseWriter)
		if len(vals) > 1 && vals[0].Kind() == reflect.Int {
			res.WriteHeader(int(vals[0].Int()))
			res.Write([]byte(vals[1].String()))
		} else if len(vals) > 0 {
			res.Write([]byte(vals[0].String()))
		}
		if r.written() {
			return
		}
	}
}
