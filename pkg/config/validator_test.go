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
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "valid config",
			data:    []byte(`{"baseImage": "golang", "images": ["app"]}`),
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			data:    []byte(`{baseImage: "golang"}`),
			wantErr: true,
		},
		{
			name:    "empty data",
			data:    []byte(``),
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
