package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type PluginMetaData struct {
	IID       string `json:"IID"`
	ClassName string `json:"className"`
	MetaData  struct {
		Name        string      `json:"Name"`
		Version     string      `json:"Version"`
		Vendor      string      `json:"Vendor"`
		Copyright   string      `json:"Copyright"`
		License     interface{} `json:"License"`
		Description string      `json:"Description"`
		Url         string      `json:"Url"`
	} `json:"MetaData"`
}

func ExecuteQtPluginInfo(qtPluginInfoPath, pluginPath string) (*PluginMetaData, error) {
	if _, err := os.Stat(qtPluginInfoPath); err != nil {
		return nil, fmt.Errorf("qtplugininfo not found at %s: %w", qtPluginInfoPath, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), CommandTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, qtPluginInfoPath, "--full-json", pluginPath)
	output, err := cmd.Output()

	if ctx.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf("qtplugininfo command timed out after %v", CommandTimeout)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to run qtplugininfo: %w", err)
	}

	var metadata PluginMetaData
	if err := json.Unmarshal(output, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse plugin metadata: %w", err)
	}

	if metadata.MetaData.Version == "" {
		return nil, fmt.Errorf("version not found in plugin metadata")
	}

	return &metadata, nil
}

func FindFileInDirectory(baseDir, targetName string, maxDepth int, matchFunc func(os.DirEntry, string) bool) (string, error) {
	if maxDepth <= 0 {
		return "", nil
	}

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return "", nil
	}

	for _, entry := range entries {
		fullPath := filepath.Join(baseDir, entry.Name())

		if matchFunc != nil {
			if matchFunc(entry, fullPath) {
				return fullPath, nil
			}
		} else if entry.Name() == targetName {
			return fullPath, nil
		}

		if entry.IsDir() {
			if found, err := FindFileInDirectory(fullPath, targetName, maxDepth-1, matchFunc); err == nil && found != "" {
				return found, nil
			}
		}
	}

	return "", nil
}

func GetQtPluginInfoPath(qtCreatorRootPath string) (string, error) {
	platformConfig, err := GetPlatformConfig()
	if err != nil {
		return "", err
	}

	for _, execName := range platformConfig.ExecutableNames {
		execPath := filepath.Join(qtCreatorRootPath, execName)
		if PathExists(execPath) {
			dir := filepath.Dir(execPath)
			qtPluginInfoPath := filepath.Join(dir, GetQtPluginInfoName())
			if PathExists(qtPluginInfoPath) {
				return qtPluginInfoPath, nil
			}
		}
	}

	return "", fmt.Errorf("qtplugininfo not found in Qt Creator directory: %s", qtCreatorRootPath)
}

func FindCorePlugin(qtCreatorRootPath string) (string, error) {
	corePluginName := GetCorePluginName()

	searchRootAbs, err := filepath.Abs(qtCreatorRootPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	var foundPath string
	err = filepath.WalkDir(qtCreatorRootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() {
			return nil
		}

		if d.Name() == corePluginName {
			pathAbs, err := filepath.Abs(path)
			if err == nil && strings.HasPrefix(pathAbs, searchRootAbs) {
				foundPath = path
				return filepath.SkipAll
			}
		}

		return nil
	})

	if err != nil && err != filepath.SkipAll {
		return "", fmt.Errorf("error while searching for core plugin: %w", err)
	}

	if foundPath == "" {
		return "", fmt.Errorf("core plugin (%s) not found in Qt Creator directory: %s", corePluginName, qtCreatorRootPath)
	}

	return foundPath, nil
}

func SanitizeFileMode(mode os.FileMode) os.FileMode {
	mode &^= os.ModeSetuid | os.ModeSetgid | os.ModeSticky
	mode = mode &^ GroupWriteMask

	if mode.IsDir() {
		return DefaultDirPermissions
	}
	if mode&ExecutableBitMask != 0 {
		return ExecutablePermissions
	}
	return DefaultFilePermissions
}

