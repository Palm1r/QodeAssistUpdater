package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

type Config struct {
	QtCreatorPath string `yaml:"qtcreator_path"`
	PluginPath    string `yaml:"plugin_path"`
}

const (
	GithubRepo = "Palm1r/QodeAssist"
	PluginName = "QodeAssist"
)

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

func (c *Config) Validate() error {
	if c.QtCreatorPath == "" {
		return fmt.Errorf("qtcreator_path is required")
	}

	if c.PluginPath == "" {
		return fmt.Errorf("plugin_path is required")
	}

	qtcInfo, err := os.Stat(c.QtCreatorPath)
	if err != nil {
		return fmt.Errorf("qtcreator_path does not exist or is not accessible: %w", err)
	}

	if qtcInfo.Mode()&0111 == 0 && filepath.Ext(c.QtCreatorPath) != ".exe" {
		if runtime.GOOS != "windows" {
			return fmt.Errorf("qtcreator_path is not executable")
		}
	}

	pluginDir := filepath.Dir(c.PluginPath)
	if pluginDir == "" {
		pluginDir = "."
	}
	
	pluginDirInfo, err := os.Stat(pluginDir)
	if err != nil {
		if err := os.MkdirAll(pluginDir, 0755); err != nil {
			return fmt.Errorf("plugin_path parent directory cannot be created: %w", err)
		}
	} else if !pluginDirInfo.IsDir() {
		return fmt.Errorf("plugin_path parent is not a directory")
	}

	return nil
}
