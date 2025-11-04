package commands

import (
	"runtime/debug"
	"testing"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
	}{
		{"version set", "v1.2.3", "v1.2.3"},
		{"version not set", "", "(devel)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Version = tt.version
			got := version()
			if tt.version == "" {
				assert.NotEmpty(t, got, "version should be set in CI")
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestVersionWithBuildInfo(t *testing.T) {
	originalVersion := Version
	defer func() { Version = originalVersion }()

	Version = ""
	i, ok := debug.ReadBuildInfo()
	if ok {
		Version = i.Main.Version
	}
	got := version()
	assert.Equal(t, Version, got)
}

func TestAddVersion(t *testing.T) {
	rootCmd := &cobra.Command{Use: "test"}
	addVersion(rootCmd)

	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "version" {
			found = true
			break
		}
	}
	assert.True(t, found, "version command should be added to the root command")
}
