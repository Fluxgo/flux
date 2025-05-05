package middleware

import (
	"fmt"
	"sync"
	"time"

	"github.com/Fluxgo/flux/pkg/flux"
	"github.com/gofiber/fiber/v2"
	"github.com/go-playground/validator/v10"
)

type MiddlewareConfig struct {
	Recover         bool
	Logger          bool
	CORS            bool
	Compression     bool
	RequestID       bool
	BodyLimit       string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	CORSConfig      flux.CORSConfig
	SecurityHeaders bool
}

type MiddlewareOption func(*MiddlewareConfig)

func WithoutRecover() MiddlewareOption {
	return func(c *MiddlewareConfig) {
		c.Recover = false
	}
}

func WithoutLogger() MiddlewareOption {
	return func(c *MiddlewareConfig) {
		c.Logger = false
	}
}

func WithoutCORS() MiddlewareOption {
	return func(c *MiddlewareConfig) {
		c.CORS = false
	}
}

func WithCORSConfig(config flux.CORSConfig) MiddlewareOption {
	return func(c *MiddlewareConfig) {
		c.CORSConfig = config
	}
}

func WithoutCompression() MiddlewareOption {
	return func(c *MiddlewareConfig) {
		c.Compression = false
	}
}

func WithoutRequestID() MiddlewareOption {
	return func(c *MiddlewareConfig) {
		c.RequestID = false
	}
}

func WithBodyLimit(limit string) MiddlewareOption {
	return func(c *MiddlewareConfig) {
		c.BodyLimit = limit
	}
}

func WithTimeouts(read, write time.Duration) MiddlewareOption {
	return func(c *MiddlewareConfig) {
		c.ReadTimeout = read
		c.WriteTimeout = write
	}
}

func WithSecurityHeaders() MiddlewareOption {
	return func(c *MiddlewareConfig) {
		c.SecurityHeaders = true
	}
}

func RequestLogger() flux.MiddlewareFunc {
	return func(next flux.HandlerFunc) flux.HandlerFunc {
		return func(ctx *flux.Context) error {
			start := time.Now()

			method := ctx.Method()
			path := ctx.Path()

			err := next(ctx)

			duration := time.Since(start)

			if err != nil {
				ctx.App().Logger().Error("[%s] %s - %v - %s", method, path, err, duration)
			} else {
				ctx.App().Logger().Info("[%s] %s - %d - %s", method, path, ctx.Response().StatusCode(), duration)
			}

			return err
		}
	}
}

func Recover() flux.MiddlewareFunc {
	return func(next flux.HandlerFunc) flux.HandlerFunc {
		return func(ctx *flux.Context) (err error) {
			defer func() {
				if r := recover(); r != nil {
					var ok bool
					if err, ok = r.(error); !ok {
						err = flux.NewAppError("Internal Server Error", 500).
							WithDetail("panic", r)
					}

					ctx.App().Logger().Error("Recovered from panic: %v", r)
				}
			}()

			return next(ctx)
		}
	}
}

func RequireAuth() flux.MiddlewareFunc {
	return func(next flux.HandlerFunc) flux.HandlerFunc {
		return func(ctx *flux.Context) error {

			token := ctx.Get("Authorization")
			if token == "" {
				return flux.ErrUnauthorized
			}

			auth := ctx.App().Auth()
			if auth == nil {
				ctx.App().Logger().Error("Auth is not initialized")
				return flux.ErrInternalError.WithDetail("message", "Authentication system not initialized")
			}

			claims, err := auth.ValidateToken(token)
			if err != nil {
				return flux.ErrUnauthorized.WithError(err)
			}

			ctx.Locals("user_id", claims["sub"])
			ctx.Locals("claims", claims)

			return next(ctx)
		}
	}
}

// CORS headers
func CORS(options flux.CORSConfig) flux.MiddlewareFunc {
	return func(next flux.HandlerFunc) flux.HandlerFunc {
		return func(ctx *flux.Context) error {

			ctx.Set("Access-Control-Allow-Origin", options.AllowOrigins)
			ctx.Set("Access-Control-Allow-Methods", options.AllowMethods)
			ctx.Set("Access-Control-Allow-Headers", options.AllowHeaders)

			if ctx.Method() == "OPTIONS" {
				return ctx.SendStatus(204)
			}

			return next(ctx)
		}
	}
}

func SecurityHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {

		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'")
		c.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), interest-cohort=()")

		if c.Protocol() == "https" {
			c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		return c.Next()
	}
}

func CacheControl(maxAge string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("Cache-Control", "public, max-age="+maxAge)
		return c.Next()
	}
}

func NoCache() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
		c.Set("Pragma", "no-cache")
		c.Set("Expires", "0")
		return c.Next()
	}
}

type Middleware struct {
	SecurityHeaders fiber.Handler
	CacheControl   func(maxAge string) fiber.Handler
	NoCache        fiber.Handler
}

func New() *Middleware {
	return &Middleware{
		SecurityHeaders: SecurityHeaders(),
		CacheControl:   CacheControl,
		NoCache:        NoCache(),
	}
}

type RateLimiterConfig struct {

	Max int

	Window time.Duration
	KeyFunc func(*flux.Context) string
	SkipFunc func(*flux.Context) bool
}

func RateLimit(config RateLimiterConfig) flux.MiddlewareFunc {
	if config.Max <= 0 {
		config.Max = 100
	}
	if config.Window <= 0 {
		config.Window = time.Minute
	}
	if config.KeyFunc == nil {
		config.KeyFunc = func(ctx *flux.Context) string {
			return ctx.IP()
		}
	}

	type windowEntry struct {
		timestamp time.Time
		count     int
	}

	type client struct {
		windows      []windowEntry
		lastAccessed time.Time
		mu           sync.Mutex
	}

	clients := struct {
		data map[string]*client
		mu   sync.RWMutex
	}{
		data: make(map[string]*client),
	}

	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			now := time.Now()

			clients.mu.Lock()
			for key, c := range clients.data {

				if now.Sub(c.lastAccessed) > 30*time.Minute {
					delete(clients.data, key)
				}
			}
			clients.mu.Unlock()
		}
	}()

	return func(next flux.HandlerFunc) flux.HandlerFunc {
		return func(ctx *flux.Context) error {

			if config.SkipFunc != nil && config.SkipFunc(ctx) {
				return next(ctx)
			}

			key := config.KeyFunc(ctx)
			now := time.Now()

			clients.mu.RLock()
			c, exists := clients.data[key]
			clients.mu.RUnlock()

			if !exists {
				clients.mu.Lock()

				if c, exists = clients.data[key]; !exists {
					c = &client{
						windows: make([]windowEntry, 0, 10),
					}
					clients.data[key] = c
				}
				clients.mu.Unlock()
			}

			c.mu.Lock()
			defer c.mu.Unlock()

			c.lastAccessed = now

			cutoff := now.Add(-config.Window)
			windowStart := 0

			for i, window := range c.windows {
				if window.timestamp.After(cutoff) {
					windowStart = i
					break
				}
			}

			if windowStart > 0 {
				c.windows = c.windows[windowStart:]
			}

			count := 0
			for _, window := range c.windows {
				count += window.count
			}

			if count >= config.Max {
				ctx.Set("Retry-After", fmt.Sprintf("%d", int(config.Window.Seconds())))
				ctx.Set("X-RateLimit-Limit", fmt.Sprintf("%d", config.Max))
				ctx.Set("X-RateLimit-Remaining", "0")
				ctx.Set("X-RateLimit-Reset", fmt.Sprintf("%d", now.Add(config.Window).Unix()))

				return flux.NewAppError("Rate limit exceeded", 429)
			}

			if len(c.windows) > 0 && now.Sub(c.windows[len(c.windows)-1].timestamp) < time.Second {

				c.windows[len(c.windows)-1].count++
			} else {

				c.windows = append(c.windows, windowEntry{
					timestamp: now,
					count:     1,
				})
			}

			remaining := config.Max - count - 1
			ctx.Set("X-RateLimit-Limit", fmt.Sprintf("%d", config.Max))
			ctx.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			ctx.Set("X-RateLimit-Reset", fmt.Sprintf("%d", now.Add(config.Window).Unix()))

			return next(ctx)
		}
	}
}

func Timeout(duration time.Duration) flux.MiddlewareFunc {
	return func(next flux.HandlerFunc) flux.HandlerFunc {
		return func(ctx *flux.Context) error {

			done := make(chan error)

			go func() {
				done <- next(ctx)
			}()

			select {
			case err := <-done:
				return err
			case <-time.After(duration):
				return flux.NewAppError("Request timeout", 408)
			}
		}
	}
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
	Tag     string `json:"tag"`
}

type ValidationResult struct {
	Errors []ValidationError `json:"errors"`
}

func (vr *ValidationResult) ValidationFailed() bool {
	return len(vr.Errors) > 0
}

func Validator() flux.MiddlewareFunc {
	return func(next flux.HandlerFunc) flux.HandlerFunc {
		return func(ctx *flux.Context) error {
			if ctx.Body() == nil {
				return next(ctx)
			}

			validate := ctx.App().Validator()
			if validate == nil {
				return next(ctx)
			}

			if err := validate.Struct(ctx.Body()); err != nil {
				
				validationErrors, ok := err.(validator.ValidationErrors)
				if !ok {
					
					return flux.NewAppError("Validation failed", 422).WithError(err)
				}
				
				result := &ValidationResult{
					Errors: make([]ValidationError, len(validationErrors)),
				}

				for i, fieldError := range validationErrors {
					field := fieldError.Field()
					tag := fieldError.Tag()
					
					message := fmt.Sprintf("Validation failed on '%s'", tag)
					switch tag {
					case "required":
						message = "This field is required"
					case "min":
						message = fmt.Sprintf("Should be at least %s", fieldError.Param())
					case "max":
						message = fmt.Sprintf("Should be at most %s", fieldError.Param())
					case "email":
						message = "Invalid email address"
					case "url":
						message = "Invalid URL format"
					case "uuid":
						message = "Invalid UUID format"
					case "json":
						message = "Invalid JSON format"
					case "numeric":
						message = "Must be a numeric value"
					case "alpha":
						message = "Must contain only alphabetic characters"
					case "alphanum":
						message = "Must contain only alphanumeric characters"
					case "datetime":
						message = fmt.Sprintf("Invalid datetime format, expected: %s", fieldError.Param())
					}
					
					result.Errors[i] = ValidationError{
						Field:   field,
						Tag:     tag,
						Value:   fmt.Sprintf("%v", fieldError.Value()),
						Message: message,
					}
				}
				
				return flux.NewAppError("Validation Error", 422).
					WithDetail("validation", result)
			}
			
			return next(ctx)
		}
	}
}
