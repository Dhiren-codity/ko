package commands

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "runtime/debug"
)

func TestVersion(t *testing.T) {
    originalVersion := Version
    defer func() { Version = originalVersion }()

    Version = "v1.2.3"
    got := version()
    assert.Equal(t, "v1.2.3", got)
}

func TestVersionWithBuildInfo(t *testing.T) {
    originalVersion := Version
    defer func() { Version = originalVersion }()

    Version = ""
    got := version()
    if got == "" {
        t.Skip("Cannot mock debug.ReadBuildInfo - stdlib function")
    }
    assert.NotEmpty(t, got)
}

func TestFormatVersion(t *testing.T) {
    tests := []struct {
        name string
        input string
        want string
    }{
        {"empty version", "", "unknown"},
        {"simple version", "v1.0.0", "v1.0.0"},
        {"version with build info", "v1.0.0+build123", "v1.0.0"},
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
        name string
        input string
        want bool
    }{
        {"empty version", "", false},
        {"unknown version", "unknown", false},
        {"valid semver", "v1.0.0", true},
        {"valid semver with pre-release", "v1.0.0-alpha", true},
        {"valid semver with build", "v1.0.0+build123", true},
        {"invalid semver", "1.0", false},
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
        name string
        version string
        want string
    }{
        {"empty version", "", "ko version unknown (non-standard)"},
        {"valid version", "v1.0.0", "ko version v1.0.0 (valid)"},
        {"non-standard version", "1.0", "ko version 1.0 (non-standard)"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            Version = tt.version
            got := getVersionInfo()
            assert.Equal(t, tt.want, got)
        })
    }
}
