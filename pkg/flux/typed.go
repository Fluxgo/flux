package flux

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// Endpoint is a handler with a compile-time response contract.
type Endpoint[Response any] func(*Context) (Response, error)

// BodyEndpoint is a handler with compile-time request and response contracts.
type BodyEndpoint[Request, Response any] func(*Context, Request) (Response, error)

// Body parses and validates a JSON, XML, or form request body into T.
func Body[T any](ctx *Context) (T, error) {
	var value T
	if err := ctx.Ctx.BodyParser(&value); err != nil {
		return value, ErrBadRequest.WithCode("invalid_body").WithError(err)
	}
	if err := validatorFor(ctx).Struct(value); err != nil {
		return value, NewAppError("Validation failed", http.StatusUnprocessableEntity).
			WithCode("validation_failed").WithError(err)
	}
	return value, nil
}

// QueryParams parses and validates query parameters into T using Fiber's
// schema tags. It is useful for typed filters, pagination, and search options.
func QueryParams[T any](ctx *Context) (T, error) {
	var value T
	if err := ctx.Ctx.QueryParser(&value); err != nil {
		return value, ErrBadRequest.WithCode("invalid_query").WithError(err)
	}
	if err := validatorFor(ctx).Struct(value); err != nil {
		return value, NewAppError("Validation failed", http.StatusUnprocessableEntity).
			WithCode("validation_failed").WithError(err)
	}
	return value, nil
}

// PathParams parses and validates named route parameters into T.
func PathParams[T any](ctx *Context) (T, error) {
	var value T
	if err := ctx.Ctx.ParamsParser(&value); err != nil {
		return value, ErrBadRequest.WithCode("invalid_path").WithError(err)
	}
	if err := validatorFor(ctx).Struct(value); err != nil {
		return value, NewAppError("Validation failed", http.StatusUnprocessableEntity).
			WithCode("validation_failed").WithError(err)
	}
	return value, nil
}

// JSONHandler adapts a typed endpoint to a Flux handler. The optional status
// defaults to 200 OK.
func JSONHandler[Response any](endpoint Endpoint[Response], status ...int) HandlerFunc {
	return func(ctx *Context) error {
		response, err := endpoint(ctx)
		if err != nil {
			return err
		}
		ctx.Status(firstStatus(status, http.StatusOK))
		return ctx.JSON(response)
	}
}

// JSONBodyHandler binds and validates Request before invoking endpoint. The
// optional success status defaults to 200 OK.
func JSONBodyHandler[Request, Response any](endpoint BodyEndpoint[Request, Response], status ...int) HandlerFunc {
	return func(ctx *Context) error {
		request, err := Body[Request](ctx)
		if err != nil {
			return err
		}
		response, err := endpoint(ctx, request)
		if err != nil {
			return err
		}
		ctx.Status(firstStatus(status, http.StatusOK))
		return ctx.JSON(response)
	}
}

func firstStatus(status []int, fallback int) int {
	if len(status) > 0 && status[0] > 0 {
		return status[0]
	}
	return fallback
}

func validatorFor(ctx *Context) interface{ Struct(interface{}) error } {
	if ctx != nil && ctx.app != nil && ctx.app.validator != nil {
		return ctx.app.validator
	}
	return validate
}

// RequestContext exposes the standard library context carried by the request,
// allowing handlers to call database drivers, SDKs, and other Go libraries.
func (c *Context) RequestContext() context.Context {
	return c.Ctx.UserContext()
}

// SetRequestContext replaces the standard library context carried by the
// request. This is commonly used by tracing and cancellation middleware.
func (c *Context) SetRequestContext(ctx context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}
	c.Ctx.SetUserContext(ctx)
}

// RouteOption enriches an explicitly registered route with documentation.
type RouteOption func(*Route)

func RouteName(name string) RouteOption {
	return func(route *Route) { route.Name = name }
}

func RouteDescription(description string) RouteOption {
	return func(route *Route) { route.Description = description }
}

func RouteRequest(model interface{}) RouteOption {
	return func(route *Route) { route.RequestBody = model }
}

func RouteResponse(model interface{}) RouteOption {
	return func(route *Route) { route.Response = model }
}

// Handle registers an explicit route without reflection. Controller-based
// routing remains available for backwards compatibility.
func (app *Application) Handle(method, path string, handler HandlerFunc, options ...RouteOption) *Route {
	method = strings.ToUpper(strings.TrimSpace(method))
	route := &Route{Method: method, Path: path, Handler: handler}
	for _, option := range options {
		if option != nil {
			option(route)
		}
	}
	if route.Name == "" {
		route.Name = fmt.Sprintf("%s %s", method, path)
	}

	app.server.Add(method, path, func(c *fiber.Ctx) error {
		return handler(NewContext(c, app))
	})
	app.routes.AddFromRoute(route, route.Name)
	return route
}

func (app *Application) GET(path string, handler HandlerFunc, options ...RouteOption) *Route {
	return app.Handle(http.MethodGet, path, handler, options...)
}

func (app *Application) POST(path string, handler HandlerFunc, options ...RouteOption) *Route {
	return app.Handle(http.MethodPost, path, handler, options...)
}

func (app *Application) PUT(path string, handler HandlerFunc, options ...RouteOption) *Route {
	return app.Handle(http.MethodPut, path, handler, options...)
}

func (app *Application) PATCH(path string, handler HandlerFunc, options ...RouteOption) *Route {
	return app.Handle(http.MethodPatch, path, handler, options...)
}

func (app *Application) DELETE(path string, handler HandlerFunc, options ...RouteOption) *Route {
	return app.Handle(http.MethodDelete, path, handler, options...)
}

func (app *Application) HEAD(path string, handler HandlerFunc, options ...RouteOption) *Route {
	return app.Handle(http.MethodHead, path, handler, options...)
}

func (app *Application) OPTIONS(path string, handler HandlerFunc, options ...RouteOption) *Route {
	return app.Handle(http.MethodOptions, path, handler, options...)
}
