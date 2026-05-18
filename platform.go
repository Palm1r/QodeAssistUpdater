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
		return &PlatformConfig{
			QtCreatorPaths: []string{
				filepath.Join(homeDir, "Qt", "Tools", "QtCreator"),
				"/usr",
				"/usr/local",
				"/opt/Qt/Tools/QtCreator",
			},
			PluginPaths: []string{
				filepath.Join(homeDir, ".local", "share", "data", "QtProject", "qtcreator", "plugins", "{qtc_version}", PluginAuthor),
				filepath.Join(homeDir, ".config", "QtProject", "qtcreator", "plugins", "{qtc_version}", PluginAuthor),
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
		userPluginsBase1 := filepath.Join(homeDir, ".local", "share", "data", "QtProject", "qtcreator", "plugins", versionStr)
		userPluginsBase2 := filepath.Join(homeDir, ".config", "QtProject", "qtcreator", "plugins", versionStr)
		searchDirs = append(searchDirs, userPluginsBase1, userPluginsBase2)
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
