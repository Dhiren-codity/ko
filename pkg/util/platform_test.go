package util

import (
	"runtime"
	"testing"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/stretchr/testify/assert"
)

func TestPlatform_String_NoVariant(t *testing.T) {
	p := Platform{
		OS:           "linux",
		Architecture: "amd64",
	}

	assert.Equal(t, "linux/amd64", p.String())
}

func TestPlatform_String_WithVariant(t *testing.T) {
	p := Platform{
		OS:           "linux",
		Architecture: "arm",
		Variant:      "v7",
	}

	assert.Equal(t, "linux/arm/v7", p.String())
}

func TestParsePlatform_Valid_NoVariant(t *testing.T) {
	got, err := ParsePlatform("linux/amd64")
	assert.NoError(t, err)
	assert.Equal(t, &Platform{
		OS:           "linux",
		Architecture: "amd64",
	}, got)
}

func TestParsePlatform_Valid_WithVariant(t *testing.T) {
	got, err := ParsePlatform("linux/arm/v7")
	assert.NoError(t, err)
	assert.Equal(t, &Platform{
		OS:           "linux",
		Architecture: "arm",
		Variant:      "v7",
	}, got)
}

func TestParsePlatform_EmptyString(t *testing.T) {
	got, err := ParsePlatform("")
	assert.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "platform cannot be empty")
}

func TestParsePlatform_NotEnoughParts(t *testing.T) {
	got, err := ParsePlatform("linux")
	assert.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "invalid platform format")
}

func TestParsePlatform_TooManyParts(t *testing.T) {
	got, err := ParsePlatform("linux/amd64/extra/part")
	assert.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "too many components")
}

func TestParsePlatform_EmptyOS(t *testing.T) {
	got, err := ParsePlatform("/amd64")
	assert.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "platform OS cannot be empty")
}

func TestParsePlatform_EmptyArch(t *testing.T) {
	got, err := ParsePlatform("linux/")
	assert.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "platform architecture cannot be empty")
}

func TestToPlatform(t *testing.T) {
	vp := v1.Platform{
		OS:           "linux",
		Architecture: "arm64",
		Variant:      "v8",
	}
	got := ToPlatform(vp)

	assert.Equal(t, Platform{
		OS:           "linux",
		Architecture: "arm64",
		Variant:      "v8",
	}, got)
}

func TestPlatform_ToV1Platform(t *testing.T) {
	p := Platform{
		OS:           "windows",
		Architecture: "amd64",
		Variant:      "v2",
	}
	got := p.ToV1Platform()

	assert.Equal(t, v1.Platform{
		OS:           "windows",
		Architecture: "amd64",
		Variant:      "v2",
	}, got)
}

func TestIsValidPlatform_ValidCases(t *testing.T) {
	tests := []string{
		"linux/amd64",
		"linux/AMD64",
		"windows/386",
		"darwin/arm64",
		"freebsd/arm",
		"linux/arm/v7", // variant is allowed but not validated
	}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			assert.True(t, IsValidPlatform(tt))
		})
	}
}

func TestIsValidPlatform_InvalidCases(t *testing.T) {
	tests := []string{
		"",                  // parse error
		"linux",             // parse error
		"unknownos/amd64",   // invalid OS
		"linux/unknownarch", // invalid arch
		"/amd64",            // parse error
		"linux/",            // parse error
	}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			assert.False(t, IsValidPlatform(tt))
		})
	}
}

func TestGetHostPlatform(t *testing.T) {
	host := GetHostPlatform()

	assert.Equal(t, runtime.GOOS, host.OS)
	assert.Equal(t, runtime.GOARCH, host.Architecture)
	// Variant is not set by GetHostPlatform; should be empty
	assert.Equal(t, "", host.Variant)
}

func TestNormalizePlatform_Valid(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "already lowercase no variant",
			input:    "linux/amd64",
			expected: "linux/amd64",
		},
		{
			name:     "mixed case no variant",
			input:    "Linux/AMD64",
			expected: "linux/amd64",
		},
		{
			name:     "mixed case with variant",
			input:    "Linux/Arm/V7",
			expected: "linux/arm/v7",
		},
		{
			name:     "upper case",
			input:    "WINDOWS/AMD64",
			expected: "windows/amd64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := NormalizePlatform(tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, out)
		})
	}
}

func TestNormalizePlatform_Invalid(t *testing.T) {
	_, err := NormalizePlatform("invalid")
	assert.Error(t, err)
}

func TestMatchesPlatform_Equal_NoVariant(t *testing.T) {
	match, err := MatchesPlatform("linux/amd64", "linux/amd64")
	assert.NoError(t, err)
	assert.True(t, match)
}

func TestMatchesPlatform_Equal_WithVariant(t *testing.T) {
	match, err := MatchesPlatform("linux/arm/v7", "linux/arm/v7")
	assert.NoError(t, err)
	assert.True(t, match)
}

func TestMatchesPlatform_DifferentOS(t *testing.T) {
	match, err := MatchesPlatform("linux/amd64", "windows/amd64")
	assert.NoError(t, err)
	assert.False(t, match)
}

func TestMatchesPlatform_DifferentArch(t *testing.T) {
	match, err := MatchesPlatform("linux/amd64", "linux/386")
	assert.NoError(t, err)
	assert.False(t, match)
}

func TestMatchesPlatform_CaseInsensitiveOSArch(t *testing.T) {
	match, err := MatchesPlatform("Linux/AMD64", "linux/amd64")
	assert.NoError(t, err)
	assert.True(t, match)
}

func TestMatchesPlatform_VariantBothEmpty(t *testing.T) {
	match, err := MatchesPlatform("linux/amd64", "LINUX/AMD64")
	assert.NoError(t, err)
	assert.True(t, match)
}

func TestMatchesPlatform_VariantBothSetEqual(t *testing.T) {
	match, err := MatchesPlatform("linux/arm/v7", "LINUX/ARM/V7")
	assert.NoError(t, err)
	assert.True(t, match)
}

func TestMatchesPlatform_VariantBothSetNotEqual(t *testing.T) {
	match, err := MatchesPlatform("linux/arm/v7", "linux/arm/v6")
	assert.NoError(t, err)
	assert.False(t, match)
}

func TestMatchesPlatform_OneVariantEmptyOtherSet(t *testing.T) {
	tests := []struct {
		name   string
		p1     string
		p2     string
		expect bool
	}{
		{
			name:   "first has variant second empty",
			p1:     "linux/arm/v7",
			p2:     "linux/arm",
			expect: false,
		},
		{
			name:   "first empty second has variant",
			p1:     "linux/arm",
			p2:     "linux/arm/v7",
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := MatchesPlatform(tt.p1, tt.p2)
			assert.NoError(t, err)
			assert.Equal(t, tt.expect, match)
		})
	}
}

func TestMatchesPlatform_InvalidFirst(t *testing.T) {
	match, err := MatchesPlatform("invalid", "linux/amd64")
	assert.Error(t, err)
	assert.False(t, match)
	assert.Contains(t, err.Error(), "invalid platform1")
}

func TestMatchesPlatform_InvalidSecond(t *testing.T) {
	match, err := MatchesPlatform("linux/amd64", "invalid")
	assert.Error(t, err)
	assert.False(t, match)
	assert.Contains(t, err.Error(), "invalid platform2")
}

func TestFilterPlatforms_MixedValidity(t *testing.T) {
	input := []string{
		"linux/amd64",
		"linux", // invalid
		"windows/386",
		"unknown/amd64",    // invalid os
		"darwin/unknown",   // invalid arch
		"linux/arm/v7",     // valid
		"invalid/format/4", // parse error
	}

	valid, invalid := FilterPlatforms(input)

	assert.ElementsMatch(t, []string{
		"linux/amd64",
		"windows/386",
		"linux/arm/v7",
	}, valid)

	assert.ElementsMatch(t, []string{
		"linux",
		"unknown/amd64",
		"darwin/unknown",
		"invalid/format/4",
	}, invalid)
}

func TestIsValidOS(t *testing.T) {
	validOS := []string{
		"linux",
		"darwin",
		"windows",
		"freebsd",
		"openbsd",
		"netbsd",
		"plan9",
		"solaris",
		"aix",
		"android",
		"ios",
		"js",
		"wasip1",
	}

	for _, os := range validOS {
		t.Run(os, func(t *testing.T) {
			assert.True(t, isValidOS(os))
			assert.True(t, isValidOS(stringsToUpper(os)))
		})
	}

	invalidOS := []string{"", "unknown", "LINUXES"}
	for _, os := range invalidOS {
		t.Run("invalid-"+os, func(t *testing.T) {
			assert.False(t, isValidOS(os))
		})
	}
}

func TestIsValidArch(t *testing.T) {
	validArch := []string{
		"386",
		"amd64",
		"arm",
		"arm64",
		"ppc64",
		"ppc64le",
		"mips",
		"mipsle",
		"mips64",
		"mips64le",
		"s390x",
		"riscv64",
		"wasm",
		"loong64",
	}

	for _, arch := range validArch {
		t.Run(arch, func(t *testing.T) {
			assert.True(t, isValidArch(arch))
			assert.True(t, isValidArch(stringsToUpper(arch)))
		})
	}

	invalidArch := []string{"", "x86_64", "unknownarch"}
	for _, arch := range invalidArch {
		t.Run("invalid-"+arch, func(t *testing.T) {
			assert.False(t, isValidArch(arch))
		})
	}
}

// helper to avoid importing strings just for ToUpper in tests
func stringsToUpper(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'a' && c <= 'z' {
			b[i] = c - 'a' + 'A'
		}
	}
	return string(b)
}
