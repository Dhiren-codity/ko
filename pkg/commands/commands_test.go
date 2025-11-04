package commands

import (
    "os/exec"
    "testing"

    "github.com/spf13/cobra"
    "github.com/stretchr/testify/assert"
)

func TestAddKubeCommands(t *testing.T) {
    topLevel := &cobra.Command{}
    AddKubeCommands(topLevel)

    commands := []string{"delete", "version", "create", "apply", "resolve", "build", "run"}
    for _, cmd := range commands {
        t.Run("CheckCommand_"+cmd, func(t *testing.T) {
            found := false
            for _, c := range topLevel.Commands() {
                if c.Name() == cmd {
                    found = true
                    break
                }
            }
            assert.True(t, found, "Command %s should be added to topLevel", cmd)
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
                execCommand = func(name string, arg ...string) *exec.Cmd {
                    return exec.Command("echo", "kubectl")
                }
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

var execCommand = exec.Command

func TestMain(m *testing.M) {
    execCommand = exec.Command
    m.Run()
}