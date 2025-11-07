package commands

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "os/exec"
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
            setup: func() {},
            teardown: func() {},
            want: true,
        },
        {
            name: "kubectl not available",
            setup: func() {
                exec.LookPath = func(file string) (string, error) {
                    return "", exec.ErrNotFound
                }
            },
            teardown: func() {
                exec.LookPath = exec.LookPath
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

func TestGetKubectlVersion(t *testing.T) {
    tests := []struct {
        name    string
        setup   func()
        teardown func()
        want    string
        wantErr bool
    }{
        {
            name: "get version successfully",
            setup: func() {
                exec.Command = func(name string, arg ...string) *exec.Cmd {
                    return &exec.Cmd{
                        Path:   "/bin/echo",
                        Args:   []string{"echo", "Client Version: v1.20.0"},
                        Stdout: nil,
                        Stderr: nil,
                    }
                }
            },
            teardown: func() {
                exec.Command = exec.Command
            },
            want:    "Client Version: v1.20.0",
            wantErr: false,
        },
        {
            name: "error getting version",
            setup: func() {
                exec.Command = func(name string, arg ...string) *exec.Cmd {
                    return &exec.Cmd{
                        Path:   "/bin/false",
                        Args:   []string{"false"},
                        Stdout: nil,
                        Stderr: nil,
                    }
                }
            },
            teardown: func() {
                exec.Command = exec.Command
            },
            want:    "",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tt.setup()
            defer tt.teardown()
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
        setup   func()
        teardown func()
        wantErr bool
    }{
        {
            name: "all commands valid",
            setup: func() {
                exec.LookPath = func(file string) (string, error) {
                    return "/usr/bin/kubectl", nil
                }
                exec.Command = func(name string, arg ...string) *exec.Cmd {
                    return &exec.Cmd{
                        Path:   "/bin/echo",
                        Args:   []string{"echo", "Client Version: v1.20.0"},
                        Stdout: nil,
                        Stderr: nil,
                    }
                }
            },
            teardown: func() {
                exec.LookPath = exec.LookPath
                exec.Command = exec.Command
            },
            wantErr: false,
        },
        {
            name: "kubectl not available",
            setup: func() {
                exec.LookPath = func(file string) (string, error) {
                    return "", exec.ErrNotFound
                }
            },
            teardown: func() {
                exec.LookPath = exec.LookPath
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tt.setup()
            defer tt.teardown()
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
        cmd     string
        setup   func()
        teardown func()
        want    bool
    }{
        {
            name: "command available",
            cmd:  "kubectl",
            setup: func() {
                exec.LookPath = func(file string) (string, error) {
                    return "/usr/bin/kubectl", nil
                }
            },
            teardown: func() {
                exec.LookPath = exec.LookPath
            },
            want: true,
        },
        {
            name: "command not available",
            cmd:  "nonexistent",
            setup: func() {
                exec.LookPath = func(file string) (string, error) {
                    return "", exec.ErrNotFound
                }
            },
            teardown: func() {
                exec.LookPath = exec.LookPath
            },
            want: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tt.setup()
            defer tt.teardown()
            got := isCommandAvailable(tt.cmd)
            assert.Equal(t, tt.want, got)
        })
    }
}

func TestValidatePrerequisites(t *testing.T) {
    tests := []struct {
        name    string
        setup   func()
        teardown func()
        want    []string
    }{
        {
            name: "all commands available",
            setup: func() {
                exec.LookPath = func(file string) (string, error) {
                    return "/usr/bin/" + file, nil
                }
            },
            teardown: func() {
                exec.LookPath = exec.LookPath
            },
            want: []string{},
        },
        {
            name: "some commands missing",
            setup: func() {
                exec.LookPath = func(file string) (string, error) {
                    if file == "kubectl" {
                        return "/usr/bin/kubectl", nil
                    }
                    return "", exec.ErrNotFound
                }
            },
            teardown: func() {
                exec.LookPath = exec.LookPath
            },
            want: []string{"docker"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tt.setup()
            defer tt.teardown()
            got := validatePrerequisites()
            assert.Equal(t, tt.want, got)
        })
    }
}
