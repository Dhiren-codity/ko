package commands

import (
	"os/exec"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestIsKubectlAvailable(t *testing.T) {
	tests := []struct {
		name  string
		setup func()
		want  bool
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

func TestAddKubeCommands(t *testing.T) {
	rootCmd := &cobra.Command{Use: "test"}
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
