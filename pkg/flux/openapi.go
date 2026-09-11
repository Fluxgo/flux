package flux

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

type OpenAPISpec struct {
	OpenAPI    string              `json:"openapi"`
	Info       OpenAPIInfo         `json:"info"`
	Paths      map[string]PathItem `json:"paths"`
	Components OpenAPIComponents   `json:"components"`
}

type OpenAPIInfo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

type PathItem struct {
	Get     *Operation `json:"get,omitempty"`
	Post    *Operation `json:"post,omitempty"`
	Put     *Operation `json:"put,omitempty"`
	Delete  *Operation `json:"delete,omitempty"`
	Patch   *Operation `json:"patch,omitempty"`
	Head    *Operation `json:"head,omitempty"`
	Options *Operation `json:"options,omitempty"`
}

type Operation struct {
	Summary     string                `json:"summary"`
	Description string                `json:"description"`
	OperationID string                `json:"operationId,omitempty"`
	Parameters  []*Parameter          `json:"parameters,omitempty"`
	RequestBody *RequestBody          `json:"requestBody,omitempty"`
	Responses   map[string]*Response  `json:"responses"`
	Tags        []string              `json:"tags,omitempty"`
	Security    []map[string][]string `json:"security,omitempty"`
}

type Parameter struct {
	Name        string  `json:"name"`
	In          string  `json:"in"`
	Description string  `json:"description"`
	Required    bool    `json:"required"`
	Schema      *Schema `json:"schema"`
}

type RequestBody struct {
	Description string                     `json:"description"`
	Required    bool                       `json:"required"`
	Content     map[string]MediaTypeObject `json:"content"`
}

type Response struct {
	Description string                     `json:"description"`
	Content     map[string]MediaTypeObject `json:"content,omitempty"`
}

type MediaTypeObject struct {
	Schema *Schema `json:"schema"`
}

type Schema struct {
	Type                 string             `json:"type,omitempty"`
	Format               string             `json:"format,omitempty"`
	Description          string             `json:"description,omitempty"`
	Properties           map[string]*Schema `json:"properties,omitempty"`
	Items                *Schema            `json:"items,omitempty"`
	Required             []string           `json:"required,omitempty"`
	AdditionalProperties *Schema            `json:"additionalProperties,omitempty"`
}

type OpenAPIComponents struct {
	Schemas         map[string]*Schema        `json:"schemas"`
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes"`
}

type SecurityScheme struct {
	Type         string `json:"type"`
	Description  string `json:"description,omitempty"`
	Name         string `json:"name,omitempty"`
	In           string `json:"in,omitempty"`
	Scheme       string `json:"scheme,omitempty"`
	BearerFormat string `json:"bearerFormat,omitempty"`
}

func (app *Application) GenerateOpenAPI() (*OpenAPISpec, error) {
	spec := &OpenAPISpec{
		OpenAPI: "3.0.0",
		Info: OpenAPIInfo{
			Title:       app.config.Name,
			Description: app.config.Description,
			Version:     app.config.Version,
		},
		Paths: make(map[string]PathItem),
		Components: OpenAPIComponents{
			Schemas:         make(map[string]*Schema),
			SecuritySchemes: make(map[string]SecurityScheme),
		},
	}

	spec.Components.SecuritySchemes["bearerAuth"] = SecurityScheme{
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: "JWT",
		Description:  "JWT token for authentication",
	}

	for _, route := range app.routes.routes {
		path, parameters := openAPIPath(route.Path)
		operation := &Operation{
			Summary:     route.Description,
			Description: route.Description,
			OperationID: route.Handler,
			Parameters:  parameters,
			Responses: map[string]*Response{
				"200": {
					Description: "Successful operation",
					Content: map[string]MediaTypeObject{
						"application/json": {Schema: schemaForValue(route.Response)},
					},
				},
			},
		}
		if operation.Summary == "" {
			operation.Summary = fmt.Sprintf("%s %s", route.Method, route.Path)
		}
		if route.RequestBody != nil {
			operation.RequestBody = &RequestBody{
				Required: true,
				Content: map[string]MediaTypeObject{
					"application/json": {Schema: schemaForValue(route.RequestBody)},
				},
			}
		}

		pathItem := spec.Paths[path]
		setOperation(&pathItem, route.Method, operation)
		spec.Paths[path] = pathItem
	}

	return spec, nil
}

func schemaForValue(value interface{}) *Schema {
	if value == nil {
		return &Schema{Type: "object"}
	}
	return generateSchemaFromType(reflect.TypeOf(value))
}

func openAPIPath(path string) (string, []*Parameter) {
	parts := strings.Split(path, "/")
	parameters := make([]*Parameter, 0)
	for index, part := range parts {
		if !strings.HasPrefix(part, ":") {
			continue
		}
		name := strings.TrimPrefix(part, ":")
		parts[index] = "{" + name + "}"
		parameters = append(parameters, &Parameter{
			Name:     name,
			In:       "path",
			Required: true,
			Schema:   &Schema{Type: "string"},
		})
	}
	return strings.Join(parts, "/"), parameters
}

func setOperation(pathItem *PathItem, method string, operation *Operation) {
	switch strings.ToUpper(method) {
	case "GET":
		pathItem.Get = operation
	case "POST":
		pathItem.Post = operation
	case "PUT":
		pathItem.Put = operation
	case "DELETE":
		pathItem.Delete = operation
	case "PATCH":
		pathItem.Patch = operation
	case "HEAD":
		pathItem.Head = operation
	case "OPTIONS":
		pathItem.Options = operation
	}
}

func GenerateSwaggerUI(spec *OpenAPISpec) (string, error) {
	specJSON, err := json.Marshal(spec)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>flux API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@3/swagger-ui.css">
    <script src="https://unpkg.com/swagger-ui-dist@3/swagger-ui-bundle.js"></script>
</head>
<body>
    <div id="swagger-ui"></div>
    <script>
        window.onload = function() {
            SwaggerUIBundle({
                spec: %s,
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIBundle.SwaggerUIStandalonePreset
                ],
            });
        }
    </script>
</body>
</html>`, string(specJSON)), nil
}

func generateSchemaFromType(t reflect.Type) *Schema {
	if t == nil {
		return nil
	}

	switch t.Kind() {
	case reflect.String:
		return &Schema{Type: "string"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &Schema{Type: "integer"}
	case reflect.Float32, reflect.Float64:
		return &Schema{Type: "number"}
	case reflect.Bool:
		return &Schema{Type: "boolean"}
	case reflect.Struct:
		schema := &Schema{
			Type:       "object",
			Properties: make(map[string]*Schema),
			Required:   make([]string, 0),
		}

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			jsonTag := field.Tag.Get("json")
			if jsonTag == "-" {
				continue
			}

			name := strings.Split(jsonTag, ",")[0]
			if name == "" {
				name = field.Name
			}

			fieldSchema := generateSchemaFromType(field.Type)
			if fieldSchema != nil {
				schema.Properties[name] = fieldSchema
				if !strings.Contains(jsonTag, "omitempty") {
					schema.Required = append(schema.Required, name)
				}
			}
		}
		return schema

	case reflect.Slice, reflect.Array:
		return &Schema{
			Type:  "array",
			Items: generateSchemaFromType(t.Elem()),
		}

	case reflect.Map:
		return &Schema{
			Type:                 "object",
			AdditionalProperties: generateSchemaFromType(t.Elem()),
		}

	case reflect.Ptr:
		return generateSchemaFromType(t.Elem())
	}

	return nil
}
