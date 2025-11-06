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
			path:     "nonexistentfile",
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
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid config",
			data:    []byte(`{"baseImage":"golang","platforms":["linux/amd64"],"images":["app"],"labels":{"version":"1.0"},"allowInsecure":false}`),
			wantErr: false,
		},
		{
			name:    "empty data",
			data:    []byte(``),
			wantErr: true,
			errMsg:  "empty configuration data",
		},
		{
			name:    "invalid JSON",
			data:    []byte(`{baseImage:golang}`),
			wantErr: true,
			errMsg:  "invalid JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseConfig(tt.data)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
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
				BaseImage: "golang",
				Platforms: []string{"linux/amd64"},
				Images:    []string{"app"},
				Labels:    map[string]string{"version": "1.0"},
			},
			expected: []string{},
		},
		{
			name: "missing baseImage",
			cfg: &Config{
				Platforms: []string{"linux/amd64"},
				Images:    []string{"app"},
			},
			expected: []string{"baseImage is required"},
		},
		{
			name: "invalid platform",
			cfg: &Config{
				BaseImage: "golang",
				Platforms: []string{"invalid/platform"},
				Images:    []string{"app"},
			},
			expected: []string{"invalid platform format: invalid/platform"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateConfigStructure(tt.cfg)
			assert.Equal(t, tt.expected, errors)
		})
	}
}

func TestValidateImageReferences(t *testing.T) {
	tests := []struct {
		name    string
		refs    []string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid references",
			refs:    []string{"golang", "alpine"},
			wantErr: false,
		},
		{
			name:    "invalid references",
			refs:    []string{"invalid/image"},
			wantErr: true,
			errMsg:  "invalid image references: invalid/image",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateImageReferences(tt.refs)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
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
			ref:      "golang",
			expected: true,
		},
		{
			name:     "invalid reference",
			ref:      "invalid/image",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidPlatform(tt.platform)
			assert.Equal(t, tt.expected, result)
		})
	}
}
