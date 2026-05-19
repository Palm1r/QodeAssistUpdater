package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func findQtPluginsDirs(root, pluginsSubPath string, maxDepth int) []string {
	if !PathExists(root) {
		return nil
	}

	var result []string
	filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}

		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return nil
		}
		depth := 0
		if rel != "." {
			depth = strings.Count(rel, string(os.PathSeparator)) + 1
		}
		if depth >= maxDepth {
			return filepath.SkipDir
		}

		if d.Name() == "QtProject" {
			pluginsDir := filepath.Join(path, pluginsSubPath)
			if info, statErr := os.Stat(pluginsDir); statErr == nil && info.IsDir() {
				result = append(result, pluginsDir)
			}
			return filepath.SkipDir
		}

		return nil
	})

	return result
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

func GetQtCreatorExecutablePath(qtCreatorRootPath string) (string, error) {
	platformConfig, err := GetPlatformConfig()
	if err != nil {
		return "", err
	}

	for _, execName := range platformConfig.ExecutableNames {
		execPath := filepath.Join(qtCreatorRootPath, execName)
		if PathExists(execPath) {
			return execPath, nil
		}
	}

	return "", fmt.Errorf("Qt Creator executable not found in directory: %s", qtCreatorRootPath)
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
