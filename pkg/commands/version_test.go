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
        setup         func()
        expected      string
    }{
        {
            name: "Version is set",
            setup: func() {
                Version = "v1.0.0"
            },
            expected: "v1.0.0",
        },
        {
            name: "Version is empty, build info available",
            setup: func() {
                Version = ""
                debug.ReadBuildInfo = func() (*debug.BuildInfo, bool) {
                    return &debug.BuildInfo{
                        Main: debug.Module{
                            Version: "v1.2.3",
                        },
                    }, true
                }
            },
            expected: "v1.2.3",
        },
        {
            name: "Version is empty, build info not available",
            setup: func() {
                Version = ""
                debug.ReadBuildInfo = func() (*debug.BuildInfo, bool) {
                    return nil, false
                }
            },
            expected: "",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tt.setup()
            result := version()
            assert.Equal(t, tt.expected, result)
        })
    }
}