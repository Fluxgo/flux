package main

import (
	"fmt"
	"log"

	"github.com/Fluxgo/flux/pkg/flux"
)

type HelloController struct {
	flux.Controller
}

// HandleGetHello handles GET /hello requests
// @route GET /hello
// @desc Get a hello message
// @response 200 object
func (c *HelloController) HandleGetHello(ctx *flux.Context) error {
	return ctx.JSON(flux.H{
		"message": "Hello from flux!",
	})
}

func main() {
	app, err := flux.New(&flux.Config{
		Name:        "Hello flux",
		Version:     "1.0.0",
		Description: "A simple Hello World example",
		Server: flux.ServerConfig{
			Host:     "localhost",
			Port:     3000,
			BasePath: "/",
		},
	})
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	app.RegisterController(&HelloController{})

	fmt.Println("Server starting on http://localhost:3000")
	if err := app.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
} 
