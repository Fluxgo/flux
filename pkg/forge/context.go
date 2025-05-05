package flux

import (
	"encoding/xml"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/go-playground/validator/v10"
)


var validate = validator.New()


type Context struct {
	*fiber.Ctx
	app *Application
}

// H is a shorthand for map[string]interface{}
type H map[string]interface{}


type ValidationErrors map[string]string


func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return ""
	}

	var errMsgs []string
	for _, msg := range ve {
		errMsgs = append(errMsgs, msg)
	}
	return strings.Join(errMsgs, "; ")
}


func NewContext(c *fiber.Ctx, app *Application) *Context {
	return &Context{
		Ctx: c,
		app: app,
	}
}


func (c *Context) App() *Application {
	return c.app
}


func (c *Context) JSON(data interface{}) error {
	return c.Ctx.JSON(data)
}


func (c *Context) XML(data interface{}) error {
	c.Ctx.Set("Content-Type", "application/xml")
	xmlData, err := xml.Marshal(data)
	if err != nil {
		return err
	}
	return c.Ctx.Send(xmlData)
}


func (c *Context) HTML(html string) error {
	c.Ctx.Set("Content-Type", "text/html")
	return c.Ctx.SendString(html)
}


func (c *Context) Text(text string) error {
	return c.Ctx.SendString(text)
}


func (c *Context) Negotiate(data interface{}) error {
	accept := c.Ctx.Get("Accept")

	switch {
	case accept == "application/xml" || accept == "text/xml":
		return c.XML(data)
	case accept == "text/plain":
		
		if str, ok := data.(string); ok {
			return c.Text(str)
		}
		return c.JSON(data) 
	default:
		return c.JSON(data) 
	}
}


func (c *Context) Bind(v interface{}) error {
	
	if err := c.Ctx.BodyParser(v); err != nil {
		return err
	}

	
	if err := validate.Struct(v); err != nil {
		return err
	}

	return nil
}


func (c *Context) Validate(v interface{}) error {
	return validate.Struct(v)
}


func (c *Context) ValidateWithDetails(i interface{}) ValidationErrors {
	if err := c.Validate(i); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors := make(ValidationErrors)
			for _, e := range validationErrors {
				fieldName := e.Field()
				
				if len(fieldName) > 0 && fieldName[0] >= 'A' && fieldName[0] <= 'Z' {
					fieldName = string(fieldName[0]+32) + fieldName[1:]
				}

				switch e.Tag() {
				case "required":
					errors[fieldName] = fmt.Sprintf("The %s field is required", fieldName)
				case "email":
					errors[fieldName] = fmt.Sprintf("The %s must be a valid email address", fieldName)
				case "min":
					errors[fieldName] = fmt.Sprintf("The %s must be at least %s characters", fieldName, e.Param())
				case "max":
					errors[fieldName] = fmt.Sprintf("The %s must not be greater than %s characters", fieldName, e.Param())
				case "url":
					errors[fieldName] = fmt.Sprintf("The %s must be a valid URL", fieldName)
				default:
					errors[fieldName] = fmt.Sprintf("The %s field is invalid (failed %s validation)", fieldName, e.Tag())
				}
			}
			return errors
		}
		return ValidationErrors{"_error": err.Error()}
	}
	return nil
}


func (c *Context) RespondWithValidationErrors(errors ValidationErrors) error {
	return c.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{
		"error":   true,
		"message": "Validation failed",
		"details": errors,
	})
}


func (c *Context) BindAndValidate(v interface{}) error {
	if err := c.Bind(v); err != nil {
		return NewAppError("Invalid request body", 400).WithError(err)
	}
	if err := c.Validate(v); err != nil {
		return NewAppError("Validation failed", 422).WithError(err)
	}
	return nil
}


func (c *Context) Param(name string) string {
	return c.Ctx.Params(name, "")
}


func (c *Context) Query(key string) string {
	return c.Ctx.Query(key)
}


func (c *Context) QueryDefault(key, defaultValue string) string {
	value := c.Ctx.Query(key)
	if value == "" {
		return defaultValue
	}
	return value
}


func (c *Context) Header(key string) string {
	return c.Ctx.Get(key)
}


func (c *Context) SetHeader(key, value string) {
	c.Ctx.Set(key, value)
}


func (c *Context) Status(code int) *Context {
	c.Ctx.Status(code)
	return c
}


func (c *Context) Redirect(url string, status ...int) error {
	return c.Ctx.Redirect(url, status...)
}


func (c *Context) SendFile(file string) error {
	return c.Ctx.SendFile(file)
}


func (c *Context) FormFile(key string) (*multipart.FileHeader, error) {
	return c.Ctx.FormFile(key)
}


func (c *Context) SaveFile(file *multipart.FileHeader, path string) error {
	return c.Ctx.SaveFile(file, path)
}


func (c *Context) FormValue(key string) string {
	return c.Ctx.FormValue(key)
}


func (c *Context) Error(err error) error {
	if appErr, ok := err.(*AppError); ok {
		code := 500
		if appErr.Code != "" {
			code = appErr.StatusCode
		}
		return c.Status(code).JSON(appErr)
	}
	return c.Status(500).JSON(NewAppError("Internal Server Error", 500).WithError(err))
}


func (c *Context) Success(data interface{}) error {
	return c.JSON(H{
		"success": true,
		"data":    data,
	})
}


func (c *Context) Stream(contentType string, reader func(w io.Writer) error) error {
	c.Ctx.Set("Content-Type", contentType)
	return reader(c.Ctx.Response().BodyWriter())
}


