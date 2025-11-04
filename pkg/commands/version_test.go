package commands

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "runtime/debug"
    "github.com/spf13/cobra"
    "bytes"
    "os"
)

func TestVersion(t *testing.T) {
    tests := []struct {
        name    string
        version string
        want    string
    }{
        {"Version set", "v1.2.3", "v1.2.3"},
        {"Version not set", "", ""},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            Version = tt.version
            got := version()
            assert.Equal(t, tt.want, got)
        })
    }
}

func TestAddVersion(t *testing.T) {
    rootCmd := &cobra.Command{}
    addVersion(rootCmd)

    found := false
    for _, cmd := range rootCmd.Commands() {
        if cmd.Use == "version" {
            found = true
            break
        }
    }
    assert.True(t, found, "version command should be added to root command")
}

func TestVersionCommandOutput(t *testing.T) {
    rootCmd := &cobra.Command{}
    addVersion(rootCmd)

    buf := new(bytes.Buffer)
    rootCmd.SetOut(buf)
    rootCmd.SetArgs([]string{"version"})

    Version = "v1.2.3"
    err := rootCmd.Execute()
    assert.NoError(t, err)
    assert.Equal(t, "v1.2.3\n", buf.String())
}

func TestVersionCommandOutputNoVersion(t *testing.T) {
    rootCmd := &cobra.Command{}
    addVersion(rootCmd)

    buf := new(bytes.Buffer)
    rootCmd.SetOut(buf)
    rootCmd.SetArgs([]string{"version"})

    Version = ""
    err := rootCmd.Execute()
    assert.NoError(t, err)
    assert.Equal(t, "could not determine build information\n", buf.String())
}

func TestVersionWithBuildInfo(t *testing.T) {
    Version = ""
    i, ok := debug.ReadBuildInfo()
    if !ok {
        t.Skip("Build info not available")
    }
    got := version()
    assert.Equal(t, i.Main.Version, got)
}

func TestVersionCommandIntegration(t *testing.T) {
    rootCmd := &cobra.Command{}
    addVersion(rootCmd)

    buf := new(bytes.Buffer)
    rootCmd.SetOut(buf)
    rootCmd.SetArgs([]string{"version"})

    Version = "v1.2.3"
    err := rootCmd.Execute()
    assert.NoError(t, err)
    assert.Equal(t, "v1.2.3\n", buf.String())
}

func TestVersionCommandIntegrationNoVersion(t *testing.T) {
    rootCmd := &cobra.Command{}
    addVersion(rootCmd)

    buf := new(bytes.Buffer)
    rootCmd.SetOut(buf)
    rootCmd.SetArgs([]string{"version"})

    Version = ""
    err := rootCmd.Execute()
    assert.NoError(t, err)
    assert.Equal(t, "could not determine build information\n", buf.String())
}

func TestMain(m *testing.M) {
    os.Exit(m.Run())
}
