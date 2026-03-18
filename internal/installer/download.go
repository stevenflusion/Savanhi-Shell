// Package installer provides dependency installation and management.
// This file implements download functionality.
package installer

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// downloadFile downloads a file from URL to the cache directory.
func (i *DefaultInstaller) downloadFile(ctx context.Context, sourceURL, cacheDir string, opts *Options) (*DownloadResult, error) {
	// Parse URL and determine filename
	parsedURL, err := url.Parse(sourceURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Handle template URLs
	sourceURL = i.expandURLTemplate(sourceURL)

	// Determine filename
	filename := filepath.Base(parsedURL.Path)
	if filename == "" || filename == "." {
		filename = "download"
	}

	// Create cache path
	cachePath := filepath.Join(cacheDir, filename)

	// Check if already cached
	if opts.UseCache {
		if info, err := os.Stat(cachePath); err == nil {
			// File exists in cache
			result := &DownloadResult{
				URL:       sourceURL,
				LocalPath: cachePath,
				Size:      info.Size(),
				Cached:    true,
			}

			// Calculate checksum
			if checksum, err := i.calculateChecksum(cachePath); err == nil {
				result.Checksum = checksum
			}

			return result, nil
		}
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: opts.Timeout,
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", sourceURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set user agent
	req.Header.Set("User-Agent", "Savanhi-Shell/1.0")

	// Report progress start
	i.reportProgress(&InstallProgress{
		Component: "download",
		Stage:     StageDownloading,
		Percent:   0,
		Message:   "Starting download...",
	})

	// Execute request with retries
	var resp *http.Response
	var lastErr error

	for attempt := 0; attempt < opts.MaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			time.Sleep(time.Duration(attempt*2) * time.Second)
		}

		resp, lastErr = client.Do(req)
		if lastErr == nil && resp.StatusCode == http.StatusOK {
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("%w: %v", ErrNetworkError, lastErr)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Create temp file for download
	tempPath := cachePath + ".tmp"
	tempFile, err := os.Create(tempPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tempFile.Close()

	// Get content length for progress
	contentLength := resp.ContentLength
	var downloaded int64

	// Download with progress
	buf := make([]byte, 32*1024) // 32KB buffer
	startTime := time.Now()

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := tempFile.Write(buf[:n]); writeErr != nil {
				os.Remove(tempPath)
				return nil, fmt.Errorf("failed to write file: %w", writeErr)
			}
			downloaded += int64(n)

			// Report progress
			percent := float64(0)
			if contentLength > 0 {
				percent = float64(downloaded) / float64(contentLength) * 100
			}

			i.reportProgress(&InstallProgress{
				Component:       "download",
				Stage:           StageDownloading,
				Percent:         percent,
				BytesDownloaded: downloaded,
				TotalBytes:      contentLength,
				Message:         "Downloading...",
				UpdatedAt:       time.Now(),
			})
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			os.Remove(tempPath)
			return nil, fmt.Errorf("download error: %w", err)
		}
	}

	// Close and rename
	tempFile.Close()
	if err := os.Rename(tempPath, cachePath); err != nil {
		os.Remove(tempPath)
		return nil, fmt.Errorf("failed to finalize download: %w", err)
	}

	// Calculate checksum
	checksum, err := i.calculateChecksum(cachePath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate checksum: %w", err)
	}

	// Cache result
	i.downloadCache["download"] = cachePath

	return &DownloadResult{
		URL:       sourceURL,
		LocalPath: cachePath,
		Size:      downloaded,
		Checksum:  checksum,
		Verified:  false,
		Cached:    false,
		Duration:  time.Since(startTime),
	}, nil
}

// expandURLTemplate expands URL templates with OS and arch.
func (i *DefaultInstaller) expandURLTemplate(templateURL string) string {
	// Replace {os} with runtime OS
	result := strings.ReplaceAll(templateURL, "{os}", runtime.GOOS)

	// Replace {arch} with runtime arch
	result = strings.ReplaceAll(result, "{arch}", runtime.GOARCH)

	// Replace {platform} with detected platform
	result = strings.ReplaceAll(result, "{platform}", string(Platform(i.context.OS)))

	return result
}

// calculateChecksum calculates SHA256 checksum of a file.
func (i *DefaultInstaller) calculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// refreshFontCache refreshes the font cache on Linux systems.
func (i *DefaultInstaller) refreshFontCache() error {
	// Only needed on Linux
	if i.context.OS == "darwin" {
		return nil
	}

	// Check if fc-cache is available
	if _, err := os.Stat("/usr/bin/fc-cache"); os.IsNotExist(err) {
		// fc-cache not available, skip
		return nil
	}

	// Run fc-cache -fv
	cmd := exec.Command("fc-cache", "-fv", i.context.FontDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to refresh font cache: %w\n%s", err, string(output))
	}

	return nil
}

// installFontFromZip installs a font from a zip archive.
func (i *DefaultInstaller) installFontFromZip(zipPath string, dep *Dependency, result *InstallResult) error {
	// Open the zip file
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}
	defer reader.Close()

	// Extract font files
	extracted := false
	for _, file := range reader.File {
		// Only extract font files
		ext := strings.ToLower(filepath.Ext(file.Name))
		if ext != ".ttf" && ext != ".otf" && ext != ".woff" && ext != ".woff2" {
			continue
		}

		// Open file in zip
		rc, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open file in zip: %w", err)
		}

		// Create target path
		targetPath := filepath.Join(i.context.FontDir, filepath.Base(file.Name))

		// Create target file
		targetFile, err := os.Create(targetPath)
		if err != nil {
			rc.Close()
			return fmt.Errorf("failed to create font file: %w", err)
		}

		// Copy content
		_, err = io.Copy(targetFile, rc)
		targetFile.Close()
		rc.Close()

		if err != nil {
			return fmt.Errorf("failed to extract font: %w", err)
		}

		// Set permissions
		if err := os.Chmod(targetPath, 0644); err != nil {
			return fmt.Errorf("failed to set permissions: %w", err)
		}

		extracted = true
		result.Warnings = append(result.Warnings, fmt.Sprintf("Extracted: %s", filepath.Base(file.Name)))
	}

	if !extracted {
		return fmt.Errorf("no font files found in archive")
	}

	result.InstalledPath = i.context.FontDir

	// Refresh font cache
	if err := i.refreshFontCache(); err != nil {
		result.Warnings = append(result.Warnings, "Failed to refresh font cache - you may need to run 'fc-cache -fv' manually")
	}

	return nil
}
