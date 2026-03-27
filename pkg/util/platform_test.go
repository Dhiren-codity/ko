package util

import (
	"runtime"
	"testing"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/stretchr/testify/assert"
)

func TestPlatform_String(t *testing.T) {
	tests := []struct {
		name string
		p    Platform
		want string
	}{
		{
			name: "no variant",
			p: Platform{
				OS:           "linux",
				Architecture: "amd64",
			},
			want: "linux/amd64",
		},
		{
			name: "with variant",
			p: Platform{
				OS:           "linux",
				Architecture: "arm",
				Variant:      "v7",
			},
			want: "linux/arm/v7",
		},
		{
			name: "empty fields produce slashes",
			p: Platform{
				OS:           "",
				Architecture: "",
				Variant:      "",
			},
			want: "/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.p.String())
		})
	}
}

func TestParsePlatform_Success(t *testing.T) {
	tests := []struct {
		in       string
		wantOS   string
		wantArch string
		wantVar  string
	}{
		{"linux/amd64", "linux", "amd64", ""},
		{"linux/arm64", "linux", "arm64", ""},
		{"linux/arm/v7", "linux", "arm", "v7"},
		{"windows/amd64", "windows", "amd64", ""},
		{"Darwin/AMD64", "Darwin", "AMD64", ""}, // case preserved by parser
	}
	for _, tt := range tests {
		got, err := ParsePlatform(tt.in)
		assert.NoError(t, err, tt.in)
		assert.Equal(t, tt.wantOS, got.OS)
		assert.Equal(t, tt.wantArch, got.Architecture)
		assert.Equal(t, tt.wantVar, got.Variant)
	}
}

func TestParsePlatform_Errors(t *testing.T) {
	tests := []struct {
		in  string
		msg string
	}{
		{"", "platform cannot be empty"},
		{"linux", "invalid platform format"},
		{"linux/amd64/variant/extra", "too many components"},
		{"/amd64", "platform OS cannot be empty"},
		{"linux/", "platform architecture cannot be empty"},
	}
	for _, tt := range tests {
		_, err := ParsePlatform(tt.in)
		assert.Error(t, err, tt.in)
		assert.Contains(t, err.Error(), tt.msg, tt.in)
	}
}

func TestToPlatformAndToV1Platform(t *testing.T) {
	vp := v1.Platform{
		OS:           "linux",
		Architecture: "arm",
		Variant:      "v7",
	}
	p := ToPlatform(vp)
	assert.Equal(t, "linux", p.OS)
	assert.Equal(t, "arm", p.Architecture)
	assert.Equal(t, "v7", p.Variant)

	back := p.ToV1Platform()
	assert.Equal(t, vp.OS, back.OS)
	assert.Equal(t, vp.Architecture, back.Architecture)
	assert.Equal(t, vp.Variant, back.Variant)
}

func TestIsValidPlatform(t *testing.T) {
	valid := []string{
		"linux/amd64",
		"LiNuX/AmD64", // case-insensitive validation
		"windows/arm64",
		"freebsd/386",
		"linux/amd64/any-variant", // variant not validated for validity
		"linux/riscv64",
		"darwin/arm64",
		"linux/wasm",
		"linux/loong64",
	}
	for _, p := range valid {
		assert.True(t, IsValidPlatform(p), p)
	}

	invalid := []string{
		"unknown/amd64",  // bad OS
		"linux/somearch", // bad arch
		"linux",          // bad format
		"/amd64",         // empty OS
		"linux/",         // empty arch
	}
	for _, p := range invalid {
		assert.False(t, IsValidPlatform(p), p)
	}
}

func TestGetHostPlatform(t *testing.T) {
	h := GetHostPlatform()
	assert.Equal(t, runtime.GOOS, h.OS)
	assert.Equal(t, runtime.GOARCH, h.Architecture)
	assert.Equal(t, "", h.Variant)
}

func TestNormalizePlatform(t *testing.T) {
	out, err := NormalizePlatform("LINUX/AMD64/V8")
	assert.NoError(t, err)
	assert.Equal(t, "linux/amd64/v8", out)

	out, err = NormalizePlatform("Linux/ArM")
	assert.NoError(t, err)
	assert.Equal(t, "linux/arm", out)

	_, err = NormalizePlatform("linux") // invalid format
	assert.Error(t, err)
}

func TestMatchesPlatform_Basics(t *testing.T) {
	tests := []struct {
		p1   string
		p2   string
		want bool
	}{
		{"linux/amd64", "linux/amd64", true},
		{"linux/amd64", "linux/arm64", false},
		{"linux/arm/v7", "linux/arm/v7", true},
		{"linux/arm/v7", "linux/arm/V7", true},  // variant case-insensitive
		{"LINUX/ARM/V7", "linux/arm/v7", true},  // os/arch case-insensitive
		{"linux/arm", "linux/arm/v7", false},    // variant mismatch (one empty)
		{"linux/arm/v7", "linux/arm", false},    // variant mismatch (one empty)
		{"windows/amd64", "linux/amd64", false}, // OS mismatch
	}
	for _, tt := range tests {
		got, err := MatchesPlatform(tt.p1, tt.p2)
		assert.NoError(t, err, tt.p1+" vs "+tt.p2)
		assert.Equal(t, tt.want, got, tt.p1+" vs "+tt.p2)
	}
}

func TestMatchesPlatform_Errors(t *testing.T) {
	_, err := MatchesPlatform("linux", "linux/amd64")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid platform1:")

	_, err = MatchesPlatform("linux/amd64", "linux")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid platform2:")
}

func TestFilterPlatforms(t *testing.T) {
	in := []string{
		"linux/amd64",
		"unknown/amd64", // invalid
		"linux/arm64",
		"linux",      // invalid
		"darwin/ppc", // invalid arch
		"windows/386",
	}
	valid, invalid := FilterPlatforms(in)

	assert.Equal(t, []string{"linux/amd64", "linux/arm64", "windows/386"}, valid)
	assert.Equal(t, []string{"unknown/amd64", "linux", "darwin/ppc"}, invalid)
}

func Test_isValidOS(t *testing.T) {
	assert.True(t, isValidOS("linux"))
	assert.True(t, isValidOS("LINUX"))
	assert.True(t, isValidOS("windows"))
	assert.True(t, isValidOS("darwin"))
	assert.True(t, isValidOS("js"))
	assert.True(t, isValidOS("wasip1"))

	assert.False(t, isValidOS("beos"))
	assert.False(t, isValidOS("plan10"))
}

func Test_isValidArch(t *testing.T) {
	assert.True(t, isValidArch("amd64"))
	assert.True(t, isValidArch("AMd64"))
	assert.True(t, isValidArch("arm"))
	assert.True(t, isValidArch("arm64"))
	assert.True(t, isValidArch("riscv64"))
	assert.True(t, isValidArch("wasm"))
	assert.True(t, isValidArch("loong64"))

	assert.False(t, isValidArch("alpha"))
	assert.False(t, isValidArch("x86_128"))
}
