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

    tests := []struct {
        name     string
        cmdName  string
        expected bool
    }{
        {"delete command", "delete", true},
        {"version command", "version", true},
        {"create command", "create", true},
        {"apply command", "apply", true},
        {"resolve command", "resolve", true},
        {"build command", "build", true},
        {"run command", "run", true},
        {"non-existent command", "nonexistent", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cmd, _, err := rootCmd.Find([]string{tt.cmdName})
            if tt.expected {
                assert.NoError(t, err)
                assert.NotNil(t, cmd)
            } else {
                assert.Error(t, err)
                assert.Nil(t, cmd)
            }
        })
    }
}

func TestIsKubectlAvailable(t *testing.T) {
    tests := []struct {
        name     string
        setup    func()
        expected bool
    }{
        {
            "kubectl available",
            func() {
                execCommand = func(name string, arg ...string) *exec.Cmd {
                    return exec.Command("echo", "kubectl")
                }
            },
            true,
        },
        {
            "kubectl not available",
            func() {
                execCommand = func(name string, arg ...string) *exec.Cmd {
                    return exec.Command("false")
                }
            },
            false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tt.setup()
            available := isKubectlAvailable()
            assert.Equal(t, tt.expected, available)
        })
    }
}

var execCommand = exec.Command

func TestMain(m *testing.M) {
    execCommand = exec.Command
    m.Run()
}
