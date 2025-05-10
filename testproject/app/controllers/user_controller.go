package controllers

import (
	"github.com/Fluxgo/flux/pkg/flux"
)

// UserController handles requests related to User
type UserController struct {
	flux.Controller
}

// HandleGetUsers handles getting all Users
// Route: GET /user
// Description: Get all users
// Response: 200 - []User
func (c *UserController) HandleGetUsers(ctx *flux.Context) error {
	var items []interface{}
	if err := c.App().DB().Find(&items).Error; err != nil {
		return ctx.Status(500).JSON(map[string]string{"error": err.Error()})
	}
	return ctx.JSON(items)
}

// HandleGetUserById handles getting a User by ID
// Route: GET /user/:id
// Description: Get a specific user by ID
// Param: id - path - int - required - User ID
// Response: 200 - User
// Response: 404 - Error message when user not found
func (c *UserController) HandleGetUserById(ctx *flux.Context) error {
	id := ctx.Param("id")
	var item interface{}
	if err := c.App().DB().First(&item, id).Error; err != nil {
		return ctx.Status(404).JSON(map[string]string{"error": "Not found"})
	}
	return ctx.JSON(item)
}

// HandleCreateUser handles creating a new User
// Route: POST /user
// Description: Create a new user
// Body: CreateUserRequest
// Response: 201 - User
// Response: 400 - Error message when request body is invalid
// Response: 500 - Error message when database operation fails
func (c *UserController) HandleCreateUser(ctx *flux.Context) error {
	var req CreateUserRequest
	
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

// HandleUpdateUser handles updating a User
// Route: PUT /user/:id
// Description: Update a specific user by ID
// Param: id - path - int - required - User ID
// Body: UpdateUserRequest
// Response: 200 - Updated User
// Response: 400 - Error message when request body is invalid
// Response: 404 - Error message when user not found
// Response: 500 - Error message when database operation fails
func (c *UserController) HandleUpdateUser(ctx *flux.Context) error {
	id := ctx.Param("id")
	
	var req UpdateUserRequest
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

// HandleDeleteUser handles deleting a User
// Route: DELETE /user/:id
// Description: Delete a specific user by ID
// Param: id - path - int - required - User ID
// Response: 204 - No content on successful deletion
// Response: 404 - Error message when user not found
// Response: 500 - Error message when database operation fails
func (c *UserController) HandleDeleteUser(ctx *flux.Context) error {
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
