package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileExists(t *testing.T) {
	// Create a temporary file
	file, err := os.CreateTemp("", "testfile")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	assert.True(t, FileExists(file.Name()))
	assert.False(t, FileExists("nonexistentfile"))
}

func TestParseConfig(t *testing.T) {
	validJSON := `{
        "baseImage": "gcr.io/distroless/base",
        "platforms": ["linux/amd64"],
        "images": ["gcr.io/my-project/my-image"],
        "labels": {"version": "1.0"},
        "allowInsecure": false
    }`

	invalidJSON := `{"baseImage": "gcr.io/distroless/base"`

	cfg, err := ParseConfig([]byte(validJSON))
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "gcr.io/distroless/base", cfg.BaseImage)

	cfg, err = ParseConfig([]byte(invalidJSON))
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestValidateConfigStructure(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			cfg: &Config{
				BaseImage: "gcr.io/distroless/base",
				Platforms: []string{"linux/amd64"},
				Images:    []string{"gcr.io/my-project/my-image"},
			},
			wantErr: false,
		},
		{
			name: "missing baseImage",
			cfg: &Config{
				Platforms: []string{"linux/amd64"},
				Images:    []string{"gcr.io/my-project/my-image"},
			},
			wantErr: true,
			errMsg:  "baseImage is required",
		},
		{
			name: "invalid platform",
			cfg: &Config{
				BaseImage: "gcr.io/distroless/base",
				Platforms: []string{"invalid/platform"},
				Images:    []string{"gcr.io/my-project/my-image"},
			},
			wantErr: true,
			errMsg:  "invalid platform format: invalid/platform",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateConfigStructure(tt.cfg)
			if tt.wantErr {
				assert.NotEmpty(t, errors)
				assert.Contains(t, errors[0], tt.errMsg)
			} else {
				assert.Empty(t, errors)
			}
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
			refs:    []string{"gcr.io/my-project/my-image"},
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
	assert.True(t, IsValidImageReference("gcr.io/my-project/my-image"))
	assert.False(t, IsValidImageReference("Invalid/Image"))
}

func TestIsValidPlatform(t *testing.T) {
	assert.True(t, IsValidPlatform("linux/amd64"))
	assert.False(t, IsValidPlatform("invalid/platform"))
}

func TestGetConfigPath(t *testing.T) {
	expectedPath := ".ko.json"
	assert.Equal(t, expectedPath, GetConfigPath())
}

func TestLoadDefaultConfig(t *testing.T) {
	// Create a temporary config file
	configContent := `{
        "baseImage": "gcr.io/distroless/base",
        "platforms": ["linux/amd64"],
        "images": ["gcr.io/my-project/my-image"]
    }`
	file, err := os.CreateTemp("", ".ko.json")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	_, err = file.WriteString(configContent)
	assert.NoError(t, err)
	file.Close()

	// Override GetConfigPath to return the temp file path
	originalGetConfigPath := GetConfigPath
	GetConfigPath = func() string { return file.Name() }
	defer func() { GetConfigPath = originalGetConfigPath }()

	cfg, err := LoadDefaultConfig()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "gcr.io/distroless/base", cfg.BaseImage)
}
