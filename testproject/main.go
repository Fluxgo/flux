package main

import (
	"fmt"
	"log"

	"github.com/Fluxgo/flux/pkg/flux"
	// Import your controllers and models as needed
	"testproject/app/controllers"
)

func main() {
	//New flux application
	app, err := flux.New(&flux.Config{
		Name:        "testproject",
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

	// Register controllers directly
	app.RegisterController(&controllers.UserController{})
	
	// Start the server
	fmt.Printf("Server starting on http://localhost:3000\n")
	if err := app.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
