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
            topLevel := &cobra.Command{}
            AddKubeCommands(topLevel)

            assert.NotNil(t, topLevel)
            assert.NotNil(t, topLevel.Commands())
            assert.Greater(t, len(topLevel.Commands()), 0)
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