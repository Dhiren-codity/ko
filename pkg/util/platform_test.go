package util

import (
	"fmt"
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
			name: "without variant",
			p:    Platform{OS: "linux", Architecture: "amd64"},
			want: "linux/amd64",
		},
		{
			name: "with variant",
			p:    Platform{OS: "linux", Architecture: "arm", Variant: "v7"},
			want: "linux/arm/v7",
		},
		{
			name: "empty fields produce slash pair",
			p:    Platform{OS: "darwin", Architecture: "arm64", Variant: ""},
			want: "darwin/arm64",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParsePlatform_Success(t *testing.T) {
	tests := []struct {
		in     string
		expect Platform
	}{
		{"linux/amd64", Platform{OS: "linux", Architecture: "amd64", Variant: ""}},
		{"linux/arm64/v8", Platform{OS: "linux", Architecture: "arm64", Variant: "v8"}},
		{"linux/amd64/", Platform{OS: "linux", Architecture: "amd64", Variant: ""}},
		{"LINUX/AMD64", Platform{OS: "LINUX", Architecture: "AMD64", Variant: ""}}, // ParsePlatform preserves case
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := ParsePlatform(tt.in)
			assert.NoError(t, err)
			assert.Equal(t, tt.expect, *got)
		})
	}
}

func TestParsePlatform_Errors(t *testing.T) {
	tests := []struct {
		in       string
		contains string
	}{
		{"", "platform cannot be empty"},
		{"linux", "invalid platform format"},
		{"a/b/c/d", "too many components"},
		{"/amd64", "platform OS cannot be empty"},
		{"linux/", "platform architecture cannot be empty"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := ParsePlatform(tt.in)
			assert.Nil(t, got)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.contains)
		})
	}
}

func TestToPlatform(t *testing.T) {
	vp := v1.Platform{OS: "linux", Architecture: "arm64", Variant: "v8"}
	p := ToPlatform(vp)
	assert.Equal(t, "linux", p.OS)
	assert.Equal(t, "arm64", p.Architecture)
	assert.Equal(t, "v8", p.Variant)
}

func TestPlatform_ToV1Platform(t *testing.T) {
	p := Platform{OS: "windows", Architecture: "386", Variant: ""}
	vp := p.ToV1Platform()
	assert.Equal(t, "windows", vp.OS)
	assert.Equal(t, "386", vp.Architecture)
	assert.Equal(t, "", vp.Variant)
}

func TestIsValidPlatform_True(t *testing.T) {
	valid := []string{
		"linux/amd64",
		"LINUX/AMD64", // case-insensitive match
		"windows/386",
		"linux/arm64/v8",
		"darwin/arm64",
		"freebsd/amd64",
		"js/wasm",
		"wasip1/wasm",
		"linux/loong64",
	}
	for _, in := range valid {
		t.Run(in, func(t *testing.T) {
			assert.True(t, IsValidPlatform(in))
		})
	}
}

func TestIsValidPlatform_False(t *testing.T) {
	invalid := []string{
		"",                      // empty
		"linux",                 // wrong format
		"linux/",                // empty arch
		"/amd64",                // empty os
		"linux/amd64/extra/seg", // too many components
		"beos/amd64",            // invalid OS
		"linux/foo",             // invalid arch
	}
	for _, in := range invalid {
		t.Run(in, func(t *testing.T) {
			assert.False(t, IsValidPlatform(in))
		})
	}
}

func TestGetHostPlatform(t *testing.T) {
	hp := GetHostPlatform()
	assert.Equal(t, runtime.GOOS, hp.OS)
	assert.Equal(t, runtime.GOARCH, hp.Architecture)
	assert.Equal(t, "", hp.Variant)

	expected := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
	assert.Equal(t, expected, hp.String())
}

func TestNormalizePlatform(t *testing.T) {
	tests := []struct {
		in       string
		want     string
		wantErr  bool
		errMatch string
	}{
		{"LINUX/AMD64", "linux/amd64", false, ""},
		{"Linux/Arm64/V8", "linux/arm64/v8", false, ""},
		{"linux/amd64/", "linux/amd64", false, ""},
		{"", "", true, "platform cannot be empty"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := NormalizePlatform(tt.in)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMatch)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMatchesPlatform(t *testing.T) {
	type tc struct {
		p1     string
		p2     string
		match  bool
		errStr string
	}
	tests := []tc{
		{"linux/amd64", "LINUX/AMD64", true, ""},
		{"linux/ARM64/v8", "linux/arm64/V8", true, ""},
		{"linux/arm64/v8", "linux/arm64", false, ""},
		{"linux/arm64/v8", "linux/arm64/v7", false, ""},
		{"linux/amd64", "windows/amd64", false, ""},
		{"linux/amd64", "linux/386", false, ""},
		{"linux/arm/v7", "linux/arm/v7", true, ""},
		{"linux/arm/v7", "linux/arm", false, ""},
		{"", "linux/amd64", false, "invalid platform1"},
		{"linux/amd64", "", false, "invalid platform2"},
		{"", "", false, "invalid platform1"},
	}
	for _, tt := range tests {
		name := tt.p1 + " vs " + tt.p2
		t.Run(name, func(t *testing.T) {
			ok, err := MatchesPlatform(tt.p1, tt.p2)
			if tt.errStr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errStr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.match, ok)
		})
	}
}

func TestFilterPlatforms(t *testing.T) {
	input := []string{
		"linux/amd64",
		"linux/foo",
		"windows/386",
		"bogus/amd64",
		"linux/amd64/", // valid even with trailing slash as empty variant
	}
	valid, invalid := FilterPlatforms(input)

	assert.Equal(t, []string{"linux/amd64", "windows/386", "linux/amd64/"}, valid)
	assert.Equal(t, []string{"linux/foo", "bogus/amd64"}, invalid)
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
		{"ios", true},
		{"beos", false},
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
		{"amd64", true},
		{"AMD64", true},
		{"386", true},
		{"arm", true},
		{"arm64", true},
		{"riscv64", true},
		{"loong64", true},
		{"x86_64", false},
		{"foo", false},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			assert.Equal(t, tt.want, isValidArch(tt.in))
		})
	}
}
