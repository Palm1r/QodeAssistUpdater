package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

type Config struct {
	QtCreatorPath string `yaml:"qtcreator_path"`
	PluginPath    string `yaml:"plugin_path"`

	mu             sync.RWMutex
	qtcInfo        *QtCreatorInfo
	versionFetched bool
}

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

func expandQtcPath(path string) []string {
	pattern := strings.ReplaceAll(path, "{qtc_version}", "*")
	if !strings.ContainsAny(pattern, "*?[") {
		return []string{pattern}
	}
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil
	}
	return matches
}

func pickNewestQtCreator(candidates []string) string {
	var bestPath string
	var bestVersion *Version
	for _, path := range candidates {
		info, err := os.Stat(path)
		if err != nil || !info.IsDir() {
			continue
		}
		version, err := GetQtCreatorVersion(path)
		if err != nil {
			continue
		}
		if version.IsNewer(bestVersion) {
			bestVersion = version
			bestPath = path
		}
	}
	return bestPath
}

func findQtCreatorOnPath() string {
	for _, name := range []string{"qtcreator", "qtcreator.exe"} {
		execPath, err := exec.LookPath(name)
		if err != nil {
			continue
		}

		if resolved, rerr := filepath.EvalSymlinks(execPath); rerr == nil {
			execPath = resolved
		}

		// The binary lives at <root>/bin/<name> or <root>/<name>.
		for _, root := range []string{
			filepath.Dir(filepath.Dir(execPath)),
			filepath.Dir(execPath),
		} {
			if _, err := GetQtCreatorVersion(root); err == nil {
				return root
			}
		}
	}

	return ""
}

func findQtCreator() (string, error) {
	if root := findQtCreatorOnPath(); root != "" {
		return root, nil
	}

	platformConfig, err := GetPlatformConfig()
	if err != nil {
		return "", err
	}

	for _, candidate := range platformConfig.QtCreatorPaths {
		if best := pickNewestQtCreator(expandQtcPath(candidate)); best != "" {
			return best, nil
		}
	}

	return "", fmt.Errorf("Qt Creator not found in any default location")
}

func NewConfigFromArgs(qtcPath, pluginPath string) (*Config, error) {
	if qtcPath == "" {
		detected, err := findQtCreator()
		if err != nil {
			return nil, fmt.Errorf("qtcreator path not provided and auto-detection failed: %w", err)
		}
		qtcPath = detected
	}

	config := &Config{QtCreatorPath: qtcPath, PluginPath: pluginPath}

	if pluginPath == "" {
		if err := config.validateQtCreatorPath(); err != nil {
			return nil, err
		}
		detected, err := findPlugin(config.QtCreatorPath)
		if err != nil {
			return nil, fmt.Errorf("plugin path not provided and auto-detection failed: %w", err)
		}
		config.PluginPath = detected
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

func findPlugin(qtCreatorPath string) (string, error) {
	version, err := GetQtCreatorVersion(qtCreatorPath)
	if err != nil {
		return "", fmt.Errorf("failed to get Qt Creator version: %w", err)
	}

	searchDirs, err := GetPluginSearchDirs(qtCreatorPath, version)
	if err != nil {
		return "", err
	}

	platformConfig, err := GetPlatformConfig()
	if err != nil {
		return "", err
	}

	versionStr := version.String()
	targetName := platformConfig.PluginFileName

	for _, searchDir := range searchDirs {
		if !PathExists(searchDir) {
			continue
		}

		foundPath, err := FindFileInDirectory(searchDir, targetName, MaxSearchDepth, nil)
		if err == nil && foundPath != "" {
			configPath := foundPath

			if strings.HasSuffix(foundPath, "Qt Creator.app") {
				configPath = filepath.Dir(foundPath)
			}

			configPath = strings.Replace(configPath, versionStr, "{qtc_version}", 1)

			return filepath.ToSlash(configPath), nil
		}
	}

	if len(platformConfig.PluginPaths) > 0 {
		defaultPath := platformConfig.PluginPaths[0]
		return defaultPath, nil
	}

	return "", fmt.Errorf("no default plugin path available for this system")
}

func CreateDefaultConfig(path string) error {
	PrintStep("Creating configuration...")

	var qtCreatorPath, pluginPath string
	var foundQtCreator, foundPlugin bool

	qtcPath, err := findQtCreator()
	if err != nil {
		foundQtCreator = false

		platformConfig, cfgErr := GetPlatformConfig()
		if cfgErr != nil {
			return fmt.Errorf("failed to get platform config: %w", cfgErr)
		}
		if len(platformConfig.QtCreatorPaths) > 0 {
			qtCreatorPath = platformConfig.QtCreatorPaths[0]
		}
	} else {
		qtCreatorPath = qtcPath
		foundQtCreator = true

		plugPath, err := findPlugin(qtCreatorPath)
		if err != nil {
			foundPlugin = false

			platformConfig, cfgErr := GetPlatformConfig()
			if cfgErr != nil {
				return fmt.Errorf("failed to get platform config: %w", cfgErr)
			}
			if len(platformConfig.PluginPaths) > 0 {
				pluginPath = platformConfig.PluginPaths[0]
			}
		} else {
			pluginPath = plugPath
			foundPlugin = true
		}
	}

	if foundQtCreator {
		if version, verr := GetQtCreatorVersion(qtCreatorPath); verr == nil {
			qtCreatorPath = strings.Replace(qtCreatorPath, version.String(), "{qtc_version}", 1)
		}
	}

	qtCreatorPath = filepath.ToSlash(qtCreatorPath)
	pluginPath = filepath.ToSlash(pluginPath)

	configContent := fmt.Sprintf(`qtcreator_path: "%s"
plugin_path: "%s"
`, qtCreatorPath, pluginPath)

	if err := os.WriteFile(path, []byte(configContent), DefaultFilePermissions); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	if !foundQtCreator || !foundPlugin {
		PrintWarning("Could not auto-detect all paths")

		if !foundQtCreator {
			PrintError("Qt Creator not found in default locations")
			fmt.Println()
		}

		if !foundPlugin {
			PrintWarning("Using default plugin path")
			fmt.Println()
		}

		platformConfig, _ := GetPlatformConfig()
		if platformConfig != nil {
			PrintVerbose(Gray("Please verify the configuration file and update paths if needed:"))
			PrintVerbose(Gray(fmt.Sprintf("  Config: %s", path)))
			fmt.Println()
			PrintVerbose(Gray("Default Qt Creator paths:"))
			for _, p := range platformConfig.QtCreatorPaths {
				PrintVerbose(Gray(fmt.Sprintf("  • %s", p)))
			}
		}

		return fmt.Errorf("configuration incomplete - please edit the config file with correct paths")
	}

	PrintSuccess("Configuration created successfully")
	PrintField("Config file", path)
	fmt.Println()

	return nil
}

func (c *Config) expandVariables(path string) (string, error) {
	c.mu.RLock()
	qtcInfo := c.qtcInfo
	c.mu.RUnlock()

	if qtcInfo == nil || qtcInfo.Version == nil {
		return path, nil
	}

	path = strings.ReplaceAll(path, "{qtc_version}", qtcInfo.Version.String())
	return path, nil
}

func (c *Config) GetQtCreatorInfo() (*QtCreatorInfo, error) {
	c.mu.RLock()
	if c.versionFetched {
		info := c.qtcInfo
		c.mu.RUnlock()
		return info, nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.versionFetched {
		return c.qtcInfo, nil
	}

	info, err := GetQtCreatorInfo(c.QtCreatorPath)
	if err != nil {
		return nil, err
	}

	c.qtcInfo = info
	c.versionFetched = true
	return info, nil
}

func (c *Config) GetQtCreatorVersion() (*Version, error) {
	info, err := c.GetQtCreatorInfo()
	if err != nil {
		return nil, err
	}
	return info.Version, nil
}

func (c *Config) GetPluginPath() (string, error) {
	if !c.versionFetched {
		if _, err := c.GetQtCreatorVersion(); err != nil {
			return "", fmt.Errorf("failed to get Qt Creator version: %w", err)
		}
	}
	return c.expandVariables(c.PluginPath)
}

func (c *Config) validateQtCreatorPath() error {
	if c.QtCreatorPath == "" {
		return fmt.Errorf("qtcreator_path is required")
	}

	expandedPath, err := ExpandPath(c.QtCreatorPath)
	if err != nil {
		return fmt.Errorf("failed to expand qtcreator_path: %w", err)
	}

	if strings.Contains(expandedPath, "{qtc_version}") {
		resolved := pickNewestQtCreator(expandQtcPath(expandedPath))
		if resolved == "" {
			return fmt.Errorf("qtcreator_path %q did not match any Qt Creator installation", c.QtCreatorPath)
		}
		expandedPath = resolved
	}

	c.QtCreatorPath = expandedPath

	qtcInfo, err := os.Stat(c.QtCreatorPath)
	if err != nil {
		return fmt.Errorf("qtcreator_path does not exist or is not accessible: %w", err)
	}

	if !qtcInfo.IsDir() {
		return fmt.Errorf("qtcreator_path should be a directory")
	}

	return nil
}

func (c *Config) validatePluginPath() error {
	if c.PluginPath == "" {
		return fmt.Errorf("plugin_path is required")
	}

	expandedPath, err := ExpandPath(c.PluginPath)
	if err != nil {
		return fmt.Errorf("failed to expand plugin_path: %w", err)
	}

	if filepath.IsAbs(expandedPath) || strings.Contains(expandedPath, "{qtc_version}") || strings.HasPrefix(c.PluginPath, "~") {
		return nil
	}

	return fmt.Errorf("plugin_path should be an absolute path or contain {qtc_version} variable")
}

func (c *Config) Validate() error {
	if err := c.validateQtCreatorPath(); err != nil {
		return err
	}

	if err := c.validatePluginPath(); err != nil {
		return err
	}

	return nil
}
