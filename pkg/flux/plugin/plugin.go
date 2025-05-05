package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"sync"
)

type AppInterface interface {
	GetConfig() interface{}
	GetDB() interface{}
	GetAuth() interface{}
	GetQueue() interface{}
	GetMailer() interface{}
	RegisterController(controller interface{})
}

type Plugin interface {
	Name() string
	Description() string
	Version() string
	Init(app AppInterface) error
	Shutdown() error
}

type Manager struct {
	plugins     map[string]Plugin
	app         AppInterface
	mu          sync.RWMutex
	pluginDir   string
	configPath  string
}

type Config struct {
	Enabled  bool                   `json:"enabled"`
	Settings map[string]interface{} `json:"settings"`
}

func NewManager(app AppInterface, pluginDir string) *Manager {
	return &Manager{
		plugins:   make(map[string]Plugin),
		app:       app,
		pluginDir: pluginDir,
	}
}

func (m *Manager) LoadPlugins() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := os.MkdirAll(m.pluginDir, 0755); err != nil {
		return fmt.Errorf("failed to create plugin directory: %w", err)
	}

	config, err := m.loadConfig()
	if err != nil {
		return err
	}

	return filepath.Walk(m.pluginDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || filepath.Ext(path) != ".so" {
			return nil
		}

		p, err := plugin.Open(path)
		if err != nil {
			return fmt.Errorf("failed to load plugin %s: %w", path, err)
		}

		newFunc, err := p.Lookup("New")
		if err != nil {
			return fmt.Errorf("plugin %s does not export New function: %w", path, err)
		}

		pluginFunc, ok := newFunc.(func() (Plugin, error))
		if !ok {
			return fmt.Errorf("plugin %s New function has invalid signature", path)
		}

		plugin, err := pluginFunc()
		if err != nil {
			return fmt.Errorf("failed to create plugin instance %s: %w", path, err)
		}

		if !config[plugin.Name()].Enabled {
			return nil
		}

		if err := plugin.Init(m.app); err != nil {
			return fmt.Errorf("failed to initialize plugin %s: %w", path, err)
		}

		m.plugins[plugin.Name()] = plugin

		return nil
	})
}


func (m *Manager) UnloadPlugins() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, plugin := range m.plugins {
		if err := plugin.Shutdown(); err != nil {
			return fmt.Errorf("failed to shutdown plugin %s: %w", name, err)
		}
		delete(m.plugins, name)
	}

	return nil
}


func (m *Manager) GetPlugin(name string) (Plugin, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugin, ok := m.plugins[name]
	return plugin, ok
}


func (m *Manager) loadConfig() (map[string]Config, error) {
	configPath := filepath.Join(m.pluginDir, "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]Config), nil
		}
		return nil, err
	}

	var config map[string]Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return config, nil
}


func (m *Manager) saveConfig(config map[string]Config) error {
	configPath := filepath.Join(m.pluginDir, "config.json")
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
} 
