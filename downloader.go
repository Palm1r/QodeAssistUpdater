package main

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/bodgit/sevenzip"
)

func DownloadPlugin(url, destPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	totalBytes := resp.ContentLength
	var downloadedBytes int64

	buf := make([]byte, 32*1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			written, writeErr := out.Write(buf[:n])
			if writeErr != nil {
				return writeErr
			}
			downloadedBytes += int64(written)
			
			if totalBytes > 0 {
				percent := float64(downloadedBytes) / float64(totalBytes) * 100
				fmt.Printf("\rDownloading... %.1f%% (%d/%d bytes)", percent, downloadedBytes, totalBytes)
			} else {
				fmt.Printf("\rDownloading... %d bytes", downloadedBytes)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	fmt.Println()

	return nil
}

func FindPluginAsset(release *GithubRelease, qtCreatorVersion *Version) (string, string, error) {
	platform := runtime.GOOS
	arch := runtime.GOARCH
	
	qtcVerFull := fmt.Sprintf("QtC%d.%d.%d", qtCreatorVersion.Major, qtCreatorVersion.Minor, qtCreatorVersion.Patch)
	qtcVerMajorMinor := fmt.Sprintf("QtC%d.%d", qtCreatorVersion.Major, qtCreatorVersion.Minor)

	var platformName string
	switch platform {
	case "windows":
		platformName = "Windows"
	case "linux":
		platformName = "Linux"
	case "darwin":
		platformName = "macOS"
	default:
		return "", "", fmt.Errorf("unsupported platform: %s", platform)
	}

	var archName string
	if platform == "darwin" {
		archName = "universal"
	} else {
		switch arch {
		case "amd64":
			archName = "x64"
		case "arm64":
			archName = "arm64"
		default:
			archName = arch
		}
	}

	var matchingAssets []struct {
		Name string
		URL  string
		FullVersionMatch bool
	}

	for _, asset := range release.Assets {
		name := asset.Name
		hasPlatform := strings.Contains(name, platformName)
		hasArch := strings.Contains(name, archName)
		hasQtcVersionFull := strings.Contains(name, qtcVerFull)
		hasQtcVersionMajorMinor := strings.Contains(name, qtcVerMajorMinor)

		if hasPlatform && hasArch && (hasQtcVersionFull || hasQtcVersionMajorMinor) {
			matchingAssets = append(matchingAssets, struct {
				Name string
				URL  string
				FullVersionMatch bool
			}{asset.Name, asset.BrowserDownloadURL, hasQtcVersionFull})
		}
	}

	if len(matchingAssets) == 0 {
		return "", "", fmt.Errorf("no matching asset found for %s %s Qt Creator %s",
			platformName, archName, qtCreatorVersion.String())
	}

	for _, asset := range matchingAssets {
		if asset.FullVersionMatch {
			return asset.Name, asset.URL, nil
		}
	}

	return matchingAssets[0].Name, matchingAssets[0].URL, nil
}

func InstallPlugin(sourcePath, pluginDir string) error {
	ext := strings.ToLower(filepath.Ext(sourcePath))
	
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return err
	}

	if ext == ".zip" {
		return extractZipArchive(sourcePath, pluginDir)
	}

	if ext == ".7z" {
		return extract7zArchive(sourcePath, pluginDir)
	}

	return fmt.Errorf("unsupported archive format: %s (supported: .zip, .7z)", ext)
}

func extractZipArchive(archivePath, destDir string) error {
	fmt.Printf("Extracting ZIP archive...\n")
	
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open ZIP archive: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		if strings.Contains(f.Name, "..") || filepath.IsAbs(f.Name) {
			return fmt.Errorf("invalid file path in archive (ZipSlip protection): %s", f.Name)
		}

		fpath := filepath.Join(destDir, f.Name)
		destDirAbs, _ := filepath.Abs(destDir)
		fpathAbs, _ := filepath.Abs(fpath)
		if !strings.HasPrefix(fpathAbs, destDirAbs+string(os.PathSeparator)) && fpathAbs != destDirAbs {
			return fmt.Errorf("invalid file path in archive (ZipSlip protection): %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fpath, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
			return fmt.Errorf("failed to create parent directory: %w", err)
		}

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("failed to open file in archive: %w", err)
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return fmt.Errorf("failed to create destination file: %w", err)
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return fmt.Errorf("failed to extract file: %w", err)
		}
	}

	fmt.Println("Archive extracted successfully")
	return nil
}

func extract7zArchive(archivePath, destDir string) error {
	fmt.Printf("Extracting 7z archive...\n")
	
	r, err := sevenzip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open 7z archive: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		if strings.Contains(f.Name, "..") || filepath.IsAbs(f.Name) {
			return fmt.Errorf("invalid file path in archive (ZipSlip protection): %s", f.Name)
		}

		fpath := filepath.Join(destDir, f.Name)
		destDirAbs, _ := filepath.Abs(destDir)
		fpathAbs, _ := filepath.Abs(fpath)
		if !strings.HasPrefix(fpathAbs, destDirAbs+string(os.PathSeparator)) && fpathAbs != destDirAbs {
			return fmt.Errorf("invalid file path in archive (ZipSlip protection): %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fpath, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
			return fmt.Errorf("failed to create parent directory: %w", err)
		}

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("failed to open file in archive: %w", err)
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return fmt.Errorf("failed to create destination file: %w", err)
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return fmt.Errorf("failed to extract file: %w", err)
		}
	}

	fmt.Println("Archive extracted successfully")
	return nil
}

