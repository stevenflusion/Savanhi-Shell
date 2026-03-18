// Package shell provides tests for shell RC file manipulation.
package shell

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFormatMarker(t *testing.T) {
	tests := []struct {
		marker    string
		wantStart string
		wantEnd   string
	}{
		{"theme", "# >>> savanhi-theme >>>", "# <<< savanhi-theme <<<"},
		{"font", "# >>> savanhi-font >>>", "# <<< savanhi-font <<<"},
		{"config", "# >>> savanhi-config >>>", "# <<< savanhi-config <<<"},
	}

	for _, tt := range tests {
		t.Run(tt.marker, func(t *testing.T) {
			start := formatStartMarker(tt.marker)
			end := formatEndMarker(tt.marker)

			if start != tt.wantStart {
				t.Errorf("formatStartMarker(%s) = %s, want %s", tt.marker, start, tt.wantStart)
			}
			if end != tt.wantEnd {
				t.Errorf("formatEndMarker(%s) = %s, want %s", tt.marker, end, tt.wantEnd)
			}
		})
	}
}

func TestInjectSection(t *testing.T) {
	tests := []struct {
		name     string
		existing string
		marker   string
		content  string
		wantErr  bool
	}{
		{
			name:     "inject into empty file",
			existing: "",
			marker:   "theme",
			content:  "export THEME=mytheme",
			wantErr:  false,
		},
		{
			name:     "inject into file with content",
			existing: "# existing content\n",
			marker:   "theme",
			content:  "export THEME=mytheme",
			wantErr:  false,
		},
		{
			name:     "inject with existing marker",
			existing: "# >>> savanhi-theme >>>\nexport OLD=value\n# <<< savanhi-theme <<<\n",
			marker:   "theme",
			content:  "export NEW=value",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpDir := t.TempDir()
			rcPath := filepath.Join(tmpDir, ".testrc")

			if tt.existing != "" {
				os.WriteFile(rcPath, []byte(tt.existing), 0644)
			}

			// Create shell
			shell, _ := NewZshShellWithPath(rcPath)

			// Inject section
			err := shell.InjectSection(tt.marker, tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("InjectSection() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify section exists
				has, _ := shell.HasSection(tt.marker)
				if !has {
					t.Error("HasSection() returned false after injection")
				}

				// Verify content
				content, err := shell.GetSection(tt.marker)
				if err != nil {
					t.Errorf("GetSection() error = %v", err)
					return
				}
				if content != tt.content {
					t.Errorf("GetSection() = %q, want %q", content, tt.content)
				}
			}
		})
	}
}

func TestRemoveSection(t *testing.T) {
	tests := []struct {
		name    string
		content string
		marker  string
		want    string
		wantErr bool
	}{
		{
			name:    "remove existing section",
			content: "# existing\n# >>> savanhi-theme >>>\nexport THEME=mytheme\n# <<< savanhi-theme <<<\n",
			marker:  "theme",
			want:    "# existing\n",
			wantErr: false,
		},
		{
			name:    "remove non-existent section",
			content: "# existing\n",
			marker:  "theme",
			want:    "# existing\n",
			wantErr: false,
		},
		{
			name:    "remove from empty file",
			content: "",
			marker:  "theme",
			want:    "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			rcPath := filepath.Join(tmpDir, ".testrc")
			os.WriteFile(rcPath, []byte(tt.content), 0644)

			shell, _ := NewZshShellWithPath(rcPath)

			err := shell.RemoveSection(tt.marker)
			if (err != nil) != tt.wantErr {
				t.Errorf("RemoveSection() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				content, _ := os.ReadFile(rcPath)
				if string(content) != tt.want {
					t.Errorf("Content after removal = %q, want %q", string(content), tt.want)
				}
			}
		})
	}
}

func TestUnclosedMarkers(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "properly closed markers",
			content: "# >>> savanhi-theme >>>\ncontent\n# <<< savanhi-theme <<<",
			wantErr: false,
		},
		{
			name:    "unclosed start marker",
			content: "# >>> savanhi-theme >>>\ncontent\n# no end marker",
			wantErr: true,
		},
		{
			name:    "end without start",
			content: "content\n# <<< savanhi-theme <<<",
			wantErr: true,
		},
		{
			name:    "multiple markers properly closed",
			content: "# >>> savanhi-theme >>>\ntheme\n# <<< savanhi-theme <<<\n# >>> savanhi-font >>>\nfont\n# <<< savanhi-font <<<",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMarkers(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMarkers() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseMarkers(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    map[string]string
		wantErr bool
	}{
		{
			name:    "single marker",
			content: "# >>> savanhi-theme >>>\nexport THEME=mytheme\n# <<< savanhi-theme <<<",
			want:    map[string]string{"theme": "export THEME=mytheme"},
			wantErr: false,
		},
		{
			name:    "multiple markers",
			content: "# >>> savanhi-theme >>>\nthemecode\n# <<< savanhi-theme <<<\n# >>> savanhi-font >>>\nfontcode\n# <<< savanhi-font <<<",
			want: map[string]string{
				"theme": "themecode",
				"font":  "fontcode",
			},
			wantErr: false,
		},
		{
			name:    "unclosed marker",
			content: "# >>> savanhi-theme >>>\nthemecode",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "empty marker",
			content: "# >>> savanhi-theme >>>\n\n\n# <<< savanhi-theme <<<",
			want:    map[string]string{"theme": "\n"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMarkers(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMarkers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				for k, v := range tt.want {
					if got[k] != v {
						t.Errorf("ParseMarkers()[%s] = %q, want %q", k, got[k], v)
					}
				}
			}
		})
	}
}

func TestHasSection(t *testing.T) {
	tests := []struct {
		name    string
		content string
		marker  string
		want    bool
	}{
		{
			name:    "section exists",
			content: "# >>> savanhi-theme >>>\ncode\n# <<< savanhi-theme <<<",
			marker:  "theme",
			want:    true,
		},
		{
			name:    "section does not exist",
			content: "# some other content",
			marker:  "theme",
			want:    false,
		},
		{
			name:    "empty file",
			content: "",
			marker:  "theme",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			rcPath := filepath.Join(tmpDir, ".testrc")
			os.WriteFile(rcPath, []byte(tt.content), 0644)

			shell, _ := NewZshShellWithPath(rcPath)

			got, err := shell.HasSection(tt.marker)
			if err != nil {
				t.Errorf("HasSection() error = %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("HasSection() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSection(t *testing.T) {
	tests := []struct {
		name    string
		content string
		marker  string
		want    string
		wantErr bool
	}{
		{
			name:    "get existing section",
			content: "# >>> savanhi-theme >>>\nexport THEME=xyz\n# <<< savanhi-theme <<<",
			marker:  "theme",
			want:    "export THEME=xyz",
			wantErr: false,
		},
		{
			name:    "get non-existent section",
			content: "# other content",
			marker:  "theme",
			want:    "",
			wantErr: false,
		},
		{
			name:    "unclosed marker",
			content: "# >>> savanhi-theme >>>\nexport THEME=xyz",
			marker:  "theme",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			rcPath := filepath.Join(tmpDir, ".testrc")
			os.WriteFile(rcPath, []byte(tt.content), 0644)

			shell, _ := NewZshShellWithPath(rcPath)

			got, err := shell.GetSection(tt.marker)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSection() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("GetSection() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPreserveUserContent(t *testing.T) {
	tests := []struct {
		name            string
		content         string
		preserveMarkers []string
		wantMarkers     map[string]string
		wantErr         bool
	}{
		{
			name:            "preserve specific marker",
			content:         "# >>> savanhi-theme >>>\nthemecode\n# <<< savanhi-theme <<<\n# >>> savanhi-font >>>\nfontcode\n# <<< savanhi-font <<<\nuser content",
			preserveMarkers: []string{"theme"},
			wantMarkers:     map[string]string{"theme": "themecode"},
			wantErr:         false,
		},
		{
			name:            "preserve multiple markers",
			content:         "# >>> savanhi-theme >>>\nthemecode\n# <<< savanhi-theme <<<\n# >>> savanhi-font >>>\nfontcode\n# <<< savanhi-font <<<",
			preserveMarkers: []string{"theme", "font"},
			wantMarkers: map[string]string{
				"theme": "themecode",
				"font":  "fontcode",
			},
			wantErr: false,
		},
		{
			name:            "preserve none",
			content:         "# >>> savanhi-theme >>>\nthemecode\n# <<< savanhi-theme <<<\nuser content",
			preserveMarkers: []string{},
			wantMarkers:     map[string]string{},
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanContent, preserved, err := PreserveUserContent(tt.content, tt.preserveMarkers)
			if (err != nil) != tt.wantErr {
				t.Errorf("PreserveUserContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Check that markers were removed from clean content
				if containsSavanhiMarker(cleanContent) {
					t.Error("cleanContent still contains Savanhi markers")
				}

				// Check preserved markers
				for k, v := range tt.wantMarkers {
					if preserved[k] != v {
						t.Errorf("preserved[%s] = %q, want %q", k, preserved[k], v)
					}
				}
			}
		})
	}
}

func containsSavanhiMarker(content string) bool {
	return len(content) > 10 && (contains(content, "savanhi-") > -1)
}

func contains(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func TestPermissionErrors(t *testing.T) {
	// Skip on Windows where permission model differs
	// This test requires root/user permission manipulation
	t.Skip("requires specific permission setup")
}

func TestZshShell_GetRCPath(t *testing.T) {
	shell, err := NewZshShell()
	if err != nil {
		t.Fatalf("NewZshShell() error = %v", err)
	}

	path, err := shell.GetRCPath()
	if err != nil {
		t.Errorf("GetRCPath() error = %v", err)
		return
	}

	if path == "" {
		t.Error("GetRCPath() returned empty path")
		return
	}

	// Should end with .zshrc
	if filepath.Base(path) != ".zshrc" {
		t.Errorf("GetRCPath() = %s, want path ending with .zshrc", path)
	}
}

func TestBashShell_GetRCPath(t *testing.T) {
	shell, err := NewBashShell()
	if err != nil {
		t.Fatalf("NewBashShell() error = %v", err)
	}

	path, err := shell.GetRCPath()
	if err != nil {
		t.Errorf("GetRCPath() error = %v", err)
		return
	}

	if path == "" {
		t.Error("GetRCPath() returned empty path")
		return
	}

	// Should end with .bashrc
	if filepath.Base(path) != ".bashrc" {
		t.Errorf("GetRCPath() = %s, want path ending with .bashrc", path)
	}
}

func TestDetectShellType(t *testing.T) {
	// This test verifies DetectShellType doesn't crash
	// The actual result depends on the environment
	shellType := DetectShellType()

	validTypes := map[ShellType]bool{
		ShellTypeBash: true,
		ShellTypeZsh:  true,
		ShellTypeFish: true,
		ShellTypePwsh: true,
	}

	if !validTypes[shellType] {
		t.Errorf("DetectShellType() returned invalid type: %s", shellType)
	}
}

func TestFindDuplicateMarkers(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
	}{
		{
			name:    "no duplicates",
			content: "# >>> savanhi-theme >>>\ntheme\n# <<< savanhi-theme <<<\n# >>> savanhi-font >>>\nfont\n# <<< savanhi-font <<<",
			want:    nil,
		},
		{
			name:    "duplicate theme markers",
			content: "# >>> savanhi-theme >>>\nent1\n# <<< savanhi-theme <<<\n# >>> savanhi-theme >>>\nent2\n# <<< savanhi-theme <<<",
			want:    []string{"theme"},
		},
		{
			name:    "empty content",
			content: "",
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindDuplicateMarkers(tt.content)
			if len(got) != len(tt.want) {
				t.Errorf("FindDuplicateMarkers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoveAllMarkers(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "remove all markers",
			content: "# >>> savanhi-theme >>>\nthemecode\n# <<< savanhi-theme <<<\nuser content\n# >>> savanhi-font >>>\nfontcode\n# <<< savanhi-font <<<",
			want:    "user content",
		},
		{
			name:    "no markers to remove",
			content: "user content\nmore content",
			want:    "user content\nmore content",
		},
		{
			name:    "empty content",
			content: "",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RemoveAllMarkers(tt.content)
			if err != nil {
				t.Errorf("RemoveAllMarkers() error = %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("RemoveAllMarkers() = %q, want %q", got, tt.want)
			}
		})
	}
}
