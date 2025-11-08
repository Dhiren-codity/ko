package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileExists(t *testing.T) {
	dir := t.TempDir()

	// Non-existent
	assert.False(t, FileExists(filepath.Join(dir, "nope.json")))

	// Directory should be false
	assert.False(t, FileExists(dir))

	// File exists
	f := filepath.Join(dir, "file.txt")
	err := os.WriteFile(f, []byte("data"), 0o644)
	assert.NoError(t, err)
	assert.True(t, FileExists(f))
}

func TestParseConfig_Success(t *testing.T) {
	data := []byte(`{
		"baseImage": "alpine:3.18",
		"platforms": ["linux/amd64", "darwin/arm64"],
		"images": ["ghcr.io/acme/app:1.0.0", "proj/svc"],
		"labels": {"env": "test"},
		"allowInsecure": true
	}`)
	cfg, err := ParseConfig(data)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "alpine:3.18", cfg.BaseImage)
	assert.ElementsMatch(t, []string{"linux/amd64", "darwin/arm64"}, cfg.Platforms)
	assert.ElementsMatch(t, []string{"ghcr.io/acme/app:1.0.0", "proj/svc"}, cfg.Images)
	assert.Equal(t, map[string]string{"env": "test"}, cfg.Labels)
	assert.Equal(t, true, cfg.AllowInsecure)
}

func TestParseConfig_EmptyData(t *testing.T) {
	cfg, err := ParseConfig(nil)
	assert.Nil(t, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty configuration data")
}

func TestParseConfig_InvalidJSON(t *testing.T) {
	data := []byte(`{ invalid json }`)
	cfg, err := ParseConfig(data)
	assert.Nil(t, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid JSON")
}

func TestValidateConfigStructure_Nil(t *testing.T) {
	errs := ValidateConfigStructure(nil)
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0], "configuration is nil")
}

func TestValidateConfigStructure_RequiredFields(t *testing.T) {
	cfg := &Config{}
	errs := ValidateConfigStructure(cfg)
	assert.Contains(t, errs, "baseImage is required")
	assert.Contains(t, errs, "at least one image must be specified")
}

func TestValidateConfigStructure_InvalidPlatformAndBaseImage(t *testing.T) {
	cfg := &Config{
		BaseImage: "INVALID^^", // invalid reference
		Platforms: []string{"linux", "foo/bar"},
		Images:    []string{"alpine"}, // required satisfied
	}
	errs := ValidateConfigStructure(cfg)
	// invalid platforms captured
	assert.Contains(t, errs, "invalid platform format: linux")
	assert.Contains(t, errs, "invalid platform format: foo/bar")
	// invalid base image captured
	found := false
	for _, e := range errs {
		if strings.Contains(e, "invalid base image reference") && strings.Contains(e, "INVALID^^") {
			found = true
			break
		}
	}
	assert.True(t, found, "expected invalid base image reference error")
}

func TestCheckRequiredFields(t *testing.T) {
	// Both missing
	errs := CheckRequiredFields(&Config{})
	assert.Len(t, errs, 2)
	assert.Contains(t, errs, "baseImage is required")
	assert.Contains(t, errs, "at least one image must be specified")

	// Only baseImage missing
	errs = CheckRequiredFields(&Config{Images: []string{"img"}})
	assert.Len(t, errs, 1)
	assert.Contains(t, errs, "baseImage is required")

	// Only images missing
	errs = CheckRequiredFields(&Config{BaseImage: "alpine:3.18"})
	assert.Len(t, errs, 1)
	assert.Contains(t, errs, "at least one image must be specified")

	// All present
	errs = CheckRequiredFields(&Config{BaseImage: "alpine:3.18", Images: []string{"img"}})
	assert.Empty(t, errs)
}

func TestValidateImageReferences(t *testing.T) {
	// No refs provided
	err := ValidateImageReferences(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no image references provided")

	// Invalid refs collected
	invalidDigest := "sha256:deadbeef"
	err = ValidateImageReferences([]string{
		"alpine:3.18",
		"BadUpper:latest",
		"img@" + invalidDigest,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid image references")
	assert.Contains(t, err.Error(), "BadUpper:latest")
	assert.Contains(t, err.Error(), "img@"+invalidDigest)

	// All valid
	validDigest := "sha256:" + strings.Repeat("a", 64)
	err = ValidateImageReferences([]string{
		"alpine",
		"alpine:3.18",
		"ghcr.io/org/app:1.2.3",
		"registry.example.com/ns/img@" + validDigest,
	})
	assert.NoError(t, err)
}

func TestIsValidPlatform(t *testing.T) {
	valid := []string{
		"linux/amd64",
		"linux/arm64",
		"linux/arm",
		"windows/386",
		"darwin/arm64",
	}
	invalid := []string{
		"",
		"linux",
		"amd64/linux",
		"solaris/amd64",
		"linux/mips",
		"linux/arm/v7",
	}

	for _, p := range valid {
		assert.True(t, IsValidPlatform(p), "expected valid: %q", p)
	}
	for _, p := range invalid {
		assert.False(t, IsValidPlatform(p), "expected invalid: %q", p)
	}
}

func TestGetConfigPath(t *testing.T) {
	expected := filepath.Join(".", ".ko.json")
	assert.Equal(t, expected, GetConfigPath())
}

func TestLoadDefaultConfig_NotFound(t *testing.T) {
	wd, err := os.Getwd()
	assert.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(wd) })

	dir := t.TempDir()
	assert.NoError(t, os.Chdir(dir))

	cfg, err := LoadDefaultConfig()
	assert.Nil(t, cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "default configuration file not found")
}

func TestLoadDefaultConfig_Success(t *testing.T) {
	wd, err := os.Getwd()
	assert.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(wd) })

	dir := t.TempDir()
	assert.NoError(t, os.Chdir(dir))

	data := []byte(`{
		"baseImage": "alpine:3.18",
		"platforms": ["linux/amd64"],
		"images": ["ghcr.io/acme/app:1.0.0"],
		"labels": {"k":"v"},
		"allowInsecure": false
	}`)
	err = os.WriteFile(GetConfigPath(), data, 0o644)
	assert.NoError(t, err)

	cfg, err := LoadDefaultConfig()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "alpine:3.18", cfg.BaseImage)
	assert.ElementsMatch(t, []string{"linux/amd64"}, cfg.Platforms)
	assert.ElementsMatch(t, []string{"ghcr.io/acme/app:1.0.0"}, cfg.Images)
	assert.Equal(t, map[string]string{"k": "v"}, cfg.Labels)
	assert.False(t, cfg.AllowInsecure)
}

func TestValidateConfig_Various(t *testing.T) {
	dir := t.TempDir()

	// Missing file
	missing := filepath.Join(dir, "nope.json")
	err := ValidateConfig(missing)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "configuration file not found")

	// Invalid JSON
	invalidJSON := filepath.Join(dir, "bad.json")
	assert.NoError(t, os.WriteFile(invalidJSON, []byte("{ bad json }"), 0o644))
	err = ValidateConfig(invalidJSON)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse configuration")
	assert.Contains(t, err.Error(), "invalid JSON")

	// Structural validation errors (missing required fields)
	structErr := filepath.Join(dir, "struct.json")
	assert.NoError(t, os.WriteFile(structErr, []byte(`{"platforms":["linux/amd64"]}`), 0o644))
	err = ValidateConfig(structErr)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "configuration validation failed")
	assert.Contains(t, err.Error(), "baseImage is required")
	assert.Contains(t, err.Error(), "at least one image must be specified")

	// Invalid image references
	invalidRefs := filepath.Join(dir, "badrefs.json")
	assert.NoError(t, os.WriteFile(invalidRefs, []byte(`{
		"baseImage": "alpine:3.18",
		"platforms": ["linux/amd64"],
		"images": ["BadUpper:latest"]
	}`), 0o644))
	err = ValidateConfig(invalidRefs)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid image references")
	assert.Contains(t, err.Error(), "BadUpper:latest")

	// Valid config
	valid := filepath.Join(dir, "good.json")
	assert.NoError(t, os.WriteFile(valid, []byte(`{
		"baseImage": "alpine:3.18",
		"platforms": ["linux/amd64", "darwin/arm64"],
		"images": ["ghcr.io/acme/app:1.0.0"],
		"labels": {"env":"prod"},
		"allowInsecure": false
	}`), 0o644))
	err = ValidateConfig(valid)
	assert.NoError(t, err)
}
