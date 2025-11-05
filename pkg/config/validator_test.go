package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		setup   func()
		wantErr bool
		errMsg  string
	}{
		{
			name:    "file does not exist",
			path:    "nonexistent.json",
			setup:   func() {},
			wantErr: true,
			errMsg:  "configuration file not found",
		},
		{
			name: "invalid JSON",
			path: "invalid.json",
			setup: func() {
				ioutil.WriteFile("invalid.json", []byte("{invalid json}"), 0644)
			},
			wantErr: true,
			errMsg:  "invalid JSON",
		},
		{
			name: "valid configuration",
			path: "valid.json",
			setup: func() {
				ioutil.WriteFile("valid.json", []byte(`{"baseImage": "golang:1.16", "images": ["app"]}`), 0644)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := ValidateConfig(tt.path)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
			os.Remove(tt.path)
		})
	}
}

func TestFileExists(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		setup  func()
		exists bool
	}{
		{
			name: "file exists",
			path: "exists.json",
			setup: func() {
				ioutil.WriteFile("exists.json", []byte("content"), 0644)
			},
			exists: true,
		},
		{
			name:   "file does not exist",
			path:   "nonexistent.json",
			setup:  func() {},
			exists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			exists := FileExists(tt.path)
			assert.Equal(t, tt.exists, exists)
			os.Remove(tt.path)
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
			name:    "empty data",
			data:    []byte(""),
			wantErr: true,
			errMsg:  "empty configuration data",
		},
		{
			name:    "invalid JSON",
			data:    []byte("{invalid json}"),
			wantErr: true,
			errMsg:  "invalid JSON",
		},
		{
			name:    "valid JSON",
			data:    []byte(`{"baseImage": "golang:1.16", "images": ["app"]}`),
			wantErr: false,
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
		name   string
		cfg    *Config
		errors []string
	}{
		{
			name:   "nil configuration",
			cfg:    nil,
			errors: []string{"configuration is nil"},
		},
		{
			name:   "missing baseImage",
			cfg:    &Config{Images: []string{"app"}},
			errors: []string{"baseImage is required"},
		},
		{
			name:   "valid configuration",
			cfg:    &Config{BaseImage: "golang:1.16", Images: []string{"app"}},
			errors: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateConfigStructure(tt.cfg)
			assert.Equal(t, tt.errors, errors)
		})
	}
}

func TestCheckRequiredFields(t *testing.T) {
	tests := []struct {
		name   string
		cfg    *Config
		errors []string
	}{
		{
			name:   "missing baseImage",
			cfg:    &Config{Images: []string{"app"}},
			errors: []string{"baseImage is required"},
		},
		{
			name:   "missing images",
			cfg:    &Config{BaseImage: "golang:1.16"},
			errors: []string{"at least one image must be specified"},
		},
		{
			name:   "all fields present",
			cfg:    &Config{BaseImage: "golang:1.16", Images: []string{"app"}},
			errors: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := CheckRequiredFields(tt.cfg)
			assert.Equal(t, tt.errors, errors)
		})
	}
}

func TestIsValidImageReference(t *testing.T) {
	tests := []struct {
		name  string
		ref   string
		valid bool
	}{
		{
			name:  "empty reference",
			ref:   "",
			valid: false,
		},
		{
			name:  "invalid reference",
			ref:   "invalid/image",
			valid: false,
		},
		{
			name:  "valid reference",
			ref:   "golang:1.16",
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := IsValidImageReference(tt.ref)
			assert.Equal(t, tt.valid, valid)
		})
	}
}

func TestIsValidPlatform(t *testing.T) {
	tests := []struct {
		name     string
		platform string
		valid    bool
	}{
		{
			name:     "empty platform",
			platform: "",
			valid:    false,
		},
		{
			name:     "invalid platform",
			platform: "invalid/platform",
			valid:    false,
		},
		{
			name:     "valid platform",
			platform: "linux/amd64",
			valid:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := IsValidPlatform(tt.platform)
			assert.Equal(t, tt.valid, valid)
		})
	}
}

func TestGetConfigPath(t *testing.T) {
	expectedPath := filepath.Join(".", ".ko.json")
	path := GetConfigPath()
	assert.Equal(t, expectedPath, path)
}

func TestLoadDefaultConfig(t *testing.T) {
	tests := []struct {
		name    string
		setup   func()
		wantErr bool
		errMsg  string
	}{
		{
			name:    "default config not found",
			setup:   func() {},
			wantErr: true,
			errMsg:  "default configuration file not found",
		},
		{
			name: "valid default config",
			setup: func() {
				ioutil.WriteFile(".ko.json", []byte(`{"baseImage": "golang:1.16", "images": ["app"]}`), 0644)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			_, err := LoadDefaultConfig()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
			os.Remove(".ko.json")
		})
	}
}
