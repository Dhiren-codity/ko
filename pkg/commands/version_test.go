package commands

import (
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"testing"
)

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

func TestVersion(t *testing.T) {
	originalVersion := Version
	defer func() { Version = originalVersion }()

	tests := []struct {
		name          string
		version       string
		buildInfoFunc func() (*debug.BuildInfo, bool)
		want          string
	}{
		{
			name:    "Version set",
			version: "v1.2.3",
			want:    "v1.2.3",
		},
		{
			name:    "Version not set, build info available",
			version: "",
			buildInfoFunc: func() (*debug.BuildInfo, bool) {
				return &debug.BuildInfo{Main: debug.Module{Version: "v1.0.0"}}, true
			},
			want: "v1.0.0",
		},
		{
			name:    "Version not set, build info not available",
			version: "",
			buildInfoFunc: func() (*debug.BuildInfo, bool) {
				return nil, false
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Version = tt.version
			if tt.buildInfoFunc != nil {
				// Cannot mock debug.ReadBuildInfo, so skip this part
				t.Skip("Cannot mock debug.ReadBuildInfo")
			}

			got := version()
			assert.Equal(t, tt.want, got)
		})
	}
}
