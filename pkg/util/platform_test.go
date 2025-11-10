package util

import (
	"runtime"
	"testing"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/stretchr/testify/assert"
)

func TestPlatform_String(t *testing.T) {
	tests := []struct {
		name     string
		p        Platform
		expected string
	}{
		{
			name:     "no variant",
			p:        Platform{OS: "linux", Architecture: "amd64"},
			expected: "linux/amd64",
		},
		{
			name:     "with variant",
			p:        Platform{OS: "linux", Architecture: "arm", Variant: "v7"},
			expected: "linux/arm/v7",
		},
		{
			name:     "with variant different case",
			p:        Platform{OS: "Windows", Architecture: "AMD64", Variant: "V8"},
			expected: "Windows/AMD64/V8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.p.String())
		})
	}
}

func TestParsePlatform_Success(t *testing.T) {
	tests := []struct {
		in       string
		expected Platform
	}{
		{
			in:       "linux/amd64",
			expected: Platform{OS: "linux", Architecture: "amd64"},
		},
		{
			in:       "linux/arm/v7",
			expected: Platform{OS: "linux", Architecture: "arm", Variant: "v7"},
		},
		{
			in:       "Linux/ARM64/v8",
			expected: Platform{OS: "Linux", Architecture: "ARM64", Variant: "v8"},
		},
		{
			in:       "windows/386",
			expected: Platform{OS: "windows", Architecture: "386"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := ParsePlatform(tt.in)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, *got)
			// String should mirror the original fields as-is
			assert.Equal(t, tt.expected.String(), got.String())
		})
	}
}

func TestParsePlatform_Errors(t *testing.T) {
	tests := []struct {
		name      string
		in        string
		errSubstr string
	}{
		{
			name:      "empty string",
			in:        "",
			errSubstr: "platform cannot be empty",
		},
		{
			name:      "too few components",
			in:        "linux",
			errSubstr: "invalid platform format",
		},
		{
			name:      "too many components",
			in:        "linux/amd64/v8/extra",
			errSubstr: "too many components",
		},
		{
			name:      "empty OS",
			in:        "/amd64",
			errSubstr: "platform OS cannot be empty",
		},
		{
			name:      "empty arch",
			in:        "linux/",
			errSubstr: "platform architecture cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePlatform(tt.in)
			assert.Nil(t, got)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errSubstr)
		})
	}
}

func TestToPlatform_And_ToV1Platform(t *testing.T) {
	// ToPlatform
	vp := v1.Platform{
		OS:           "linux",
		Architecture: "arm",
		Variant:      "v7",
	}
	p := ToPlatform(vp)
	assert.Equal(t, Platform{OS: "linux", Architecture: "arm", Variant: "v7"}, p)

	// ToV1Platform
	p2 := Platform{OS: "windows", Architecture: "amd64"}
	vp2 := p2.ToV1Platform()
	assert.Equal(t, v1.Platform{OS: "windows", Architecture: "amd64", Variant: ""}, vp2)
}

func TestIsValidPlatform(t *testing.T) {
	tests := []struct {
		in     string
		valid  bool
		reason string
	}{
		{"linux/amd64", true, "standard linux/amd64"},
		{"darwin/arm64", true, "standard darwin/arm64"},
		{"windows/386", true, "standard windows/386"},
		{"freebsd/arm64", true, "supported os and arch"},
		{"android/arm64", true, "supported os and arch"},
		{"linux/arm/v7", true, "variant should not affect validity"},
		{"linux/arm/v9", true, "unknown variant still valid per current implementation"},
		{"LINUX/AMD64", true, "case-insensitive valid"},
		{"unknownos/amd64", false, "unknown os"},
		{"linux/unknownarch", false, "unknown arch"},
		{"", false, "empty string invalid"},
		{"linux", false, "too few components invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			assert.Equal(t, tt.valid, IsValidPlatform(tt.in), tt.reason)
		})
	}
}

func TestGetHostPlatform(t *testing.T) {
	host := GetHostPlatform()
	assert.Equal(t, runtime.GOOS, host.OS)
	assert.Equal(t, runtime.GOARCH, host.Architecture)
	// variant is not set by GetHostPlatform
	assert.Equal(t, "", host.Variant)

	// Ensure the returned host platform string is considered valid
	hostStr := host.String()
	assert.True(t, IsValidPlatform(hostStr))
}

func TestNormalizePlatform(t *testing.T) {
	tests := []struct {
		in       string
		expected string
		wantErr  bool
	}{
		{"Linux/AMD64", "linux/amd64", false},
		{"linux/arm/V7", "linux/arm/v7", false},
		{"WINDOWS/386", "windows/386", false},
		{"", "", true},
		{"linux", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			out, err := NormalizePlatform(tt.in)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, out)
		})
	}
}

func TestMatchesPlatform(t *testing.T) {
	tests := []struct {
		name       string
		p1         string
		p2         string
		wantMatch  bool
		wantErrSub string
	}{
		{
			name:      "exact match no variant",
			p1:        "linux/amd64",
			p2:        "linux/amd64",
			wantMatch: true,
		},
		{
			name:      "os mismatch",
			p1:        "linux/amd64",
			p2:        "windows/amd64",
			wantMatch: false,
		},
		{
			name:      "arch mismatch",
			p1:        "linux/amd64",
			p2:        "linux/arm64",
			wantMatch: false,
		},
		{
			name:      "both variants empty match",
			p1:        "linux/arm",
			p2:        "linux/arm",
			wantMatch: true,
		},
		{
			name:      "same variants match case-insensitive",
			p1:        "linux/arm/v7",
			p2:        "linux/arm/V7",
			wantMatch: true,
		},
		{
			name:      "one variant missing no match",
			p1:        "linux/arm/v7",
			p2:        "linux/arm",
			wantMatch: false,
		},
		{
			name:      "different variants no match",
			p1:        "linux/arm/v7",
			p2:        "linux/arm/v6",
			wantMatch: false,
		},
		{
			name:       "invalid platform1 error",
			p1:         "",
			p2:         "linux/amd64",
			wantErrSub: "invalid platform1",
		},
		{
			name:       "invalid platform2 error",
			p1:         "linux/amd64",
			p2:         "linux",
			wantErrSub: "invalid platform2",
		},
		{
			name:      "case-insensitive os and arch",
			p1:        "LINUX/ARM64",
			p2:        "linux/arm64",
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := MatchesPlatform(tt.p1, tt.p2)
			if tt.wantErrSub != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrSub)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantMatch, match)
		})
	}
}

func TestFilterPlatforms(t *testing.T) {
	input := []string{
		"linux/amd64",
		"LINUX/ARM64",
		"nos/amd64",
		"linux/unknown",
		"darwin/arm64",
		"linux/arm/v7",
	}
	valid, invalid := FilterPlatforms(input)

	assert.Equal(t, []string{"linux/amd64", "LINUX/ARM64", "darwin/arm64", "linux/arm/v7"}, valid)
	assert.Equal(t, []string{"nos/amd64", "linux/unknown"}, invalid)
}

func Test_isValidOS(t *testing.T) {
	tests := []struct {
		os    string
		valid bool
	}{
		{"linux", true},
		{"darwin", true},
		{"windows", true},
		{"freebsd", true},
		{"openbsd", true},
		{"netbsd", true},
		{"plan9", true},
		{"solaris", true},
		{"aix", true},
		{"android", true},
		{"ios", true},
		{"js", true},
		{"wasip1", true},
		{"LINUX", true}, // case-insensitive
		{"unknownos", false},
	}

	for _, tt := range tests {
		t.Run(tt.os, func(t *testing.T) {
			assert.Equal(t, tt.valid, isValidOS(tt.os))
		})
	}
}

func Test_isValidArch(t *testing.T) {
	tests := []struct {
		arch  string
		valid bool
	}{
		{"386", true},
		{"amd64", true},
		{"arm", true},
		{"arm64", true},
		{"ppc64", true},
		{"ppc64le", true},
		{"mips", true},
		{"mipsle", true},
		{"mips64", true},
		{"mips64le", true},
		{"s390x", true},
		{"riscv64", true},
		{"wasm", true},
		{"loong64", true},
		{"AMD64", true}, // case-insensitive
		{"unknownarch", false},
	}

	for _, tt := range tests {
		t.Run(tt.arch, func(t *testing.T) {
			assert.Equal(t, tt.valid, isValidArch(tt.arch))
		})
	}
}
