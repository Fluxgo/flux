package flux

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode"

	"github.com/Fluxgo/flux/pkg/flux/auth"
	"github.com/Fluxgo/flux/pkg/flux/logger"
	"github.com/Fluxgo/flux/pkg/flux/mailer"
	"github.com/Fluxgo/flux/pkg/flux/plugin"
	"github.com/Fluxgo/flux/pkg/flux/queue"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	fiblogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"gorm.io/gorm"
)

type Application struct {
	config      *Config
	server      *fiber.App
	validator   *validator.Validate
	database    *Database
	auth        *auth.Auth
	mailer      *mailer.Mailer
	queue       *queue.Queue
	plugins     *plugin.Manager
	logger      *logger.Logger
	routes      *RouteManager
	mu          sync.RWMutex
	controllers []interface{}
	startTime   time.Time
}

type Config struct {
	Name        string
	Version     string
	Description string
	Server      ServerConfig
	Database    DatabaseConfig
	Auth        auth.Config
	Mailer      mailer.Config
	Queue       queue.Config
	CORS        CORSConfig
	LogLevel    string
}

type ServerConfig struct {
	Host     string
	Port     int
	BasePath string
}

type CORSConfig struct {
	AllowOrigins     string `yaml:"allow_origins" json:"allow_origins"`
	AllowMethods     string `yaml:"allow_methods" json:"allow_methods"`
	AllowHeaders     string `yaml:"allow_headers" json:"allow_headers"`
	AllowCredentials bool   `yaml:"allow_credentials" json:"allow_credentials"`
	ExposeHeaders    string `yaml:"expose_headers" json:"expose_headers"`
	MaxAge           int    `yaml:"max_age" json:"max_age"`
}

func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS,PATCH",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Requested-With",
		AllowCredentials: false, 
		ExposeHeaders:    "",
		MaxAge:           86400,
	}
}

type RateLimitConfig struct {
	Enabled            bool          `yaml:"enabled" json:"enabled"`
	Max                int           `yaml:"max" json:"max"`
	Duration           time.Duration `yaml:"duration" json:"duration"`
	KeyGenerator       func(*fiber.Ctx) string
	Storage            fiber.Storage
	SkipFailedRequests bool     `yaml:"skip_failed" json:"skip_failed"`
	SkipPaths          []string `yaml:"skip_paths" json:"skip_paths"`
	LimitReached       func(*fiber.Ctx) error
}

func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Enabled:            true,
		Max:                100,
		Duration:           time.Duration(1) * time.Minute,
		SkipFailedRequests: false,
		SkipPaths:          []string{},
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   true,
				"message": "Rate limit exceeded",
			})
		},
	}
}

func New(config *Config) (*Application, error) {
	fiberConfig := fiber.Config{
		AppName:             config.Name,
		ServerHeader:        "flux", 
		ErrorHandler:        defaultErrorHandler,
		DisableStartupMessage: true, 
	}

	app := &Application{
		config:    config,
		server:    fiber.New(fiberConfig),
		validator: validator.New(),
		startTime: time.Now(),
	}

	// Initialize the route manager
	app.routes = NewRouteManager(app)

	
	logLevel := logger.LevelInfo
	if config.LogLevel != "" {
		logLevel = logger.ParseLevel(config.LogLevel)
	}

	log := logger.New(logger.Config{
		Level: logLevel,
	})
	log.Info("Initializing flux application: %s v%s", config.Name, config.Version)
	app.logger = log

	app.server.Use(recover.New())
	app.server.Use(fiblogger.New())

	corsConfig := config.CORS
	if corsConfig.AllowOrigins == "" {
		corsConfig = DefaultCORSConfig()
	}
	app.server.Use(cors.New(cors.Config{
		AllowOrigins:     corsConfig.AllowOrigins,
		AllowMethods:     corsConfig.AllowMethods,
		AllowHeaders:     corsConfig.AllowHeaders,
		AllowCredentials: corsConfig.AllowCredentials,
		ExposeHeaders:    corsConfig.ExposeHeaders,
		MaxAge:           corsConfig.MaxAge,
	}))

	app.server.Use(SecurityHeaders())

	if config.Database.Driver != "" {
		log.Info("Initializing database connection: %s", config.Database.Driver)
		db, err := NewDatabase(&config.Database)
		if err != nil {
			log.Error("Failed to initialize database: %v", err)
			return nil, fmt.Errorf("failed to initialize database: %w", err)
		}
		app.database = db
		log.Info("Database connection established")
	}

	if config.Auth.SecretKey != "" {
		log.Info("Initializing authentication")
		auth, err := auth.New(config.Auth)
		if err != nil {
			log.Error("Failed to initialize auth: %v", err)
			return nil, fmt.Errorf("failed to initialize auth: %w", err)
		}
		app.auth = auth
		log.Info("Authentication initialized")
	}

	if config.Mailer.Host != "" {
		log.Info("Initializing mailer")
		mailer, err := mailer.New(config.Mailer)
		if err != nil {
			log.Error("Failed to initialize mailer: %v", err)
			return nil, fmt.Errorf("failed to initialize mailer: %w", err)
		}
		app.mailer = mailer
		log.Info("Mailer initialized")
	}

	if config.Queue.Host != "" {
		log.Info("Initializing message queue")
		queue, err := queue.New(config.Queue.Host, config.Queue.Password, config.Queue.DB)
		if err != nil {
			log.Error("Failed to initialize queue: %v", err)
			return nil, fmt.Errorf("failed to initialize queue: %w", err)
		}
		app.queue = queue
		log.Info("Message queue initialized")
	}

	log.Info("Loading plugins")
	plugins := plugin.NewManager(app, "plugins")
	if err := plugins.LoadPlugins(); err != nil {
		log.Error("Failed to load plugins: %v", err)
		return nil, fmt.Errorf("failed to load plugins: %w", err)
	}
	app.plugins = plugins
	log.Info("Plugins loaded successfully")

	app.server.Get("/", func(c *fiber.Ctx) error {
		return c.Type("html").SendString(`
			<!DOCTYPE html>
			<html lang="en">
			<head>
				<meta charset="UTF-8">
				<title>Goflux</title>
				<style>
					body { font-family: sans-serif; text-align: center; padding: 50px; background-color:rgb(11, 10, 10); }
					h1 { font-size: 2.5em; color:rgb(209, 219, 231); }
					p { font-size: 1.2em; color: #fff; }
					a { color: #007BFF; text-decoration: none; }
					a:hover { text-decoration: underline; }
				</style>
			</head>
			<body>
				<h1>Welcome to flux</h1>
				<p>Built with passion by <strong>Yemi Ogunrinde</strong></p>
				<p>Version: <strong>` + config.Version + `</strong></p>
			</body>
			</html>
		`)
	})

	log.Info("flux application initialized successfully")
	return app, nil
}

func defaultErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	return c.Status(code).JSON(fiber.Map{
		"error":   true,
		"message": err.Error(),
	})
}

func (app *Application) GetConfig() interface{} {
	return app.config
}

func (app *Application) GetDB() interface{} {
	return app.database.DB
}

func (app *Application) GetAuth() interface{} {
	return app.auth
}

func (app *Application) GetQueue() interface{} {
	return app.queue
}

func (app *Application) GetMailer() interface{} {
	return app.mailer
}

func (app *Application) Validator() *validator.Validate {
	return app.validator
}

func (app *Application) RegisterController(controller interface{}) {
	app.mu.Lock()
	defer app.mu.Unlock()

	if c, ok := controller.(interface{ SetApplication(*Application) }); ok {
		c.SetApplication(app)
	}

	app.controllers = append(app.controllers, controller)

	controllerType := reflect.TypeOf(controller)
	controllerValue := reflect.ValueOf(controller)

	controllerName := controllerType.Elem().Name()
	controllerBaseName := strings.TrimSuffix(controllerName, "Controller")
	basePath := "/" + strings.ToLower(controllerBaseName)

	for i := 0; i < controllerType.NumMethod(); i++ {
		method := controllerType.Method(i)

		if !strings.HasPrefix(method.Name, "Handle") {
			continue
		}

		routeInfo := parseRouteFromMethodName(method.Name, basePath)
		handler := createHandlerFunc(method, controllerValue)

		
		description := descriptionFromMethod(controllerBaseName, method.Name)

		
		app.routes.Add(
			routeInfo.HTTPMethod,
			routeInfo.Path,
			fmt.Sprintf("%s.%s", controllerName, method.Name),
			description,
		)

		switch routeInfo.HTTPMethod {
		case "GET":
			app.server.Get(routeInfo.Path, handler)
		case "POST":
			app.server.Post(routeInfo.Path, handler)
		case "PUT":
			app.server.Put(routeInfo.Path, handler)
		case "DELETE":
			app.server.Delete(routeInfo.Path, handler)
		case "PATCH":
			app.server.Patch(routeInfo.Path, handler)
		case "OPTIONS":
			app.server.Options(routeInfo.Path, handler)
		case "HEAD":
			app.server.Head(routeInfo.Path, handler)
		}
	}

	
	if err := app.GenerateRouteFiles(); err != nil {
		app.logger.Error("Failed to generate route files: %v", err)
	}
}


func descriptionFromMethod(controllerName string, methodName string) string {
	
	actionName := strings.TrimPrefix(methodName, "Handle")

	
	for _, method := range []string{"Get", "Post", "Put", "Delete", "Patch", "Options", "Head"} {
		if strings.HasPrefix(actionName, method) {
			actionName = strings.TrimPrefix(actionName, method)
			break
		}
	}

	
	var description strings.Builder
	for i, r := range actionName {
		if i > 0 && r >= 'A' && r <= 'Z' {
			description.WriteRune(' ')
		}
		description.WriteRune(r)
	}

	
	if description.String() == "Index" {
		return fmt.Sprintf("List all %ss", strings.ToLower(controllerName))
	} else if description.String() == "ById" || description.String() == "By Id" {
		return fmt.Sprintf("Get a specific %s by ID", strings.ToLower(controllerName))
	} else if strings.Contains(strings.ToLower(description.String()), "create") {
		return fmt.Sprintf("Create a new %s", strings.ToLower(controllerName))
	} else if strings.Contains(strings.ToLower(description.String()), "update") {
		return fmt.Sprintf("Update a %s", strings.ToLower(controllerName))
	} else if strings.Contains(strings.ToLower(description.String()), "delete") {
		return fmt.Sprintf("Delete a %s", strings.ToLower(controllerName))
	}

	return description.String()
}


func (app *Application) GenerateRouteFiles() error {
	
	if err := os.MkdirAll("routes", 0755); err != nil {
		return fmt.Errorf("failed to create routes directory: %w", err)
	}

	
	if err := app.routes.GenerateRoutesFile("."); err != nil {
		return fmt.Errorf("failed to generate routes.go: %w", err)
	}

	app.logger.Info("Route files generated successfully")
	return nil
}


func (app *Application) Routes() *RouteManager {
	return app.routes
}

type RouteInfo struct {
	HTTPMethod string
	Path       string
}

func parseRouteFromMethodName(methodName string, basePath string) RouteInfo {

	actionName := strings.TrimPrefix(methodName, "Handle")

	httpMethod := "GET"

	for _, method := range []string{"Get", "Post", "Put", "Delete", "Patch", "Options", "Head"} {
		if strings.HasPrefix(actionName, method) {
			httpMethod = strings.ToUpper(method)
			actionName = strings.TrimPrefix(actionName, method)
			break
		}
	}

	if actionName != "" {

		var path strings.Builder
		for i, r := range actionName {
			if i > 0 && r >= 'A' && r <= 'Z' {
				path.WriteRune('-')
			}
			path.WriteRune(unicode.ToLower(r))
		}

		actionPath := path.String()

		if actionPath == "index" || actionPath == "" {
			return RouteInfo{
				HTTPMethod: httpMethod,
				Path:       basePath,
			}
		}

		if strings.Contains(actionPath, "by-id") {
			return RouteInfo{
				HTTPMethod: httpMethod,
				Path:       fmt.Sprintf("%s/:id", basePath),
			}
		}

		return RouteInfo{
			HTTPMethod: httpMethod,
			Path:       fmt.Sprintf("%s/%s", basePath, actionPath),
		}
	}

	return RouteInfo{
		HTTPMethod: httpMethod,
		Path:       basePath,
	}
}

func createHandlerFunc(method reflect.Method, controllerValue reflect.Value) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := &Context{Ctx: c}
		result := method.Func.Call([]reflect.Value{controllerValue, reflect.ValueOf(ctx)})
		if len(result) > 0 && !result[0].IsNil() {
			if err, ok := result[0].Interface().(error); ok {
				return err
			}
		}
		return nil
	}
}

func (app *Application) Start() error {
	app.logger.Info("flux server started on %s:%d", app.config.Server.Host, app.config.Server.Port)
	app.logger.Info("Version: %s", app.config.Version)
	app.logger.Info("Environment: %s", getEnvironment())

	if app.queue != nil {
		app.queue.Start()
	}

	return app.server.Listen(fmt.Sprintf("%s:%d", app.config.Server.Host, app.config.Server.Port))
}

func (app *Application) Listen(addr string) error {
	addrInfo := strings.Split(addr, ":")
	if len(addrInfo) == 2 {
		app.logger.Info("flux server started on %s:%s", addrInfo[0], addrInfo[1])
	} else {
		app.logger.Info("flux server started on %s", addr)
	}
	app.logger.Info("Version: %s", app.config.Version)
	app.logger.Info("Environment: %s", getEnvironment())

	if app.queue != nil {
		app.queue.Start()
	}

	return app.server.Listen(addr)
}


func getEnvironment() string {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = os.Getenv("ENV")
	}
	if env == "" {
		env = "development"
	}
	return env
}


func (app *Application) Serve() error {
	return app.Start()
}

func (app *Application) Shutdown() error {
	if app.queue != nil {
		app.queue.Stop()
	}

	if app.plugins != nil {
		if err := app.plugins.UnloadPlugins(); err != nil {
			return fmt.Errorf("failed to unload plugins: %w", err)
		}
	}

	if app.database != nil {
		if err := app.database.Close(); err != nil {
			return fmt.Errorf("failed to close database: %w", err)
		}
	}

	return app.server.Shutdown()
}

func (app *Application) DB() *gorm.DB {
	return app.database.DB
}

func (app *Application) Auth() *auth.JWTManager {
	return app.auth.JWTManager
}

func (app *Application) Queue() *queue.Queue {
	return app.queue
}

func (app *Application) Mailer() *mailer.Mailer {
	return app.mailer
}

func (app *Application) Plugins() *plugin.Manager {
	return app.plugins
}

func (app *Application) Logger() *logger.Logger {
	return app.logger
}

func (app *Application) WithLogField(key string, value interface{}) *logger.Logger {
	return app.logger.WithField(key, value)
}

func (a *Application) Group(prefix string) fiber.Router {
	return a.server.Group(prefix)
}

func (a *Application) Use(middleware ...interface{}) {
	a.server.Use(middleware...)
}

func (a *Application) Get() *fiber.App {
	return a.server
}

func (app *Application) Test(req *http.Request) (*http.Response, error) {
	return app.server.Test(req)
}

func (app *Application) AddTracing() {
	app.server.Use(func(c *fiber.Ctx) error {

		traceID := generateTraceID()
		c.Locals("trace_id", traceID)

		c.Set("X-Trace-ID", traceID)

		requestLogger := app.logger.WithField("trace_id", traceID)
		c.Locals("logger", requestLogger)

		requestLogger.Info("Received %s %s from %s", c.Method(), c.Path(), c.IP())

		startTime := time.Now()

		err := c.Next()

		duration := time.Since(startTime)

		if err != nil {
			requestLogger.Error("Request failed: %v (took %v)", err, duration)
		} else {
			requestLogger.Info("Request completed with status %d (took %v)", c.Response().StatusCode(), duration)
		}

		return err
	})
}

func generateTraceID() string {

	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func (app *Application) EnableGracefulShutdown() {

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		app.logger.Info("Shutdown signal received, shutting down gracefully...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if app.queue != nil {

			_ = ctx
			app.logger.Debug("Shutting down job queue...")

			if err := app.queue.Shutdown(); err != nil {
				app.logger.Error("Error shutting down job queue: %v", err)
			}
		}

		if err := app.Shutdown(); err != nil {
			app.logger.Error("Failed to shut down server gracefully: %v", err)
		}

		app.logger.Info("Server shutdown complete")
		os.Exit(0)
	}()
}

func (app *Application) EnableHealthCheck(path string) {
	if path == "" {
		path = "/health"
	}

	app.server.Get(path, func(c *fiber.Ctx) error {
		health := map[string]interface{}{
			"status":      "ok",
			"version":     app.config.Version,
			"name":        app.config.Name,
			"timestamp":   time.Now().Format(time.RFC3339),
			"connections": runtime.NumGoroutine(),
		}

		if app.database != nil && app.database.DB != nil {
			sqlDB, err := app.database.DB.DB()
			if err == nil {
				health["database"] = "ok"
				health["database_connections"] = map[string]interface{}{
					"open":  sqlDB.Stats().OpenConnections,
					"idle":  sqlDB.Stats().Idle,
					"inUse": sqlDB.Stats().InUse,
				}
			} else {
				health["database"] = "error"
				health["database_error"] = err.Error()
			}
		}

		if app.queue != nil {
			if app.queue.IsRunning() {
				health["queue"] = "ok"
			} else {
				health["queue"] = "not running"
			}
		}

		return c.JSON(health)
	})

	app.logger.Info("Health check endpoint enabled at %s", path)
}

func (a *Application) ConfigureMiddleware(options ...interface{}) {
	config := &MiddlewareConfig{
		Compress:        true,
		CORS:            true,
		CORSConfig:      DefaultCORSConfig(),
		Logger:          true,
		Recover:         true,
		RequestID:       true,
		RateLimit:       false,
		RateLimitConfig: DefaultRateLimitConfig(),
		BodyLimit:       "1MB",
		ReadTimeout:     60 * time.Second,
		WriteTimeout:    60 * time.Second,
		IdleTimeout:     120 * time.Second,
	}

	
	for _, option := range options {
		switch opt := option.(type) {
		case MiddlewareOption:
			opt(config)
		case func(*MiddlewareConfig):
			opt(config)
		}
	}

	if config.Recover {
		a.server.Use(recover.New())
	}

	if config.RequestID {
		a.server.Use(requestid.New())
	}

	if config.Compress {
		a.server.Use(compress.New())
	}

	if config.CORS {
		a.server.Use(cors.New(cors.Config{
			AllowOrigins:     config.CORSConfig.AllowOrigins,
			AllowMethods:     config.CORSConfig.AllowMethods,
			AllowHeaders:     config.CORSConfig.AllowHeaders,
			AllowCredentials: config.CORSConfig.AllowCredentials,
			ExposeHeaders:    config.CORSConfig.ExposeHeaders,
			MaxAge:           config.CORSConfig.MaxAge,
		}))
	}

	if config.Logger {
		a.server.Use(fiblogger.New())
	}

	if config.RateLimit {
		rlConfig := config.RateLimitConfig
		a.server.Use(limiter.New(limiter.Config{
			Max:                    rlConfig.Max,
			Expiration:             rlConfig.Duration,
			KeyGenerator:           rlConfig.KeyGenerator,
			LimitReached:           rlConfig.LimitReached,
			SkipFailedRequests:     rlConfig.SkipFailedRequests,
			SkipSuccessfulRequests: false,
			Storage:                rlConfig.Storage,
			Next: func(c *fiber.Ctx) bool {
				// Skip rate limiting for specified paths
				path := c.Path()
				for _, skipPath := range rlConfig.SkipPaths {
					if strings.HasPrefix(path, skipPath) {
						return true
					}
				}
				return false
			},
		}))
	}

	
	a.server.Server().ReadTimeout = config.ReadTimeout
	a.server.Server().WriteTimeout = config.WriteTimeout
	a.server.Server().IdleTimeout = config.IdleTimeout

	
	a.server.Server().MaxRequestBodySize = fiberBodyLimitToInt(config.BodyLimit)
}


func fiberBodyLimitToInt(bodyLimit string) int {
	units := map[string]int{
		"B":  1,
		"KB": 1024,
		"MB": 1024 * 1024,
		"GB": 1024 * 1024 * 1024,
	}

	var value int
	var unit string

	if _, err := fmt.Sscanf(bodyLimit, "%d%s", &value, &unit); err != nil {
		return 4 * 1024 * 1024
	}

	unitValue, found := units[strings.ToUpper(unit)]
	if !found {
		return 4 * 1024 * 1024
	}

	return value * unitValue
}

type MiddlewareConfig struct {
	Compress        bool
	CORS            bool
	CORSConfig      CORSConfig
	Logger          bool
	LoggerConfig    fiber.Config
	Limiter         bool
	LimiterConfig   limiter.Config
	Recover         bool
	RequestID       bool
	RateLimit       bool
	RateLimitConfig RateLimitConfig
	BodyLimit       string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
}

type MiddlewareOption func(*MiddlewareConfig)

func WithCompress() MiddlewareOption {
	return func(c *MiddlewareConfig) {
		c.Compress = true
	}
}

func WithCORS(config CORSConfig) MiddlewareOption {
	return func(c *MiddlewareConfig) {
		c.CORS = true
		c.CORSConfig = config
	}
}

func WithLogger(config ...fiber.Config) MiddlewareOption {
	return func(c *MiddlewareConfig) {
		c.Logger = true
		if len(config) > 0 {
			c.LoggerConfig = config[0]
		}
	}
}

func WithLimiter(config ...limiter.Config) MiddlewareOption {
	return func(c *MiddlewareConfig) {
		c.Limiter = true
		if len(config) > 0 {
			c.LimiterConfig = config[0]
		}
	}
}

func WithRecover() MiddlewareOption {
	return func(c *MiddlewareConfig) {
		c.Recover = true
	}
}

func WithRequestID() MiddlewareOption {
	return func(c *MiddlewareConfig) {
		c.RequestID = true
	}
}

func WithBodyLimit(limit string) MiddlewareOption {
	return func(c *MiddlewareConfig) {
		c.BodyLimit = limit
	}
}

func WithTimeout(read, write time.Duration) MiddlewareOption {
	return func(c *MiddlewareConfig) {
		c.ReadTimeout = read
		c.WriteTimeout = write
	}
}

func WithIdleTimeout(idle time.Duration) MiddlewareOption {
	return func(c *MiddlewareConfig) {
		c.IdleTimeout = idle
	}
}

func SecurityHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {

		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "SAMEORIGIN")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		c.Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; script-src 'self'")

		if os.Getenv("ENVIRONMENT") == "production" {
			c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		c.Set("Permissions-Policy", "accelerometer=(), camera=(), geolocation=(), gyroscope=(), magnetometer=(), microphone=(), payment=(), usb=()")

		return c.Next()
	}
}
