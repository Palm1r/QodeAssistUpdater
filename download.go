package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var httpClient = &http.Client{
	Timeout: HTTPTimeout,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return fmt.Errorf("too many redirects")
		}
		if err := validateURL(req.URL.String()); err != nil {
			return fmt.Errorf("redirect to invalid URL: %w", err)
		}
		return nil
	},
}

func validateURL(rawURL string) error {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme != "https" {
		return fmt.Errorf("only HTTPS URLs are allowed, got: %s", parsedURL.Scheme)
	}

	allowedHosts := []string{"github.com", "githubusercontent.com"}
	validHost := false
	for _, host := range allowedHosts {
		if parsedURL.Hostname() == host || strings.HasSuffix(parsedURL.Hostname(), "."+host) {
			validHost = true
			break
		}
	}

	if !validHost {
		return fmt.Errorf("only GitHub URLs are allowed, got: %s", parsedURL.Host)
	}

	if strings.Contains(parsedURL.Path, "..") {
		return fmt.Errorf("path traversal detected in URL")
	}

	return nil
}

func downloadWithRetry(downloadURL, destPath string, maxRetries int) error {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			if backoff > 30*time.Second {
				backoff = 30 * time.Second
			}
			time.Sleep(backoff)
		}

		err := downloadFile(downloadURL, destPath)
		if err == nil {
			return nil
		}

		lastErr = err
	}

	return fmt.Errorf("download failed after %d attempts: %w", maxRetries+1, lastErr)
}

func downloadFile(downloadURL, destPath string) error {
	if err := validateURL(downloadURL); err != nil {
		return fmt.Errorf("URL validation failed: %w", err)
	}

	resp, err := httpClient.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d %s", resp.StatusCode, resp.Status)
	}

	totalBytes := resp.ContentLength
	if totalBytes > MaxDownloadSize {
		return fmt.Errorf("file size (%d bytes) exceeds maximum allowed size (%d bytes)", totalBytes, MaxDownloadSize)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	pw := &ProgressWriter{
		Total:  totalBytes,
		Writer: out,
		OnProgress: func(current, total int64) {
			PrintProgress(current, total, "Downloading...")
		},
	}

	limitedReader := io.LimitReader(resp.Body, MaxDownloadSize)

	written, err := io.Copy(pw, limitedReader)
	if err != nil {
		os.Remove(destPath)
		return fmt.Errorf("download interrupted: %w", err)
	}

	if written > MaxDownloadSize {
		os.Remove(destPath)
		return fmt.Errorf("file size exceeds maximum allowed size (%d bytes)", MaxDownloadSize)
	}

	fmt.Println()
	return nil
}

func DownloadPlugin(downloadURL, destPath string) error {
	const maxRetries = 3
	return downloadWithRetry(downloadURL, destPath, maxRetries)
}
