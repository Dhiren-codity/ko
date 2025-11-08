package commands

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)

	os.Stdout = w
	defer func() { os.Stdout = origStdout }()

	fn()

	require.NoError(t, w.Close())
	out, err := io.ReadAll(r)
	require.NoError(t, err)
	require.NoError(t, r.Close())
	return string(out)
}

func TestIsValidVersion_Table(t *testing.T) {
	tests := []struct {
		name string
		in   string
		ok   bool
	}{
		{"v-prefix valid", "v1.2.3", true},
		{"no prefix valid", "1.2.3", true},
		{"missing patch", "1.2", false},
		{"empty", "", false},
		{"unknown literal", "unknown", false},
		{"prerelease", "v1.2.3-beta.1", true},
		{"build metadata", "v1.2.3+build.1", true},
		{"prerelease and build", "v1.2.3-beta.1+build.1", true},
		{"nonsense", "not-a-version", false},
		{"dangling prerelease dash", "v1.2.3-", false},
		{"dangling build plus", "v1.2.3+", false},
		{"leading zeros allowed", "v01.02.03", true},
		{"missing dash before rc", "v1.2.3rc1", false},
		{"underscore in prerelease", "v1.2.3-rc_1", true},
		{"dot-ending prerelease", "v1.2.3-rc.", true},
		{"multiple build plus segments invalid", "v1.2.3+build+extra", false},
		{"devel", "devel", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidVersion(tt.in)
			assert.Equal(t, tt.ok, got, "input: %q", tt.in)
		})
	}
}

func TestFormatVersion_Table(t *testing.T) {
	tests := []struct {
		name string
		in   string
		out  string
	}{
		{"empty -> unknown", "", "unknown"},
		{"plain semver", "v1.2.3", "v1.2.3"},
		{"plain no prefix", "1.2.3", "1.2.3"},
		{"with build metadata", "v1.2.3+exp.sha.5114f85", "v1.2.3"},
		{"prerelease with build metadata", "1.2.3-beta.1+build.5", "1.2.3-beta.1"},
		{"non-semver passthrough", "devel", "devel"},
		{"multiple plus signs", "custom+abc+def", "custom"},
		{"trailing plus", "abc+", "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatVersion(tt.in)
			assert.Equal(t, tt.out, got)
		})
	}
}

func TestFormatVersion_DropsOnlyFirstPlus(t *testing.T) {
	in := "x+y+z"
	out := formatVersion(in)
	assert.Equal(t, "x", out)
}

func TestIsValidVersion_VPrefixOptional(t *testing.T) {
	assert.True(t, isValidVersion("v1.0.0"))
	assert.True(t, isValidVersion("1.0.0"))
}

func TestIsValidVersion_InvalidExtraPlusOrDangling(t *testing.T) {
	assert.False(t, isValidVersion("v1.0.0+"))
	assert.False(t, isValidVersion("v1.0.0-"))
	assert.False(t, isValidVersion("v1.0.0+build+extra"))
}

func TestVersion_UsesGlobalWhenSet(t *testing.T) {
	orig := Version
	defer func() { Version = orig }()

	Version = "test-version-123"
	got := version()
	assert.Equal(t, "test-version-123", got)
	assert.Equal(t, "test-version-123", Version)
}

func TestVersion_ReadsBuildInfoWhenEmpty(t *testing.T) {
	orig := Version
	defer func() { Version = orig }()

	Version = ""
	got1 := version()
	// It should set the global Version to whatever build info provides (possibly "(devel)")
	assert.Equal(t, Version, got1)

	// Subsequent calls should return the same value since Version is now set
	got2 := version()
	assert.Equal(t, got1, got2)
}

func TestGetVersionInfo_WithValidSemver(t *testing.T) {
	orig := Version
	defer func() { Version = orig }()

	Version = "v1.2.3"
	got := getVersionInfo()
	assert.Equal(t, "ko version v1.2.3 (valid)", got)
}

func TestGetVersionInfo_WithBuildMetadata(t *testing.T) {
	orig := Version
	defer func() { Version = orig }()

	Version = "v1.2.3+meta.build-1"
	got := getVersionInfo()
	assert.Equal(t, "ko version v1.2.3 (valid)", got)
}

func TestGetVersionInfo_WithPrereleaseAndBuild(t *testing.T) {
	orig := Version
	defer func() { Version = orig }()

	Version = "1.2.3-beta.2+build.5"
	got := getVersionInfo()
	assert.Equal(t, "ko version 1.2.3-beta.2 (valid)", got)
}

func TestGetVersionInfo_NonStandard_Devel(t *testing.T) {
	orig := Version
	defer func() { Version = orig }()

	Version = "devel"
	got := getVersionInfo()
	assert.Equal(t, "ko version devel (non-standard)", got)
}

func TestGetVersionInfo_UnknownLiteral(t *testing.T) {
	orig := Version
	defer func() { Version = orig }()

	Version = "unknown"
	got := getVersionInfo()
	assert.Equal(t, "ko version unknown (non-standard)", got)
}

func TestAddVersion_AddsSubcommand(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	addVersion(root)

	var found *cobra.Command
	for _, c := range root.Commands() {
		if c.Name() == "version" {
			found = c
			break
		}
	}
	require.NotNil(t, found, "version subcommand should be registered")
	assert.Equal(t, "version", found.Use)
	assert.Equal(t, "Print ko version.", found.Short)
}

func TestVersionCommand_PrintsGlobalVersion(t *testing.T) {
	orig := Version
	defer func() { Version = orig }()

	Version = "v9.9.9"

	root := &cobra.Command{Use: "root"}
	addVersion(root)
	root.SetArgs([]string{"version"})

	out := captureStdout(t, func() {
		err := root.Execute()
		require.NoError(t, err)
	})

	assert.Equal(t, "v9.9.9\n", out)
}

func TestVersionCommand_PrintsBuildInfoOrMessageWhenEmpty(t *testing.T) {
	orig := Version
	defer func() { Version = orig }()

	// Force reading from build info
	Version = ""

	root := &cobra.Command{Use: "root"}
	addVersion(root)
	root.SetArgs([]string{"version"})

	out := captureStdout(t, func() {
		err := root.Execute()
		require.NoError(t, err)
	})

	out = strings.TrimSpace(out)
	// Typically will be something like "(devel)" when running tests, but fallback message is also acceptable.
	if out == "could not determine build information" {
		assert.Equal(t, "could not determine build information", out)
	} else {
		assert.NotEmpty(t, out)
	}
}
