package commands

import (
    "os/exec"
    "testing"

    "github.com/spf13/cobra"
    "github.com/stretchr/testify/assert"
)

func TestAddKubeCommands(t *testing.T) {
    rootCmd := &cobra.Command{Use: "root"}
    AddKubeCommands(rootCmd)

    commands := []string{"delete", "version", "create", "apply", "resolve", "build", "run"}
    for _, cmd := range commands {
        t.Run("CheckCommand_"+cmd, func(t *testing.T) {
            found := false
            for _, c := range rootCmd.Commands() {
                if c.Use == cmd {
                    found = true
                    break
                }
            }
            assert.True(t, found, "Command %s should be added to root command", cmd)
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
