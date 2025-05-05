package flux

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)


type AppMode string

const (
	
	MonolithMode AppMode = "monolith"
	
	MicroserviceMode AppMode = "microservice"
)

type HotReloader struct {
	app           *Application
	watcher       *fsnotify.Watcher
	cmd           *exec.Cmd
	done          chan bool
	debounce      *time.Timer
	mode          AppMode
	microservice  string
	entrypoint    string
	projectRoot   string
	buildCommands []string
	runCommands   []string
}


func NewHotReloader(app *Application) (*HotReloader, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	return &HotReloader{
		app:         app,
		watcher:     watcher,
		done:        make(chan bool),
		mode:        MonolithMode,
		projectRoot: ".",
		entrypoint:  ".",
		buildCommands: []string{
			"go", "build", "-o", getTempBinaryName(), ".",
		},
		runCommands: []string{
			"go", "run", ".",
		},
	}, nil
}


func NewMicroserviceHotReloader(app *Application, microserviceName, entrypointPath string) (*HotReloader, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	
	if entrypointPath == "" {
		entrypointPath = filepath.Join("cmd", microserviceName, "main.go")
	}

	entryDir := filepath.Dir(entrypointPath)

	return &HotReloader{
		app:          app,
		watcher:      watcher,
		done:         make(chan bool),
		mode:         MicroserviceMode,
		microservice: microserviceName,
		projectRoot:  ".",
		entrypoint:   entryDir,
		buildCommands: []string{
			"go", "build", "-o", getTempBinaryName(), entrypointPath,
		},
		runCommands: []string{
			"go", "run", entrypointPath,
		},
	}, nil
}

// getTempBinaryName returns an appropriate temporary binary name based on OS
func getTempBinaryName() string {
	if runtime.GOOS == "windows" {
		return "tmp_flux_app.exe"
	}
	return "tmp_flux_app"
}

func (h *HotReloader) Start() error {
	if err := h.startApp(); err != nil {
		return err
	}

	
	if err := h.setupWatcher(); err != nil {
		return fmt.Errorf("failed to setup file watcher: %w", err)
	}

	go h.watch()

	return nil
}

func (h *HotReloader) Stop() error {
	close(h.done)
	if h.cmd != nil && h.cmd.Process != nil {
		
		if runtime.GOOS == "windows" {
			h.cmd.Process.Signal(os.Interrupt)
			
			time.Sleep(100 * time.Millisecond)
		}
		if err := h.cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process: %w", err)
		}
	}
	return h.watcher.Close()
}


func (h *HotReloader) setupWatcher() error {
	var dirsToWatch []string

	if h.mode == MicroserviceMode {
		
		dirsToWatch = []string{
			h.projectRoot,
			filepath.Join(h.projectRoot, "api"),
			filepath.Join(h.projectRoot, "cmd"),
			filepath.Join(h.projectRoot, "internal"),
			filepath.Join(h.projectRoot, "pkg"),
		}
		
		
		if h.microservice != "" {
			dirsToWatch = append(dirsToWatch, 
				filepath.Join(h.projectRoot, "cmd", h.microservice),
				filepath.Join(h.projectRoot, "api", h.microservice),
				filepath.Join(h.projectRoot, "internal", h.microservice),
			)
		}
	} else {
		
		dirsToWatch = []string{
			h.projectRoot,
			filepath.Join(h.projectRoot, "cmd"),
			filepath.Join(h.projectRoot, "pkg"),
			filepath.Join(h.projectRoot, "control"),
			filepath.Join(h.projectRoot, "plugins"),
			filepath.Join(h.projectRoot, "app"),
		}
	}

	for _, dir := range dirsToWatch {
		
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil 
			}

			// Skip vendor, .git and other non-essential directories
			if info.IsDir() && (strings.HasPrefix(info.Name(), ".") || 
				info.Name() == "vendor" || 
				info.Name() == "node_modules" ||
				info.Name() == "tmp") {
				return filepath.SkipDir
			}

			
			if info.IsDir() {
				return h.watcher.Add(path)
			}

			return nil
		}); err != nil {
			fmt.Printf("Warning: Error walking directory %s: %v\n", dir, err)
		}
	}

	return nil
}

func (h *HotReloader) startApp() error {
	if h.cmd != nil && h.cmd.Process != nil {
		
		if runtime.GOOS == "windows" {
			h.cmd.Process.Signal(os.Interrupt)
			
			time.Sleep(100 * time.Millisecond)
		}

		if err := h.cmd.Process.Kill(); err != nil {
			fmt.Printf("Warning: failed to kill existing process: %v\n", err)
			
		}

		
		h.cmd.Wait()
	}

	
	buildCmd := exec.Command(h.buildCommands[0], h.buildCommands[1:]...)
	buildCmd.Dir = h.projectRoot 
	buildCmd.Stderr = os.Stderr

	if err := buildCmd.Run(); err != nil {
		fmt.Printf("Build failed, not reloading: %v\n", err)
		return nil 
	}

	
	defer os.Remove(filepath.Join(h.projectRoot, getTempBinaryName()))

	
	h.cmd = exec.Command(h.runCommands[0], h.runCommands[1:]...)
	h.cmd.Dir = h.projectRoot 
	h.cmd.Stdout = os.Stdout
	h.cmd.Stderr = os.Stderr

	
	modeString := "monolith"
	if h.mode == MicroserviceMode {
		modeString = fmt.Sprintf("microservice (%s)", h.microservice)
	}
	
	commandString := strings.Join(h.runCommands, " ")
	fmt.Printf(" flux: Starting %s with command: %s\n", modeString, commandString)

	if err := h.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}

	return nil
}

func (h *HotReloader) watch() {
	for {
		select {
		case event, ok := <-h.watcher.Events:
			if !ok {
				return
			}

			
			if !strings.HasSuffix(event.Name, ".go") || 
			   strings.HasSuffix(event.Name, ".tmp") || 
			   strings.HasSuffix(event.Name, "_test.go") {
				continue
			}

			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) != 0 {
				
				if h.debounce != nil {
					h.debounce.Stop()
				}

				h.debounce = time.AfterFunc(500*time.Millisecond, func() {
					fmt.Printf(" flux: Changes detected in %s, reloading application...\n", filepath.Base(event.Name))
					if err := h.startApp(); err != nil {
						fmt.Printf(" Error restarting application: %v\n", err)
					}
				})
			}

		case err, ok := <-h.watcher.Errors:
			if !ok {
				return
			}

			fmt.Printf("Error watching files: %v\n", err)

		case <-h.done:
			return
		}
	}
}


func (h *HotReloader) SetCustomBuildCommand(cmd ...string) {
	h.buildCommands = cmd
}


func (h *HotReloader) SetCustomRunCommand(cmd ...string) {
	h.runCommands = cmd
}


func (h *HotReloader) SetProjectRoot(path string) {
	h.projectRoot = path
}
