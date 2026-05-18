package main

import (
	"fmt"
	"os"
	"path/filepath"
)

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
