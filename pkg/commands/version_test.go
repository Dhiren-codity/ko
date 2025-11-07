package commands

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	defer func() { os.Stdout = orig }()

	r, w, err := os.Pipe()
	assert.NoError(t, err)

	os.Stdout = w

	fn()

	err = w.Close()
	assert.NoError(t, err)

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	assert.NoError(t, err)
	_ = r.Close()

	return buf.String()
}

func TestFormatVersion_Table(t *testing.T) {
	tests := []struct {
		name string
		in   string
		out  string
	}{
		{"empty", "", "unknown"},
		{"plain", "1.2.3", "1.2.3"},
		{"with_v_prefix", "v1.2.3", "v1.2.3"},
		{"with_build_meta", "1.2.3+build", "1.2.3"},
		{"with_build_meta_complex", "v1.2.3+build.meta", "v1.2.3"},
		{"with_prerelease", "1.2.3-alpha.1", "1.2.3-alpha.1"},
		{"unknown_literal", "unknown", "unknown"},
		{"multiple_plus", "1.2.3+build+meta", "1.2.3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatVersion(tt.in)
			assert.Equal(t, tt.out, got)
		})
	}
}

func TestIsValidVersion_Table(t *testing.T) {
	tests := []struct {
		name string
		in   string
		ok   bool
	}{
		{"plain", "1.2.3", true},
		{"with_v_prefix", "v1.2.3", true},
		{"with_prerelease", "1.2.3-alpha", true},
		{"with_build_meta", "1.2.3+build.5", true},
		{"with_prerelease_and_build_meta", "1.2.3-alpha+build.5", true},
		{"empty", "", false},
		{"unknown", "unknown", false},
		{"too_few_segments", "1.2", false},
		{"too_many_segments", "1.2.3.4", false},
		{"just_v", "v", false},
		{"nonsense", "foo", false},
		{"dangling_dash", "1.2.3-", false},
		{"dangling_plus", "1.2.3+build+", false},
		{"invalid_patch", "1.2.x", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidVersion(tt.in)
			assert.Equal(t, tt.ok, got, "input: %q", tt.in)
		})
	}
}

func TestGetVersionInfo_ValidSimple(t *testing.T) {
	old := Version
	defer func() { Version = old }()

	Version = "v1.2.3"
	got := getVersionInfo()
	assert.Equal(t, "ko version v1.2.3 (valid)", got)
}

func TestGetVersionInfo_FormatsBuildMetadataButKeepsValidStatus(t *testing.T) {
	old := Version
	defer func() { Version = old }()

	Version = "1.2.3+build.meta"
	got := getVersionInfo()
	assert.Equal(t, "ko version 1.2.3 (valid)", got)
}

func TestGetVersionInfo_NonStandardUnknown(t *testing.T) {
	old := Version
	defer func() { Version = old }()

	Version = "unknown"
	got := getVersionInfo()
	assert.Equal(t, "ko version unknown (non-standard)", got)
}

func TestGetVersionInfo_PreReleaseValid(t *testing.T) {
	old := Version
	defer func() { Version = old }()

	Version = "v2.0.0-rc.1"
	got := getVersionInfo()
	assert.Equal(t, "ko version v2.0.0-rc.1 (valid)", got)
}

func TestVersion_ReturnsVarWhenSet(t *testing.T) {
	old := Version
	defer func() { Version = old }()

	Version = "v0.0.1"
	got := version()
	assert.Equal(t, "v0.0.1", got)
	assert.Equal(t, "v0.0.1", Version, "version() should not modify Version when already set")
}

func TestAddVersionCommand_AddsCommandWithExpectedProperties(t *testing.T) {
	root := &cobra.Command{Use: "ko"}
	addVersion(root)

	cmd, _, err := root.Find([]string{"version"})
	assert.NoError(t, err)
	assert.NotNil(t, cmd)
	if cmd == nil {
		t.Fatal("version command should be added")
	}

	assert.Equal(t, "version", cmd.Name())
	assert.Equal(t, "Print ko version.", cmd.Short)
}

func TestAddVersionCommand_PrintsRawVersion(t *testing.T) {
	old := Version
	defer func() { Version = old }()

	Version = "1.2.3+build"
	root := &cobra.Command{Use: "ko"}
	addVersion(root)
	root.SetArgs([]string{"version"})

	out := captureStdout(t, func() {
		err := root.Execute()
		assert.NoError(t, err)
	})

	assert.Contains(t, out, "1.2.3+build")
}

func TestAddVersionCommand_PrintsSimpleVersion(t *testing.T) {
	old := Version
	defer func() { Version = old }()

	Version = "v9.9.9"
	root := &cobra.Command{Use: "ko"}
	addVersion(root)
	root.SetArgs([]string{"version"})

	out := captureStdout(t, func() {
		err := root.Execute()
		assert.NoError(t, err)
	})

	assert.True(t, strings.HasPrefix(out, "v9.9.9"))
}

func TestFormatVersion_DoesNotAlterPrerelease(t *testing.T) {
	got := formatVersion("1.2.3-beta.2")
	assert.Equal(t, "1.2.3-beta.2", got)
}

func TestIsValidVersion_VersionWithHyphensInPrerelease(t *testing.T) {
	assert.True(t, isValidVersion("1.2.3-alpha-beta"))
}

func TestIsValidVersion_VersionWithDotsInPrerelease(t *testing.T) {
	assert.True(t, isValidVersion("1.2.3-alpha.beta"))
}

func TestIsValidVersion_VersionWithDashesInBuild(t *testing.T) {
	assert.True(t, isValidVersion("1.2.3+build-20250101"))
}

func TestGetVersionInfo_WithVPrefixAndBuild(t *testing.T) {
	old := Version
	defer func() { Version = old }()

	Version = "v3.4.5+sha.abcd1234"
	got := getVersionInfo()
	assert.Equal(t, "ko version v3.4.5 (valid)", got)
}
