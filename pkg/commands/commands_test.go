package commands

import (
	"errors"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsKubectlAvailable(t *testing.T) {
	tests := []struct {
		name  string
		setup func() error
		want  bool
	}{
		{
			name: "kubectl available",
			setup: func() error {
				return nil
			},
			want: true,
		},
		{
			name: "kubectl not available",
			setup: func() error {
				return errors.New("not found")
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execLookPath = func(file string) (string, error) {
				return "", tt.setup()
			}
			got := isKubectlAvailable()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetKubectlVersion(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() ([]byte, error)
		want    string
		wantErr bool
	}{
		{
			name: "get version successfully",
			setup: func() ([]byte, error) {
				return []byte("v1.20.0"), nil
			},
			want:    "v1.20.0",
			wantErr: false,
		},
		{
			name: "error getting version",
			setup: func() ([]byte, error) {
				return nil, errors.New("error")
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execCommand = func(name string, arg ...string) *exec.Cmd {
				return &exec.Cmd{
					Output: tt.setup,
				}
			}
			got, err := getKubectlVersion()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidateKubeCommands(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() (bool, string, error)
		wantErr bool
	}{
		{
			name: "all valid",
			setup: func() (bool, string, error) {
				return true, "v1.20.0", nil
			},
			wantErr: false,
		},
		{
			name: "kubectl not available",
			setup: func() (bool, string, error) {
				return false, "", nil
			},
			wantErr: true,
		},
		{
			name: "error getting version",
			setup: func() (bool, string, error) {
				return true, "", errors.New("error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isKubectlAvailable = func() bool {
				available, _, _ := tt.setup()
				return available
			}
			getKubectlVersion = func() (string, error) {
				_, version, err := tt.setup()
				return version, err
			}
			err := validateKubeCommands()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestIsCommandAvailable(t *testing.T) {
	tests := []struct {
		name    string
		command string
		setup   func() error
		want    bool
	}{
		{
			name:    "command available",
			command: "docker",
			setup: func() error {
				return nil
			},
			want: true,
		},
		{
			name:    "command not available",
			command: "nonexistent",
			setup: func() error {
				return errors.New("not found")
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execLookPath = func(file string) (string, error) {
				return "", tt.setup()
			}
			got := isCommandAvailable(tt.command)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidatePrerequisites(t *testing.T) {
	tests := []struct {
		name  string
		setup func() []string
		want  []string
	}{
		{
			name: "all commands available",
			setup: func() []string {
				return []string{}
			},
			want: []string{},
		},
		{
			name: "some commands missing",
			setup: func() []string {
				return []string{"docker"}
			},
			want: []string{"docker"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validatePrerequisites = func() []string {
				return tt.setup()
			}
			got := validatePrerequisites()
			assert.Equal(t, tt.want, got)
		})
	}
}
