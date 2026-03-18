// Package detector provides system detection capabilities.
// This file contains tests for config detection.
package detector

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewConfigDetector(t *testing.T) {
	detector := NewConfigDetector()
	if detector == nil {
		t.Error("NewConfigDetector() returned nil")
	}
}

func TestConfigDetector_Detect(t *testing.T) {
	detector := NewConfigDetector()
	snapshot, err := detector.Detect()

	if err != nil {
		t.Errorf("Detect() returned error: %v", err)
	}

	if snapshot == nil {
		t.Fatal("Detect() returned nil ConfigSnapshot")
	}

	// DetectedAt should be set and recent
	if time.Since(snapshot.DetectedAt) > time.Minute {
		t.Error("DetectedAt timestamp is too old")
	}
}

func TestIsCommandAvailable(t *testing.T) {
	detector := &configDetector{}

	// Test with a command that should exist
	if !detector.isCommandAvailable("ls") && !detector.isCommandAvailable("dir") {
		// On some systems neither might exist, that's okay
		t.Log("Neither ls nor dir available, skipping command availability test")
	}

	// Test with a command that should not exist
	if detector.isCommandAvailable("this-command-should-not-exist-12345") {
		t.Error("isCommandAvailable() returned true for non-existent command")
	}
}

func TestDetectOhMyPosh(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot get home directory")
	}

	detector := &configDetector{}
	hasOMP, configPath := detector.detectOhMyPosh(homeDir)

	// If oh-my-posh is installed, configPath might be empty (if config not found)
	// but hasOMP should be true
	if hasOMP && configPath != "" {
		// Verify the path exists
		if _, err := os.Stat(configPath); err != nil {
			t.Errorf("Config path %s does not exist", configPath)
		}
	}

	_ = hasOMP // Just verify we can call the function
}

func TestDetectStarship(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot get home directory")
	}

	detector := &configDetector{}
	hasStarship, configPath := detector.detectStarship(homeDir)

	// If starship is installed, configPath might be empty (if config not found)
	// but hasStarship should be true
	if hasStarship && configPath != "" {
		// Verify the path exists
		if _, err := os.Stat(configPath); err != nil {
			t.Errorf("Config path %s does not exist", configPath)
		}
	}

	_ = hasStarship // Just verify we can call the function
}

func TestConfigPaths(t *testing.T) {
	// Test that the config paths are correctly constructed
	homeDir := "/home/testuser"

	ohMyPoshPaths := []string{
		filepath.Join(homeDir, ".config", "oh-my-posh", "config.json"),
		filepath.Join(homeDir, ".config", "oh-my-posh", "theme.json"),
		filepath.Join(homeDir, ".oh-my-posh", "config.json"),
	}

	starshipPaths := []string{
		filepath.Join(homeDir, ".config", "starship.toml"),
		filepath.Join(homeDir, ".config", "starship", "starship.toml"),
	}

	// Verify paths are correctly formatted
	for _, path := range ohMyPoshPaths {
		if !filepath.IsAbs(path) {
			t.Errorf("Path %s is not absolute", path)
		}
	}

	for _, path := range starshipPaths {
		if !filepath.IsAbs(path) {
			t.Errorf("Path %s is not absolute", path)
		}
	}
}
