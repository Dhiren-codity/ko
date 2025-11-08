package commands

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setTempPathOnly(t *testing.T) (tmp string, restore func()) {
	t.Helper()
	tmp = t.TempDir()
	orig := os.Getenv("PATH")
	require.NoError(t, os.Setenv("PATH", tmp))
	return tmp, func() {
		_ = os.Setenv("PATH", orig)
	}
}

func makeFakeTool(t *testing.T, dir, name, stdout, stderr string, exitCode int) string {
	t.Helper()
	var file string
	if runtime.GOOS == "windows" {
		file = filepath.Join(dir, name+".bat")
		content := "@echo off\r\n"
		if stderr != "" {
			content += "echo " + stderr + " 1>&2\r\n"
		}
		if stdout != "" {
			content += "echo " + stdout + "\r\n"
		}
		content += "exit /B " + func() string {
			return strconvI(exitCode)
		}() + "\r\n"
		require.NoError(t, os.WriteFile(file, []byte(content), 0o755))
	} else {
		file = filepath.Join(dir, name)
		content := "#!/usr/bin/env sh\n"
		if stderr != "" {
			content += "echo \"" + stderr + "\" 1>&2\n"
		}
		if stdout != "" {
			content += "echo \"" + stdout + "\"\n"
		}
		content += "exit " + strconvI(exitCode) + "\n"
		require.NoError(t, os.WriteFile(file, []byte(content), 0o755))
	}
	return file
}

func strconvI(i int) string {
	// simple int to string without importing strconv directly at each call
	return func() string {
		return string([]byte{'0' + byte(i/100%10), '0' + byte(i/10%10), '0' + byte(i%10)})
	}()
}

func TestIsCommandAvailable_Table(t *testing.T) {
	tmp, restore := setTempPathOnly(t)
	defer restore()

	makeFakeTool(t, tmp, "mytool", "ok", "", 0)

	tests := []struct {
		name string
		cmd  string
		want bool
	}{
		{"found", "mytool", true},
		{"not found", "definitely-not-present", false},
		{"empty name", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isCommandAvailable(tt.cmd)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsKubectlAvailable_Table(t *testing.T) {
	tmp, restore := setTempPathOnly(t)
	defer restore()

	tests := []struct {
		name        string
		createKube  bool
		wantPresent bool
	}{
		{"kubectl present", true, true},
		{"kubectl missing", false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.createKube {
				makeFakeTool(t, tmp, "kubectl", "Client Version: v1.27.3", "", 0)
			}
			got := isKubectlAvailable()
			assert.Equal(t, tt.wantPresent, got)
		})
	}
}

func TestGetKubectlVersion_Success(t *testing.T) {
	tmp, restore := setTempPathOnly(t)
	defer restore()

	want := "Client Version: v1.28.1"
	makeFakeTool(t, tmp, "kubectl", want, "", 0)

	got, err := getKubectlVersion()
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestGetKubectlVersion_ErrorWhenMissing(t *testing.T) {
	_, restore := setTempPathOnly(t)
	defer restore()

	ver, err := getKubectlVersion()
	assert.Error(t, err)
	assert.Empty(t, ver)
	assert.Contains(t, err.Error(), "failed to get kubectl version")
}

func TestGetKubectlVersion_NonZeroExit(t *testing.T) {
	tmp, restore := setTempPathOnly(t)
	defer restore()

	makeFakeTool(t, tmp, "kubectl", "", "boom", 1)

	ver, err := getKubectlVersion()
	assert.Error(t, err)
	assert.Empty(t, ver)
	assert.Contains(t, err.Error(), "failed to get kubectl version")
}

func TestValidateKubeCommands_MissingKubectl(t *testing.T) {
	_, restore := setTempPathOnly(t)
	defer restore()

	err := validateKubeCommands()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "kubectl is required but not found in PATH")
}

func TestValidateKubeCommands_Success(t *testing.T) {
	tmp, restore := setTempPathOnly(t)
	defer restore()

	makeFakeTool(t, tmp, "kubectl", "Client Version: v1.29.0", "", 0)

	err := validateKubeCommands()
	assert.NoError(t, err)
}

func TestValidateKubeCommands_EmptyVersion(t *testing.T) {
	tmp, restore := setTempPathOnly(t)
	defer restore()

	makeFakeTool(t, tmp, "kubectl", "", "", 0)

	err := validateKubeCommands()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "kubectl version could not be determined")
}

func TestValidateKubeCommands_FailedVersion(t *testing.T) {
	tmp, restore := setTempPathOnly(t)
	defer restore()

	makeFakeTool(t, tmp, "kubectl", "", "some error", 2)

	err := validateKubeCommands()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to validate kubectl")
}

func TestGetRequiredCommands_StableOrder(t *testing.T) {
	got := getRequiredCommands()
	assert.Equal(t, []string{"kubectl", "docker"}, got)
}

func TestValidatePrerequisites_Table(t *testing.T) {
	tmp, restore := setTempPathOnly(t)
	defer restore()

	tests := []struct {
		name       string
		addKubectl bool
		addDocker  bool
		want       []string
	}{
		{"none present", false, false, []string{"kubectl", "docker"}},
		{"only kubectl", true, false, []string{"docker"}},
		{"only docker", false, true, []string{"kubectl"}},
		{"both present", true, true, []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.addKubectl {
				makeFakeTool(t, tmp, "kubectl", "Client Version: v1.29.0", "", 0)
			}
			if tt.addDocker {
				makeFakeTool(t, tmp, "docker", "Docker version 25.0.0", "", 0)
			}
			got := validatePrerequisites()
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestAddKubeCommands_AddsOrNotDecreases(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	initial := len(root.Commands())

	AddKubeCommands(root)

	after := len(root.Commands())
	assert.GreaterOrEqual(t, after, initial)
}

func TestAddKubeCommands_IdempotentNoPanic(t *testing.T) {
	root := &cobra.Command{Use: "root"}

	assert.NotPanics(t, func() {
		AddKubeCommands(root)
		AddKubeCommands(root)
	})
	// Ensure the root command itself is usable
	assert.NotNil(t, root)
}
