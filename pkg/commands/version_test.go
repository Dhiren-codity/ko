package commands

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/spf13/cobra"
    "runtime/debug"
)

func TestVersion(t *testing.T) {
    originalVersion := Version
    defer func() { Version = originalVersion }()

    tests := []struct {
        name       string
        version    string
        buildInfo  *debug.BuildInfo
        buildInfoOk bool
        want       string
    }{
        {"Version set", "v1.0.0", nil, false, "v1.0.0"},
        {"Version not set, build info available", "", &debug.BuildInfo{Main: debug.Module{Version: "v1.2.3"}}, true, "v1.2.3"},
        {"Version not set, build info not available", "", nil, false, ""},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            Version = tt.version
            if tt.buildInfo != nil {
                debug.ReadBuildInfo = func() (*debug.BuildInfo, bool) {
                    return tt.buildInfo, tt.buildInfoOk
                }
            } else {
                debug.ReadBuildInfo = func() (*debug.BuildInfo, bool) {
                    return nil, tt.buildInfoOk
                }
            }

            got := version()
            assert.Equal(t, tt.want, got)
        })
    }
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

func TestVersionCommandOutput(t *testing.T) {
    originalVersion := Version
    defer func() { Version = originalVersion }()

    tests := []struct {
        name    string
        version string
        want    string
    }{
        {"Version set", "v1.0.0", "v1.0.0\n"},
        {"Version not set", "", "could not determine build information\n"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            Version = tt.version
            cmd := &cobra.Command{
                Use: "version",
                Run: func(cmd *cobra.Command, args []string) {
                    v := version()
                    if v == "" {
                        cmd.Println("could not determine build information")
                    } else {
                        cmd.Println(v)
                    }
                },
            }

            output, err := executeCommand(cmd)
            assert.NoError(t, err)
            assert.Equal(t, tt.want, output)
        })
    }
}

func executeCommand(cmd *cobra.Command, args ...string) (string, error) {
    buf := new(bytes.Buffer)
    cmd.SetOut(buf)
    cmd.SetErr(buf)
    cmd.SetArgs(args)

    _, err := cmd.ExecuteC()
    return buf.String(), err
}