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

// qtplugininfo is used only on Windows, where running `qtcreator.exe --version`

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

func GetQtPluginInfoPath(qtCreatorRootPath string) (string, error) {
	platformConfig, err := GetPlatformConfig()
	if err != nil {
		return "", err
	}

	qtPluginInfoName := GetQtPluginInfoName()

	for _, execName := range platformConfig.ExecutableNames {
		execPath := filepath.Join(qtCreatorRootPath, execName)
		if PathExists(execPath) {
			dir := filepath.Dir(execPath)
			qtPluginInfoPath := filepath.Join(dir, qtPluginInfoName)
			if PathExists(qtPluginInfoPath) {
				return qtPluginInfoPath, nil
			}
		}
	}

	if found, err := exec.LookPath(qtPluginInfoName); err == nil {
		return found, nil
	}

	matchExecutable := func(entry os.DirEntry, fullPath string) bool {
		return !entry.IsDir() && entry.Name() == qtPluginInfoName
	}
	if found, _ := FindFileInDirectory(qtCreatorRootPath, qtPluginInfoName, MaxSearchDepth, matchExecutable); found != "" {
		return found, nil
	}

	return "", fmt.Errorf("qtplugininfo not found in Qt Creator directory or PATH: %s", qtCreatorRootPath)
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
