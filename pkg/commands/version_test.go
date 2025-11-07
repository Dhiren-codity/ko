package commands

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestVersion(t *testing.T) {
	originalVersion := Version
	defer func() { Version = originalVersion }()

	Version = "v1.2.3"
	got := version()
	assert.Equal(t, "v1.2.3", got)

	Version = ""
	got = version()
	assert.NotEmpty(t, got)
}

func TestFormatVersion(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty version", "", "unknown"},
		{"simple version", "v1.2.3", "v1.2.3"},
		{"version with build info", "v1.2.3+build.123", "v1.2.3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatVersion(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsValidVersion(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"empty version", "", false},
		{"unknown version", "unknown", false},
		{"valid semver", "v1.2.3", true},
		{"valid semver with pre-release", "v1.2.3-alpha", true},
		{"valid semver with build", "v1.2.3+build.123", true},
		{"invalid semver", "version1.2.3", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidVersion(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetVersionInfo(t *testing.T) {
	originalVersion := Version
	defer func() { Version = originalVersion }()

	tests := []struct {
		name    string
		version string
		want    string
	}{
		{"empty version", "", "ko version unknown (non-standard)"},
		{"valid version", "v1.2.3", "ko version v1.2.3 (valid)"},
		{"non-standard version", "version1.2.3", "ko version version1.2.3 (non-standard)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Version = tt.version
			got := getVersionInfo()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAddVersion(t *testing.T) {
	cmd := &cobra.Command{}
	addVersion(cmd)

	found := false
	for _, c := range cmd.Commands() {
		if c.Use == "version" {
			found = true
			break
		}
	}
	assert.True(t, found, "version command should be added")
}
