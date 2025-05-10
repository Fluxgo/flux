package controllers

// CreateUserRequest represents the request body for creating a user
type CreateUserRequest struct {
	Name string `json:"name" validate:"required"`
	// Add more fields as needed for creation
}

// UpdateUserRequest represents the request body for updating a user
type UpdateUserRequest struct {
	Name string `json:"name" validate:"required"`
	// Add more fields as needed for updates
}
