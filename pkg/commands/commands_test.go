package commands

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func setupPATH(t *testing.T, dir string) func() {
	t.Helper()
	oldPath := os.Getenv("PATH")
	err := os.Setenv("PATH", dir)
	assert.NoError(t, err)

	var oldPathext string
	if runtime.GOOS == "windows" {
		oldPathext = os.Getenv("PATHEXT")
		_ = os.Setenv("PATHEXT", ".COM;.EXE;.BAT;.CMD")
	}

	return func() {
		_ = os.Setenv("PATH", oldPath)
		if runtime.GOOS == "windows" {
			_ = os.Setenv("PATHEXT", oldPathext)
		}
	}
}

func writeFile(t *testing.T, path, content string, mode os.FileMode) {
	t.Helper()
	err := os.WriteFile(path, []byte(content), mode)
	assert.NoError(t, err)
}

func createFakeCommand(t *testing.T, dir, name, content string) string {
	t.Helper()

	filename := name
	if runtime.GOOS == "windows" {
		if !strings.Contains(name, ".") {
			filename = name + ".bat"
		}
		if content == "" {
			content = "@echo off\r\n"
		}
	} else {
		if content == "" {
			content = "#!/bin/sh\n:"
		} else if !strings.HasPrefix(content, "#!") {
			content = "#!/bin/sh\n" + content
		}
	}

	path := filepath.Join(dir, filename)

	mode := os.FileMode(0644)
	if runtime.GOOS != "windows" {
		mode = 0755
	}
	writeFile(t, path, content, mode)
	if runtime.GOOS != "windows" {
		err := os.Chmod(path, 0755)
		assert.NoError(t, err)
	}
	return path
}

func kubectlScriptSuccessOutput(version string) string {
	if runtime.GOOS == "windows" {
		return "@echo off\r\necho Client Version: " + version + "\r\n"
	}
	return "#!/bin/sh\necho \"Client Version: " + version + "\"\n"
}

func kubectlScriptExitFailure() string {
	if runtime.GOOS == "windows" {
		return "@echo off\r\nexit /b 1\r\n"
	}
	return "#!/bin/sh\nexit 1\n"
}

func kubectlScriptNoOutput() string {
	if runtime.GOOS == "windows" {
		return "@echo off\r\nrem no output\r\n"
	}
	return "#!/bin/sh\n: # no output\n"
}

func TestIsCommandAvailable(t *testing.T) {
	dir := t.TempDir()
	restore := setupPATH(t, dir)
	defer restore()

	createFakeCommand(t, dir, "foo", "")

	if runtime.GOOS == "windows" {
		writeFile(t, filepath.Join(dir, "bar.txt"), "not executable", 0644)
	} else {
		writeFile(t, filepath.Join(dir, "bar"), "not executable", 0644)
	}

	tests := []struct {
		name    string
		command string
		want    bool
	}{
		{"present executable", "foo", true},
		{"absent command", "does-not-exist-12345", false},
		{"present but non-executable or wrong extension", "bar", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isCommandAvailable(tt.command)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsKubectlAvailable(t *testing.T) {
	t.Run("kubectl present", func(t *testing.T) {
		dir := t.TempDir()
		restore := setupPATH(t, dir)
		defer restore()

		createFakeCommand(t, dir, "kubectl", "")
		assert.True(t, isKubectlAvailable())
	})

	t.Run("kubectl absent", func(t *testing.T) {
		dir := t.TempDir()
		restore := setupPATH(t, dir)
		defer restore()

		assert.False(t, isKubectlAvailable())
	})
}

func TestGetKubectlVersion(t *testing.T) {
	tests := []struct {
		name        string
		script      string
		want        string
		expectError bool
	}{
		{"success returns version string", kubectlScriptSuccessOutput("v1.28.3"), "Client Version: v1.28.3", false},
		{"command fails returns error", kubectlScriptExitFailure(), "", true},
		{"empty output returns empty string and no error", kubectlScriptNoOutput(), "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			restore := setupPATH(t, dir)
			defer restore()

			createFakeCommand(t, dir, "kubectl", tt.script)

			got, err := getKubectlVersion()
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, "", got)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidateKubeCommands(t *testing.T) {
	tests := []struct {
		name        string
		setupScript string
		expectErr   bool
		errContains []string
		expectNoErr bool
	}{
		{
			name:        "kubectl missing",
			setupScript: "",
			expectErr:   true,
			errContains: []string{"kubectl is required but not found in PATH"},
		},
		{
			name:        "kubectl present and version ok",
			setupScript: kubectlScriptSuccessOutput("v1.30.0"),
			expectNoErr: true,
		},
		{
			name:        "kubectl present but version empty",
			setupScript: kubectlScriptNoOutput(),
			expectErr:   true,
			errContains: []string{"kubectl version could not be determined"},
		},
		{
			name:        "kubectl present but command fails",
			setupScript: kubectlScriptExitFailure(),
			expectErr:   true,
			errContains: []string{"failed to validate kubectl", "failed to get kubectl version"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			restore := setupPATH(t, dir)
			defer restore()

			if tt.setupScript != "" {
				createFakeCommand(t, dir, "kubectl", tt.setupScript)
			}

			err := validateKubeCommands()
			if tt.expectNoErr {
				assert.NoError(t, err)
				return
			}
			if tt.expectErr {
				assert.Error(t, err)
				for _, frag := range tt.errContains {
					assert.Contains(t, err.Error(), frag)
				}
				return
			}
		})
	}
}

func TestGetRequiredCommands(t *testing.T) {
	got := getRequiredCommands()
	want := []string{"kubectl", "docker"}
	assert.Equal(t, want, got)
}

func TestValidatePrerequisites(t *testing.T) {
	tests := []struct {
		name           string
		createKubectl  bool
		createDocker   bool
		expectedMisses []string
	}{
		{"both present", true, true, []string{}},
		{"only kubectl present", true, false, []string{"docker"}},
		{"only docker present", false, true, []string{"kubectl"}},
		{"neither present", false, false, []string{"kubectl", "docker"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			restore := setupPATH(t, dir)
			defer restore()

			if tt.createKubectl {
				createFakeCommand(t, dir, "kubectl", "")
			}
			if tt.createDocker {
				createFakeCommand(t, dir, "docker", "")
			}

			missing := validatePrerequisites()
			assert.ElementsMatch(t, tt.expectedMisses, missing)
			assert.Equal(t, len(tt.expectedMisses), len(missing))
		})
	}
}

func TestAddKubeCommands_NoPanicAndAddsSomething(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	assert.NotNil(t, root)

	AddKubeCommands(root)

	assert.NotNil(t, root.Commands())

	expected := map[string]struct{}{
		"delete":  {},
		"version": {},
		"create":  {},
		"apply":   {},
		"resolve": {},
		"build":   {},
		"run":     {},
	}
	foundAny := false
	for _, c := range root.Commands() {
		if _, ok := expected[c.Use]; ok {
			foundAny = true
			break
		}
	}
	if len(root.Commands()) > 0 {
		assert.True(t, foundAny, "expected at least one known subcommand to be added")
	}
}
