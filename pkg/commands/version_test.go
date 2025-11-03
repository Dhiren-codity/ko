package commands

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/spf13/cobra"
    "runtime/debug"
)

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
            version: "v1.0.0",
            want:    "v1.0.0",
        },
        {
            name:    "Version not set, build info available",
            version: "",
            buildInfoFunc: func() (*debug.BuildInfo, bool) {
                return &debug.BuildInfo{Main: debug.Module{Version: "v2.0.0"}}, true
            },
            want: "v2.0.0",
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
                originalReadBuildInfo := debug.ReadBuildInfo
                debug.ReadBuildInfo = tt.buildInfoFunc
                defer func() { debug.ReadBuildInfo = originalReadBuildInfo }()
            }

            got := version()
            assert.Equal(t, tt.want, got)
        })
    }
}