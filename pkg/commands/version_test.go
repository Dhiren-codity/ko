package commands

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "runtime/debug"
)

func TestVersion(t *testing.T) {
    tests := []struct {
        name       string
        version    string
        buildInfo  *debug.BuildInfo
        buildInfoOk bool
        expected   string
    }{
        {
            name:       "Version set",
            version:    "v1.0.0",
            buildInfo:  nil,
            buildInfoOk: false,
            expected:   "v1.0.0",
        },
        {
            name:       "Version not set, build info available",
            version:    "",
            buildInfo:  &debug.BuildInfo{Main: debug.Module{Version: "v1.2.3"}},
            buildInfoOk: true,
            expected:   "v1.2.3",
        },
        {
            name:       "Version not set, build info not available",
            version:    "",
            buildInfo:  nil,
            buildInfoOk: false,
            expected:   "",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            Version = tt.version
            if tt.buildInfo != nil {
                debug.ReadBuildInfo = func() (*debug.BuildInfo, bool) {
                    return tt.buildInfo, tt.buildInfoOk
                }
            }
            got := version()
            assert.Equal(t, tt.expected, got)
        })
    }
}
