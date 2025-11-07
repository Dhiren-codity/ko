package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileExists(t *testing.T) {
	dir := t.TempDir()

	// Non-existent file
	assert.False(t, FileExists(filepath.Join(dir, "missing.json")))

	// Directory should return false
	assert.False(t, FileExists(dir))

	// Existing file
	f, err := os.CreateTemp(dir, "cfg-*.json")
	assert.NoError(t, err)
	defer f.Close()
	assert.True(t, FileExists(f.Name()))
}

func TestParseConfig(t *testing.T) {
	// Empty data
	_, err := ParseConfig(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty configuration data")

	// Invalid JSON
	_, err = ParseConfig([]byte("{"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid JSON")

	// Valid JSON
	digest := strings.Repeat("a", 64)
	data := []byte(fmt.Sprintf(`{
		"baseImage": "gcr.io/distroless/static:nonroot",
		"platforms": ["linux/amd64", "darwin/arm64"],
		"images": ["nginx:1.25", "alpine@sha256:%s"],
		"labels": {"app":"demo"},
		"allowInsecure": true
	}`, digest))
	cfg, err := ParseConfig(data)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "gcr.io/distroless/static:nonroot", cfg.BaseImage)
	assert.Equal(t, []string{"linux/amd64", "darwin/arm64"}, cfg.Platforms)
	assert.Equal(t, 2, len(cfg.Images))
	assert.Equal(t, "demo", cfg.Labels["app"])
	assert.True(t, cfg.AllowInsecure)
}

func TestIsValidPlatform(t *testing.T) {
	valid := []string{
		"linux/amd64",
		"linux/arm64",
		"linux/arm",
		"windows/386",
		"darwin/arm64",
	}
	for _, v := range valid {
		assert.Truef(t, IsValidPlatform(v), "expected valid platform: %s", v)
	}

	invalid := []string{
		"",
		"linux",
		"linux/ppc64le",
		"freebsd/amd64",
		"linux/arm/v7",
		"linux/Arm64",
		"linux//amd64",
		"linux/amd64/extra",
		"linux/",
		"/amd64",
	}
	for _, v := range invalid {
		assert.Falsef(t, IsValidPlatform(v), "expected invalid platform: %s", v)
	}
}

func TestIsValidImageReference(t *testing.T) {
	digest := strings.Repeat("a", 64)
	valid := []string{
		"nginx",
		"library/ubuntu:22.04",
		"gcr.io/distroless/static:nonroot",
		"alpine:3",
		"alpine:LATEST",
		"ghcr.io/org/repo@sha256:" + digest,
		"my_repo/name",
		"repo/name:tag-with.dots_and-hyphens",
		"docker.io/library/busybox",
		"gcr.io/project/image:tag@sha256:" + digest,
	}
	for _, v := range valid {
		assert.Truef(t, IsValidImageReference(v), "expected valid image reference: %s", v)
	}

	invalid := []string{
		"",
		"UPPER/repo",
		"alpine:tag!",
		"localhost:5000/repo",
		"nginx@sha256:short",
		"/leading/slash",
		"trailing/slash/",
	}
	for _, v := range invalid {
		assert.Falsef(t, IsValidImageReference(v), "expected invalid image reference: %s", v)
	}
}

func TestValidateImageReferences(t *testing.T) {
	// No refs provided
	err := ValidateImageReferences(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no image references provided")

	// Valid refs
	digest := strings.Repeat("a", 64)
	valid := []string{"nginx", "alpine:3.19", "ghcr.io/org/repo@sha256:" + digest}
	err = ValidateImageReferences(valid)
	assert.NoError(t, err)

	// Some invalid refs
	refs := []string{"nginx", "UPPER/NAME", "alpine:ok", "Invalid", "ghcr.io/org/repo@"}
	err = ValidateImageReferences(refs)
	assert.Error(t, err)
	msg := err.Error()
	assert.Contains(t, msg, "invalid image references:")
	assert.Contains(t, msg, "UPPER/NAME")
	assert.Contains(t, msg, "Invalid")
	assert.Contains(t, msg, "ghcr.io/org/repo@")
}

func TestCheckRequiredFields(t *testing.T) {
	// Missing both
	cfg := &Config{}
	errs := CheckRequiredFields(cfg)
	assert.Contains(t, errs, "baseImage is required")
	assert.Contains(t, errs, "at least one image must be specified")

	// BaseImage present, images empty
	cfg = &Config{BaseImage: "gcr.io/distroless/static:nonroot"}
	errs = CheckRequiredFields(cfg)
	assert.Len(t, errs, 1)
	assert.Contains(t, errs, "at least one image must be specified")

	// Images present, baseImage empty
	cfg = &Config{Images: []string{"nginx"}}
	errs = CheckRequiredFields(cfg)
	assert.Len(t, errs, 1)
	assert.Contains(t, errs, "baseImage is required")

	// All good
	cfg = &Config{BaseImage: "gcr.io/distroless/static:nonroot", Images: []string{"nginx"}}
	errs = CheckRequiredFields(cfg)
	assert.Empty(t, errs)
}

func TestValidateConfigStructure(t *testing.T) {
	// Nil cfg
	var nilCfg *Config
	errs := ValidateConfigStructure(nilCfg)
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0], "configuration is nil")

	// Valid cfg
	cfg := &Config{
		BaseImage: "gcr.io/distroless/static:nonroot",
		Images:    []string{"nginx"},
		Platforms: []string{"linux/amd64", "darwin/arm64"},
	}
	errs = ValidateConfigStructure(cfg)
	assert.Empty(t, errs)

	// Invalid platform
	cfg = &Config{
		BaseImage: "gcr.io/distroless/static:nonroot",
		Images:    []string{"nginx"},
		Platforms: []string{"linux/ppc64le", "linux/amd64"},
	}
	errs = ValidateConfigStructure(cfg)
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0], "invalid platform format: linux/ppc64le")

	// Invalid base image
	cfg = &Config{
		BaseImage: "UPPER/IMAGE",
		Images:    []string{"nginx"},
	}
	errs = ValidateConfigStructure(cfg)
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0], "invalid base image reference: UPPER/IMAGE")

	// Missing required fields should be reported
	cfg = &Config{}
	errs = ValidateConfigStructure(cfg)
	assert.Contains(t, errs, "baseImage is required")
	assert.Contains(t, errs, "at least one image must be specified")
}

func TestGetConfigPath(t *testing.T) {
	expected := filepath.Join(".", ".ko.json")
	assert.Equal(t, expected, GetConfigPath())
}

func TestLoadDefaultConfig_NoFile(t *testing.T) {
	// Work within an isolated temp dir
	origWD, err := os.Getwd()
	assert.NoError(t, err)
	defer func() { _ = os.Chdir(origWD) }()

	tmp := t.TempDir()
	assert.NoError(t, os.Chdir(tmp))

	_, err = LoadDefaultConfig()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "default configuration file not found")
}

func TestLoadDefaultConfig_Success(t *testing.T) {
	origWD, err := os.Getwd()
	assert.NoError(t, err)
	defer func() { _ = os.Chdir(origWD) }()

	tmp := t.TempDir()
	assert.NoError(t, os.Chdir(tmp))

	digest := strings.Repeat("a", 64)
	content := []byte(fmt.Sprintf(`{
		"baseImage": "gcr.io/distroless/static:nonroot",
		"platforms": ["linux/amd64"],
		"images": ["nginx", "alpine@sha256:%s"],
		"labels": {"env":"test"},
		"allowInsecure": false
	}`, digest))
	err = os.WriteFile(GetConfigPath(), content, 0o644)
	assert.NoError(t, err)

	cfg, err := LoadDefaultConfig()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "gcr.io/distroless/static:nonroot", cfg.BaseImage)
	assert.Equal(t, []string{"linux/amd64"}, cfg.Platforms)
	assert.Equal(t, []string{"nginx", "alpine@sha256:" + digest}, cfg.Images)
	assert.Equal(t, "test", cfg.Labels["env"])
	assert.False(t, cfg.AllowInsecure)
}

func TestValidateConfig(t *testing.T) {
	dir := t.TempDir()

	// Missing file
	missing := filepath.Join(dir, "missing.json")
	err := ValidateConfig(missing)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "configuration file not found")

	// Invalid JSON
	p1 := filepath.Join(dir, "bad.json")
	assert.NoError(t, os.WriteFile(p1, []byte("{"), 0o644))
	err = ValidateConfig(p1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse configuration")
	assert.Contains(t, err.Error(), "invalid JSON")

	// Invalid structure (missing required fields)
	p2 := filepath.Join(dir, "invalid-structure.json")
	assert.NoError(t, os.WriteFile(p2, []byte(`{"images":[]}`), 0o644))
	err = ValidateConfig(p2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "configuration validation failed")
	assert.Contains(t, err.Error(), "baseImage is required")
	assert.Contains(t, err.Error(), "at least one image must be specified")

	// Invalid image references
	p3 := filepath.Join(dir, "invalid-images.json")
	cfgBadImages := `{
		"baseImage": "gcr.io/distroless/static:nonroot",
		"images": ["nginx", "UPPER/NAME"]
	}`
	assert.NoError(t, os.WriteFile(p3, []byte(cfgBadImages), 0o644))
	err = ValidateConfig(p3)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid image references")
	assert.Contains(t, err.Error(), "UPPER/NAME")

	// Valid config
	digest := strings.Repeat("a", 64)
	p4 := filepath.Join(dir, "ok.json")
	cfgOK := fmt.Sprintf(`{
		"baseImage": "gcr.io/distroless/static:nonroot",
		"platforms": ["linux/amd64", "windows/386"],
		"images": ["nginx", "alpine@sha256:%s"],
		"labels": {"team":"core"},
		"allowInsecure": true
	}`, digest)
	assert.NoError(t, os.WriteFile(p4, []byte(cfgOK), 0o644))
	err = ValidateConfig(p4)
	assert.NoError(t, err)
}
