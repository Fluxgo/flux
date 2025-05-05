package main

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/Fluxgo/flux/pkg/flux"
	"github.com/stretchr/testify/assert"
)

func TestHelloEndpoint(t *testing.T) {
	app, err := flux.New(&flux.Config{
		Name:        "Hello flux Test",
		Version:     "1.0.0",
		Description: "Test instance of Hello World",
		Server: flux.ServerConfig{
			Host:     "localhost",
			Port:     3000,
			BasePath: "/",
		},
	})
	assert.NoError(t, err)

	
	app.RegisterController(&HelloController{})

	
	req := httptest.NewRequest("GET", "/hello", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	assert.NoError(t, err)
	assert.Equal(t, "Hello from flux!", response["message"])
} 
