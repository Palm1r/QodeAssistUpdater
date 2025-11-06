package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Version struct {
	Major int
	Minor int
	Patch int
}

func ParseVersion(s string) (*Version, error) {
	s = strings.TrimPrefix(s, "v")
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid version format: %s (expected X.Y.Z)", s)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid major version: %w", err)
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid minor version: %w", err)
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid patch version: %w", err)
	}

	return &Version{Major: major, Minor: minor, Patch: patch}, nil
}

func (v *Version) IsNewer(other *Version) bool {
	if other == nil {
		return true
	}
	if v.Major != other.Major {
		return v.Major > other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor > other.Minor
	}
	return v.Patch > other.Patch
}

func (v *Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func GetQtCreatorVersion(qtCreatorRootPath string) (*Version, error) {
	qtPluginInfoPath, err := GetQtPluginInfoPath(qtCreatorRootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to find qtplugininfo: %w", err)
	}

	corePluginPath, err := FindCorePlugin(qtCreatorRootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to find core plugin: %w", err)
	}

	metadata, err := ExecuteQtPluginInfo(qtPluginInfoPath, corePluginPath)
	if err != nil {
		return nil, err
	}

	version, err := ParseVersion(metadata.MetaData.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to parse version: %w", err)
	}

	return version, nil
}

func findPluginFile(pluginPath, pluginName string) string {
	info, err := os.Stat(pluginPath)
	if err != nil {
		return ""
	}

	var searchDir string
	if !info.IsDir() {
		searchDir = filepath.Dir(pluginPath)
	} else {
		searchDir = pluginPath
	}

	pluginNameLower := strings.ToLower(pluginName)
	var pluginFile string

	platformConfig, _ := GetPlatformConfig()
	if platformConfig == nil {
		return ""
	}

	err = filepath.WalkDir(searchDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}

		fileName := strings.ToLower(d.Name())
		if strings.Contains(fileName, pluginNameLower) {
			ext := filepath.Ext(fileName)
			if ext == platformConfig.LibExtension {
				pluginFile = path
				return filepath.SkipAll
			}
		}
		return nil
	})

	return pluginFile
}

func CheckPluginInstalled(pluginPath, pluginName string) bool {
	return findPluginFile(pluginPath, pluginName) != ""
}

func GetInstalledPluginVersionFromPath(pluginPath, pluginName, qtCreatorPath string) (*Version, error) {
	pluginFile := findPluginFile(pluginPath, pluginName)
	if pluginFile == "" {
		return nil, fmt.Errorf("plugin library file not found in directory: %s", pluginPath)
	}

	qtPluginInfoPath, err := GetQtPluginInfoPath(qtCreatorPath)
	if err != nil {
		return nil, fmt.Errorf("failed to find qtplugininfo: %w", err)
	}

	metadata, err := ExecuteQtPluginInfo(qtPluginInfoPath, pluginFile)
	if err != nil {
		return nil, err
	}

	version, err := ParseVersion(metadata.MetaData.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to parse version: %w", err)
	}

	return version, nil
}
