// Package shell provides marker utilities for RC file manipulation.
package shell

import (
	"bufio"
	"strings"
)

// MarkerType represents the type of marker section.
type MarkerType string

const (
	// MarkerTypeTheme is for theme configuration.
	MarkerTypeTheme MarkerType = "theme"
	// MarkerTypeFont is for font configuration.
	MarkerTypeFont MarkerType = "font"
	// MarkerTypeAlias is for shell aliases.
	MarkerTypeAlias MarkerType = "alias"
	// MarkerTypePath is for PATH modifications.
	MarkerTypePath MarkerType = "path"
	// MarkerTypePlugin is for plugin initialization.
	MarkerTypePlugin MarkerType = "plugin"
	// MarkerTypeConfig is for general configuration.
	MarkerTypeConfig MarkerType = "config"
)

// ParseMarkers extracts all Savanhi markers from RC content.
// Returns a map of marker names to their content.
func ParseMarkers(content string) (map[string]string, error) {
	markers := make(map[string]string)

	lines := strings.Split(content, "\n")
	var currentMarker string
	var currentContent []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for start marker
		if strings.HasPrefix(trimmed, MarkerStartPrefix) && strings.HasSuffix(trimmed, MarkerStartSuffix) {
			if currentMarker != "" {
				return nil, ErrUnclosedMarker
			}
			// Extract marker name
			currentMarker = extractMarkerName(trimmed, MarkerStartPrefix, MarkerStartSuffix)
			currentContent = nil
			continue
		}

		// Check for end marker
		if strings.HasPrefix(trimmed, MarkerEndPrefix) && strings.HasSuffix(trimmed, MarkerEndSuffix) {
			if currentMarker == "" {
				// End marker without start - malformed
				return nil, ErrUnclosedMarker
			}
			expectedEnd := MarkerEndPrefix + currentMarker + MarkerEndSuffix
			if trimmed != expectedEnd {
				return nil, ErrUnclosedMarker
			}
			// Save marker content
			markers[currentMarker] = strings.Join(currentContent, "\n")
			currentMarker = ""
			currentContent = nil
			continue
		}

		// Accumulate content if inside a marker
		if currentMarker != "" {
			currentContent = append(currentContent, line)
		}
	}

	// Check for unclosed marker at end of file
	if currentMarker != "" {
		return nil, ErrUnclosedMarker
	}

	return markers, nil
}

// ValidateMarkers validates that all markers are properly closed.
func ValidateMarkers(content string) error {
	scanner := bufio.NewScanner(strings.NewReader(content))
	stack := make([]string, 0)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Check for start marker
		if strings.HasPrefix(line, MarkerStartPrefix) && strings.HasSuffix(line, MarkerStartSuffix) {
			marker := extractMarkerName(line, MarkerStartPrefix, MarkerStartSuffix)
			stack = append(stack, marker)
		}

		// Check for end marker
		if strings.HasPrefix(line, MarkerEndPrefix) && strings.HasSuffix(line, MarkerEndSuffix) {
			if len(stack) == 0 {
				return ErrUnclosedMarker
			}
			marker := extractMarkerName(line, MarkerEndPrefix, MarkerEndSuffix)
			if stack[len(stack)-1] != marker {
				return ErrUnclosedMarker
			}
			stack = stack[:len(stack)-1]
		}
	}

	if len(stack) > 0 {
		return ErrUnclosedMarker
	}

	return nil
}

// FindDuplicateMarkers finds duplicate marker sections in content.
func FindDuplicateMarkers(content string) []string {
	seen := make(map[string]int)
	duplicates := make([]string, 0)

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, MarkerStartPrefix) && strings.HasSuffix(line, MarkerStartSuffix) {
			marker := extractMarkerName(line, MarkerStartPrefix, MarkerStartSuffix)
			seen[marker]++
			if seen[marker] > 1 {
				duplicates = append(duplicates, marker)
			}
		}
	}

	return duplicates
}

// RemoveAllMarkers removes all Savanhi markers from content.
func RemoveAllMarkers(content string) (string, error) {
	result := make([]string, 0)
	inMarker := false

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for start marker
		if strings.HasPrefix(trimmed, MarkerStartPrefix) && strings.HasSuffix(trimmed, MarkerStartSuffix) {
			inMarker = true
			continue
		}

		// Check for end marker
		if strings.HasPrefix(trimmed, MarkerEndPrefix) && strings.HasSuffix(trimmed, MarkerEndSuffix) {
			inMarker = false
			continue
		}

		// Add line if not inside a marker
		if !inMarker {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n"), nil
}

// PreserveUserContent preserves user modifications outside markers.
// This is used when updating sections while keeping user changes.
func PreserveUserContent(content string, preservedMarkers []string) (string, map[string]string, error) {
	// Extract content we want to preserve
	preserved := make(map[string]string)

	// Parse all markers
	allMarkers, err := ParseMarkers(content)
	if err != nil {
		return "", nil, err
	}

	// Keep only preserved markers
	for marker := range allMarkers {
		shouldPreserve := false
		for _, pm := range preservedMarkers {
			if pm == marker {
				shouldPreserve = true
				break
			}
		}
		if shouldPreserve {
			preserved[marker] = allMarkers[marker]
		}
	}

	// Remove all markers from content
	cleanContent, err := RemoveAllMarkers(content)
	if err != nil {
		return "", nil, err
	}

	return cleanContent, preserved, nil
}

// extractMarkerName extracts the marker name from a marker line.
func extractMarkerName(line, prefix, suffix string) string {
	// Remove prefix and suffix
	start := len(prefix)
	end := len(line) - len(suffix)
	if start >= end {
		return ""
	}
	return strings.TrimSpace(line[start:end])
}
