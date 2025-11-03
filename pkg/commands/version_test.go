package commands

import (
    "bytes"
    "fmt"
    "testing"

    "github.com/spf13/cobra"
    "github.com/stretchr/testify/assert"
)

func TestAddVersion(t *testing.T) {
    var buf bytes.Buffer
    rootCmd := &cobra.Command{
        Use: "test",
        Run: func(cmd *cobra.Command, args []string) {
            fmt.Fprint(&buf, "root command")
        },
    }

    addVersion(rootCmd)

    tests := []struct {
        name    string
        args    []string
        wantOut string
    }{
        {"version command", []string{"version"}, "could not determine build information\n"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            buf.Reset()
            rootCmd.SetArgs(tt.args)
            err := rootCmd.Execute()
            assert.NoError(t, err)
            assert.Equal(t, tt.wantOut, buf.String())
        })
    }
}

func TestVersion(t *testing.T) {
    originalVersion := Version
    defer func() { Version = originalVersion }()

    tests := []struct {
        name    string
        version string
        want    string
    }{
        {"empty version", "", ""},
        {"non-empty version", "v1.0.0", "v1.0.0"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            Version = tt.version
            got := version()
            assert.Equal(t, tt.want, got)
        })
    }
}