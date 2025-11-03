package commands

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/spf13/cobra"
    "runtime/debug"
    "strings"
)

func TestVersion(t *testing.T) {
    tests := []struct {
        name       string
        version    string
        buildInfo  *debug.BuildInfo
        wantOutput string
    }{
        {
            name:       "Version set",
            version:    "v1.2.3",
            buildInfo:  nil,
            wantOutput: "v1.2.3",
        },
        {
            name:       "Version not set, build info available",
            version:    "",
            buildInfo:  &debug.BuildInfo{Main: debug.Module{Version: "v1.0.0"}},
            wantOutput: "v1.0.0",
        },
        {
            name:       "Version not set, build info not available",
            version:    "",
            buildInfo:  nil,
            wantOutput: "",
        },
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
            assert.Equal(t, tt.wantOutput, got)
        })
    }
}

func TestAddVersion(t *testing.T) {
    var output string
    cmd := &cobra.Command{
        Use: "test",
        Run: func(cmd *cobra.Command, args []string) {
            output = strings.Join(args, " ")
        },
    }

    addVersion(cmd)

    versionCmd, _, err := cmd.Find([]string{"version"})
    assert.NoError(t, err)
    assert.NotNil(t, versionCmd)

    versionCmd.Run(versionCmd, []string{})
    assert.Contains(t, output, "could not determine build information")
}
