package main

import (
	"fmt"
	"os"
	"strings"
	"path/filepath"
)


func generateAdvancedController(name string) error {
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

	
	mainPath := "main.go"
	if _, err := os.Stat(mainPath); err == nil {
		mainContent, err := os.ReadFile(mainPath)
		if err != nil {
			return fmt.Errorf("failed to read main.go: %w", err)
		}

		mainContentStr := string(mainContent)
		
		
		importSection := "import ("
		controllersImport := "\"" + getCurrentModuleName() + "/app/controllers\""
		
		
		if !strings.Contains(mainContentStr, controllersImport) && strings.Contains(mainContentStr, importSection) {
			importEnd := strings.Index(mainContentStr, ")")
			if importEnd > 0 {
				mainContentStr = mainContentStr[:importEnd] + "\t" + controllersImport + "\n" + mainContentStr[importEnd:]
			}
		}
		
		
		startServerComment := "// Start the server"
		controllerRegistration := "\t// Register " + strings.TrimSuffix(name, "Controller") + " controller\n\tapp.RegisterController(&controllers." + name + "{})\n\n\t"
		
		if strings.Contains(mainContentStr, startServerComment) && !strings.Contains(mainContentStr, "app.RegisterController(&controllers."+name+"{}") {
			mainContentStr = strings.Replace(mainContentStr, startServerComment, controllerRegistration+startServerComment, 1)
		}
		
		
		if err := os.WriteFile(mainPath, []byte(mainContentStr), 0644); err != nil {
			return fmt.Errorf("failed to update main.go: %w", err)
		}
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

	fmt.Printf("Generated controller: %s\n", name)
	fmt.Printf("Generated model: %s\n", strings.TrimSuffix(name, "Controller"))
	fmt.Printf("Updated main.go to register the controller\n")

	return nil
}
