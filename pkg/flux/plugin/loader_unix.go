//go:build linux || darwin || freebsd

package plugin

import (
	"fmt"
	stdplugin "plugin"
)

func (m *Manager) loadSharedObject(path string, config map[string]Config) error {
	sharedObject, err := stdplugin.Open(path)
	if err != nil {
		return fmt.Errorf("failed to load plugin %s: %w", path, err)
	}

	newSymbol, err := sharedObject.Lookup("New")
	if err != nil {
		return fmt.Errorf("plugin %s does not export New function: %w", path, err)
	}
	newPlugin, ok := newSymbol.(func() (Plugin, error))
	if !ok {
		return fmt.Errorf("plugin %s New function has invalid signature", path)
	}

	loaded, err := newPlugin()
	if err != nil {
		return fmt.Errorf("failed to create plugin instance %s: %w", path, err)
	}
	if !config[loaded.Name()].Enabled {
		return nil
	}
	if err := loaded.Init(m.app); err != nil {
		return fmt.Errorf("failed to initialize plugin %s: %w", path, err)
	}
	m.plugins[loaded.Name()] = loaded
	return nil
}
