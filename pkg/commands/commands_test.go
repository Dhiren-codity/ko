package commands

import (
    "testing"
    "github.com/spf13/cobra"
    "github.com/stretchr/testify/assert"
)

func TestAddKubeCommands(t *testing.T) {
    rootCmd := &cobra.Command{Use: "root"}
    AddKubeCommands(rootCmd)

    expectedCommands := []string{"delete", "version", "create", "apply", "resolve", "build", "run"}
    for _, cmdName := range expectedCommands {
        t.Run("command "+cmdName, func(t *testing.T) {
            cmd, _, err := rootCmd.Find([]string{cmdName})
            assert.NoError(t, err)
            assert.NotNil(t, cmd)
            assert.Equal(t, cmdName, cmd.Name())
        })
    }
}
