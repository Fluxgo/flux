package flux

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"runtime"

	"github.com/Fluxgo/flux/pkg/flux/logger"
	"github.com/gofiber/fiber/v2"
)


type Microservice struct {
	Name        string
	Version     string
	Description string
	app         *Application
	logger      *logger.Logger
	config      *MicroserviceConfig
	routes      []Route
	isSetup     bool
}

type MicroserviceConfig struct {
	Name          string        `yaml:"name" json:"name"`
	Port          int           `yaml:"port" json:"port"`
	Host          string        `yaml:"host" json:"host"`
	Description   string        `yaml:"description" json:"description"`
	ReadTimeout   time.Duration `yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout  time.Duration `yaml:"write_timeout" json:"write_timeout"`
	BodyLimit     string        `yaml:"body_limit" json:"body_limit"`
	CORS          CORSConfig    `yaml:"cors" json:"cors"`
	LogLevel      string        `yaml:"log_level" json:"log_level"`
	EnableTracing bool          `yaml:"enable_tracing" json:"enable_tracing"`
	Metrics       bool          `yaml:"metrics" json:"metrics"`
	HealthCheck   bool          `yaml:"health_check" json:"health_check"`
	WithDB        bool          `yaml:"with_db" json:"with_db"`
	WithCache     bool          `yaml:"with_cache" json:"with_cache"`
	WithQueue     bool          `yaml:"with_queue" json:"with_queue"`
	WithAuth      bool          `yaml:"with_auth" json:"with_auth"`
}

func DefaultMicroserviceConfig() *MicroserviceConfig {
	return &MicroserviceConfig{
		Host:          "0.0.0.0",
		Port:          3000,
		ReadTimeout:   30 * time.Second,
		WriteTimeout:  30 * time.Second,
		BodyLimit:     "1MB",
		CORS:          DefaultCORSConfig(),
		LogLevel:      "info",
		EnableTracing: true,
		Metrics:       true,
		HealthCheck:   true,
	}
}

func NewMicroservice(name, version, description string) *Microservice {
	return &Microservice{
		Name:        name,
		Version:     version,
		Description: description,
		config:      DefaultMicroserviceConfig(),
		routes:      make([]Route, 0),
	}
}

func (ms *Microservice) WithConfig(config *MicroserviceConfig) *Microservice {
	ms.config = config
	return ms
}

func (ms *Microservice) AddRoute(method, path, description string, handler HandlerFunc) *Microservice {
	ms.routes = append(ms.routes, Route{
		Method:      method,
		Path:        path,
		Description: description,
		Handler:     handler,
	})
	return ms
}

func (ms *Microservice) GET(path, description string, handler HandlerFunc) *Microservice {
	return ms.AddRoute("GET", path, description, handler)
}

func (ms *Microservice) POST(path, description string, handler HandlerFunc) *Microservice {
	return ms.AddRoute("POST", path, description, handler)
}

func (ms *Microservice) PUT(path, description string, handler HandlerFunc) *Microservice {
	return ms.AddRoute("PUT", path, description, handler)
}

func (ms *Microservice) PATCH(path, description string, handler HandlerFunc) *Microservice {
	return ms.AddRoute("PATCH", path, description, handler)
}

func (ms *Microservice) DELETE(path, description string, handler HandlerFunc) *Microservice {
	return ms.AddRoute("DELETE", path, description, handler)
}

func (ms *Microservice) Setup() error {
	if ms.isSetup {
		return nil
	}

	config := &Config{
		Name:        ms.Name,
		Version:     ms.Version,
		Description: ms.Description,
		Server: ServerConfig{
			Host: ms.config.Host,
			Port: ms.config.Port,
		},
		LogLevel: ms.config.LogLevel,
		CORS:     ms.config.CORS,
	}

	app, err := New(config)
	if err != nil {
		return fmt.Errorf("failed to create application: %w", err)
	}
	ms.app = app
	ms.logger = app.Logger()

	app.ConfigureMiddleware(
		func(c *MiddlewareConfig) {
			c.BodyLimit = ms.config.BodyLimit
			c.ReadTimeout = ms.config.ReadTimeout
			c.WriteTimeout = ms.config.WriteTimeout
			c.CORSConfig = ms.config.CORS
		},
	)

	if ms.config.EnableTracing {
		app.AddTracing()
	}

	if ms.config.HealthCheck {
		app.EnableHealthCheck("/health")
	}

	if ms.config.Metrics {
		app.server.Get("/metrics", func(c *fiber.Ctx) error {
			metrics := map[string]interface{}{
				"uptime":      time.Since(app.startTime),
				"connections": ms.GetOpenConnections(),
				"routes":      len(app.server.Stack()),
			}
			return c.JSON(metrics)
		})
	}

	for _, route := range ms.routes {
		app.server.Add(route.Method, route.Path, func(c *fiber.Ctx) error {
			ctx := NewContext(c, app)
			return route.Handler(ctx)
		})
	}

	ms.isSetup = true
	return nil
}

// Start the microservice
func (ms *Microservice) Start() error {
	if !ms.isSetup {
		if err := ms.Setup(); err != nil {
			return err
		}
	}

	ms.EnableGracefulShutdown()

	addr := fmt.Sprintf("%s:%d", ms.config.Host, ms.config.Port)
	ms.logger.Info("Starting %s v%s on %s", ms.Name, ms.Version, addr)
	return ms.app.server.Listen(addr)
}

func (ms *Microservice) StartWithHotReload() error {
	if !ms.isSetup {
		if err := ms.Setup(); err != nil {
			return err
		}
	}

	ms.EnableGracefulShutdown()

	reloader, err := NewMicroserviceHotReloader(ms.app, ms.Name, filepath.Join("cmd", ms.Name, "main.go"))
	if err != nil {
		return fmt.Errorf("failed to create hot reloader: %w", err)
	}

	if err := reloader.Start(); err != nil {
		return fmt.Errorf("failed to start hot reloader: %w", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	ms.logger.Info("Shutting down hot reloader...")
	if err := reloader.Stop(); err != nil {
		ms.logger.Error("Error stopping hot reloader: %v", err)
	}

	return nil
}

func (ms *Microservice) Stop() error {
	if ms.app == nil {
		return nil
	}
	return ms.app.Shutdown()
}

func (ms *Microservice) EnableGracefulShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		ms.logger.Info("Shutdown signal received, shutting down gracefully...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		_ = ctx

		if err := ms.Stop(); err != nil {
			ms.logger.Error("Failed to shutdown server gracefully: %v", err)
		} else {
			ms.logger.Info("Server shutdown complete")
		}
		os.Exit(0)
	}()
}

func (ms *Microservice) GetOpenConnections() int {
	if ms.app == nil || ms.app.server == nil {
		return 0
	}

	return runtime.NumGoroutine() - 10
}

func (ms *Microservice) App() *Application {
	return ms.app
}

func (ms *Microservice) Logger() *logger.Logger {
	return ms.logger
}

func fluxHandlerToFiberHandler(handler HandlerFunc, app *Application) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := NewContext(c, app)
		return handler(ctx)
	}
}

func CreateNewProject(projectName string) {
	fmt.Println("Creating new flux project:", projectName)
	os.MkdirAll(projectName+"/controllers", os.ModePerm)
	os.MkdirAll(projectName+"/models", os.ModePerm)
	os.MkdirAll(projectName+"/routes", os.ModePerm)
	os.MkdirAll(projectName+"/services", os.ModePerm)
	fmt.Println("Project scaffold created.")
}

func GenerateController(name string, service string) {
	var path string
	if service != "" {
		path = fmt.Sprintf("services/%s/controllers", service)
	} else {
		path = "controllers"
	}
	os.MkdirAll(path, os.ModePerm)
	controllerFile := fmt.Sprintf("%s/%sController.go", path, name)
	content := fmt.Sprintf(`package controllers

import "fmt"

func %sController() {
	fmt.Println("%s controller logic here")
}`, name, name)
	os.WriteFile(controllerFile, []byte(content), 0644)
	fmt.Println("Controller created at:", controllerFile)
}

func GenerateModel(name string, service string) {
	var path string
	if service != "" {
		path = fmt.Sprintf("services/%s/models", service)
	} else {
		path = "models"
	}
	os.MkdirAll(path, os.ModePerm)
	modelFile := fmt.Sprintf("%s/%s.go", path, name)
	content := fmt.Sprintf(`package models

// %s represents the model structure
type %s struct {
	ID   int
	Name string
}`, name, name)
	os.WriteFile(modelFile, []byte(content), 0644)
	fmt.Println("Model created at:", modelFile)
}

func CreateMicroservice(serviceName string) {
	base := fmt.Sprintf("services/%s", serviceName)
	folders := []string{
		base + "/controllers",
		base + "/models",
		base + "/routes",
		base + "/config",
	}
	for _, folder := range folders {
		os.MkdirAll(folder, os.ModePerm)
	}
	mainFile := base + "/main.go"
	mainContent := fmt.Sprintf(`package main

import "fmt"

func main() {
	fmt.Println("%s microservice started")
}`, serviceName)
	os.WriteFile(mainFile, []byte(mainContent), 0644)
	fmt.Println("Microservice scaffold created at:", base)
}

func CreateMicroserviceProject(config *MicroserviceConfig) error {
	if config == nil {
		config = DefaultMicroserviceConfig()
	}

	name := config.Name
	if name == "" {
		return fmt.Errorf("microservice name cannot be empty")
	}

	if err := os.MkdirAll(name, 0755); err != nil {
		return fmt.Errorf("failed to create microservice directory: %w", err)
	}

	dirs := []string{
		filepath.Join(name, "api"),
		filepath.Join(name, "api", "handlers"),
		filepath.Join(name, "api", "middleware"),
		filepath.Join(name, "internal", "models"),
		filepath.Join(name, "internal", "services"),
		filepath.Join(name, "internal", "repositories"),
		filepath.Join(name, "pkg", "logger"),
		filepath.Join(name, "config"),
		filepath.Join(name, "cmd", name),
	}

	if config.WithDB {
		dirs = append(dirs, filepath.Join(name, "internal", "database"))
		dirs = append(dirs, filepath.Join(name, "migrations"))
	}

	if config.WithCache {
		dirs = append(dirs, filepath.Join(name, "internal", "cache"))
	}

	if config.WithQueue {
		dirs = append(dirs, filepath.Join(name, "internal", "queue"))
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	mainContent := generateMicroserviceMainFile(config)
	if err := os.WriteFile(filepath.Join(name, "cmd", name, "main.go"), []byte(mainContent), 0644); err != nil {
		return fmt.Errorf("failed to create main.go: %w", err)
	}

	configContent := generateMicroserviceConfigFile(config)
	if err := os.WriteFile(filepath.Join(name, "config", "config.yaml"), []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to create config.yaml: %w", err)
	}

	dockerfileContent := generateDockerfile(config)
	if err := os.WriteFile(filepath.Join(name, "Dockerfile"), []byte(dockerfileContent), 0644); err != nil {
		return fmt.Errorf("failed to create Dockerfile: %w", err)
	}

	dockerComposeContent := generateDockerCompose(config)
	if err := os.WriteFile(filepath.Join(name, "docker-compose.yml"), []byte(dockerComposeContent), 0644); err != nil {
		return fmt.Errorf("failed to create docker-compose.yml: %w", err)
	}

	modContent := fmt.Sprintf(`module github.com/%s

go 1.20

require (
	github.com/Fluxgo/flux v0.1.5
	github.com/gofiber/fiber/v2 v2.52.6
)
`, name)
	if err := os.WriteFile(filepath.Join(name, "go.mod"), []byte(modContent), 0644); err != nil {
		return fmt.Errorf("failed to create go.mod: %w", err)
	}

	handlerContent := generateSampleHandler(config)
	if err := os.WriteFile(filepath.Join(name, "api", "handlers", "health.go"), []byte(handlerContent), 0644); err != nil {
		return fmt.Errorf("failed to create sample handler: %w", err)
	}

	readmeContent := generateMicroserviceReadme(config)
	if err := os.WriteFile(filepath.Join(name, "README.md"), []byte(readmeContent), 0644); err != nil {
		return fmt.Errorf("failed to create README.md: %w", err)
	}

	fmt.Printf("Created new flux microservice: %s\n", name)
	return nil
}

func generateMicroserviceMainFile(config *MicroserviceConfig) string {
	return fmt.Sprintf(`package main

import (
	"fmt"
	"log"

	"github.com/Fluxgo/flux/pkg/flux"
	"github.com/gofiber/fiber/v2"
)

func main() {
	// flux application
	app, err := flux.New(&flux.Config{
		Name:        "%s",
		Version:     "1.0.0",
		Description: "%s",
		Server: flux.ServerConfig{
			Host:     "%s",
			Port:     %d,
			BasePath: "/",
		},
	})
	if err != nil {
		log.Fatalf("Failed to create application: %%v", err)
	}

	// Sample routes
	app.Get().Get("/", func(c *fiber.Ctx) error {
		ctx := flux.NewContext(c, app)
		return ctx.JSON(map[string]interface{}{
			"message": "Welcome to %s microservice",
			"version": "1.0.0",
		})
	})

	// health check endpoint
	app.Get().Get("/health", func(c *fiber.Ctx) error {
		ctx := flux.NewContext(c, app)
		return ctx.JSON(map[string]interface{}{
			"status": "ok",
			"service": "%s",
		})
	})

	// Start the server
	fmt.Printf("Starting %%s microservice on %%s:%%d\n", "%s", "%s", %d)
	if err := app.Start(); err != nil {
		log.Fatalf("Failed to start server: %%v", err)
	}
}
`,
		config.Name,
		config.Description,
		config.Host,
		config.Port,
		config.Name,
		config.Name,
		config.Name,
		config.Host,
		config.Port)
}

func generateConfigOptions(config *MicroserviceConfig) string {
	options := ""

	if config.WithDB {
		options += `Database: flux.DatabaseConfig{
			Driver: "sqlite",
			Name:   "flux.db",
			// Uncomment these for production use
			// Driver:   "postgres",  
			// Host:     "db",
			// Port:     5432,
			// Username: "postgres",
			// Password: "postgres",
			// Name:     "flux",
		},`
	}

	return options
}

func generateMicroserviceConfigFile(config *MicroserviceConfig) string {
	return fmt.Sprintf(`# %s Microservice Configuration

# Service Settings
service:
  name: "%s"
  version: "1.0.0"
  description: "%s"
  environment: "development"
  debug: true

# Server Configuration
server:
  host: "0.0.0.0"
  port: %d
  base_path: "/api"
  read_timeout: 10s
  write_timeout: 10s
  idle_timeout: 120s

%s
`,
		config.Name,
		config.Name,
		config.Description,
		config.Port,
		generateAdditionalConfig(config))
}

func generateAdditionalConfig(config *MicroserviceConfig) string {
	var additionalConfig string

	if config.WithDB {
		additionalConfig += `# Database Configuration
database:
  driver: "sqlite"
  name: "flux.db"
  # For production:
  # driver: "postgres"
  # host: "db"
  # port: 5432
  # username: "postgres"
  # password: "postgres"
  # name: "flux"
  max_open_conns: 20
  max_idle_conns: 5
  conn_max_life: 300s

`
	}

	if config.WithCache {
		additionalConfig += `# Cache Configuration
cache:
  driver: "redis"
  host: "cache"
  port: 6379
  prefix: "flux:"
  ttl: 3600s

`
	}

	if config.WithQueue {
		additionalConfig += `# Queue Configuration
queue:
  driver: "redis"
  host: "queue"
  port: 6379
  db: 1

`
	}

	if config.WithAuth {
		additionalConfig += `# Authentication Config
auth:
  jwt:
    secret: "change-this-to-your-personal-secret-in-prod"
    expiration: 86400s # 24 hours
    refresh_expiration: 604800s 

`
	}

	return additionalConfig
}

func generateDockerfile(config *MicroserviceConfig) string {
	return `# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o service ./cmd/` + config.Name + `

# Final stage
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/service .
COPY --from=builder /app/config ./config

RUN chmod +x service

EXPOSE ` + fmt.Sprintf("%d", config.Port) + `

ENTRYPOINT ["./service"]
`
}

func generateDockerCompose(config *MicroserviceConfig) string {
	services := `version: '3.8'

services:
  api:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "` + fmt.Sprintf("%d:%d", config.Port, config.Port) + `"
    restart: unless-stopped
`

	if config.WithDB {
		services += `    depends_on:
      - db
    environment:
      - DB_HOST=db
      - DB_PORT=5432
      - DB_USER=postgres
      - DB_PASSWORD=postgres
      - DB_NAME=flux

  db:
    image: postgres:14-alpine
    volumes:
      - postgres-data:/var/lib/postgresql/data
    environment:
      - POSTGRES_PASSWORD=postgres
      - POSTGRES_USER=postgres
      - POSTGRES_DB=flux
    ports:
      - "5432:5432"
`
	}

	if config.WithCache {
		services += `
  cache:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
`
	}

	if config.WithQueue {
		services += `  queue:
    image: redis:alpine
    ports:
      - "6379:6379"
    command: redis-server --appendonly yes
    volumes:
      - queue-data:/data
    restart: unless-stopped

`
	}

	volumes := `
volumes:`

	if config.WithDB {
		volumes += `
  postgres-data:`
	}

	if config.WithCache {
		volumes += `
  redis-data:`
	}

	if !config.WithDB && !config.WithCache {
		volumes = ""
	}

	return services + volumes
}

func generateSampleHandler(config *MicroserviceConfig) string {
	return fmt.Sprintf(`package handlers

import (
	"github.com/Fluxgo/flux/pkg/flux"
)

//  health check requests
func Health(ctx *flux.Context) error {
	return ctx.JSON(map[string]interface{}{
		"status":  "ok",
		"service": "%s",
		"version": "1.0.0",
	})
}
`, config.Name)
}

func generateMicroserviceReadme(config *MicroserviceConfig) string {
	return fmt.Sprintf(`# %s Microservice

## Description

%s

## Getting Started

### Prerequisites

- Go 1.23 or higher
- Docker and Docker Compose (optional)

### Installation

1. Clone the repository
2. Build and run the service:

`+"```bash"+`
go mod download
go run cmd/%s/main.go
`+"```"+`

### Using Docker

Build and run with Docker:

`+"```bash"+`
docker build -t %s .
docker run -p %d:%d %s
`+"```"+`

Or use Docker Compose:

`+"```bash"+`
docker-compose up
`+"```"+`

## API Endpoints

- GET /api/health - Health check endpoint
- Additional endpoints will be documented here

## Configuration

Configuration is loaded from config/config.yaml. See this file for available options.

## Development

### Project Structure

- cmd/%s/ - Main application entry point
- api/ - API handlers and middleware
- internal/ - Internal packages and business logic
- config/ - Configuration files

### Testing

`+"```bash"+`
go test ./...
`+"```"+`
`,
		config.Name,
		config.Description,
		config.Name,
		config.Name,
		config.Port, config.Port,
		config.Name,
		config.Name,
	)
}
