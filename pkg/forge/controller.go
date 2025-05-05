package flux

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/gofiber/fiber/v2"
)


type Controller struct {
	app        *Application
	middleware []MiddlewareFunc
	name       string
	routes     map[string]*Route
}


type Route struct {
	Name        string
	Method      string
	Path        string
	Description string
	RequestBody interface{}
	Response    interface{}
	Handler     HandlerFunc
}

type HandlerFunc func(*Context) error

type MiddlewareFunc func(HandlerFunc) HandlerFunc

func (c *Controller) Use(middleware ...MiddlewareFunc) {
	c.middleware = append(c.middleware, middleware...)
}


func (c *Controller) RegisterRoutes(router fiber.Router) {
	if c.routes == nil {
		c.routes = make(map[string]*Route)
	}
	
	t := reflect.TypeOf(c)
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		if strings.HasPrefix(method.Name, "Handle") {
			c.registerRoute(router, method)
		}
	}
}

func (c *Controller) RegisterRoute(method, path, description string, handler HandlerFunc) *Route {
	if c.routes == nil {
		c.routes = make(map[string]*Route)
	}
	
	route := &Route{
		Method:      strings.ToUpper(method),
		Path:        path,
		Description: description,
		Handler:     handler,
	}
	
	routeKey := fmt.Sprintf("%s:%s", route.Method, route.Path)
	c.routes[routeKey] = route
	
	
	if c.app != nil && c.app.routes != nil {
		handlerName := fmt.Sprintf("%s.CustomHandler", c.Name())
		c.app.routes.Add(route.Method, route.Path, handlerName, description)
	}
	
	return route
}

func (r *Route) SetName(name string) *Route {
	r.Name = name
	return r
}

func (r *Route) SetRequestBody(model interface{}) *Route {
	r.RequestBody = model
	return r
}

func (r *Route) SetResponse(model interface{}) *Route {
	r.Response = model
	return r
}

func (c *Controller) registerRoute(router fiber.Router, method reflect.Method) {
	name := strings.TrimPrefix(method.Name, "Handle")
	parts := splitCamelCase(name)

	var httpMethod string
	var path string

	if len(parts) > 0 {
		httpMethod = strings.ToUpper(parts[0])
		if len(parts) > 1 {
			path = "/" + strings.ToLower(strings.Join(parts[1:], "/"))
		}
	}

	route := &Route{
		Method:      httpMethod,
		Path:        path,
		Name:        method.Name,
		Description: "", // We could extract from comments in the future not now
	}
	
	routeKey := fmt.Sprintf("%s:%s", route.Method, route.Path)
	if c.routes == nil {
		c.routes = make(map[string]*Route)
	}
	c.routes[routeKey] = route
	
	
	if c.app != nil && c.app.routes != nil {
		handlerName := fmt.Sprintf("%s.%s", c.Name(), method.Name)
		c.app.routes.Add(route.Method, route.Path, handlerName, route.Description)
	}

	handler := func(ctx *fiber.Ctx) error {
		fluxCtx := NewContext(ctx, c.app)
		finalHandler := func(ctx *Context) error {
			result := method.Func.Call([]reflect.Value{
				reflect.ValueOf(c),
				reflect.ValueOf(ctx),
			})
			
			if !result[0].IsNil() {
				return result[0].Interface().(error)
			}
			return nil
		}

		// Apply middleware in reverse order (LIFO)
		if len(c.middleware) > 0 {
			chain := finalHandler
			for i := len(c.middleware) - 1; i >= 0; i-- {
				chain = c.middleware[i](chain)
			}
			return chain(fluxCtx)
		}

		return finalHandler(fluxCtx)
	}

	switch httpMethod {
	case "GET":
		router.Get(path, handler)
	case "POST":
		router.Post(path, handler)
	case "PUT":
		router.Put(path, handler)
	case "DELETE":
		router.Delete(path, handler)
	case "PATCH":
		router.Patch(path, handler)
	case "OPTIONS":
		router.Options(path, handler)
	case "HEAD":
		router.Head(path, handler)
	}
}

func splitCamelCase(s string) []string {
	var parts []string
	var current strings.Builder

	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			parts = append(parts, current.String())
			current.Reset()
		}
		current.WriteRune(r)
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

func (c *Controller) SetApplication(app *Application) {
	c.app = app
}

func (c *Controller) App() *Application {
	return c.app
}

func (c *Controller) SetName(name string) {
	c.name = name
}

func (c *Controller) Name() string {
	if c.name != "" {
		return c.name
	}
	
	t := reflect.TypeOf(c)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}

func (c *Controller) GetRoutes() map[string]*Route {
	return c.routes
}

func (c *Controller) GetRouteByName(name string) *Route {
	for _, route := range c.routes {
		if route.Name == name {
			return route
		}
	}
	return nil
}

func (c *Controller) Group(prefix string) *ControllerGroup {
	return &ControllerGroup{
		prefix:     prefix,
		middleware: c.middleware,
	}
}

type ControllerGroup struct {
	prefix      string
	middleware  []MiddlewareFunc
	controllers []interface{}
	name        string
}

func (g *ControllerGroup) Use(middleware ...MiddlewareFunc) *ControllerGroup {
	g.middleware = append(g.middleware, middleware...)
	return g
}


func (g *ControllerGroup) Add(controller interface{}) *ControllerGroup {
	g.controllers = append(g.controllers, controller)

	if c, ok := controller.(*Controller); ok {
		c.middleware = append(c.middleware, g.middleware...)
	}

	return g
}

func (g *ControllerGroup) SetName(name string) *ControllerGroup {
	g.name = name
	return g
}


func (g *ControllerGroup) Register(app *Application) {
	router := app.Group(g.prefix)
	
	for _, controller := range g.controllers {
		if c, ok := controller.(*Controller); ok {
			
			for _, mw := range g.middleware {
				router.Use(func(c *fiber.Ctx) error {
					fluxCtx := NewContext(c, app)
					var err error
					wrapped := mw(func(ctx *Context) error {
						return ctx.Next()
					})
					err = wrapped(fluxCtx)
					if err != nil {
						return err
					}
					return c.Next()
				})
			}
			c.RegisterRoutes(router)
		} else {
			app.RegisterController(controller)
		}
	}
}
