package commands

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/spf13/cobra"
    "runtime/debug"
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
    tests := []struct {
        name    string
        version string
        buildInfo *debug.BuildInfo
        want    string
    }{
        {"Version set", "v1.0.0", nil, "v1.0.0"},
        {"Version not set, build info available", "", &debug.BuildInfo{Main: debug.Module{Version: "v1.2.3"}}, "v1.2.3"},
        {"Version not set, build info not available", "", nil, ""},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            Version = tt.version
            if tt.buildInfo != nil {
                debug.ReadBuildInfo = func() (*debug.BuildInfo, bool) {
                    return tt.buildInfo, true
                }
            } else {
                debug.ReadBuildInfo = func() (*debug.BuildInfo, bool) {
                    return nil, false
                }
            }

            got := version()
            assert.Equal(t, tt.want, got)
        })
    }
}
