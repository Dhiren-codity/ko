package commands

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "runtime/debug"
)

func TestVersion(t *testing.T) {
    originalVersion := Version
    defer func() { Version = originalVersion }()

    tests := []struct {
        name          string
        version       string
        buildInfoFunc func() (*debug.BuildInfo, bool)
        expected      string
    }{
        {
            name:    "Version set",
            version: "v1.2.3",
            buildInfoFunc: func() (*debug.BuildInfo, bool) {
                return nil, false
            },
            expected: "v1.2.3",
        },
        {
            name:    "Version not set, build info available",
            version: "",
            buildInfoFunc: func() (*debug.BuildInfo, bool) {
                return &debug.BuildInfo{Main: debug.Module{Version: "v1.0.0"}}, true
            },
            expected: "v1.0.0",
        },
        {
            name:    "Version not set, build info not available",
            version: "",
            buildInfoFunc: func() (*debug.BuildInfo, bool) {
                return nil, false
            },
            expected: "",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            Version = tt.version
            originalReadBuildInfo := debug.ReadBuildInfo
            debug.ReadBuildInfo = tt.buildInfoFunc
            defer func() { debug.ReadBuildInfo = originalReadBuildInfo }()

            got := version()
            assert.Equal(t, tt.expected, got)
        })
    }
}