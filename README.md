# flux

flux is a modern, full-stack web framework for Go — designed to combine developer happiness, performance, and structure.

## Features

- **Type-Safe Request Handling**: Struct-based binding & validation (like FastAPI)
- **Auto-generated Swagger Docs**: OpenAPI docs from your handlers
- **Modular MVC Architecture**: Controllers, Services, Models (like NestJS)
- **Microservices Support**: First-class tools for building distributed systems
- **CLI Scaffolding**: `flux make:controller`, `make:model`, `make:microservice`, etc.
- **Go-Level Performance**: Fiber/Gin speed under the hood
- **Built-in Auth**: CLI generator for login/register flow
- **Extensible Plugins**: File uploads, RBAC, jobs, and more coming
- **Full-Stack Ready**: With template support or HTMX/SPA integration

## Implementation Details

flux is built on top of the [Fiber](https://github.com/gofiber/fiber) web framework for several reasons:

1. **Performance**: Fiber is built on top of [fasthttp](https://github.com/valyala/fasthttp), which is significantly faster than Go's standard net/http package
2. **Express-like API**: Familiar API design for developers coming from Node.js/Express
3. **Middleware ecosystem**: Rich middleware ecosystem that we can leverage
4. **Low memory footprint**: Optimized for minimal memory usage and high concurrency

While we could have built directly on Go's standard library, we chose Fiber to provide better performance and developer experience. The flux framework abstracts away most Fiber-specific details, allowing you to work with a clean, consistent API.

## flux vs Fiber: Why Choose flux?

While flux is built on top of Fiber for its performance benefits, it offers several significant advantages:

1. **Convention over Configuration Architecture**
   - Opinionated MVC structure with clear separation of concerns
   - Controller-based routing with automatic route generation
   - Standardized project layout for sustainable development

2. **Powerful Middleware System**
   - Express.js-like middleware with next() handler functionality
   - Controller-level middleware for route-specific handling
   - Middleware groups for sharing behavior across controllers
   - Built-in middleware for common tasks (logging, auth, rate limiting)

3. **Full-Stack Development Framework**
   - Complete solution beyond just HTTP handling
   - Database integration with GORM (ORM)
   - Authentication system with JWT
   - Background job processing with queues
   - Mailing capabilities
   - Extensible plugin architecture

4. **Dual Architecture Support**
   - Monolithic applications with MVC pattern
   - Microservices with modern containerized structure
   - Shared tools and patterns across both architectures

5. **Developer-Friendly Tooling**
   - CLI for scaffolding new projects, controllers, models, and microservices
   - Hot reloading for rapid development
   - Automatic OpenAPI documentation generation

## Installation

```bash
go install github.com/Fluxgo/flux/cmd/flux@latest
```

## Quick Start

### Monolithic Application

Create a new flux project:

```bash
flux new myapp
cd myapp
```

Generate a controller:

```bash
flux make:controller User
```

Generate a model:

```bash
flux make:model Post --migration
```

Start the development server:

```bash
flux serve
```

### Microservice Application

Create a new microservice:

```bash
flux make:microservice user-service
cd user-service
```

Optional flags for microservice creation:
- `--with-db`: Include database integration
- `--with-auth`: Include authentication support
- `--with-cache`: Include Redis cache integration
- `--with-queue`: Include task queue support

Sample with options:

```bash
flux make:microservice payment-service --with-db --with-auth
```

## Architecture Options

flux supports two primary architectural patterns:

### 1. Monolithic Architecture

Best for:
- Smaller teams and projects
- Rapid prototyping
- Applications with simpler domains

Structure:
```
myapp/
├── app/
│   ├── controllers/      # Route handlers
│   ├── services/         # Business logic
│   ├── models/           # DB schemas
├── config/               # App/env config
├── database/             # Migrations/seeders
├── routes/               # Route groups
├── templates/            # Optional views
├── fluxflux.yaml            # Project config
└── main.go
```

### 2. Microservice Architecture

Best for:
- Larger teams and projects
- Complex domain boundaries
- Scalable, distributed systems

Structure:
```
service-name/
├── api/              # API layer
│   ├── handlers/     # HTTP request handlers
│   └── middleware/   # HTTP middleware
├── cmd/              # Application entry points
│   └── service-name/ # Main service executable
├── config/           # Configuration files
├── internal/         # Private application code
│   ├── models/       # Data models
│   ├── services/     # Business logic
│   └── repositories/ # Data access layer
└── pkg/              # Public libraries
    └── logger/       # Logging utilities
```

## Sample Controller

```go
package controllers

import (
	"github.com/Fluxgo/flux/pkg/flux"
)

// UserController handles user-related requests
type UserController struct {
	flux.Controller
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// HandlePostLogin handles user login
// @route POST /login
// @desc Authenticate a user
// @body LoginRequest
// @response 200 { message: string }
func (c *UserController) HandlePostLogin(ctx *flux.Context) error {
	var req LoginRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.Status(400).JSON(flux.H{
			"error": "Invalid request body",
		})
	}

	// Implement actual login logic
	return ctx.JSON(flux.H{
		"message": "Welcome",
	})
}
```

## Routing and Controllers

flux uses a convention-based approach to routing inspired by Ruby on Rails and Laravel. Controllers and their methods automatically map to HTTP routes.

### Controller Naming Convention

Controllers should be named with the `Controller` suffix:

```go
// UserController -> maps to "/user" route prefix
type UserController struct {
	flux.Controller
}

// AuthController -> maps to "/auth" route prefix
type AuthController struct {
	flux.Controller
}
```

### Method Naming Convention

Controller methods should follow this pattern:

```
Handle[HTTP Method][Action]
```

For example:

```go
// HandleGetUsers maps to GET /user
func (c *UserController) HandleGetUsers(ctx *flux.Context) error {
    // ...continue with implementation inside this wrapper
}

// HandlePostUser maps to POST /user
func (c *UserController) HandlePostUser(ctx *flux.Context) error {
    // ...continue with implementation inside this wrapper
}

// HandlePutUserById maps to PUT /user/:id
func (c *UserController) HandlePutUserById(ctx *flux.Context) error {
    id := ctx.Param("id")
    // ...continue with implementation inside this wrapper
}

// HandleDeleteUser maps to DELETE /user
func (c *UserController) HandleDeleteUser(ctx *flux.Context) error {
    // ...continue with implementation inside this wrapper
}
```

### Special Path Rules

1. `ById` in the method name automatically maps to the path pattern with `:id` parameter
2. For nested resources, use camel case: `HandleGetUserPosts` maps to GET /user/posts

### Registering Controllers

To register a controller with your flux application:

```go
app.RegisterController(&UserController{})
app.RegisterController(&AuthController{})
```

### Complete Example: Auth Controller

Here's an example of a complete authentication controller:

```go
package controllers

import (
	"github.com/Fluxgo/flux/pkg/flux"
)

// AuthController handles authentication-related requests
type AuthController struct {
	flux.Controller
}

// RegisterRequest represents the registration request body
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name" validate:"required"`
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// HandlePostRegister handles user registration
// Maps to: POST /auth/register
func (c *AuthController) HandlePostRegister(ctx *flux.Context) error {
	var req RegisterRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.Status(400).JSON(flux.H{
			"error": "Invalid request body",
		})
	}
	
	if err := ctx.Validate(&req); err != nil {
		return ctx.Status(400).JSON(flux.H{
			"error": "Validation failed",
			"details": err.Error(),
		})
	}
	
	// Add user registration logic here...
	
	return ctx.Status(201).JSON(flux.H{
		"message": "User registered successfully",
	})
}

// HandlePostLogin handles user login
// Maps to: POST /auth/login
func (c *AuthController) HandlePostLogin(ctx *flux.Context) error {
	var req LoginRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.Status(400).JSON(flux.H{
			"error": "Invalid request body",
		})
	}
	
	if err := ctx.Validate(&req); err != nil {
		return ctx.Status(400).JSON(flux.H{
			"error": "Validation failed", 
			"details": err.Error(),
		})
	}
	
	// user authentication logic here...
	
	token := "sample-jwt-token" 
	
	return ctx.JSON(flux.H{
		"token": token,
		"message": "Login successful",
	})
}
```

### Starting the Server

You can impmentment start flux application server using any of these methods:

```go
// Option 1: Using Start (recommended)
if err := app.Start(); err != nil {
    log.Fatalf("Failed to start server: %v", err)
}

// Option 2: Using Listen (Fiber-style)
if err := app.Listen(":3000"); err != nil {
    log.Fatalf("Failed to start server: %v", err)
}

// Option 3: Using Serve (net/http-style)
if err := app.Serve(); err != nil {
    log.Fatalf("Failed to start server: %v", err)
}
```

## Middleware System

flux provides a powerful middleware system inspired by Express.js. Middleware functions have access to the request/response cycle and can:

- Execute any code
- Make changes to the request and response objects
- End the request-response cycle
- Call the next middleware in the stack

### Defining Middleware

```go
// Simple middleware function
func LoggingMiddleware(next flux.HandlerFunc) flux.HandlerFunc {
    return func(ctx *flux.Context) error {
        start := time.Now()
        
        // Call the next handler in the chain
        err := next(ctx)
        
        // Log after the request is processed
        duration := time.Since(start)
        ctx.App().Logger().Info("Request processed in %s", duration)
        
        return err
    }
}
```

### Using Middleware

Middleware can be applied at multiple levels:

#### 1. Controller-level middleware

```go
// Apply middleware to a controller
userController := &UserController{}
userController.Use(middleware.RequestLogger(), middleware.RequireAuth())

// Register the controller
app.RegisterController(userController)
```

#### 2. Controller group middleware

```go
// Create a group with shared middleware
api := (&flux.Controller{}).Group("/api")
api.Use(middleware.Recover(), middleware.RequestLogger())

// Add controllers to the group
api.Add(&UserController{})
api.Add(&ProductController{})

// Register all controllers in the group
api.Register(app)
```

#### 3. Global middleware

```go
// Apply middleware to all routes
app.Use(middleware.Recover(), middleware.RequestLogger())
```

### Built-in Middleware

flux comes with several built-in middleware functions:

- `middleware.RequestLogger()` - Logs request information and timing
- `middleware.Recover()` - Catches panics and converts them to errors
- `middleware.RequireAuth()` - Handles authentication checks
- `middleware.CORS(options)` - Configures CORS headers
- `middleware.RateLimit(limit)` - Limits request rates
- `middleware.Timeout(duration)` - Sets a timeout for request handling

## CORS Configuration

flux includes built-in CORS support. Configure it in your application:

```go
app, err := flux.New(&flux.Config{
    
    CORS: flux.CORSConfig{
        AllowOrigins:     "http://localhost:3000,https://ffg.com",
        AllowMethods:     "GET,POST,PUT,DELETE",
        AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
        AllowCredentials: true,
        MaxAge:           86400, 
    },
})
```

If not specified, flux uses a permissive default CORS configuration that allows all origins.

## Database Integration

flux uses GORM for database operations. Here's an example model:

```go
package models

// User represents a user entity
type User struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Email     string `json:"email" gorm:"uniqueIndex"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt string `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the table name for the model
func (User) TableName() string {
	return "users"
}
```

## Configuration

Configure your application in `flux.yaml`:

```yaml
app:
  name: "myapp"
  version: "0.1.0"
  description: "A flux application"

server:
  port: 3000
  host: localhost
  base_path: /

database:
  driver: sqlite
  name: flux.db
```

## CLI Commands

- `flux new [name]`: Create a new monolithic flux project
- `flux make:controller [name]`: Generate a new controller
- `flux make:model [name]`: Generate a new model
- `flux make:microservice [name]`: Generate a new microservice project
- `flux serve`: Start the development server with hot reload
- `flux db:migrate`: Run database migrations
- `flux doc:generate`: Generate OpenAPI documentation

## Microservices with flux

flux provides first-class support for building microservices with a modern, production-ready structure:

### Features

- **Containerization**: Docker and docker-compose configurations included
- **API-First Design**: Structured API handlers and middleware
- **Configuration Management**: Environment-based configuration
- **Health Checks**: Built-in health check endpoint
- **Modern Project Layout**: Following Go best practices for project structure

### Development Workflow

1. Create a new microservice: `flux make:microservice my-service`
2. Add your business logic in the internal/services directory
3. Expose your API endpoints in the api/handlers directory
4. Run locally: `go run cmd/my-service/main.go`
5. Deploy with Docker: `docker-compose up --build`

## Contributing

Contributions are welcome! Raise any noticed issue and Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.