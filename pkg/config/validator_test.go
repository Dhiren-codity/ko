package config

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

// Mock for os.Stat
type MockFileInfo struct{}

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
