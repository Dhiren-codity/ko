package commands

import (
    "os/exec"
    "testing"

    "github.com/spf13/cobra"
    "github.com/stretchr/testify/assert"
)

func TestIsKubectlAvailable(t *testing.T) {
    tests := []struct {
        name    string
        setup   func()
        teardown func()
        want    bool
    }{
        {
            name: "kubectl available",
            setup: func() {
                execCommand = func(name string, arg ...string) *exec.Cmd {
                    return exec.Command("echo", "kubectl")
                }
            },
            teardown: func() {
                execCommand = exec.Command
            },
            want: true,
        },
        {
            name: "kubectl not available",
            setup: func() {
                execCommand = func(name string, arg ...string) *exec.Cmd {
                    return exec.Command("false")
                }
            },
            teardown: func() {
                execCommand = exec.Command
            },
            want: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tt.setup()
            defer tt.teardown()
            got := isKubectlAvailable()
            assert.Equal(t, tt.want, got)
        })
    }
}

func TestAddKubeCommands(t *testing.T) {
    rootCmd := &cobra.Command{Use: "test"}
    AddKubeCommands(rootCmd)

    subCommands := []string{"delete", "version", "create", "apply", "resolve", "build", "run"}
    for _, cmd := range subCommands {
        t.Run("check "+cmd+" command", func(t *testing.T) {
            found := false
            for _, c := range rootCmd.Commands() {
                if c.Use == cmd {
                    found = true
                    break
                }
            }
            assert.True(t, found, "command "+cmd+" should be added to root command")
        })
    }
}