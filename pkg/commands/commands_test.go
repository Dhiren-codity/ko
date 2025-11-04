package commands

import (
    "os/exec"
    "testing"

    "github.com/spf13/cobra"
    "github.com/stretchr/testify/assert"
)

func TestAddKubeCommands(t *testing.T) {
    tests := []struct {
        name string
    }{
        {"basic test"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            rootCmd := &cobra.Command{}
            AddKubeCommands(rootCmd)

            assert.NotNil(t, rootCmd.Commands(), "commands should be added to the root command")
            assert.NotNil(t, rootCmd.Find("delete"), "delete command should be available")
            assert.NotNil(t, rootCmd.Find("version"), "version command should be available")
            assert.NotNil(t, rootCmd.Find("create"), "create command should be available")
            assert.NotNil(t, rootCmd.Find("apply"), "apply command should be available")
            assert.NotNil(t, rootCmd.Find("resolve"), "resolve command should be available")
            assert.NotNil(t, rootCmd.Find("build"), "build command should be available")
            assert.NotNil(t, rootCmd.Find("run"), "run command should be available")
        })
    }
}

func TestIsKubectlAvailable(t *testing.T) {
    tests := []struct {
        name    string
        setup   func()
        want    bool
    }{
        {
            name: "kubectl available",
            setup: func() {
                exec.Command("touch", "/usr/local/bin/kubectl").Run()
            },
            want: true,
        },
        {
            name: "kubectl not available",
            setup: func() {
                exec.Command("rm", "-f", "/usr/local/bin/kubectl").Run()
            },
            want: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tt.setup()
            got := isKubectlAvailable()
            assert.Equal(t, tt.want, got)
        })
    }
}
