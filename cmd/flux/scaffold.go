package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Fluxgo/flux/pkg/flux"
)

type ProjectTemplate struct {
	Name        string
	Description string
	Files       map[string]string
}

type ControllerTemplate struct {
	Name        string
	Description string
	Methods     []string
}

type ModelTemplate struct {
	Name        string
	Description string
	Fields      []string
}

func createNewProject(name string) error {

	if err := os.MkdirAll(name, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	dirs := []string{
		filepath.Join(name, "app", "controllers"),
		filepath.Join(name, "app", "models"),
		filepath.Join(name, "app", "services"),
		filepath.Join(name, "config"),
		filepath.Join(name, "database", "migrations"),
		filepath.Join(name, "database", "seeders"),
		filepath.Join(name, "routes"),
		filepath.Join(name, "templates"),
		filepath.Join(name, "storage", "logs"),
		filepath.Join(name, "storage", "uploads"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	mainContent := `package main

import (
	"fmt"
	"log"

	"github.com/Fluxgo/flux/pkg/flux"
	"` + name + `/routes" // Import the routes package
)

func main() {
	//New flux application
	app, err := flux.New(&flux.Config{
		Name:        "` + name + `",
		Version:     "1.0.0",
		Description: "A flux application",
		Server: flux.ServerConfig{
			Host:     "localhost",
			Port:     3000,
			BasePath: "/",
		},
		Database: flux.DatabaseConfig{
			Driver: "sqlite",  
			Name:   "flux.db",
			// Uncomment these for other database types
			// Host:     "localhost",
			// Port:     3306,  
			// Username: "flux_user",
			// Password: "flux_password",
		},
	})
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	// Register all routes
	routes.RegisterAllRoutes(app)
	
	// Register individual controllers if needed
	// app.RegisterController(&UserController{})

	// Start the server
	fmt.Printf("Server starting on http://localhost:3000\n")
	if err := app.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
`

	if err := os.WriteFile(filepath.Join(name, "main.go"), []byte(mainContent), 0644); err != nil {
		return fmt.Errorf("failed to create main.go: %w", err)
	}

	configContent := `# Configuration

# Application Settings
app:
  name: "` + name + `"
  version: "1.0.0"
  description: "A powerful web application built with flux Framework"
  environment: "development" 
  debug: true
  timezone: "UTC"
  secret_key: "change-this-to-your-own-secure-secret-key"
  log_level: "info" 

# Server Configuration
server:
  host: "localhost"
  port: 3000
  base_path: "/"
  read_timeout: 10s
  write_timeout: 10s
  idle_timeout: 120s

# Database Configuration
database:
  # Main database connection
  default:
    driver: "sqlite" 
    name: "flux.db"
    # Uncomment these below for other database types
    # host: "localhost"
    # port: 3306  
    # username: "flux_user"
    # password: "flux_password"
    # ssl_mode: "disable" 
    # charset: "utf8mb4"
    # timezone: "Local"
    max_open_conns: 100
    max_idle_conns: 10
    conn_max_life: 3600s 
    slow_threshold: 200ms
    log_level: "info" 
    debug: false


auth:
  jwt:
    secret_key: "change-this-to-your-own-personal-jwt-secret-key"
    expiration: 86400 
    refresh_expiration: 604800 
    signing_method: "HS256" 


view:
  engine: "go-template" 
  directory: "templates"
  extension: ".gohtml"
  cache: true
`

	if err := os.WriteFile(filepath.Join(name, "config", "flux.yaml"), []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to create flux.yaml: %w", err)
	}

	modContent := `module ` + name + `

go 1.20

require (
	github.com/Fluxgo/flux v0.1.5
)
`

	if err := os.WriteFile(filepath.Join(name, "go.mod"), []byte(modContent), 0644); err != nil {
		return fmt.Errorf("failed to create go.mod: %w", err)
	}

	readmeContent := `# ` + name + `

A web application built with flux Framework.

## Getting Started

1. Run the development server:
   
   ` + "```" + `bash
   flux serve
   ` + "```" + `

2. Open [http://localhost:3000](http://localhost:3000) in your browser.

## Database Configuration

This project uses SQLite by default, which requires no additional setup. To use other databases:

1. Edit the database configuration in ` + "`config/flux.yaml`" + `
2. Choose from: sqlite, mysql, postgres, sqlserver
3. Provide connection details as required

## Creating Controllers and Models

Generate new controllers:

` + "```" + `bash
flux generate controller User
` + "```" + `

Generate new models:

` + "```" + `bash
flux generate model User
` + "```" + `

## Learn More

To learn more about flux Framework, check out the documentation at flux Framework Documentation(https://github.com/Fluxgo/flux).
`

	if err := os.WriteFile(filepath.Join(name, "README.md"), []byte(readmeContent), 0644); err != nil {
		return fmt.Errorf("failed to create README.md: %w", err)
	}

	fmt.Printf("Created new flux project: %s\n", name)
	return nil
}

func generateController(name string) error {
	name = strings.ToUpper(name[:1]) + name[1:]
	if !strings.HasSuffix(name, "Controller") {
		name += "Controller"
	}

	if err := os.MkdirAll(filepath.Join("app", "controllers"), 0755); err != nil {
		return fmt.Errorf("failed to create controllers directory: %w", err)
	}

	if err := os.MkdirAll(filepath.Join("app", "models"), 0755); err != nil {
		return fmt.Errorf("failed to create models directory: %w", err)
	}

	if err := os.MkdirAll(filepath.Join("routes"), 0755); err != nil {
		return fmt.Errorf("failed to create routes directory: %w", err)
	}

	controllerContent := `package controllers

import (
	"github.com/Fluxgo/flux/pkg/flux"
)

// ` + name + ` handles requests related to ` + strings.TrimSuffix(name, "Controller") + `
type ` + name + ` struct {
	flux.Controller
}

// HandleGet` + strings.TrimSuffix(name, "Controller") + `s handles getting all ` + strings.TrimSuffix(name, "Controller") + `s
// Route: GET /` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + `
// Description: Get all ` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + `s
// Response: 200 - []` + strings.TrimSuffix(name, "Controller") + `
func (c *` + name + `) HandleGet` + strings.TrimSuffix(name, "Controller") + `s(ctx *flux.Context) error {
	var items []interface{}
	if err := c.App().DB().Find(&items).Error; err != nil {
		return ctx.Status(500).JSON(map[string]string{"error": err.Error()})
	}
	return ctx.JSON(items)
}

// HandleGet` + strings.TrimSuffix(name, "Controller") + `ById handles getting a ` + strings.TrimSuffix(name, "Controller") + ` by ID
// Route: GET /` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + `/:id
// Description: Get a specific ` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + ` by ID
// Param: id - path - int - required - ` + strings.TrimSuffix(name, "Controller") + ` ID
// Response: 200 - ` + strings.TrimSuffix(name, "Controller") + `
// Response: 404 - Error message when ` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + ` not found
func (c *` + name + `) HandleGet` + strings.TrimSuffix(name, "Controller") + `ById(ctx *flux.Context) error {
	id := ctx.Param("id")
	var item interface{}
	if err := c.App().DB().First(&item, id).Error; err != nil {
		return ctx.Status(404).JSON(map[string]string{"error": "Not found"})
	}
	return ctx.JSON(item)
}

// HandleCreate` + strings.TrimSuffix(name, "Controller") + ` handles creating a new ` + strings.TrimSuffix(name, "Controller") + `
// Route: POST /` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + `
// Description: Create a new ` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + `
// Body: Create` + strings.TrimSuffix(name, "Controller") + `Request
// Response: 201 - ` + strings.TrimSuffix(name, "Controller") + `
// Response: 400 - Error message when request body is invalid
// Response: 500 - Error message when database operation fails
func (c *` + name + `) HandleCreate` + strings.TrimSuffix(name, "Controller") + `(ctx *flux.Context) error {
	var req Create` + strings.TrimSuffix(name, "Controller") + `Request
	
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(400).JSON(map[string]string{"error": err.Error()})
	}
	
	// Validate the request
	if err := ctx.Validate(req); err != nil {
		return ctx.Status(400).JSON(map[string]string{"error": err.Error()})
	}
	
	// Create record (replace with your model)
	item := map[string]interface{}{"name": req.Name}
	
	if err := c.App().DB().Create(&item).Error; err != nil {
		return ctx.Status(500).JSON(map[string]string{"error": err.Error()})
	}
	
	return ctx.Status(201).JSON(item)
}

// HandleUpdate` + strings.TrimSuffix(name, "Controller") + ` handles updating a ` + strings.TrimSuffix(name, "Controller") + `
// Route: PUT /` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + `/:id
// Description: Update a specific ` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + ` by ID
// Param: id - path - int - required - ` + strings.TrimSuffix(name, "Controller") + ` ID
// Body: Update` + strings.TrimSuffix(name, "Controller") + `Request
// Response: 200 - Updated ` + strings.TrimSuffix(name, "Controller") + `
// Response: 400 - Error message when request body is invalid
// Response: 404 - Error message when ` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + ` not found
// Response: 500 - Error message when database operation fails
func (c *` + name + `) HandleUpdate` + strings.TrimSuffix(name, "Controller") + `(ctx *flux.Context) error {
	id := ctx.Param("id")
	
	var req Update` + strings.TrimSuffix(name, "Controller") + `Request
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(400).JSON(map[string]string{"error": err.Error()})
	}
	
	// Validate the request
	if err := ctx.Validate(req); err != nil {
		return ctx.Status(400).JSON(map[string]string{"error": err.Error()})
	}
	
	// Update record (based on your model)
	var item interface{}
	if err := c.App().DB().First(&item, id).Error; err != nil {
		return ctx.Status(404).JSON(map[string]string{"error": "Not found"})
	}
	
	// Update fields based on request
	
	if err := c.App().DB().Save(&item).Error; err != nil {
		return ctx.Status(500).JSON(map[string]string{"error": err.Error()})
	}
	
	return ctx.JSON(item)
}

// HandleDelete` + strings.TrimSuffix(name, "Controller") + ` handles deleting a ` + strings.TrimSuffix(name, "Controller") + `
// Route: DELETE /` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + `/:id
// Description: Delete a specific ` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + ` by ID
// Param: id - path - int - required - ` + strings.TrimSuffix(name, "Controller") + ` ID
// Response: 204 - No content on successful deletion
// Response: 404 - Error message when ` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + ` not found
// Response: 500 - Error message when database operation fails
func (c *` + name + `) HandleDelete` + strings.TrimSuffix(name, "Controller") + `(ctx *flux.Context) error {
	id := ctx.Param("id")
	
	// Delete record 
	var item interface{}
	if err := c.App().DB().First(&item, id).Error; err != nil {
		return ctx.Status(404).JSON(map[string]string{"error": "Not found"})
	}
	
	if err := c.App().DB().Delete(&item).Error; err != nil {
		return ctx.Status(500).JSON(map[string]string{"error": err.Error()})
	}
	
	return ctx.Status(204).Send([]byte{})
}
`

	modelContent := `package models

import (
	"time"
)

// ` + strings.TrimSuffix(name, "Controller") + ` represents a ` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + ` entity
type ` + strings.TrimSuffix(name, "Controller") + ` struct {
	ID        uint      ` + "`json:\"id\" gorm:\"primaryKey\"`" + `
	CreatedAt time.Time ` + "`json:\"created_at\" gorm:\"autoCreateTime\"`" + `
	UpdatedAt time.Time ` + "`json:\"updated_at\" gorm:\"autoUpdateTime\"`" + `
	// Add your custom fields here
	Name string ` + "`json:\"name\" gorm:\"size:255;not null\"`" + `
	// Add more fields as needed
}

// TableName overrides the default table name
func (` + strings.TrimSuffix(name, "Controller") + `) TableName() string {
	return "` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + `s"
}
`

	typesContent := `package controllers

// Create` + strings.TrimSuffix(name, "Controller") + `Request represents the request body for creating a ` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + `
type Create` + strings.TrimSuffix(name, "Controller") + `Request struct {
	Name string ` + "`json:\"name\" validate:\"required\"`" + `
	// Add more fields as needed for creation
}

// Update` + strings.TrimSuffix(name, "Controller") + `Request represents the request body for updating a ` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + `
type Update` + strings.TrimSuffix(name, "Controller") + `Request struct {
	Name string ` + "`json:\"name\" validate:\"required\"`" + `
	// Add more fields as needed for updates
}
`

	// Create functional routes file
	routesContent := `package routes

import (
	"` + getCurrentModuleName() + `/app/controllers"
	"github.com/Fluxgo/flux/pkg/flux"
	"github.com/gofiber/fiber/v2"
)

// Register` + strings.TrimSuffix(name, "Controller") + `Routes registers all ` + strings.TrimSuffix(name, "Controller") + ` routes with the app
func Register` + strings.TrimSuffix(name, "Controller") + `Routes(app *flux.Application) {
	// Register controller with the app
	controller := &controllers.` + name + `{}
	app.RegisterController(controller)
	
	// If you prefer manual route registration instead of automatic registration:
	/*
	` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + `Group := app.Group("/` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + `")
	{
		// GET all ` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + `s
		` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + `Group.Get("/", func(c *fiber.Ctx) error {
			ctx := flux.NewContext(c, app)
			return controller.HandleGet` + strings.TrimSuffix(name, "Controller") + `s(ctx)
		})
		
		// GET ` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + ` by ID
		` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + `Group.Get("/:id", func(c *fiber.Ctx) error {
			ctx := flux.NewContext(c, app)
			return controller.HandleGet` + strings.TrimSuffix(name, "Controller") + `ById(ctx)
		})
		
		// POST new ` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + `
		` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + `Group.Post("/", func(c *fiber.Ctx) error {
			ctx := flux.NewContext(c, app)
			return controller.HandleCreate` + strings.TrimSuffix(name, "Controller") + `(ctx)
		})
		
		// PUT update ` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + `
		` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + `Group.Put("/:id", func(c *fiber.Ctx) error {
			ctx := flux.NewContext(c, app)
			return controller.HandleUpdate` + strings.TrimSuffix(name, "Controller") + `(ctx)
		})
		
		// DELETE ` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + `
		` + strings.ToLower(strings.TrimSuffix(name, "Controller")) + `Group.Delete("/:id", func(c *fiber.Ctx) error {
			ctx := flux.NewContext(c, app)
			return controller.HandleDelete` + strings.TrimSuffix(name, "Controller") + `(ctx)
		})
	}
	*/
}
`

	mainRoutesPath := filepath.Join("routes", "main.go")
	var mainRoutesContent string

	if _, err := os.Stat(mainRoutesPath); os.IsNotExist(err) {

		mainRoutesContent = `package routes

import (
	"github.com/Fluxgo/flux/pkg/flux"
)

// RegisterAllRoutes registers all application routes
func RegisterAllRoutes(app *flux.Application) {
	// Register ` + strings.TrimSuffix(name, "Controller") + ` routes
	Register` + strings.TrimSuffix(name, "Controller") + `Routes(app)
}
`
	} else {

		existingContent, err := os.ReadFile(mainRoutesPath)
		if err != nil {
			return fmt.Errorf("failed to read main routes file: %w", err)
		}

		contentStr := string(existingContent)

		if !strings.Contains(contentStr, "Register"+strings.TrimSuffix(name, "Controller")+"Routes") {

			registerFuncIndex := strings.Index(contentStr, "func RegisterAllRoutes")
			if registerFuncIndex > 0 {
				closingBraceIndex := strings.Index(contentStr[registerFuncIndex:], "}") + registerFuncIndex

				if closingBraceIndex > registerFuncIndex {

					mainRoutesContent = contentStr[:closingBraceIndex] +
						"\n\t// Register " + strings.TrimSuffix(name, "Controller") + " routes\n" +
						"\tRegister" + strings.TrimSuffix(name, "Controller") + "Routes(app)\n" +
						contentStr[closingBraceIndex:]
				} else {

					mainRoutesContent = contentStr
				}
			} else {

				mainRoutesContent = contentStr
			}
		} else {

			mainRoutesContent = contentStr
		}
	}

	if err := os.MkdirAll(filepath.Join("app", "controllers"), 0755); err != nil {
		return fmt.Errorf("failed to create controllers directory: %w", err)
	}

	if err := os.MkdirAll(filepath.Join("app", "models"), 0755); err != nil {
		return fmt.Errorf("failed to create models directory: %w", err)
	}

	if err := os.MkdirAll(filepath.Join("routes"), 0755); err != nil {
		return fmt.Errorf("failed to create routes directory: %w", err)
	}

	if err := os.WriteFile(filepath.Join("app", "controllers", strings.ToLower(strings.TrimSuffix(name, "Controller"))+"_controller.go"), []byte(controllerContent), 0644); err != nil {
		return fmt.Errorf("failed to create controller file: %w", err)
	}

	if err := os.WriteFile(filepath.Join("app", "models", strings.ToLower(strings.TrimSuffix(name, "Controller"))+".go"), []byte(modelContent), 0644); err != nil {
		return fmt.Errorf("failed to create model file: %w", err)
	}

	if err := os.WriteFile(filepath.Join("app", "controllers", strings.ToLower(strings.TrimSuffix(name, "Controller"))+"_types.go"), []byte(typesContent), 0644); err != nil {
		return fmt.Errorf("failed to create types file: %w", err)
	}

	if err := os.WriteFile(filepath.Join("routes", strings.ToLower(strings.TrimSuffix(name, "Controller"))+"_routes.go"), []byte(routesContent), 0644); err != nil {
		return fmt.Errorf("failed to create functional routes file: %w", err)
	}

	if mainRoutesContent != "" {
		if err := os.WriteFile(mainRoutesPath, []byte(mainRoutesContent), 0644); err != nil {
			return fmt.Errorf("failed to update main routes file: %w", err)
		}
	}

	fmt.Printf("Generated controller: %s\n", name)
	fmt.Printf("Generated model: %s\n", strings.TrimSuffix(name, "Controller"))
	fmt.Printf("Generated functional routes: routes/%s_routes.go\n", strings.ToLower(strings.TrimSuffix(name, "Controller")))

	return nil
}

func generateModel(name string) error {
	name = strings.ToUpper(name[:1]) + name[1:]
	modelContent := `package models

import (
	"time"
)

// ` + name + ` represents a ` + strings.ToLower(name) + ` entity
type ` + name + ` struct {
	ID        uint      ` + "`json:\"id\" gorm:\"primaryKey\"`" + `
	CreatedAt time.Time ` + "`json:\"created_at\" gorm:\"autoCreateTime\"`" + `
	UpdatedAt time.Time ` + "`json:\"updated_at\" gorm:\"autoUpdateTime\"`" + `
	
	
	// template fields (uncomment and modify as needed):
	// Name        string    ` + "`json:\"name\" gorm:\"size:255;not null\"`" + `
	// Description string    ` + "`json:\"description\" gorm:\"type:text\"`" + `
	// Status      string    ` + "`json:\"status\" gorm:\"size:50;default:'active'\"`" + `
	// Amount      float64   ` + "`json:\"amount\" gorm:\"type:decimal(10,2);default:0\"`" + `
	// IsActive    bool      ` + "`json:\"is_active\" gorm:\"default:true\"`" + `
	// ExpiresAt   time.Time ` + "`json:\"expires_at\" gorm:\"index\"`" + `
}

// TableName overrides the table name
func (` + name + `) TableName() string {
	return "` + strings.ToLower(name) + `s"
}

// BeforeCreate hook called before record creation
func (m *` + name + `) BeforeCreate() error {
	// Add custom validation or data preparation logic here
	return nil
}
`

	migrationContent := `package migrations

import (
	"` + getCurrentModuleName() + `/app/models"
	"gorm.io/gorm"
)

// Create` + name + `Table creates the ` + strings.ToLower(name) + `s table
func Create` + name + `Table(db *gorm.DB) error {
	return db.AutoMigrate(&models.` + name + `{})
}

// Drop` + name + `Table drops the ` + strings.ToLower(name) + `s table
func Drop` + name + `Table(db *gorm.DB) error {
	return db.Migrator().DropTable(&models.` + name + `{})
}
`

	repositoryContent := `package repositories

import (
	"` + getCurrentModuleName() + `/app/models"
	"gorm.io/gorm"
)

// ` + name + `Repository provides database operations for ` + name + ` model
type ` + name + `Repository struct {
	DB *gorm.DB
}

// New` + name + `Repository creates a new repository instance
func New` + name + `Repository(db *gorm.DB) *` + name + `Repository {
	return &` + name + `Repository{
		DB: db,
	}
}

// Create inserts a new ` + name + ` record
func (r *` + name + `Repository) Create(` + strings.ToLower(name) + ` *models.` + name + `) error {
	return r.DB.Create(` + strings.ToLower(name) + `).Error
}

// FindByID retrieves a ` + name + ` by ID
func (r *` + name + `Repository) FindByID(id uint) (*models.` + name + `, error) {
	var ` + strings.ToLower(name) + ` models.` + name + `
	err := r.DB.First(&` + strings.ToLower(name) + `, id).Error
	return &` + strings.ToLower(name) + `, err
}

// FindAll retrieves all ` + name + ` records
func (r *` + name + `Repository) FindAll() ([]models.` + name + `, error) {
	var ` + strings.ToLower(name) + `s []models.` + name + `
	err := r.DB.Find(&` + strings.ToLower(name) + `s).Error
	return ` + strings.ToLower(name) + `s, err
}

// Update updates a ` + name + ` record
func (r *` + name + `Repository) Update(` + strings.ToLower(name) + ` *models.` + name + `) error {
	return r.DB.Save(` + strings.ToLower(name) + `).Error
}

// Delete removes a ` + name + ` record
func (r *` + name + `Repository) Delete(id uint) error {
	return r.DB.Delete(&models.` + name + `{}, id).Error
}

// Count returns the total number of ` + name + ` records
func (r *` + name + `Repository) Count() (int64, error) {
	var count int64
	err := r.DB.Model(&models.` + name + `{}).Count(&count).Error
	return count, err
}

// Custom queries can be added below
`

	if err := os.WriteFile(filepath.Join("app", "models", strings.ToLower(name)+".go"), []byte(modelContent), 0644); err != nil {
		return fmt.Errorf("failed to create model file: %w", err)
	}

	if err := os.WriteFile(filepath.Join("database", "migrations", strings.ToLower(name)+"_migration.go"), []byte(migrationContent), 0644); err != nil {
		return fmt.Errorf("failed to create migration file: %w", err)
	}

	if err := os.MkdirAll(filepath.Join("app", "repositories"), 0755); err != nil {
		return fmt.Errorf("failed to create repositories directory: %w", err)
	}

	if err := os.WriteFile(filepath.Join("app", "repositories", strings.ToLower(name)+"_repository.go"), []byte(repositoryContent), 0644); err != nil {
		return fmt.Errorf("failed to create repository file: %w", err)
	}

	fmt.Printf("Generated model: %s\n", name)
	return nil
}

func generateMiddleware(name string) error {
	name = strings.ToUpper(name[:1]) + name[1:]
	if !strings.HasSuffix(name, "Middleware") {
		name += "Middleware"
	}

	// Create middleware directory if it doesn't exist
	if err := os.MkdirAll(filepath.Join("app", "middleware"), 0755); err != nil {
		return fmt.Errorf("failed to create middleware directory: %w", err)
	}

	middlewareContent := `package middleware

import (
	"github.com/Fluxgo/flux/pkg/flux"
	"github.com/gofiber/fiber/v2"
)

// ` + name + ` is a middleware that handles ` + strings.TrimSuffix(name, "Middleware") + ` functionality
func ` + name + `() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Middleware logic here
		// Example: Authentication check, request validation, logging, etc.
		
		// Get the flux context
		ctx := &flux.Context{Ctx: c}
		
		// Example middleware implementation:
		// 1. Extract data or validate request
		// requestID := c.Get("X-Request-ID")
		
		// 2. Set values in context if needed
		// c.Locals("request_id", requestID)
		
		// 3. Perform checks
		// if !someCondition {
		//     return ctx.Status(401).JSON(map[string]string{"error": "Unauthorized"})
		// }
		
		// 4. Continue to next middleware or route handler
		return c.Next()
	}
}

// Register` + name + ` registers the middleware with the application
func Register` + name + `(app *flux.Application) {
	// Global middleware registration
	// app.Use(` + name + `())
	
	// Or group-specific middleware
	// apiGroup := app.Group("/api")
	// apiGroup.Use(` + name + `())
}
`

	if err := os.WriteFile(filepath.Join("app", "middleware", strings.ToLower(strings.TrimSuffix(name, "Middleware"))+"_middleware.go"), []byte(middlewareContent), 0644); err != nil {
		return fmt.Errorf("failed to create middleware file: %w", err)
	}

	fmt.Printf("Generated middleware: %s\n", name)
	return nil
}

func generateService(name string) error {
	name = strings.ToUpper(name[:1]) + name[1:]
	if !strings.HasSuffix(name, "Service") {
		name += "Service"
	}

	// Create services directory if it doesn't exist
	if err := os.MkdirAll(filepath.Join("app", "services"), 0755); err != nil {
		return fmt.Errorf("failed to create services directory: %w", err)
	}

	serviceContent := `package services

import (
	"` + getCurrentModuleName() + `/app/models"
	"` + getCurrentModuleName() + `/app/repositories"
	"github.com/Fluxgo/flux/pkg/flux"
)

// ` + name + ` provides business logic for ` + strings.TrimSuffix(name, "Service") + ` operations
type ` + name + ` struct {
	app *flux.Application
	// Add repositories or other dependencies here
	// repo *repositories.` + strings.TrimSuffix(name, "Service") + `Repository
}

// New` + name + ` creates a new instance of ` + name + `
func New` + name + `(app *flux.Application) *` + name + ` {
	return &` + name + `{
		app: app,
		// Initialize repositories or other dependencies here
		// repo: repositories.New` + strings.TrimSuffix(name, "Service") + `Repository(app.DB()),
	}
}

// Get` + strings.TrimSuffix(name, "Service") + ` retrieves a ` + strings.TrimSuffix(name, "Service") + ` by ID
func (s *` + name + `) Get` + strings.TrimSuffix(name, "Service") + `(id uint) (interface{}, error) {
	// Example service method implementation
	// return s.repo.FindByID(id)
	
	// Placeholder implementation
	return map[string]interface{}{
		"id":   id,
		"name": "Sample ` + strings.TrimSuffix(name, "Service") + `",
	}, nil
}

// GetAll` + strings.TrimSuffix(name, "Service") + `s retrieves all ` + strings.TrimSuffix(name, "Service") + `s
func (s *` + name + `) GetAll` + strings.TrimSuffix(name, "Service") + `s() ([]interface{}, error) {
	// Example service method implementation
	// return s.repo.FindAll()
	
	// Placeholder implementation
	return []interface{}{
		map[string]interface{}{
			"id":   1,
			"name": "Sample ` + strings.TrimSuffix(name, "Service") + ` 1",
		},
		map[string]interface{}{
			"id":   2,
			"name": "Sample ` + strings.TrimSuffix(name, "Service") + ` 2",
		},
	}, nil
}

// Create` + strings.TrimSuffix(name, "Service") + ` creates a new ` + strings.TrimSuffix(name, "Service") + `
func (s *` + name + `) Create` + strings.TrimSuffix(name, "Service") + `(data map[string]interface{}) (interface{}, error) {
	// Example service method implementation
	// newItem := &models.` + strings.TrimSuffix(name, "Service") + `{
	//     Name: data["name"].(string),
	// }
	// err := s.repo.Create(newItem)
	// return newItem, err
	
	// Placeholder implementation
	return map[string]interface{}{
		"id":   3,
		"name": data["name"],
	}, nil
}

// Update` + strings.TrimSuffix(name, "Service") + ` updates an existing ` + strings.TrimSuffix(name, "Service") + `
func (s *` + name + `) Update` + strings.TrimSuffix(name, "Service") + `(id uint, data map[string]interface{}) (interface{}, error) {
	// Example service method implementation
	// item, err := s.repo.FindByID(id)
	// if err != nil {
	//     return nil, err
	// }
	// 
	// // Update fields
	// if name, ok := data["name"].(string); ok {
	//     item.Name = name
	// }
	// 
	// err = s.repo.Update(item)
	// return item, err
	
	// Placeholder implementation
	return map[string]interface{}{
		"id":   id,
		"name": data["name"],
	}, nil
}

// Delete` + strings.TrimSuffix(name, "Service") + ` deletes a ` + strings.TrimSuffix(name, "Service") + ` by ID
func (s *` + name + `) Delete` + strings.TrimSuffix(name, "Service") + `(id uint) error {
	// Example service method implementation
	// return s.repo.Delete(id)
	
	// Placeholder implementation
	return nil
}
`

	// Create the service file
	if err := os.WriteFile(filepath.Join("app", "services", strings.ToLower(strings.TrimSuffix(name, "Service"))+"_service.go"), []byte(serviceContent), 0644); err != nil {
		return fmt.Errorf("failed to create service file: %w", err)
	}

	fmt.Printf("Generated service: %s\n", name)
	return nil
}

func generateDocumentation() error {
	outputDir := filepath.Join("docs")

	
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create docs directory: %w", err)
	}

	
	app, err := flux.New(&flux.Config{
		Name:        getCurrentModuleName(),
		Version:     "1.0.0",
		Description: "API Documentation",
		Server: flux.ServerConfig{
			Host: "localhost",
			Port: 3000,
		},
	})

	if err != nil {
		return fmt.Errorf("failed to create application for documentation: %w", err)
	}

	
	app.Routes()

	
	spec, err := app.GenerateOpenAPI()
	if err != nil {
		return fmt.Errorf("failed to generate OpenAPI specification: %w", err)
	}

	
	jsonSpec, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Save OpenAPI specification to file
	if err := os.WriteFile(filepath.Join(outputDir, "openapi.json"), jsonSpec, 0644); err != nil {
		return fmt.Errorf("failed to write OpenAPI specification: %w", err)
	}

	// Generate Swagger UI HTML
	swaggerUI, err := flux.GenerateSwaggerUI(spec)
	if err != nil {
		return fmt.Errorf("failed to generate Swagger UI: %w", err)
	}

	
	if err := os.WriteFile(filepath.Join(outputDir, "swagger.html"), []byte(swaggerUI), 0644); err != nil {
		return fmt.Errorf("failed to write Swagger UI: %w", err)
	}

	
	markdown := generateMarkdownDocumentation(app, spec)

	
	if err := os.WriteFile(filepath.Join(outputDir, "api.md"), []byte(markdown), 0644); err != nil {
		return fmt.Errorf("failed to write markdown documentation: %w", err)
	}

	fmt.Printf("Documentation generated in %s directory:\n", outputDir)
	fmt.Println("- OpenAPI specification: openapi.json")
	fmt.Println("- Swagger UI: swagger.html")
	fmt.Println("- Markdown documentation: api.md")

	return nil
}

func generateMarkdownDocumentation(app *flux.Application, spec *flux.OpenAPISpec) string {
	var doc strings.Builder

	doc.WriteString("# API Documentation\n\n")
	doc.WriteString(fmt.Sprintf("## %s\n\n", spec.Info.Title))
	doc.WriteString(fmt.Sprintf("%s\n\n", spec.Info.Description))
	doc.WriteString(fmt.Sprintf("Version: %s\n\n", spec.Info.Version))

	doc.WriteString("## Endpoints\n\n")

	
	var paths []string
	for path := range spec.Paths {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	for _, path := range paths {
		pathItem := spec.Paths[path]

		
		for _, method := range []struct {
			Name      string
			Operation *flux.Operation
		}{
			{"GET", pathItem.Get},
			{"POST", pathItem.Post},
			{"PUT", pathItem.Put},
			{"DELETE", pathItem.Delete},
			{"PATCH", pathItem.Patch},
		} {
			if method.Operation != nil {
				doc.WriteString(fmt.Sprintf("### %s %s\n\n", method.Name, path))
				doc.WriteString(fmt.Sprintf("**Description:** %s\n\n", method.Operation.Description))

				
				if len(method.Operation.Parameters) > 0 {
					doc.WriteString("**Parameters:**\n\n")
					doc.WriteString("| Name | Located in | Description | Required | Schema |\n")
					doc.WriteString("| ---- | ---------- | ----------- | -------- | ------ |\n")

					for _, param := range method.Operation.Parameters {
						doc.WriteString(fmt.Sprintf("| %s | %s | %s | %t | %s |\n",
							param.Name, param.In, param.Description, param.Required, getSchemaType(param.Schema)))
					}
					doc.WriteString("\n")
				}

				
				if method.Operation.RequestBody != nil {
					doc.WriteString("**Request Body:**\n\n")
					doc.WriteString(fmt.Sprintf("Description: %s\n\n", method.Operation.RequestBody.Description))
					doc.WriteString("Content Type: application/json\n\n")

					
					for contentType, mediaType := range method.Operation.RequestBody.Content {
						if mediaType.Schema != nil && mediaType.Schema.Properties != nil {
							doc.WriteString(fmt.Sprintf("Schema (%s):\n\n", contentType))
							doc.WriteString("| Property | Type | Description | Required |\n")
							doc.WriteString("| -------- | ---- | ----------- | -------- |\n")

							var required map[string]bool = make(map[string]bool)
							for _, req := range mediaType.Schema.Required {
								required[req] = true
							}

							for propName, propSchema := range mediaType.Schema.Properties {
								doc.WriteString(fmt.Sprintf("| %s | %s | %s | %t |\n",
									propName, getSchemaType(propSchema), propSchema.Description, required[propName]))
							}
							doc.WriteString("\n")
						}
					}
				}

				
				doc.WriteString("**Responses:**\n\n")
				doc.WriteString("| Status Code | Description | Schema |\n")
				doc.WriteString("| ----------- | ----------- | ------ |\n")

				var codes []string
				for code := range method.Operation.Responses {
					codes = append(codes, code)
				}
				sort.Strings(codes)

				for _, code := range codes {
					response := method.Operation.Responses[code]

					
					schemaType := "No content"
					for _, mediaType := range response.Content {
						if mediaType.Schema != nil {
							schemaType = getSchemaType(mediaType.Schema)
							break
						}
					}

					doc.WriteString(fmt.Sprintf("| %s | %s | %s |\n", code, response.Description, schemaType))
				}
				doc.WriteString("\n---\n\n")
			}
		}
	}

	
	if len(spec.Components.Schemas) > 0 {
		doc.WriteString("## Models\n\n")

		var schemaNames []string
		for name := range spec.Components.Schemas {
			schemaNames = append(schemaNames, name)
		}
		sort.Strings(schemaNames)

		for _, name := range schemaNames {
			schema := spec.Components.Schemas[name]
			doc.WriteString(fmt.Sprintf("### %s\n\n", name))

			if schema.Description != "" {
				doc.WriteString(fmt.Sprintf("%s\n\n", schema.Description))
			}

			if len(schema.Properties) > 0 {
				doc.WriteString("| Property | Type | Description | Required |\n")
				doc.WriteString("| -------- | ---- | ----------- | -------- |\n")

				var required map[string]bool = make(map[string]bool)
				for _, req := range schema.Required {
					required[req] = true
				}

				var propNames []string
				for propName := range schema.Properties {
					propNames = append(propNames, propName)
				}
				sort.Strings(propNames)

				for _, propName := range propNames {
					propSchema := schema.Properties[propName]
					doc.WriteString(fmt.Sprintf("| %s | %s | %s | %t |\n",
						propName, getSchemaType(propSchema), propSchema.Description, required[propName]))
				}
				doc.WriteString("\n")
			}
		}
	}

	return doc.String()
}

func getSchemaType(schema *flux.Schema) string {
	if schema == nil {
		return "unknown"
	}

	switch schema.Type {
	case "array":
		if schema.Items != nil {
			return fmt.Sprintf("array of %s", getSchemaType(schema.Items))
		}
		return "array"
	case "object":
		if len(schema.Properties) > 0 {
			return "object"
		}
		if schema.AdditionalProperties != nil {
			return fmt.Sprintf("map[string]%s", getSchemaType(schema.AdditionalProperties))
		}
		return "object"
	default:
		if schema.Format != "" {
			return fmt.Sprintf("%s (%s)", schema.Type, schema.Format)
		}
		return schema.Type
	}
}

func getCurrentModuleName() string {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return "app"
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) > 0 {
		parts := strings.Fields(lines[0])
		if len(parts) >= 2 && parts[0] == "module" {
			return parts[1]
		}
	}

	return "app"
}