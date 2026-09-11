package flux

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type createWidgetRequest struct {
	Name string `json:"name" validate:"required,min=3"`
}

type widgetResponse struct {
	Name string `json:"name"`
}

type widgetParams struct {
	ID int `params:"id" validate:"required"`
}

type widgetQuery struct {
	Page int `query:"page" validate:"min=1"`
}

func newTestApplication(t *testing.T) *Application {
	t.Helper()
	app, err := New(&Config{Name: "test", Version: "test"})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	t.Cleanup(func() { _ = app.Shutdown() })
	return app
}

func TestJSONBodyHandler(t *testing.T) {
	app := newTestApplication(t)
	app.POST("/widgets", JSONBodyHandler(
		func(_ *Context, request createWidgetRequest) (widgetResponse, error) {
			return widgetResponse{Name: request.Name}, nil
		},
		http.StatusCreated,
	))

	request := httptest.NewRequest(http.MethodPost, "/widgets", bytes.NewBufferString(`{"name":"flux"}`))
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusCreated)
	}
	var body widgetResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Name != "flux" {
		t.Fatalf("name = %q, want flux", body.Name)
	}
}

func TestJSONBodyHandlerValidationError(t *testing.T) {
	app := newTestApplication(t)
	app.POST("/widgets", JSONBodyHandler(
		func(_ *Context, request createWidgetRequest) (widgetResponse, error) {
			return widgetResponse{Name: request.Name}, nil
		},
	))

	request := httptest.NewRequest(http.MethodPost, "/widgets", bytes.NewBufferString(`{"name":""}`))
	request.Header.Set("Content-Type", "application/json")
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusUnprocessableEntity {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("status = %d, want %d; body = %s", response.StatusCode, http.StatusUnprocessableEntity, body)
	}
	var body map[string]interface{}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body["code"] != "validation_failed" {
		t.Fatalf("code = %v, want validation_failed", body["code"])
	}
}

func TestRequestContextInterop(t *testing.T) {
	type contextKey struct{}
	app := newTestApplication(t)
	app.GET("/context", JSONHandler(func(ctx *Context) (widgetResponse, error) {
		ctx.SetRequestContext(context.WithValue(context.Background(), contextKey{}, "standard-library"))
		name, _ := ctx.RequestContext().Value(contextKey{}).(string)
		return widgetResponse{Name: name}, nil
	}))

	request := httptest.NewRequest(http.MethodGet, "/context", nil)
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer response.Body.Close()

	var body widgetResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Name != "standard-library" {
		t.Fatalf("name = %q, want standard-library", body.Name)
	}
}

func TestTypedPathAndQueryParams(t *testing.T) {
	app := newTestApplication(t)
	app.GET("/widgets/:id", JSONHandler(func(ctx *Context) (widgetResponse, error) {
		params, err := PathParams[widgetParams](ctx)
		if err != nil {
			return widgetResponse{}, err
		}
		query, err := QueryParams[widgetQuery](ctx)
		if err != nil {
			return widgetResponse{}, err
		}
		return widgetResponse{Name: fmt.Sprintf("%d:%d", params.ID, query.Page)}, nil
	}))

	response, err := app.Test(httptest.NewRequest(http.MethodGet, "/widgets/42?page=3", nil))
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer response.Body.Close()
	var body widgetResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Name != "42:3" {
		t.Fatalf("name = %q, want 42:3", body.Name)
	}
}

func TestUnknownErrorsAreNotExposed(t *testing.T) {
	app := newTestApplication(t)
	app.GET("/failure", JSONHandler(func(_ *Context) (widgetResponse, error) {
		return widgetResponse{}, errors.New("database password leaked")
	}))

	response, err := app.Test(httptest.NewRequest(http.MethodGet, "/failure", nil))
	if err != nil {
		t.Fatalf("Test() error = %v", err)
	}
	defer response.Body.Close()
	var body map[string]interface{}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.StatusCode != http.StatusInternalServerError || body["message"] != "Internal Server Error" {
		t.Fatalf("unsafe error response: status=%d body=%v", response.StatusCode, body)
	}
}

func TestExplicitRouteGeneratesOpenAPIContract(t *testing.T) {
	app := newTestApplication(t)
	app.POST("/widgets/:id", JSONBodyHandler(
		func(_ *Context, request createWidgetRequest) (widgetResponse, error) {
			return widgetResponse{Name: request.Name}, nil
		},
	),
		RouteName("widgets.update"),
		RouteDescription("Update a widget"),
		RouteRequest(createWidgetRequest{}),
		RouteResponse(widgetResponse{}),
	)

	spec, err := app.GenerateOpenAPI()
	if err != nil {
		t.Fatalf("GenerateOpenAPI() error = %v", err)
	}
	operation := spec.Paths["/widgets/{id}"].Post
	if operation == nil {
		t.Fatal("POST /widgets/{id} missing from OpenAPI specification")
	}
	if operation.OperationID != "widgets.update" {
		t.Fatalf("operation ID = %q, want widgets.update", operation.OperationID)
	}
	if len(operation.Parameters) != 1 || operation.Parameters[0].Name != "id" {
		t.Fatalf("path parameters = %+v, want id", operation.Parameters)
	}
	requestSchema := operation.RequestBody.Content["application/json"].Schema
	if requestSchema.Properties["name"].Type != "string" {
		t.Fatalf("request schema = %+v, want string name", requestSchema)
	}
	responseSchema := operation.Responses["200"].Content["application/json"].Schema
	if responseSchema.Properties["name"].Type != "string" {
		t.Fatalf("response schema = %+v, want string name", responseSchema)
	}
}

func TestAppErrorIsImmutableAndUnwraps(t *testing.T) {
	cause := errors.New("database unavailable")
	first := ErrBadRequest.WithDetail("field", "name")
	second := ErrBadRequest.WithDetail("field", "email").WithError(cause)

	if got := first.Details["field"]; got != "name" {
		t.Fatalf("first detail = %v, want name", got)
	}
	if got := second.Details["field"]; got != "email" {
		t.Fatalf("second detail = %v, want email", got)
	}
	if !errors.Is(second, cause) {
		t.Fatal("AppError does not unwrap its cause")
	}
}

func TestNewAcceptsNilConfig(t *testing.T) {
	app, err := New(nil)
	if err != nil {
		t.Fatalf("New(nil) error = %v", err)
	}
	defer app.Shutdown()
	if app.config.Name != "flux" || app.config.Server.Port != 3000 {
		t.Fatalf("unexpected defaults: %+v", app.config)
	}
}
