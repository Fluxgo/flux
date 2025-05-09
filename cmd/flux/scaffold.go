package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	github.com/Fluxgo/flux v0.1.3
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

	if err := os.MkdirAll(filepath.Join("app", "middleware"), 0755); err != nil {
		return fmt.Errorf("failed to create middleware directory: %w", err)
	}

	middlewareContent := `package middleware

import (
	"fmt"
	"time"

	"github.com/Fluxgo/flux/pkg/flux"
)

// ` + name + ` is a middleware that performs checks before request handling
func ` + name + `(options ...interface{}) flux.MiddlewareFunc {
	// Configure middleware with options if provided
	config := parseOptions(options...)

	return func(next flux.HandlerFunc) flux.HandlerFunc {
		return func(ctx *flux.Context) error {
			// Store the start time for measuring request duration
			startTime := time.Now()
			
			// Get request information
			method := ctx.Method()
			path := ctx.Path()
			
			// Log the incoming request if enabled
			if config.LogRequest {
				ctx.App().Logger().Info("Request started: %s %s", method, path)
			}
			
			// Add request ID to the context for tracking
			requestID := generateRequestID()
			ctx.SetLocal("request_id", requestID)
			ctx.Set("X-Request-ID", requestID)
			
			// You can implement custom authentication logic here
			// Example: JWT Token verification
			// token := ctx.Get("Authorization")
			// if token != "" {
			//     // Validate token and set user in context
			//     user, err := validateToken(token)
			//     if err != nil {
			//         return ctx.Status(401).JSON(map[string]string{"error": "Invalid token"})
			//     }
			//     ctx.SetLocal("user", user)
			// }
			
			// Continue to the next middleware or the actual route handler
			err := next(ctx)
			
			// Calculate request duration
			duration := time.Since(startTime)
			
			// Log the completion of the request
			if config.LogRequest {
				statusCode := ctx.Response().StatusCode()
				ctx.App().Logger().Info("Request completed: %s %s [%d] - %v", method, path, statusCode, duration)
			}
			
			// You can add custom response headers here
			ctx.Set("X-Response-Time", fmt.Sprintf("%v", duration))
			
			// Return the error (if any) from the handler chain
			return err
		}
	}
}

// Configuration options for the middleware
type middlewareConfig struct {
	LogRequest bool
	// Add more configuration options as needed
}

// Parse middleware options
func parseOptions(options ...interface{}) middlewareConfig {
	config := middlewareConfig{
		LogRequest: true, // Default to true
	}
	
	// Process provided options
	for _, opt := range options {
		switch o := opt.(type) {
		case bool:
			config.LogRequest = o
		// Add more option types as needed
		}
	}
	
	return config
}

// Generate a unique request ID
func generateRequestID() string {
	// Simple implementation, can be replaced with a more robust one
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
`

	// Also create a sample usage file
	exampleContent := `package middleware

import (
	"github.com/Fluxgo/flux/pkg/flux"
)

// Example showing how to use ` + name + `

/*
func SetupRoutes(app *flux.Application) {
	// Apply middleware globally to all routes
	app.Use(` + name + `())
	
	// Or with custom options
	app.Use(` + name + `(false)) // Disable request logging

	// Apply to a specific controller
	userController := &controllers.UserController{}
	userController.Use(` + name + `())
	app.RegisterController(userController)
	
	// Apply to a route group
	api := app.Group("/api")
	api.Use(` + name + `())
}
*/
`

	if err := os.WriteFile(filepath.Join("app", "middleware", strings.ToLower(strings.TrimSuffix(name, "Middleware"))+"_middleware.go"), []byte(middlewareContent), 0644); err != nil {
		return fmt.Errorf("failed to create middleware file: %w", err)
	}

	if err := os.WriteFile(filepath.Join("app", "middleware", strings.ToLower(strings.TrimSuffix(name, "Middleware"))+"_example.go"), []byte(exampleContent), 0644); err != nil {
		return fmt.Errorf("failed to create middleware example file: %w", err)
	}

	fmt.Printf("Generated middleware: %s\n", name)
	fmt.Println("Created middleware files:")
	fmt.Printf("  - app/middleware/%s_middleware.go\n", strings.ToLower(strings.TrimSuffix(name, "Middleware")))
	fmt.Printf("  - app/middleware/%s_example.go\n", strings.ToLower(strings.TrimSuffix(name, "Middleware")))
	return nil
}

func generateService(name string) error {
	name = strings.ToUpper(name[:1]) + name[1:]
	if !strings.HasSuffix(name, "Service") {
		name += "Service"
	}

	if err := os.MkdirAll(filepath.Join("app", "services"), 0755); err != nil {
		return fmt.Errorf("failed to create services directory: %w", err)
	}

	// Check if there's a corresponding model
	modelName := strings.TrimSuffix(name, "Service")
	modelPath := filepath.Join("app", "models", strings.ToLower(modelName)+".go")
	hasModel := false
	if _, err := os.Stat(modelPath); err == nil {
		hasModel = true
	}

	serviceContent := `package services

import (
	"context"
	"errors"
	"time"

	"` + getCurrentModuleName() + `/app/models"
	"` + getCurrentModuleName() + `/app/repositories"
	"gorm.io/gorm"
)

// ` + name + ` provides business logic for ` + strings.TrimSuffix(name, "Service") + ` operations
type ` + name + ` struct {
	db         *gorm.DB
	repository *repositories.` + modelName + `Repository
	// You can add more dependencies here like:
	// cache      *cache.Client
	// events     *events.Publisher
	// config     *config.Config
}

// New` + name + ` creates a new instance of the ` + name + `
func New` + name + `(db *gorm.DB) *` + name + ` {
	return &` + name + `{
		db:         db,
		repository: repositories.New` + modelName + `Repository(db),
	}
}

// Get` + modelName + ` retrieves a ` + modelName + ` by ID with business logic
func (s *` + name + `) Get` + modelName + `(ctx context.Context, id uint) (*models.` + modelName + `, error) {
	// Create a context with timeout for database operations
	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Add any business logic before fetching
	// For example, permission checks, logging, metrics, etc.

	// Fetch the entity from the repository
	entity, err := s.repository.FindByID(id)
	if err != nil {
		// You can transform specific database errors to domain-specific errors
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("` + modelName + ` not found")
		}
		return nil, err
	}

	// Add post-processing if needed
	// ...

	return entity, nil
}

// GetAll` + modelName + `s retrieves all ` + modelName + ` records with pagination
func (s *` + name + `) GetAll` + modelName + `s(ctx context.Context, page, pageSize int) ([]models.` + modelName + `, int64, error) {
	dbCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20 // Default page size
	}

	// Get total count
	totalCount, err := s.repository.Count()
	if err != nil {
		return nil, 0, err
	}

	// Get data with pagination
	entities, err := s.repository.FindAllPaginated(page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	return entities, totalCount, nil
}

// Create` + modelName + ` creates a new ` + modelName + ` with validation
func (s *` + name + `) Create` + modelName + `(ctx context.Context, ` + strings.ToLower(modelName) + ` *models.` + modelName + `) error {
	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Perform business validation here
	if err := s.validate` + modelName + `(` + strings.ToLower(modelName) + `); err != nil {
		return err
	}

	// Additional business logic before creation
	// ...

	// Create in database
	err := s.repository.Create(` + strings.ToLower(modelName) + `)
	if err != nil {
		return err
	}

	// Post-creation processing
	// Example: s.publishCreationEvent(` + strings.ToLower(modelName) + `)

	return nil
}

// Update` + modelName + ` updates an existing ` + modelName + ` with validation
func (s *` + name + `) Update` + modelName + `(ctx context.Context, ` + strings.ToLower(modelName) + ` *models.` + modelName + `) error {
	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Verify that the entity exists
	existing, err := s.repository.FindByID(` + strings.ToLower(modelName) + `.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("cannot update non-existent ` + modelName + `")
		}
		return err
	}

	// Perform business validation
	if err := s.validate` + modelName + `(` + strings.ToLower(modelName) + `); err != nil {
		return err
	}

	// Update in database
	return s.repository.Update(` + strings.ToLower(modelName) + `)
}

// Delete` + modelName + ` deletes a ` + modelName + ` by ID
func (s *` + name + `) Delete` + modelName + `(ctx context.Context, id uint) error {
	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Check if entity exists before deletion
	existing, err := s.repository.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("cannot delete non-existent ` + modelName + `")
		}
		return err
	}

	// Perform deletion
	if err := s.repository.Delete(id); err != nil {
		return err
	}

	// Post-deletion processing
	// Example: s.publishDeletionEvent(id)

	return nil
}

// SearchBy` + modelName + `Name searches for ` + modelName + ` entities by name
func (s *` + name + `) SearchBy` + modelName + `Name(ctx context.Context, name string, page, pageSize int) ([]models.` + modelName + `, error) {
	dbCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Implement search functionality
	// This is an example - actual implementation depends on your repository
	// return s.repository.FindByName(name, page, pageSize)
	
	// Placeholder until repository is fully implemented
	return []models.` + modelName + `{}, nil
}

// Private helper methods

// validate` + modelName + ` contains business validation logic
func (s *` + name + `) validate` + modelName + `(` + strings.ToLower(modelName) + ` *models.` + modelName + `) error {
	// Add business validation rules here
	// Example: if ` + strings.ToLower(modelName) + `.Name == "" {
	//    return errors.New("name cannot be empty")
	// }
	return nil
}

// Add more business methods here as needed
`

	// Create a repository implementation if one doesn't exist yet
	repositoryDir := filepath.Join("app", "repositories")
	if err := os.MkdirAll(repositoryDir, 0755); err != nil {
		return fmt.Errorf("failed to create repositories directory: %w", err)
	}

	repositoryPath := filepath.Join(repositoryDir, strings.ToLower(modelName)+"_repository.go")
	if _, err := os.Stat(repositoryPath); os.IsNotExist(err) {
		// Create repository implementation
		repositoryContent := `package repositories

import (
	"` + getCurrentModuleName() + `/app/models"
	"gorm.io/gorm"
)

// ` + modelName + `Repository provides database operations for ` + modelName + ` model
type ` + modelName + `Repository struct {
	DB *gorm.DB
}

// New` + modelName + `Repository creates a new repository instance
func New` + modelName + `Repository(db *gorm.DB) *` + modelName + `Repository {
	return &` + modelName + `Repository{
		DB: db,
	}
}

// Create inserts a new ` + modelName + ` record
func (r *` + modelName + `Repository) Create(` + strings.ToLower(modelName) + ` *models.` + modelName + `) error {
	return r.DB.Create(` + strings.ToLower(modelName) + `).Error
}

// FindByID retrieves a ` + modelName + ` by ID
func (r *` + modelName + `Repository) FindByID(id uint) (*models.` + modelName + `, error) {
	var ` + strings.ToLower(modelName) + ` models.` + modelName + `
	err := r.DB.First(&` + strings.ToLower(modelName) + `, id).Error
	return &` + strings.ToLower(modelName) + `, err
}

// FindAll retrieves all ` + modelName + ` records
func (r *` + modelName + `Repository) FindAll() ([]models.` + modelName + `, error) {
	var ` + strings.ToLower(modelName) + `s []models.` + modelName + `
	err := r.DB.Find(&` + strings.ToLower(modelName) + `s).Error
	return ` + strings.ToLower(modelName) + `s, err
}

// FindAllPaginated retrieves ` + modelName + ` records with pagination
func (r *` + modelName + `Repository) FindAllPaginated(page, pageSize int) ([]models.` + modelName + `, error) {
	var ` + strings.ToLower(modelName) + `s []models.` + modelName + `
	offset := (page - 1) * pageSize
	err := r.DB.Offset(offset).Limit(pageSize).Find(&` + strings.ToLower(modelName) + `s).Error
	return ` + strings.ToLower(modelName) + `s, err
}

// Update updates a ` + modelName + ` record
func (r *` + modelName + `Repository) Update(` + strings.ToLower(modelName) + ` *models.` + modelName + `) error {
	return r.DB.Save(` + strings.ToLower(modelName) + `).Error
}

// Delete removes a ` + modelName + ` record
func (r *` + modelName + `Repository) Delete(id uint) error {
	return r.DB.Delete(&models.` + modelName + `{}, id).Error
}

// Count returns the total number of ` + modelName + ` records
func (r *` + modelName + `Repository) Count() (int64, error) {
	var count int64
	err := r.DB.Model(&models.` + modelName + `{}).Count(&count).Error
	return count, err
}

// FindByName finds records by name (partial match)
func (r *` + modelName + `Repository) FindByName(name string, page, pageSize int) ([]models.` + modelName + `, error) {
	var ` + strings.ToLower(modelName) + `s []models.` + modelName + `
	offset := (page - 1) * pageSize
	err := r.DB.Where("name LIKE ?", "%" + name + "%").Offset(offset).Limit(pageSize).Find(&` + strings.ToLower(modelName) + `s).Error
	return ` + strings.ToLower(modelName) + `s, err
}

// Add more custom query methods below as needed
`
		if err := os.WriteFile(repositoryPath, []byte(repositoryContent), 0644); err != nil {
			return fmt.Errorf("failed to create repository file: %w", err)
		}
	}

	// Create a usage sample file
	serviceExampleContent := `package main

import (
	"context"
	"fmt"
	"log"

	"` + getCurrentModuleName() + `/app/models"
	"` + getCurrentModuleName() + `/app/services"
)

/*
The following example shows how to use the ` + name + `:

func main() {
	// Get database connection (simplified for example)
	db := connectToDatabase()
	
	// Create a new service
	` + strings.ToLower(modelName) + `Service := services.New` + name + `(db)
	
	// Example: Create a new ` + modelName + `
	new` + modelName + ` := &models.` + modelName + `{
		Name: "Example Name",
		// Set other fields...
	}
	
	ctx := context.Background()
	
	if err := ` + strings.ToLower(modelName) + `Service.Create` + modelName + `(ctx, new` + modelName + `); err != nil {
		log.Fatalf("Failed to create ` + modelName + `: %v", err)
	}
	
	// Example: Get all ` + modelName + `s with pagination
	items, total, err := ` + strings.ToLower(modelName) + `Service.GetAll` + modelName + `s(ctx, 1, 10)
	if err != nil {
		log.Fatalf("Failed to get ` + modelName + `s: %v", err)
	}
	
	fmt.Printf("Found %d ` + modelName + `s (total: %d)\n", len(items), total)
	
	// Example: Update an existing ` + modelName + `
	if len(items) > 0 {
		item := items[0]
		item.Name = "Updated Name"
		
		if err := ` + strings.ToLower(modelName) + `Service.Update` + modelName + `(ctx, &item); err != nil {
			log.Fatalf("Failed to update ` + modelName + `: %v", err)
		}
	}
	
	// Example: Delete a ` + modelName + `
	if len(items) > 0 {
		if err := ` + strings.ToLower(modelName) + `Service.Delete` + modelName + `(ctx, items[0].ID); err != nil {
			log.Fatalf("Failed to delete ` + modelName + `: %v", err)
		}
	}
}
*/
`

	servicePath := filepath.Join("app", "services", strings.ToLower(strings.TrimSuffix(name, "Service"))+"_service.go")
	if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		return fmt.Errorf("failed to create service file: %w", err)
	}

	examplePath := filepath.Join("app", "services", strings.ToLower(strings.TrimSuffix(name, "Service"))+"_example.go")
	if err := os.WriteFile(examplePath, []byte(serviceExampleContent), 0644); err != nil {
		return fmt.Errorf("failed to create service example file: %w", err)
	}

	fmt.Printf("Generated service: %s\n", name)
	fmt.Println("Created service files:")
	fmt.Printf("  - app/services/%s_service.go\n", strings.ToLower(strings.TrimSuffix(name, "Service")))
	fmt.Printf("  - app/services/%s_example.go\n", strings.ToLower(strings.TrimSuffix(name, "Service")))
	
	if !hasModel {
		fmt.Println("\nNotice: No matching model found for this service. Consider creating one with:")
		fmt.Printf("  flux make:model %s\n", strings.TrimSuffix(name, "Service"))
	}
	
	return nil
}

func generateDocumentation() error {
	fmt.Println("Starting API documentation generation...")
	
	
	if err := os.MkdirAll(filepath.Join("docs"), 0755); err != nil {
		return fmt.Errorf("failed to create docs directory: %w", err)
	}

	config := &flux.Config{
		Name:        "API Documentation Generator",
		Version:     "1.0.0",
		Description: "Generated API Documentation",
		Server: flux.ServerConfig{
			Host:     "localhost",
			Port:     3000,
			BasePath: "/",
		},
	}

	app, err := flux.New(config)
	if err != nil {
		return fmt.Errorf("failed to create temporary app: %w", err)
	}

	// Look for controllers to auto-register
	controllersDir := filepath.Join("app", "controllers")
	if _, err := os.Stat(controllersDir); err == nil {
		
		entries, err := os.ReadDir(controllersDir)
		if err != nil {
			fmt.Printf("Warning: Could not read controllers directory: %v\n", err)
		} else {
			fmt.Printf("Found %d controller files to scan\n", len(entries))

			
			routeComments := extractRouteCommentsFromControllers(controllersDir, entries)
			
			
			for _, route := range routeComments {
				app.AddDocumentedRoute(route.Method, route.Path, route.Handler, route.Description, route.Params)
			}
		}
	}

	
	fmt.Println("Processing API routes and generating OpenAPI specification...")
	spec := app.GenerateOpenAPISpec()

	
	specJSON, err := app.OpenAPISpecToJSON()
	if err != nil {
		return fmt.Errorf("failed to generate OpenAPI JSON: %w", err)
	}

	openAPIPath := filepath.Join("docs", "openapi.json")
	if err := os.WriteFile(openAPIPath, []byte(specJSON), 0644); err != nil {
		return fmt.Errorf("failed to write OpenAPI JSON file: %w", err)
	}
	fmt.Printf("Generated OpenAPI specification: %s\n", openAPIPath)

	
	openAPIYAMLPath := filepath.Join("docs", "openapi.yaml")
	specYAML, err := app.OpenAPISpecToYAML()
	if err != nil {
		fmt.Printf("Warning: Could not generate YAML format: %v\n", err)
	} else {
		if err := os.WriteFile(openAPIYAMLPath, []byte(specYAML), 0644); err != nil {
			fmt.Printf("Warning: Could not write YAML file: %v\n", err)
		} else {
			fmt.Printf("Generated OpenAPI YAML: %s\n", openAPIYAMLPath)
		}
	}

	
	swaggerUIPath := filepath.Join("docs", "swagger.html")
	swaggerUI := generateSwaggerUIHTML()
	if err := os.WriteFile(swaggerUIPath, []byte(swaggerUI), 0644); err != nil {
		return fmt.Errorf("failed to write Swagger UI file: %w", err)
	}
	fmt.Printf("Generated Swagger UI: %s\n", swaggerUIPath)

	
	redocUIPath := filepath.Join("docs", "redoc.html")
	redocUI := generateRedocUIHTML()
	if err := os.WriteFile(redocUIPath, []byte(redocUI), 0644); err != nil {
		fmt.Printf("Warning: Could not write Redoc UI file: %v\n", err)
	} else {
		fmt.Printf("Generated Redoc UI: %s\n", redocUIPath)
	}

	
	serverFilePath := filepath.Join("docs", "serve.go")
	serverContent := generateServerCode()
	if err := os.WriteFile(serverFilePath, []byte(serverContent), 0644); err != nil {
		fmt.Printf("Warning: Could not write documentation server file: %v\n", err)
	} else {
		fmt.Printf("Generated documentation server: %s\n", serverFilePath)
		fmt.Println("You can run the documentation server with: go run docs/serve.go")
	}

	
	readmePath := filepath.Join("docs", "README.md")
	readmeContent := generateReadmeContent()
	if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		fmt.Printf("Warning: Could not write documentation README: %v\n", err)
	} else {
		fmt.Printf("Generated documentation README: %s\n", readmePath)
	}

	fmt.Println("\nAPI Documentation generation complete!")
	fmt.Println("To view the documentation, open docs/swagger.html in your browser")
	fmt.Println("or run the documentation server with: go run docs/serve.go")
	
	return nil
}

// Helper functions to generate documentation artifacts
func generateSwaggerUIHTML() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>API Documentation - Swagger UI</title>
  <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@4.5.0/swagger-ui.css" />
  <link rel="icon" type="image/png" href="https://unpkg.com/swagger-ui-dist@4.5.0/favicon-32x32.png" sizes="32x32" />
  <link rel="icon" type="image/png" href="https://unpkg.com/swagger-ui-dist@4.5.0/favicon-16x16.png" sizes="16x16" />
  <style>
    html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
    *, *:before, *:after { box-sizing: inherit; }
    body { margin: 0; background: #fafafa; }
    .topbar { display: none; }
  </style>
</head>

<body>
  <div id="swagger-ui"></div>

  <script src="https://unpkg.com/swagger-ui-dist@4.5.0/swagger-ui-bundle.js"></script>
  <script src="https://unpkg.com/swagger-ui-dist@4.5.0/swagger-ui-standalone-preset.js"></script>
  <script>
    window.onload = function() {
      const ui = SwaggerUIBundle({
        url: "openapi.json",
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "StandaloneLayout"
      });
      window.ui = ui;
    };
  </script>
</body>
</html>`
}


func generateRedocUIHTML() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>API Documentation - Redoc</title>
  <link href="https://fonts.googleapis.com/css?family=Montserrat:300,400,700|Roboto:300,400,700" rel="stylesheet">
  <style>
    body { margin: 0; padding: 0; }
  </style>
</head>
<body>
  <redoc spec-url="openapi.json"></redoc>
  <script src="https://cdn.jsdelivr.net/npm/redoc@next/bundles/redoc.standalone.js"></script>
</body>
</html>`
}


func generateServerCode() string {
	return `package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	port := "8080"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to get executable path: %v", err)
	}
	docsDir := filepath.Dir(exePath)

	
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		
		http.Redirect(w, r, "/swagger", http.StatusMovedPermanently)
	})

	
	http.HandleFunc("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(docsDir, "swagger.html"))
	})

	
	http.HandleFunc("/redoc", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(docsDir, "redoc.html"))
	})

	
	http.HandleFunc("/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		http.ServeFile(w, r, filepath.Join(docsDir, "openapi.json"))
	})

	
	http.HandleFunc("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		http.ServeFile(w, r, filepath.Join(docsDir, "openapi.yaml"))
	})

	fmt.Printf("API Documentation server started at http://localhost:%s\n", port)
	fmt.Printf("Swagger UI available at http://localhost:%s/swagger\n", port)
	fmt.Printf("Redoc UI available at http://localhost:%s/redoc\n", port)
	fmt.Printf("OpenAPI spec available at http://localhost:%s/openapi.json\n", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}`
}


func generateReadmeContent() string {
	return `# API Documentation

This directory contains the API documentation for the application.

## Available Documentation

- [OpenAPI Specification (JSON)](openapi.json)
- [OpenAPI Specification (YAML)](openapi.yaml)
- [Swagger UI](swagger.html)
- [Redoc UI](redoc.html)

## Viewing the Documentation

You can view the documentation in several ways:

### 1. Using the Documentation Server

Run the documentation server:

` + "```bash" + `
go run serve.go [port]
` + "```" + `

The default port is 8080. Once running, you can access:

- Swagger UI: http://localhost:8080/swagger
- Redoc UI: http://localhost:8080/redoc
- OpenAPI JSON: http://localhost:8080/openapi.json

### 2. Directly Opening HTML Files

You can open the HTML files directly in your browser:

- Open ` + "`swagger.html`" + ` for Swagger UI
- Open ` + "`redoc.html`" + ` for Redoc UI

### 3. Use with External Tools

The ` + "`openapi.json`" + ` and ` + "`openapi.yaml`" + ` files can be imported into tools like Postman, Insomnia, or any OpenAPI-compatible tool.

## Updating the Documentation

The documentation is generated from the API routes and controller comments. To update it, run:

` + "```bash" + `
flux doc:generate
` + "```" + `

This will scan your application's routes and controllers to generate updated documentation.`
}

// Helper function to extract route documentation from controller files
func extractRouteCommentsFromControllers(controllersDir string, entries []os.DirEntry) []flux.RouteDoc {
	var routes []flux.RouteDoc

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}

		filePath := filepath.Join(controllersDir, entry.Name())
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Warning: Could not read controller file %s: %v\n", entry.Name(), err)
			continue
		}

		
		fileRoutes := extractRoutesFromFileContent(string(fileContent), entry.Name())
		routes = append(routes, fileRoutes...)
	}

	return routes
}

// Helper function to extract route information from file content
func extractRoutesFromFileContent(content, fileName string) []flux.RouteDoc {
	var routes []flux.RouteDoc
	lines := strings.Split(content, "\n")

	var currentRoute *flux.RouteDoc
	var collectingParams bool
	var collectingResponses bool

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		
		if strings.HasPrefix(trimmedLine, "// Route:") {
			
			if currentRoute != nil {
				routes = append(routes, *currentRoute)
			}

			currentRoute = &flux.RouteDoc{}
			collectingParams = false
			collectingResponses = false

			
			routeParts := strings.SplitN(strings.TrimPrefix(trimmedLine, "// Route:"), " ", 3)
			if len(routeParts) >= 2 {
				currentRoute.Method = strings.TrimSpace(routeParts[0])
				currentRoute.Path = strings.TrimSpace(routeParts[1])
			}
		} else if currentRoute != nil && strings.HasPrefix(trimmedLine, "// Description:") {
			
			currentRoute.Description = strings.TrimSpace(strings.TrimPrefix(trimmedLine, "// Description:"))
		} else if currentRoute != nil && strings.HasPrefix(trimmedLine, "// Param:") {
			
			collectingParams = true
			collectingResponses = false

			
			paramInfo := strings.TrimSpace(strings.TrimPrefix(trimmedLine, "// Param:"))
			if currentRoute.Params == nil {
				currentRoute.Params = []map[string]string{}
			}
			
			
			paramParts := strings.SplitN(paramInfo, " - ", 5)
			if len(paramParts) >= 4 {
				param := map[string]string{
					"name":     strings.TrimSpace(paramParts[0]),
					"in":       strings.TrimSpace(paramParts[1]),
					"type":     strings.TrimSpace(paramParts[2]),
					"required": strings.Contains(strings.TrimSpace(paramParts[3]), "required") ? "true" : "false",
				}
				
				if len(paramParts) >= 5 {
					param["description"] = strings.TrimSpace(paramParts[4])
				}
				
				currentRoute.Params = append(currentRoute.Params, param)
			}
		} else if currentRoute != nil && strings.HasPrefix(trimmedLine, "// Response:") {
			
			collectingParams = false
			collectingResponses = true

			
			responseInfo := strings.TrimSpace(strings.TrimPrefix(trimmedLine, "// Response:"))
			if currentRoute.Responses == nil {
				currentRoute.Responses = []map[string]string{}
			}
			
			
			responseParts := strings.SplitN(responseInfo, " - ", 2)
			if len(responseParts) >= 2 {
				response := map[string]string{
					"status":      strings.TrimSpace(responseParts[0]),
					"description": strings.TrimSpace(responseParts[1]),
				}
				
				currentRoute.Responses = append(currentRoute.Responses, response)
			}
		} else if strings.HasPrefix(trimmedLine, "func (") && strings.Contains(trimmedLine, ") ") {
			
			if currentRoute != nil {
				
				handlerMatch := extractHandlerName(trimmedLine)
				if handlerMatch != "" {
					currentRoute.Handler = handlerMatch
					routes = append(routes, *currentRoute)
					currentRoute = nil
				}
			}
		}
	}

	
	if currentRoute != nil && currentRoute.Method != "" && currentRoute.Path != "" {
		routes = append(routes, *currentRoute)
	}

	return routes
}

// Helper function to extract handler name from function def
func extractHandlerName(line string) string {
	
	parts := strings.Split(line, "func ")
	if len(parts) < 2 {
		return ""
	}
	
	funcParts := strings.Split(parts[1], "(")
	if len(funcParts) < 2 {
		return ""
	}
	
	receiverAndName := strings.Split(funcParts[0], " ")
	if len(receiverAndName) < 2 {
		return ""
	}
	
	return receiverAndName[1]
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
