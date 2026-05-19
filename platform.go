package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type PlatformConfig struct {
	QtCreatorPaths  []string
	PluginPaths     []string
	ExecutableNames []string
	PluginFileName  string
	LibExtension    string
}

func GetPlatformConfig() (*PlatformConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	switch runtime.GOOS {
	case "darwin":
		return &PlatformConfig{
			QtCreatorPaths: []string{
				filepath.Join(homeDir, "Qt", "Qt Creator.app"),
				"/Applications/Qt Creator.app",
				filepath.Join(homeDir, "Applications", "Qt Creator.app"),
			},
			PluginPaths: []string{
				filepath.Join(homeDir, "Library", "Application Support", "QtProject", "Qt Creator", "plugins", "{qtc_version}", PluginAuthor),
			},
			ExecutableNames: []string{
				filepath.Join("Contents", "MacOS", "Qt Creator"),
				filepath.Join("Contents", "MacOS", "qtcreator"),
			},
			PluginFileName: "Qt Creator.app",
			LibExtension:   ".dylib",
		}, nil

	case "linux":
		dataHome, configHome := linuxXDGDirs(homeDir)
		return &PlatformConfig{
			QtCreatorPaths: []string{
				filepath.Join(homeDir, "Qt", "Tools", "QtCreator"),
				filepath.Join(homeDir, "qtcreator-{qtc_version}"),
				"/usr",
				"/usr/local",
				"/opt/Qt/Tools/QtCreator",
			},
			PluginPaths: []string{
				filepath.Join(dataHome, "data", "QtProject", "qtcreator", "plugins", "{qtc_version}", PluginAuthor),
				filepath.Join(configHome, "QtProject", "qtcreator", "plugins", "{qtc_version}", PluginAuthor),
			},
			ExecutableNames: []string{
				filepath.Join("bin", "qtcreator"),
				"qtcreator",
			},
			PluginFileName: "libQodeAssist.so",
			LibExtension:   ".so",
		}, nil

	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			localAppData = filepath.Join(homeDir, "AppData", "Local")
		}

		return &PlatformConfig{
			QtCreatorPaths: []string{
				"C:/Qt/Tools/QtCreator",
				"C:/Program Files/Qt/Tools/QtCreator",
				filepath.Join(homeDir, "Qt", "Tools", "QtCreator"),
			},
			PluginPaths: []string{
				filepath.Join(localAppData, "QtProject", "qtcreator", "plugins", "{qtc_version}", PluginAuthor),
			},
			ExecutableNames: []string{
				filepath.Join("bin", "qtcreator.exe"),
				"qtcreator.exe",
			},
			PluginFileName: "QodeAssist.dll",
			LibExtension:   ".dll",
		}, nil

	default:
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func GetPlatformArchName() (platform, arch string, err error) {
	switch runtime.GOOS {
	case "windows":
		platform = "Windows"
	case "linux":
		platform = "Linux"
	case "darwin":
		platform = "macOS"
	default:
		return "", "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	if runtime.GOOS == "darwin" {
		arch = "universal"
	} else {
		switch runtime.GOARCH {
		case "amd64":
			arch = "x64"
		case "arm64":
			arch = "arm64"
		default:
			arch = runtime.GOARCH
		}
	}

	return platform, arch, nil
}

func linuxXDGDirs(homeDir string) (dataHome, configHome string) {
	dataHome = os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		dataHome = filepath.Join(homeDir, ".local", "share")
	}
	configHome = os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		configHome = filepath.Join(homeDir, ".config")
	}
	return dataHome, configHome
}

func dedupStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, v := range values {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		result = append(result, v)
	}
	return result
}

func GetPluginSearchDirs(qtCreatorPath string, version *Version) ([]string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	versionStr := version.String()
	var searchDirs []string

	switch runtime.GOOS {
	case "darwin":
		userPluginsBase := filepath.Join(homeDir, "Library", "Application Support", "QtProject", "Qt Creator", "plugins", versionStr)
		searchDirs = append(searchDirs, userPluginsBase)
		qtcBuiltinPlugins := filepath.Join(qtCreatorPath, "Contents", "PlugIns", "qtcreator")
		searchDirs = append(searchDirs, qtcBuiltinPlugins)

	case "linux":
		pluginsSubPath := filepath.Join("qtcreator", "plugins")
		dataHome, configHome := linuxXDGDirs(homeDir)
		for _, root := range dedupStrings([]string{dataHome, configHome}) {
			for _, pluginsDir := range findQtPluginsDirs(root, pluginsSubPath, 4) {
				searchDirs = append(searchDirs, filepath.Join(pluginsDir, versionStr))
			}
		}
		qtcBuiltinPlugins := filepath.Join(qtCreatorPath, "lib", "qtcreator", "plugins")
		searchDirs = append(searchDirs, qtcBuiltinPlugins)

	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			localAppData = filepath.Join(homeDir, "AppData", "Local")
		}
		userPluginsBase := filepath.Join(localAppData, "QtProject", "qtcreator", "plugins", versionStr)
		searchDirs = append(searchDirs, userPluginsBase)
		qtcBuiltinPlugins := filepath.Join(qtCreatorPath, "lib", "qtcreator", "plugins")
		searchDirs = append(searchDirs, qtcBuiltinPlugins)

	default:
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return searchDirs, nil
}
