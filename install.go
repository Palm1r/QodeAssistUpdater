package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bodgit/sevenzip"
)

type archiveFile interface {
	GetName() string
	GetMode() os.FileMode
	IsDir() bool
	GetSize() int64
	Open() (io.ReadCloser, error)
}

type zipFileWrapper struct {
	*zip.File
}

func (z zipFileWrapper) GetName() string              { return z.Name }
func (z zipFileWrapper) GetMode() os.FileMode         { return z.Mode() }
func (z zipFileWrapper) IsDir() bool                  { return z.FileInfo().IsDir() }
func (z zipFileWrapper) GetSize() int64               { return int64(z.UncompressedSize64) }
func (z zipFileWrapper) Open() (io.ReadCloser, error) { return z.File.Open() }

type sevenzipFileWrapper struct {
	*sevenzip.File
}

func (s sevenzipFileWrapper) GetName() string              { return s.Name }
func (s sevenzipFileWrapper) GetMode() os.FileMode         { return s.Mode() }
func (s sevenzipFileWrapper) IsDir() bool                  { return s.FileInfo().IsDir() }
func (s sevenzipFileWrapper) GetSize() int64               { return s.FileInfo().Size() }
func (s sevenzipFileWrapper) Open() (io.ReadCloser, error) { return s.File.Open() }

func InstallPlugin(sourcePath, pluginDir string) error {
	ext := strings.ToLower(filepath.Ext(sourcePath))

	if err := EnsureDir(pluginDir); err != nil {
		return err
	}

	switch ext {
	case ".zip":
		return extractZipArchive(sourcePath, pluginDir)
	case ".7z":
		return extract7zArchive(sourcePath, pluginDir)
	default:
		return fmt.Errorf("unsupported archive format: %s (supported: .zip, .7z)", ext)
	}
}

func RemoveOldQodeAssistFiles(pluginDir string, interactive bool) error {
	if !PathExists(pluginDir) {
		return nil
	}

	platformConfig, err := GetPlatformConfig()
	if err != nil {
		return err
	}

	filesToRemove, err := FindPluginFiles(pluginDir, platformConfig.PluginFileName)
	if err != nil {
		return fmt.Errorf("failed to find plugin files: %w", err)
	}

	if len(filesToRemove) == 0 {
		return nil
	}

	if interactive {
		fmt.Println()
		PrintVerbose("The following old plugin files will be removed before update:")
		for _, file := range filesToRemove {
			PrintListItem(file)
		}

		if !ConfirmAction("\nDo you want to proceed with removing old files? (yes/no): ") {
			return fmt.Errorf("update cancelled by user")
		}
	}

	for _, file := range filesToRemove {
		if err := os.RemoveAll(file); err != nil {
			continue
		}

		if interactive {
			PrintVerbose(fmt.Sprintf("Removed: %s", file))
		}
	}

	return nil
}

func validateAndExtractFile(f archiveFile, destDir string, totalExtractedSize *int64, fileCount *int) error {
	*fileCount++
	if *fileCount > MaxFiles {
		return fmt.Errorf("archive contains too many files (max %d)", MaxFiles)
	}

	name := f.GetName()
	if strings.Contains(name, "..") || filepath.IsAbs(name) {
		return fmt.Errorf("invalid file path in archive (ZipSlip protection): %s", name)
	}

	fpath := filepath.Join(destDir, name)
	destDirAbs, err := filepath.Abs(destDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path of dest: %w", err)
	}

	fpathAbs, err := filepath.Abs(fpath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path of file: %w", err)
	}

	if !strings.HasPrefix(fpathAbs, destDirAbs+string(os.PathSeparator)) && fpathAbs != destDirAbs {
		return fmt.Errorf("invalid file path in archive (ZipSlip protection): %s", name)
	}

	if f.IsDir() {
		if err := EnsureDir(fpath); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
		return nil
	}

	fileSize := f.GetSize()
	if fileSize > 0 && fileSize > MaxExtractSize-*totalExtractedSize {
		return fmt.Errorf("file %s would exceed maximum extraction size", name)
	}

	if err := EnsureDir(filepath.Dir(fpath)); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("failed to open file in archive: %w", err)
	}
	defer rc.Close()

	safeMode := SanitizeFileMode(f.GetMode())
	outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, safeMode)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer outFile.Close()

	var limitedReader io.Reader
	if fileSize > 0 {
		limitedReader = io.LimitReader(rc, fileSize)
	} else {
		limitedReader = io.LimitReader(rc, MaxExtractSize-*totalExtractedSize)
	}

	copied, err := io.Copy(outFile, limitedReader)
	if err != nil {
		return fmt.Errorf("failed to extract file: %w", err)
	}

	*totalExtractedSize += copied
	if *totalExtractedSize > MaxExtractSize {
		return fmt.Errorf("total extraction size exceeds maximum allowed size (%d bytes)", MaxExtractSize)
	}

	return nil
}

func extractArchive(files []archiveFile, destDir, archiveType string) error {
	var totalExtractedSize int64
	fileCount := 0

	for _, f := range files {
		if err := validateAndExtractFile(f, destDir, &totalExtractedSize, &fileCount); err != nil {
			return err
		}
	}

	PrintSuccess("Archive extracted successfully")
	return nil
}

func extractZipArchive(archivePath, destDir string) error {
	PrintStep("Extracting ZIP archive...")

	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open ZIP archive: %w", err)
	}
	defer r.Close()

	files := make([]archiveFile, len(r.File))
	for i, f := range r.File {
		files[i] = zipFileWrapper{f}
	}

	return extractArchive(files, destDir, "ZIP")
}

func extract7zArchive(archivePath, destDir string) error {
	PrintStep("Extracting 7z archive...")

	r, err := sevenzip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open 7z archive: %w", err)
	}
	defer r.Close()

	files := make([]archiveFile, len(r.File))
	for i, f := range r.File {
		files[i] = sevenzipFileWrapper{f}
	}

	return extractArchive(files, destDir, "7z")
}

func RemovePlugin(pluginDir string) error {
	if !PathExists(pluginDir) {
		return fmt.Errorf("plugin directory does not exist: %s", pluginDir)
	}

	info, err := os.Stat(pluginDir)
	if err != nil {
		return fmt.Errorf("failed to check plugin directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("plugin path is not a directory: %s", pluginDir)
	}

	platformConfig, err := GetPlatformConfig()
	if err != nil {
		return err
	}

	filesToRemove, err := FindPluginFiles(pluginDir, platformConfig.PluginFileName)
	if err != nil {
		return err
	}

	if len(filesToRemove) == 0 {
		return fmt.Errorf("no plugin files found in: %s", pluginDir)
	}

	fmt.Println()
	PrintVerbose("The following files will be removed:")
	for _, file := range filesToRemove {
		relPath, err := filepath.Rel(pluginDir, file)
		if err != nil || relPath == "" {
			relPath = filepath.Base(file)
		}
		PrintListItem(relPath)
	}

	if !ConfirmAction("\nDo you want to proceed with removal? (yes/no): ") {
		return fmt.Errorf("removal cancelled by user")
	}

	for _, file := range filesToRemove {
		if err := os.RemoveAll(file); err != nil {
			return fmt.Errorf("failed to remove %s: %w", file, err)
		}

		relPath, err := filepath.Rel(pluginDir, file)
		if err != nil || relPath == "" {
			relPath = filepath.Base(file)
		}
		PrintVerbose(fmt.Sprintf("Removed: %s", relPath))
	}

	cleanupEmptyDirectories(pluginDir)

	fmt.Println()
	PrintSuccess("Plugin removed successfully")
	return nil
}

func cleanupEmptyDirectories(pluginDir string) {
	var dirs []string

	filepath.WalkDir(pluginDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if path != pluginDir && d.IsDir() {
			dirs = append(dirs, path)
		}
		return nil
	})

	sort.Slice(dirs, func(i, j int) bool {
		depthI := len(strings.Split(dirs[i], string(os.PathSeparator)))
		depthJ := len(strings.Split(dirs[j], string(os.PathSeparator)))
		return depthI > depthJ
	})

	for _, dir := range dirs {
		os.Remove(dir)
	}

	os.Remove(pluginDir)
}
