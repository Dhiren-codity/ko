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
			name: "with variant",
			p:    Platform{OS: "linux", Architecture: "arm", Variant: "v7"},
			want: "linux/arm/v7",
		},
		{
			name: "without variant",
			p:    Platform{OS: "linux", Architecture: "amd64"},
			want: "linux/amd64",
		},
		{
			name: "preserve case",
			p:    Platform{OS: "LINUX", Architecture: "AMD64"},
			want: "LINUX/AMD64",
		},
		{
			name: "empty variant omitted",
			p:    Platform{OS: "windows", Architecture: "arm64", Variant: ""},
			want: "windows/arm64",
		},
	}
	for _, tt := range tests {
		got := tt.p.String()
		assert.Equal(t, tt.want, got, tt.name)
	}
}

func TestParsePlatform_Success(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want Platform
	}{
		{
			name: "basic os/arch",
			in:   "linux/amd64",
			want: Platform{OS: "linux", Architecture: "amd64"},
		},
		{
			name: "with variant",
			in:   "linux/arm/v7",
			want: Platform{OS: "linux", Architecture: "arm", Variant: "v7"},
		},
		{
			name: "case preserved",
			in:   "LINUX/AMD64",
			want: Platform{OS: "LINUX", Architecture: "AMD64"},
		},
	}
	for _, tt := range tests {
		got, err := ParsePlatform(tt.in)
		assert.NoError(t, err, tt.name)
		assert.NotNil(t, got, tt.name)
		assert.Equal(t, tt.want, *got, tt.name)
	}
}

func TestParsePlatform_Errors(t *testing.T) {
	tests := []struct {
		name        string
		in          string
		errContains string
	}{
		{
			name:        "empty",
			in:          "",
			errContains: "platform cannot be empty",
		},
		{
			name:        "missing arch",
			in:          "linux",
			errContains: "invalid platform format",
		},
		{
			name:        "too many components",
			in:          "linux/amd64/v8/extra",
			errContains: "too many components",
		},
		{
			name:        "empty os component",
			in:          "/amd64",
			errContains: "platform OS cannot be empty",
		},
		{
			name:        "empty arch component",
			in:          "linux/",
			errContains: "platform architecture cannot be empty",
		},
	}
	for _, tt := range tests {
		got, err := ParsePlatform(tt.in)
		assert.Nil(t, got, tt.name)
		assert.Error(t, err, tt.name)
		assert.Contains(t, err.Error(), tt.errContains, tt.name)
	}
}

func TestToPlatform_ToV1Platform(t *testing.T) {
	// ToPlatform
	vp := v1.Platform{OS: "linux", Architecture: "arm", Variant: "v7"}
	got := ToPlatform(vp)
	assert.Equal(t, Platform{OS: "linux", Architecture: "arm", Variant: "v7"}, got)

	// ToV1Platform
	p := Platform{OS: "windows", Architecture: "amd64"}
	gotV1 := p.ToV1Platform()
	assert.Equal(t, v1.Platform{OS: "windows", Architecture: "amd64"}, gotV1)
}

func TestIsValidPlatform(t *testing.T) {
	valid := []string{
		"linux/amd64",
		"LiNuX/AMd64",
		"windows/arm64",
		"linux/arm/v7",
		"darwin/amd64",
		"wasip1/wasm",
		"ios/arm64",
		"js/wasm",
		"aix/ppc64",
	}
	for _, s := range valid {
		assert.True(t, IsValidPlatform(s), "expected valid: %s", s)
	}

	invalid := []string{
		"",                     // empty
		"linux",                // missing arch
		"unknown/amd64",        // unknown os
		"linux/unknown",        // unknown arch
		"/amd64",               // empty os
		"linux/",               // empty arch
		"linux/amd64/v7/extra", // too many parts
	}
	for _, s := range invalid {
		assert.False(t, IsValidPlatform(s), "expected invalid: %s", s)
	}
}

func TestGetHostPlatform(t *testing.T) {
	h := GetHostPlatform()
	assert.Equal(t, runtime.GOOS, h.OS)
	assert.Equal(t, runtime.GOARCH, h.Architecture)
	assert.Equal(t, "", h.Variant)
}

func TestNormalizePlatform(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "lowercase os and arch",
			in:   "Linux/AMD64",
			want: "linux/amd64",
		},
		{
			name: "lowercase variant",
			in:   "linux/arm/V7",
			want: "linux/arm/v7",
		},
		{
			name: "unknown components lowercased",
			in:   "MyOS/MyArch",
			want: "myos/myarch",
		},
	}
	for _, tt := range tests {
		got, err := NormalizePlatform(tt.in)
		assert.NoError(t, err, tt.name)
		assert.Equal(t, tt.want, got, tt.name)
	}

	// Error case
	_, err := NormalizePlatform("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "platform cannot be empty")
}

func TestMatchesPlatform_Success(t *testing.T) {
	tests := []struct {
		name string
		p1   string
		p2   string
		want bool
	}{
		{
			name: "same no variant diff case",
			p1:   "Linux/Amd64",
			p2:   "linux/amd64",
			want: true,
		},
		{
			name: "same with variant",
			p1:   "linux/arm/v7",
			p2:   "linux/arm/v7",
			want: true,
		},
		{
			name: "variant mismatch",
			p1:   "linux/arm/v7",
			p2:   "linux/arm/v6",
			want: false,
		},
		{
			name: "one has variant other not",
			p1:   "linux/arm/v7",
			p2:   "linux/arm",
			want: false,
		},
		{
			name: "os mismatch",
			p1:   "linux/amd64",
			p2:   "windows/amd64",
			want: false,
		},
		{
			name: "arch mismatch",
			p1:   "linux/amd64",
			p2:   "linux/arm64",
			want: false,
		},
		{
			name: "both no variant",
			p1:   "windows/amd64",
			p2:   "windows/amd64",
			want: true,
		},
	}
	for _, tt := range tests {
		got, err := MatchesPlatform(tt.p1, tt.p2)
		assert.NoError(t, err, tt.name)
		assert.Equal(t, tt.want, got, tt.name)
	}
}

func TestMatchesPlatform_Errors(t *testing.T) {
	// invalid p1
	matched, err := MatchesPlatform("", "linux/amd64")
	assert.Error(t, err)
	assert.False(t, matched)
	assert.Contains(t, err.Error(), "invalid platform1:")

	// invalid p2
	matched, err = MatchesPlatform("linux/amd64", "")
	assert.Error(t, err)
	assert.False(t, matched)
	assert.Contains(t, err.Error(), "invalid platform2:")
}

func TestFilterPlatforms(t *testing.T) {
	in := []string{
		"linux/amd64",   // valid
		"linux",         // invalid
		"/amd64",        // invalid
		"windows/arm64", // valid
		"linux/unknown", // invalid
		"darwin/amd64",  // valid
		"wasip1/wasm",   // valid
		"LINUX/AMD64",   // valid (case-insensitive)
	}
	valid, invalid := FilterPlatforms(in)

	expectedValid := []string{
		"linux/amd64",
		"windows/arm64",
		"darwin/amd64",
		"wasip1/wasm",
		"LINUX/AMD64",
	}
	expectedInvalid := []string{
		"linux",
		"/amd64",
		"linux/unknown",
	}

	assert.Equal(t, expectedValid, valid)
	assert.Equal(t, expectedInvalid, invalid)
}
