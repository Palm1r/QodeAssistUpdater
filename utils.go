package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func CalculateFileChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to calculate checksum: %w", err)
	}

	checksum := hex.EncodeToString(hash.Sum(nil))
	return checksum, nil
}

func VerifyFileChecksum(filePath, expectedChecksum string) error {
	if expectedChecksum == "" {
		return nil
	}

	PrintStep("Verifying checksum...")
	actualChecksum, err := CalculateFileChecksum(filePath)
	if err != nil {
		return err
	}

	expectedChecksum = normalizeChecksum(expectedChecksum)
	actualChecksum = normalizeChecksum(actualChecksum)

	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch!\nExpected: %s\nActual:   %s", expectedChecksum, actualChecksum)
	}

	PrintSuccess(fmt.Sprintf("Checksum verified: %s", actualChecksum))
	return nil
}

func normalizeChecksum(checksum string) string {
	normalized := ""
	for _, ch := range checksum {
		if (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F') {
			if ch >= 'A' && ch <= 'F' {
				normalized += string(ch + 32)
			} else {
				normalized += string(ch)
			}
		}
	}
	return normalized
}

func ConfirmAction(prompt string) bool {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read user input: %v\n", err)
		return false
	}
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "yes" || response == "y"
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func ExpandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		return filepath.Join(homeDir, path[2:]), nil
	}
	return os.ExpandEnv(path), nil
}

func EnsureDir(path string) error {
	if err := os.MkdirAll(path, DefaultDirPermissions); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", path, err)
	}
	return nil
}

func FindPluginFiles(baseDir, targetFileName string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(baseDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if path == baseDir {
			return nil
		}

		if d.Name() == targetFileName {
			files = append(files, path)
			if d.IsDir() {
				return filepath.SkipDir
			}
		}

		return nil
	})

	return files, err
}
