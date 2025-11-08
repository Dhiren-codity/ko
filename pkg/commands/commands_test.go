package commands

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsCommandAvailable_NotFound(t *testing.T) {
	assert.False(t, isCommandAvailable("this-command-should-not-exist-xyz-123"))
}

func TestIsCommandAvailable_FoundWithTempExecutable(t *testing.T) {
	dir := t.TempDir()
	makeFakeExecutable(t, dir, "kubectl", "echo ok", 0)

	origPath := os.Getenv("PATH")
	t.Setenv("PATH", dir+string(os.PathListSeparator)+origPath)

	assert.True(t, isCommandAvailable("kubectl"))
}

func TestIsKubectlAvailable(t *testing.T) {
	t.Run("false when not in PATH", func(t *testing.T) {
		// Remove PATH so nothing can be found
		t.Setenv("PATH", "")
		assert.False(t, isKubectlAvailable())
	})

	t.Run("true when in PATH", func(t *testing.T) {
		dir := t.TempDir()
		makeFakeExecutable(t, dir, "kubectl", "echo ok", 0)
		t.Setenv("PATH", dir)
		assert.True(t, isKubectlAvailable())
	})
}

func TestGetKubectlVersion_Success(t *testing.T) {
	dir := t.TempDir()
	expected := "v1.28.3"
	makeFakeExecutable(t, dir, "kubectl", echoCommand(expected), 0)
	t.Setenv("PATH", dir)

	got, err := getKubectlVersion()
	require.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestGetKubectlVersion_FailureFromExitCode(t *testing.T) {
	dir := t.TempDir()
	makeFakeExecutable(t, dir, "kubectl", echoCommand("ignored output"), 1)
	t.Setenv("PATH", dir)

	got, err := getKubectlVersion()
	assert.Error(t, err)
	assert.Empty(t, got)
	assert.Contains(t, err.Error(), "failed to get kubectl version")
}

func TestValidateKubeCommands(t *testing.T) {
	t.Run("missing kubectl", func(t *testing.T) {
		t.Setenv("PATH", "")
		err := validateKubeCommands()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "kubectl is required but not found in PATH")
	})

	t.Run("kubectl present returns non-empty version", func(t *testing.T) {
		dir := t.TempDir()
		makeFakeExecutable(t, dir, "kubectl", echoCommand("Client Version: v1.29.0"), 0)
		t.Setenv("PATH", dir)

		err := validateKubeCommands()
		assert.NoError(t, err)
	})

	t.Run("kubectl version empty string", func(t *testing.T) {
		dir := t.TempDir()
		// Print empty line
		makeFakeExecutable(t, dir, "kubectl", echoCommand(""), 0)
		t.Setenv("PATH", dir)

		err := validateKubeCommands()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "kubectl version could not be determined")
	})
}

func TestGetRequiredCommands(t *testing.T) {
	want := []string{"kubectl", "docker"}
	got := getRequiredCommands()
	assert.Equal(t, want, got)
}

func TestValidatePrerequisites_AllMissing(t *testing.T) {
	t.Setenv("PATH", "")
	missing := validatePrerequisites()
	// Order should be as defined in getRequiredCommands
	assert.Equal(t, []string{"kubectl", "docker"}, missing)
}

func TestAddKubeCommands_NoPanic(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	assert.NotPanics(t, func() {
		AddKubeCommands(root)
	})
}

/************ helpers ************/

// makeFakeExecutable creates a fake executable named `name` in `dir` with the provided
// shell/batch body and exit code.
// On Unix it writes a sh script; on Windows it writes a .bat script.
func makeFakeExecutable(t *testing.T, dir, name, body string, exitCode int) string {
	t.Helper()

	if runtime.GOOS == "windows" {
		// Write a .bat file
		path := filepath.Join(dir, name+".bat")
		content := "@echo off\r\n"
		if body != "" {
			content += body + "\r\n"
		}
		// Ensure proper exit code
		content += "exit /b " + itoa(exitCode) + "\r\n"
		require.NoError(t, os.WriteFile(path, []byte(content), 0o755))
		return path
	}

	// Unix-like
	path := filepath.Join(dir, name)
	content := "#!/bin/sh\n"
	if body != "" {
		content += body + "\n"
	}
	content += "exit " + itoa(exitCode) + "\n"
	require.NoError(t, os.WriteFile(path, []byte(content), 0o755))
	require.NoError(t, os.Chmod(path, 0o755))
	return path
}

// echoCommand returns a platform-appropriate echo command to print the given message.
func echoCommand(msg string) string {
	if runtime.GOOS == "windows" {
		// In batch, echo without quotes prints as-is
		if msg == "" {
			// echo. prints a blank line
			return "echo."
		}
		// Escape any special characters minimally
		escaped := strings.ReplaceAll(msg, "%", "%%")
		return "echo " + escaped
	}
	// Unix shell: use printf to avoid adding extra newline unless desired
	// But we want a newline so getKubectlVersion TrimSpace trims it
	return "echo \"" + escapeDoubleQuotes(msg) + "\""
}

func escapeDoubleQuotes(s string) string {
	return strings.ReplaceAll(s, "\"", "\\\"")
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	sign := ""
	if n < 0 {
		sign = "-"
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return sign + string(b[i:])
}
