package util

import (
	"runtime"
	"testing"

	"github.com/google/go-containerregistry/pkg/v1"
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
			name: "empty fields still printed",
			p: Platform{
				OS:           "",
				Architecture: "amd64",
			},
			want: "/amd64",
		},
		{
			name: "empty variant omitted",
			p: Platform{
				OS:           "windows",
				Architecture: "arm64",
				Variant:      "",
			},
			want: "windows/arm64",
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
		in      string
		os      string
		arch    string
		variant string
	}{
		{"linux/amd64", "linux", "amd64", ""},
		{"Linux/Arm/v7", "Linux", "Arm", "v7"},
		{"windows/386", "windows", "386", ""},
		{"darwin/arm64", "darwin", "arm64", ""},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := ParsePlatform(tt.in)
			assert.NoError(t, err)
			assert.Equal(t, tt.os, got.OS)
			assert.Equal(t, tt.arch, got.Architecture)
			assert.Equal(t, tt.variant, got.Variant)
		})
	}
}

func TestParsePlatform_Errors(t *testing.T) {
	tests := []struct {
		name          string
		in            string
		errContains   string
	}{
		{"empty string", "", "platform cannot be empty"},
		{"too few parts", "linux", "invalid platform format"},
		{"too many parts", "linux/arm/v7/extra", "too many components"},
		{"empty os", "/amd64", "platform OS cannot be empty"},
		{"empty arch", "linux/", "platform architecture cannot be empty"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := ParsePlatform(tt.in)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errContains)
			assert.Nil(t, p)
		})
	}
}

func TestToPlatform_ToV1Platform_RoundTrip(t *testing.T) {
	// v1 -> util
	vp := v1.Platform{
		OS:           "linux",
		Architecture: "amd64",
		Variant:      "",
	}
	p := ToPlatform(vp)
	assert.Equal(t, "linux", p.OS)
	assert.Equal(t, "amd64", p.Architecture)
	assert.Equal(t, "", p.Variant)

	// util -> v1
	up := Platform{OS: "linux", Architecture: "arm", Variant: "v7"}
	vp2 := up.ToV1Platform()
	assert.Equal(t, "linux", vp2.OS)
	assert.Equal(t, "arm", vp2.Architecture)
	assert.Equal(t, "v7", vp2.Variant)

	// round-trip
	assert.Equal(t, up, ToPlatform(vp2))
}

func TestIsValidPlatform(t *testing.T) {
	valid := []string{
		"linux/amd64",
		"Darwin/arm64",
		"windows/386",
		"freebsd/arm",
		"js/wasm",
		"wasip1/wasm",
		"linux/loong64",
		"linux/arm/v7", // variant allowed (not validated)
	}
	for _, s := range valid {
		t.Run("valid_"+s, func(t *testing.T) {
			assert.True(t, IsValidPlatform(s), "expected valid: %s", s)
		})
	}

	invalid := []string{
		"",
		"linux",
		"unknown/amd64",
		"linux/unknown",
		"/amd64",
		"linux/",
		"linux/arm/v7/extra",
	}
	for _, s := range invalid {
		t.Run("invalid_"+s, func(t *testing.T) {
			assert.False(t, IsValidPlatform(s), "expected invalid: %s", s)
		})
	}
}

func TestGetHostPlatform(t *testing.T) {
	host := GetHostPlatform()
	assert.Equal(t, runtime.GOOS, host.OS)
	assert.Equal(t, runtime.GOARCH, host.Architecture)
}

func TestNormalizePlatform(t *testing.T) {
	tests := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"LINUX/AMD64", "linux/amd64", false},
		{"LiNuX/ArM/V7", "linux/arm/v7", false},
		{"DARWIN/ARM64", "darwin/arm64", false},
		{"", "", true},
		{"linux", "", true},
		{"linux/arm/v7/extra", "", true},
		{"/amd64", "", true},
		{"linux/", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := NormalizePlatform(tt.in)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMatchesPlatform(t *testing.T) {
	tests := []struct {
		name string
		p1   string
		p2   string
		want bool
	}{
		{"exact match", "linux/amd64", "linux/amd64", true},
		{"case-insensitive match", "LINUX/AMD64", "linux/amd64", true},
		{"os mismatch", "linux/amd64", "windows/amd64", false},
		{"arch mismatch", "linux/amd64", "linux/arm64", false},
		{"arm no variant both", "linux/arm", "linux/arm", true},
		{"arm same variant", "linux/arm/v7", "linux/arm/v7", true},
		{"arm one missing variant", "linux/arm/v7", "linux/arm", false},
		{"arm variant mismatch", "linux/arm/v7", "linux/arm/v6", false},
		{"non-arm variant mismatch treated same", "linux/amd64/v2", "linux/amd64", false},
		{"variant case-insensitive", "linux/arm/V7", "linux/arm/v7", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MatchesPlatform(tt.p1, tt.p2)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMatchesPlatform_Errors(t *testing.T) {
	tests := []struct {
		name        string
		p1          string
		p2          string
		errContains string
	}{
		{"invalid p1 empty", "", "linux/amd64", "invalid platform1"},
		{"invalid p2 bad format", "linux/amd64", "linux", "invalid platform2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := MatchesPlatform(tt.p1, tt.p2)
			assert.False(t, ok)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errContains)
		})
	}
}

func TestFilterPlatforms(t *testing.T) {
	in := []string{
		"linux/amd64",
		"windows/amd64",
		"linux/unknown",
		"",
		"darwin/arm64",
		"linux",
	}
	valid, invalid := FilterPlatforms(in)

	assert.Equal(t, []string{"linux/amd64", "windows/amd64", "darwin/arm64"}, valid)
	assert.Equal(t, []string{"linux/unknown", "", "linux"}, invalid)
}

func Test_isValidOS(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"linux", true},
		{"LINUX", true},
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
		{"bogus", false},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			assert.Equal(t, tt.want, isValidOS(tt.in))
		})
	}
}

func Test_isValidArch(t *testing.T) {
	tests := []struct {
		in   string
		want bool
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
		{"X86_64", false}, // alias not supported by isValidArch
		{"bogus", false},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			assert.Equal(t, tt.want, isValidArch(tt.in))
		})
	}
}
