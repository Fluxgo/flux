package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/Fluxgo/flux/pkg/flux"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/urfave/cli/v2"
)

var rootCmd = &cobra.Command{
	Use:   "flux",
	Short: "flux - The GoPowerhouse",
	Long: `flux is a modern, full-stack web framework for Go — 
designed to combine developer happiness, performance, and structure.`,
}

func init() {
	newCmd := &cobra.Command{
		Use:   "new [project-name]",
		Short: "Create a new flux project",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := createNewProject(args[0]); err != nil {
				fmt.Printf("Error creating project: %v\n", err)
				os.Exit(1)
			}
			installSuccessMessage()
		},
	}

	makeControllerCmd := &cobra.Command{
		Use:   "make:controller [name]",
		Short: "Generate a new controller",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := generateController(args[0]); err != nil {
				fmt.Printf("Error generating controller: %v\n", err)
				os.Exit(1)
			}
		},
	}

	makeModelCmd := &cobra.Command{
		Use:   "make:model [name]",
		Short: "Generate a new model",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := generateModel(args[0]); err != nil {
				fmt.Printf("Error generating model: %v\n", err)
				os.Exit(1)
			}
		},
	}

	makeMigrationCmd := &cobra.Command{
		Use:   "make:migration [name]",
		Short: "Generate a new database migration",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			modelFlag, _ := cmd.Flags().GetString("model")
			if err := generateMigration(args[0], modelFlag); err != nil {
				fmt.Printf("Error generating migration: %v\n", err)
				os.Exit(1)
			}
		},
	}
	makeMigrationCmd.Flags().StringP("model", "m", "", "Generate migration for an existing model")

	makeMicroserviceCmd := &cobra.Command{
		Use:   "make:microservice [name]",
		Short: "Generate a new microservice",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			config := &flux.MicroserviceConfig{
				Name:        args[0],
				Description: "A flux microservice",
				Port:        8080,
				WithDB:      cmd.Flag("with-db").Changed,
				WithAuth:    cmd.Flag("with-auth").Changed,
				WithCache:   cmd.Flag("with-cache").Changed,
				WithQueue:   cmd.Flag("with-queue").Changed,
			}

			if err := flux.CreateMicroserviceProject(config); err != nil {
				fmt.Printf("Error generating microservice: %v\n", err)
				os.Exit(1)
			}

			microserviceSuccessMessage(args[0])
		},
	}

	// flags for the microservice
	makeMicroserviceCmd.Flags().Bool("with-db", false, "Include database support")
	makeMicroserviceCmd.Flags().Bool("with-auth", false, "Include authentication support")
	makeMicroserviceCmd.Flags().Bool("with-cache", false, "Include cache support")
	makeMicroserviceCmd.Flags().Bool("with-queue", false, "Include queue support")

	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the development server with hot reload",
		Run: func(cmd *cobra.Command, args []string) {
			microservice, _ := cmd.Flags().GetString("microservice")
			port, _ := cmd.Flags().GetInt("port")
			startServer(microservice, port)
		},
	}
	serveCmd.Flags().StringP("microservice", "m", "", "Name of the microservice to run (if in a microservices project)")
	serveCmd.Flags().IntP("port", "p", 3000, "Port to run the server on")

	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(makeControllerCmd)
	rootCmd.AddCommand(makeModelCmd)
	rootCmd.AddCommand(makeMigrationCmd)
	rootCmd.AddCommand(makeMicroserviceCmd)
	rootCmd.AddCommand(serveCmd)
}

func detectProjectStructure() (isMicroservice bool, microserviceNames []string) {
	cmdDir := "cmd"
	if _, err := os.Stat(cmdDir); err == nil {
		entries, err := os.ReadDir(cmdDir)
		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() && entry.Name() != "flux" {
					microservicePath := filepath.Join(cmdDir, entry.Name(), "main.go")
					if _, err := os.Stat(microservicePath); err == nil {
						microserviceNames = append(microserviceNames, entry.Name())
						isMicroservice = true
					}
				}
			}
		}
	}

	return
}

func startServer(microserviceName string, port int) {
	isMicroserviceProject, microserviceNames := detectProjectStructure()

	// Handle microservice mode
	if isMicroserviceProject && microserviceName == "" && len(microserviceNames) > 0 {
		fmt.Println("🔍 Detected microservice project with the following services:")
		for i, name := range microserviceNames {
			fmt.Printf("  %d. %s\n", i+1, name)
		}
		fmt.Println("\nPlease specify which microservice to run using:")
		fmt.Println("  flux serve -m <service-name>")
		fmt.Println("\nOr run each service in a separate terminal.")
		os.Exit(0)
	}

	if microserviceName != "" {
		startMicroservice(microserviceName, port)
		return
	}

	startMonolith(port)
}

func startMicroservice(name string, port int) {
	microservicePath := filepath.Join("cmd", name, "main.go")
	if _, err := os.Stat(microservicePath); os.IsNotExist(err) {
		fmt.Printf("Error: Microservice '%s' not found at path %s\n", name, microservicePath)
		os.Exit(1)
	}

	cyan := color.New(color.FgCyan).SprintFunc()
	fmt.Printf(" %s Starting microservice: %s on port %d\n", cyan("[flux]"), name, port)
	fmt.Printf(" Using configuration from: %s\n", filepath.Join("config", "config.yaml"))
	fmt.Println(" Hot reload is enabled - your changes will apply automatically.")

	os.Setenv("flux_HOT_RELOAD", "true")
	if port != 0 {
		os.Setenv("PORT", fmt.Sprintf("%d", port))
	}

	app, err := flux.New(&flux.Config{
		Name:        name,
		Version:     "1.0.0",
		Description: "A flux microservice",
		Server: flux.ServerConfig{
			Host:     "localhost",
			Port:     port,
			BasePath: "/api",
		},
	})
	if err != nil {
		fmt.Printf("Error creating application: %v\n", err)
		os.Exit(1)
	}

	reloader, err := flux.NewMicroserviceHotReloader(app, name, filepath.Join("cmd", name, "main.go"))
	if err != nil {
		fmt.Printf("Error creating hot reloader: %v\n", err)
		os.Exit(1)
	}

	if err := reloader.Start(); err != nil {
		fmt.Printf("Error starting hot reloader: %v\n", err)
		os.Exit(1)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	if err := reloader.Stop(); err != nil {
		fmt.Printf("Error stopping hot reloader: %v\n", err)
		os.Exit(1)
	}
}

func startMonolith(port int) {
	cyan := color.New(color.FgCyan).SprintFunc()
	fmt.Printf(" %s Starting monolith application on port %d\n", cyan("[flux]"), port)
	fmt.Printf(" Using configuration from: %s\n", filepath.Join("config", "flux.yaml"))
	fmt.Println(" Hot reload is enabled - your changes will apply automatically.")

	app, err := flux.New(&flux.Config{
		Name:        "flux App",
		Version:     "1.0.0",
		Description: "A flux application",
		Server: flux.ServerConfig{
			Host:     "localhost",
			Port:     port,
			BasePath: "/",
		},
	})
	if err != nil {
		fmt.Printf("Error creating application: %v\n", err)
		os.Exit(1)
	}

	reloader, err := flux.NewHotReloader(app)
	if err != nil {
		fmt.Printf("Error creating hot reloader: %v\n", err)
		os.Exit(1)
	}

	if err := reloader.Start(); err != nil {
		fmt.Printf("Error starting hot reloader: %v\n", err)
		os.Exit(1)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	if err := reloader.Stop(); err != nil {
		fmt.Printf("Error stopping hot reloader: %v\n", err)
		os.Exit(1)
	}
}

func installSuccessMessage() {
	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()

	fmt.Println("\n " + bold("Welcome to flux: The GoPowerhouse Web Framework!") + " ")
	fmt.Println("🔨 Created with passion by " + green("Yemi Ogunrinde"))
	fmt.Println(cyan("\nLet's build something amazing together! 🛠️"))
	fmt.Println(yellow("Happy Coding! 👨‍💻"))
	fmt.Println("☕ Like it? " + bold("Buy me a coffee") + " at: https://buymeacoffee.com/BisiOlaYemi\n")
}

func microserviceSuccessMessage(name string) {
	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()

	fmt.Println("\n " + bold("Microservice Created Successfully!") + " ")
	fmt.Println("⚙️ " + green(name) + " microservice is ready for development")
	fmt.Println(cyan("\nNext steps:"))
	fmt.Printf(" 1. cd %s\n", name)
	fmt.Println(" 2. go mod tidy")
	fmt.Println(" 3. Run with hot reload: flux serve -m " + name)
	fmt.Println(" 4. Or standard run: go run cmd/" + name + "/main.go")
	fmt.Println(yellow("\nHappy Crafting flux Microservices! 👨‍💻\n"))
	fmt.Println("☕ Like it? " + bold("Buy me a coffee") + " at: https://buymeacoffee.com/BisiOlaYemi\n")
}

func generateMigration(name string, modelName string) error {
	// this is currently a placeholder we'll have to improve it
	fmt.Printf("Generated migration: %s\n", name)
	if modelName != "" {
		fmt.Printf("Linked to model: %s\n", modelName)
	}
	return nil
}


func microserviceCommand(c *cli.Context) error {
	name := c.String("name")
	if name == "" {
		return errors.New("microservice name is required")
	}

	// Create new microservice configuration
	config := &flux.MicroserviceConfig{
		Name:        name,
		Port:        c.Int("port"),
		Description: fmt.Sprintf("Microservice for %s", name),
		WithDB:      c.Bool("with-db"),
		WithCache:   c.Bool("with-cache"),
		WithQueue:   c.Bool("with-queue"),
		WithAuth:    c.Bool("with-auth"),
	}

	
	err := flux.CreateMicroserviceProject(config)
	if err != nil {
		return err
	}

	fmt.Printf("Microservice '%s' created successfully\n", name)
	return nil
}


func serveCommand(c *cli.Context) error {
	port := c.Int("port")
	host := c.String("host")
	microserviceName := c.String("microservice") 

	if microserviceName != "" {
		
		fmt.Printf(" flux Starting microservice: %s on port %d\n", microserviceName, port)
		fmt.Println(" Using configuration from: config\\config.yaml")
		
		
		useHotReload := !c.Bool("no-reload")
		if useHotReload {
			fmt.Println(" Hot reload is enabled - your changes will apply automatically.")
			os.Setenv("flux_HOT_RELOAD", "true")
		}
		
		
		cmdDir := filepath.Join("cmd", microserviceName)
		if _, err := os.Stat(cmdDir); os.IsNotExist(err) {
			return fmt.Errorf("microservice directory %s does not exist", cmdDir)
		}

		mainFile := filepath.Join(cmdDir, "main.go")
		if _, err := os.Stat(mainFile); os.IsNotExist(err) {
			return fmt.Errorf("microservice main file %s does not exist", mainFile)
		}

		if useHotReload {
			// For hot reload, we use go run
			cmd := exec.Command("go", "run", mainFile)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		} else {
			// For regular run, we build and then execute
			buildCmd := exec.Command("go", "build", "-o", microserviceName+".exe", mainFile)
			buildCmd.Stdout = os.Stdout
			buildCmd.Stderr = os.Stderr
			if err := buildCmd.Run(); err != nil {
				return fmt.Errorf("failed to build microservice: %w", err)
			}

			runCmd := exec.Command("./"+microserviceName+".exe")
			runCmd.Stdout = os.Stdout
			runCmd.Stderr = os.Stderr
			return runCmd.Run()
		}
	}

	
	fmt.Printf(" [flux] Starting server on %s:%d\n", host, port)
	
	useHotReload := !c.Bool("no-reload")
	if useHotReload {
		fmt.Println(" Hot reload is enabled - your changes will apply automatically.")
	}
	
	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
