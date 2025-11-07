package commands

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateKubeCommands(t *testing.T) {
	tests := []struct {
		name     string
		setup    func()
		teardown func()
		wantErr  bool
	}{
		{
			name: "all commands valid",
			setup: func() {
				originalLookPath := exec.LookPath
				exec.LookPath = func(file string) (string, error) {
					return "/usr/bin/kubectl", nil
				}
				originalCommand := exec.Command
				exec.Command = func(name string, arg ...string) *exec.Cmd {
					return exec.Command("echo", "Client Version: v1.20.0")
				}
				defer func() {
					exec.LookPath = originalLookPath
					exec.Command = originalCommand
				}()
			},
			teardown: func() {},
			wantErr:  false,
		},
		{
			name: "kubectl not available",
			setup: func() {
				originalLookPath := exec.LookPath
				exec.LookPath = func(file string) (string, error) {
					return "", exec.ErrNotFound
				}
				defer func() { exec.LookPath = originalLookPath }()
			},
			teardown: func() {},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			err := validateKubeCommands()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			tt.teardown()
		})
	}
}

func TestIsCommandAvailable(t *testing.T) {
	tests := []struct {
		name     string
		cmd      string
		setup    func()
		teardown func()
		want     bool
	}{
		{
			name: "command available",
			cmd:  "kubectl",
			setup: func() {
				originalLookPath := exec.LookPath
				exec.LookPath = func(file string) (string, error) {
					return "/usr/bin/kubectl", nil
				}
				defer func() { exec.LookPath = originalLookPath }()
			},
			teardown: func() {},
			want:     true,
		},
		{
			name: "command not available",
			cmd:  "nonexistent",
			setup: func() {
				originalLookPath := exec.LookPath
				exec.LookPath = func(file string) (string, error) {
					return "", exec.ErrNotFound
				}
				defer func() { exec.LookPath = originalLookPath }()
			},
			teardown: func() {},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			got := isCommandAvailable(tt.cmd)
			assert.Equal(t, tt.want, got)
			tt.teardown()
		})
	}
}

func TestValidatePrerequisites(t *testing.T) {
	tests := []struct {
		name     string
		setup    func()
		teardown func()
		want     []string
	}{
		{
			name: "all commands available",
			setup: func() {
				originalLookPath := exec.LookPath
				exec.LookPath = func(file string) (string, error) {
					return "/usr/bin/" + file, nil
				}
				defer func() { exec.LookPath = originalLookPath }()
			},
			teardown: func() {},
			want:     []string{},
		},
		{
			name: "some commands missing",
			setup: func() {
				originalLookPath := exec.LookPath
				exec.LookPath = func(file string) (string, error) {
					if file == "kubectl" {
						return "/usr/bin/kubectl", nil
					}
					return "", exec.ErrNotFound
				}
				defer func() { exec.LookPath = originalLookPath }()
			},
			teardown: func() {},
			want:     []string{"docker"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			got := validatePrerequisites()
			assert.Equal(t, tt.want, got)
			tt.teardown()
		})
	}
}
