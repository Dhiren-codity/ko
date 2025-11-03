package commands

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/spf13/cobra"
    "runtime/debug"
    "fmt"
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
            buildInfoFunc: func() (*debug.BuildInfo, bool) {
                return nil, false
            },
            want: "v1.0.0",
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
            debug.ReadBuildInfo = tt.buildInfoFunc

            got := version()
            assert.Equal(t, tt.want, got)
        })
    }
}

func TestVersionCommandOutput(t *testing.T) {
    originalVersion := Version
    defer func() { Version = originalVersion }()

    tests := []struct {
        name          string
        version       string
        buildInfoFunc func() (*debug.BuildInfo, bool)
        wantOutput    string
    }{
        {
            name:    "Version set",
            version: "v1.0.0",
            buildInfoFunc: func() (*debug.BuildInfo, bool) {
                return nil, false
            },
            wantOutput: "v1.0.0\n",
        },
        {
            name:    "Version not set, build info available",
            version: "",
            buildInfoFunc: func() (*debug.BuildInfo, bool) {
                return &debug.BuildInfo{Main: debug.Module{Version: "v2.0.0"}}, true
            },
            wantOutput: "v2.0.0\n",
        },
        {
            name:    "Version not set, build info not available",
            version: "",
            buildInfoFunc: func() (*debug.BuildInfo, bool) {
                return nil, false
            },
            wantOutput: "could not determine build information\n",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            Version = tt.version
            debug.ReadBuildInfo = tt.buildInfoFunc

            cmd := &cobra.Command{
                Use: "version",
                Run: func(cmd *cobra.Command, args []string) {
                    v := version()
                    if v == "" {
                        fmt.Println("could not determine build information")
                    } else {
                        fmt.Println(v)
                    }
                },
            }

            output := captureOutput(func() {
                cmd.Execute()
            })

            assert.Equal(t, tt.wantOutput, output)
        })
    }
}

func captureOutput(f func()) string {
    old := fmt.Print
    defer func() { fmt.Print = old }()
    var buf []byte
    fmt.Print = func(a ...interface{}) (n int, err error) {
        buf = append(buf, []byte(fmt.Sprint(a...))...)
        return len(buf), nil
    }
    f()
    return string(buf)
}