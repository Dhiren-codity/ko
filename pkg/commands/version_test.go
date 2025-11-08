package commands

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func restoreVersion(v string) func() {
	old := Version
	Version = v
	return func() { Version = old }
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	outCh := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		outCh <- buf.String()
	}()

	fn()

	_ = w.Close()
	os.Stdout = old
	out := <-outCh
	_ = r.Close()
	return out
}

func Test_formatVersion_Table(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", "unknown"},
		{"simple", "v1.2.3", "v1.2.3"},
		{"with_build", "v1.2.3+build.1", "v1.2.3"},
		{"with_prerelease_and_build", "1.0.0-alpha+001", "1.0.0-alpha"},
		{"non_semver", "(devel)", "(devel)"},
		{"multiple_plus", "v1.0.0+abc+def", "v1.0.0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatVersion(tt.in)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_isValidVersion_Table(t *testing.T) {
	tests := []struct {
		name string
		in   string
		ok   bool
	}{
		{"valid_with_v", "v1.2.3", true},
		{"valid_without_v", "1.2.3", true},
		{"valid_prerelease", "1.2.3-beta", true},
		{"valid_prerelease_dotted", "v1.2.3-rc.1", true},
		{"valid_build", "1.2.3+build", true},
		{"valid_both", "v1.2.3-rc.1+build.5", true},
		{"invalid_empty", "", false},
		{"invalid_unknown", "unknown", false},
		{"invalid_short", "v1.2", false},
		{"invalid_text", "foo", false},
		{"invalid_too_many_parts", "1.2.3.4", false},
		{"invalid_whitespace", " 1.2.3", false},
		{"invalid_no_separator_for_prerelease", "v1.2.3beta", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidVersion(tt.in)
			assert.Equal(t, tt.ok, got)
		})
	}
}

func Test_version_ReturnsGlobalWhenSet(t *testing.T) {
	defer restoreVersion(Version)()
	Version = "v9.9.9"
	got := version()
	assert.Equal(t, "v9.9.9", got)
}

func Test_version_ReadsBuildInfoWhenEmptyAndCaches(t *testing.T) {
	defer restoreVersion(Version)()
	Version = ""
	first := version()
	assert.NotEmpty(t, first, "version() should return non-empty when build info is available")

	// Ensure it cached into the global
	assert.Equal(t, first, Version)

	// Second call should return the same value
	second := version()
	assert.Equal(t, first, second)
}

func Test_getVersionInfo_StandardSemverValidStatus(t *testing.T) {
	defer restoreVersion(Version)()
	Version = "v1.2.3+build.1"
	got := getVersionInfo()
	// formatted should drop build metadata, status should be valid
	assert.Equal(t, "ko version v1.2.3 (valid)", got)
}

func Test_getVersionInfo_NonStandardVersionStatus(t *testing.T) {
	defer restoreVersion(Version)()
	Version = "foobar"
	got := getVersionInfo()
	assert.Equal(t, "ko version foobar (non-standard)", got)
}

func Test_getVersionInfo_UsesFormattedWithoutBuildMetadata(t *testing.T) {
	defer restoreVersion(Version)()
	// This input is valid semver but contains build metadata
	verWithBuild := "1.0.0+meta.info"
	Version = verWithBuild

	got := getVersionInfo()
	// Output should not contain '+', should be marked valid
	assert.NotContains(t, got, "+")
	assert.Equal(t, "ko version 1.0.0 (valid)", got)
}

func Test_addVersion_AttachesSubcommand(t *testing.T) {
	root := &cobra.Command{Use: "app"}
	addVersion(root)

	cmd, _, err := root.Find([]string{"version"})
	require.NoError(t, err)
	assert.NotNil(t, cmd)
	assert.Equal(t, "version", cmd.Name())
}

func Test_addVersion_CommandPrintsProvidedVersion(t *testing.T) {
	defer restoreVersion(Version)()
	Version = "v1.2.3"
	root := &cobra.Command{Use: "app"}
	addVersion(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"version"})
		err := root.Execute()
		require.NoError(t, err)
	})
	assert.Equal(t, "v1.2.3\n", out)
}

func Test_addVersion_CommandPrintsVersionWithBuildMetadata(t *testing.T) {
	defer restoreVersion(Version)()
	Version = "v1.2.3+meta.1"
	root := &cobra.Command{Use: "app"}
	addVersion(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"version"})
		err := root.Execute()
		require.NoError(t, err)
	})
	// addVersion prints raw version(), which should include build metadata
	assert.Equal(t, "v1.2.3+meta.1\n", out)
}

func Test_addVersion_CommandPrintsDynamicVersionWhenUnset(t *testing.T) {
	defer restoreVersion(Version)()
	// Unset global and rely on build info path
	Version = ""
	// Capture expected value by calling version()
	want := version()

	root := &cobra.Command{Use: "app"}
	addVersion(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"version"})
		err := root.Execute()
		require.NoError(t, err)
	})
	assert.Equal(t, want+"\n", out)
}
