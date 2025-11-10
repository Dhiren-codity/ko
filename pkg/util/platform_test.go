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
			p:    Platform{OS: "linux", Architecture: "amd64"},
			want: "linux/amd64",
		},
		{
			name: "with variant",
			p:    Platform{OS: "linux", Architecture: "arm", Variant: "v7"},
			want: "linux/arm/v7",
		},
		{
			name: "empty fields",
			p:    Platform{},
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
		{"linux/arm/v7", "linux", "arm", "v7"},
		{"Linux/AMD64", "Linux", "AMD64", ""},
		{"linux/arm/", "linux", "arm", ""},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			p, err := ParsePlatform(tt.in)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantOS, p.OS)
			assert.Equal(t, tt.wantArch, p.Architecture)
			assert.Equal(t, tt.wantVar, p.Variant)
		})
	}
}

func TestParsePlatform_Errors(t *testing.T) {
	tests := []struct {
		name        string
		in          string
		errContains string
	}{
		{"empty", "", "platform cannot be empty"},
		{"too few components", "linux", "invalid platform format"},
		{"too many components", "linux/amd64/extra/too", "too many components"},
		{"empty os", "/amd64", "platform OS cannot be empty"},
		{"empty arch", "linux/", "platform architecture cannot be empty"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := ParsePlatform(tt.in)
			assert.Error(t, err)
			assert.Nil(t, p)
			assert.Contains(t, err.Error(), tt.errContains)
		})
	}
}

func TestToPlatform_FromV1(t *testing.T) {
	vp := v1.Platform{
		OS:           "linux",
		Architecture: "arm64",
		Variant:      "v8",
	}
	got := ToPlatform(vp)
	assert.Equal(t, "linux", got.OS)
	assert.Equal(t, "arm64", got.Architecture)
	assert.Equal(t, "v8", got.Variant)
}

func TestPlatform_ToV1Platform(t *testing.T) {
	p := Platform{OS: "windows", Architecture: "amd64", Variant: ""}
	vp := p.ToV1Platform()
	assert.Equal(t, "windows", vp.OS)
	assert.Equal(t, "amd64", vp.Architecture)
	assert.Equal(t, "", vp.Variant)
}

func TestIsValidPlatform(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"linux/amd64", true},
		{"Linux/AMD64", true},
		{"darwin/arm64", true},
		{"plan9/amd64", true},
		{"windows/386", true},
		{"linux/arm/v7", true}, // variant supported but not required by validator
		{"linux/arm", true},    // still considered valid as arch is recognized
		{"foo/bar", false},
		{"linux/unknown", false},
		{"unknown/amd64", false},
		{"", false},
		{"linux", false},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			assert.Equal(t, tt.want, IsValidPlatform(tt.in))
		})
	}
}

func TestGetHostPlatform(t *testing.T) {
	p := GetHostPlatform()
	assert.Equal(t, runtime.GOOS, p.OS)
	assert.Equal(t, runtime.GOARCH, p.Architecture)
	// Current implementation does not set Variant
	assert.Equal(t, "", p.Variant)
}

func TestNormalizePlatform_Success(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"Linux/AMD64", "linux/amd64"},
		{"LINUX/ARM/v7", "linux/arm/v7"},
		{"linux/arm/V7", "linux/arm/v7"},
		{"linux/arm/", "linux/arm"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := NormalizePlatform(tt.in)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNormalizePlatform_Error(t *testing.T) {
	_, err := NormalizePlatform("")
	assert.Error(t, err)
}

func TestMatchesPlatform_SuccessAndFailure(t *testing.T) {
	tests := []struct {
		name  string
		p1    string
		p2    string
		match bool
	}{
		{"exact match no variant", "linux/amd64", "linux/amd64", true},
		{"case insensitive", "Linux/AMD64", "linux/amd64", true},
		{"os mismatch", "linux/amd64", "windows/amd64", false},
		{"arch mismatch", "linux/arm64", "linux/amd64", false},
		{"both variants equal", "linux/arm/v7", "linux/arm/v7", true},
		{"variant case insensitive", "linux/arm/V7", "linux/arm/v7", true},
		{"one has variant, one does not", "linux/arm/v7", "linux/arm", false},
		{"different variants", "linux/arm/v7", "linux/arm/v6", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MatchesPlatform(tt.p1, tt.p2)
			assert.NoError(t, err)
			assert.Equal(t, tt.match, got)
		})
	}
}

func TestMatchesPlatform_InvalidInputs(t *testing.T) {
	_, err := MatchesPlatform("", "linux/amd64")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid platform1")

	_, err = MatchesPlatform("linux/amd64", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid platform2")
}

func TestFilterPlatforms(t *testing.T) {
	input := []string{
		"linux/amd64",
		"windows/arm64",
		"foo/bar",
		"linux/unknown",
		"darwin/amd64",
	}
	valid, invalid := FilterPlatforms(input)

	assert.Equal(t, []string{"linux/amd64", "windows/arm64", "darwin/amd64"}, valid)
	assert.Equal(t, []string{"foo/bar", "linux/unknown"}, invalid)
}

func Test_isValidOS(t *testing.T) {
	assert.True(t, isValidOS("linux"))
	assert.True(t, isValidOS("LINUX"))
	assert.True(t, isValidOS("darwin"))
	assert.False(t, isValidOS("not-an-os"))
}

func Test_isValidArch(t *testing.T) {
	assert.True(t, isValidArch("amd64"))
	assert.True(t, isValidArch("AMD64"))
	assert.True(t, isValidArch("arm64"))
	assert.True(t, isValidArch("386"))
	assert.False(t, isValidArch("unknown-arch"))
}
