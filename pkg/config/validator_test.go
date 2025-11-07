package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock for os.Stat
type MockFileInfo struct {
	mock.Mock
}

func (m *MockFileInfo) Name() string       { return "" }
func (m *MockFileInfo) Size() int64        { return 0 }
func (m *MockFileInfo) Mode() os.FileMode  { return 0 }
func (m *MockFileInfo) ModTime() time.Time { return time.Time{} }
func (m *MockFileInfo) IsDir() bool        { return false }
func (m *MockFileInfo) Sys() interface{}   { return nil }

func TestFileExists(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		setup    func()
		expected bool
	}{
		{
			name: "file exists",
			path: "testfile",
			setup: func() {
				os.Create("testfile")
			},
			expected: true,
		},
		{
			name:     "file does not exist",
			path:     "nonexistent",
			setup:    func() {},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			defer os.Remove(tt.path)
			result := FileExists(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseConfig(t *testing.T) {
	validConfig := `{"baseImage": "golang:1.16", "platforms": ["linux/amd64"], "images": ["app"], "labels": {"version": "1.0"}, "allowInsecure": false}`
	invalidConfig := `{"baseImage": "golang:1.16", "platforms": ["linux/amd64"], "images": ["app"], "labels": {"version": "1.0"}, "allowInsecure": false`

	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "valid config",
			data:    []byte(validConfig),
			wantErr: false,
		},
		{
			name:    "invalid config",
			data:    []byte(invalidConfig),
			wantErr: true,
		},
		{
			name:    "empty config",
			data:    []byte(""),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseConfig(tt.data)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateConfigStructure(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *Config
		expected []string
	}{
		{
			name: "valid config",
			cfg: &Config{
				BaseImage: "golang:1.16",
				Platforms: []string{"linux/amd64"},
				Images:    []string{"app"},
				Labels:    map[string]string{"version": "1.0"},
			},
			expected: []string{},
		},
		{
			name: "missing base image",
			cfg: &Config{
				Platforms: []string{"linux/amd64"},
				Images:    []string{"app"},
				Labels:    map[string]string{"version": "1.0"},
			},
			expected: []string{"baseImage is required"},
		},
		{
			name: "invalid platform",
			cfg: &Config{
				BaseImage: "golang:1.16",
				Platforms: []string{"invalid/platform"},
				Images:    []string{"app"},
				Labels:    map[string]string{"version": "1.0"},
			},
			expected: []string{"invalid platform format: invalid/platform"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateConfigStructure(tt.cfg)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateImageReferences(t *testing.T) {
	tests := []struct {
		name    string
		refs    []string
		wantErr bool
	}{
		{
			name:    "valid references",
			refs:    []string{"golang:1.16", "alpine:latest"},
			wantErr: false,
		},
		{
			name:    "invalid references",
			refs:    []string{"invalid/image"},
			wantErr: true,
		},
		{
			name:    "empty references",
			refs:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateImageReferences(tt.refs)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsValidImageReference(t *testing.T) {
	tests := []struct {
		name     string
		ref      string
		expected bool
	}{
		{
			name:     "valid reference",
			ref:      "golang:1.16",
			expected: true,
		},
		{
			name:     "invalid reference",
			ref:      "invalid/image",
			expected: false,
		},
		{
			name:     "empty reference",
			ref:      "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidImageReference(tt.ref)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidPlatform(t *testing.T) {
	tests := []struct {
		name     string
		platform string
		expected bool
	}{
		{
			name:     "valid platform",
			platform: "linux/amd64",
			expected: true,
		},
		{
			name:     "invalid platform",
			platform: "invalid/platform",
			expected: false,
		},
		{
			name:     "empty platform",
			platform: "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidPlatform(tt.platform)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetConfigPath(t *testing.T) {
	expected := ".ko.json"
	result := GetConfigPath()
	assert.Equal(t, expected, result)
}

func TestLoadDefaultConfig(t *testing.T) {
	validConfig := `{"baseImage": "golang:1.16", "platforms": ["linux/amd64"], "images": ["app"], "labels": {"version": "1.0"}, "allowInsecure": false}`
	invalidConfig := `{"baseImage": "golang:1.16", "platforms": ["linux/amd64"], "images": ["app"], "labels": {"version": "1.0"}, "allowInsecure": false`

	tests := []struct {
		name    string
		setup   func()
		wantErr bool
	}{
		{
			name: "valid default config",
			setup: func() {
				os.WriteFile(GetConfigPath(), []byte(validConfig), 0644)
			},
			wantErr: false,
		},
		{
			name: "invalid default config",
			setup: func() {
				os.WriteFile(GetConfigPath(), []byte(invalidConfig), 0644)
			},
			wantErr: true,
		},
		{
			name: "no default config",
			setup: func() {
				os.Remove(GetConfigPath())
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			defer os.Remove(GetConfigPath())
			_, err := LoadDefaultConfig()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
