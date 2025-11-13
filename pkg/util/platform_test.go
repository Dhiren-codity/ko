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
			name: "empty fields",
			p:    Platform{},
			want: "/",
		},
		{
			name: "empty variant not included",
			p:    Platform{OS: "windows", Architecture: "arm64", Variant: ""},
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
		in   string
		want Platform
	}{
		{"linux/amd64", Platform{OS: "linux", Architecture: "amd64"}},
		{"linux/arm/v7", Platform{OS: "linux", Architecture: "arm", Variant: "v7"}},
		{"DARWIN/AMD64", Platform{OS: "DARWIN", Architecture: "AMD64"}}, // ParsePlatform preserves case and doesn't validate values
		{"windows/arm64", Platform{OS: "windows", Architecture: "arm64"}},
		{"freebsd/386", Platform{OS: "freebsd", Architecture: "386"}},
	}
	for _, tt := range tests {
		got, err := ParsePlatform(tt.in)
		assert.NoError(t, err, "ParsePlatform(%q) unexpected error", tt.in)
		assert.Equal(t, tt.want.OS, got.OS)
		assert.Equal(t, tt.want.Architecture, got.Architecture)
		assert.Equal(t, tt.want.Variant, got.Variant)
	}
}

func TestParsePlatform_Errors(t *testing.T) {
	tests := []struct {
		in       string
		errParts []string
	}{
		{"", []string{"platform cannot be empty"}},
		{"linux", []string{"invalid platform format"}},
		{"a/b/c/d", []string{"invalid platform format", "too many components"}},
		{"/amd64", []string{"platform OS cannot be empty"}},
		{"linux/", []string{"platform architecture cannot be empty"}},
	}
	for _, tt := range tests {
		_, err := ParsePlatform(tt.in)
		if assert.Error(t, err, "expected error for %q", tt.in) {
			for _, part := range tt.errParts {
				assert.Contains(t, err.Error(), part)
			}
		}
	}
}

func TestToPlatformAndToV1Platform(t *testing.T) {
	// v1 -> our Platform
	vp := v1.Platform{OS: "linux", Architecture: "arm", Variant: "v7"}
	p := ToPlatform(vp)
	assert.Equal(t, "linux", p.OS)
	assert.Equal(t, "arm", p.Architecture)
	assert.Equal(t, "v7", p.Variant)

	// our Platform -> v1
	p2 := Platform{OS: "windows", Architecture: "amd64", Variant: ""}
	vp2 := p2.ToV1Platform()
	assert.Equal(t, "windows", vp2.OS)
	assert.Equal(t, "amd64", vp2.Architecture)
	assert.Equal(t, "", vp2.Variant)
}

func TestIsValidPlatform(t *testing.T) {
	valids := []string{
		"linux/amd64",
		"WINDOWS/arm64",
		"freebsd/386",
		"darwin/amd64",
		"linux/arm",       // variant optional and ignored by validator
		"linux/arm/v9",    // variant ignored by validator
		"openbsd/ppc64le", // os/arch recognized
		"wasip1/wasm",     // os/arch recognized
		"js/wasm",         // os/arch recognized
	}
	for _, v := range valids {
		assert.True(t, IsValidPlatform(v), "expected valid platform: %q", v)
	}

	invalids := []string{
		"foo/amd64",   // unknown OS
		"linux/foo",   // unknown arch
		"linux",       // invalid format
		"/amd64",      // empty OS
		"linux/",      // empty arch
		"plan9/sparc", // unknown arch
	}
	for _, inv := range invalids {
		assert.False(t, IsValidPlatform(inv), "expected invalid platform: %q", inv)
	}
}

func TestGetHostPlatform(t *testing.T) {
	host := GetHostPlatform()
	assert.Equal(t, runtime.GOOS, host.OS)
	assert.Equal(t, runtime.GOARCH, host.Architecture)
	assert.Equal(t, "", host.Variant)
}

func TestNormalizePlatform(t *testing.T) {
	tests := []struct {
		in       string
		want     string
		wantErr  bool
		errMatch string
	}{
		{"LiNuX/ARM/V7", "linux/arm/v7", false, ""},
		{"WINDOWS/AMD64", "windows/amd64", false, ""},
		{"darwin/amd64", "darwin/amd64", false, ""},
		{"linux", "", true, "invalid platform format"},
		{"", "", true, "platform cannot be empty"},
	}
	for _, tt := range tests {
		got, err := NormalizePlatform(tt.in)
		if tt.wantErr {
			if assert.Error(t, err, "expected error for %q", tt.in) {
				if tt.errMatch != "" {
					assert.Contains(t, err.Error(), tt.errMatch)
				}
			}
			continue
		}
		assert.NoError(t, err, "NormalizePlatform(%q) unexpected error", tt.in)
		assert.Equal(t, tt.want, got)
	}
}

func TestMatchesPlatform(t *testing.T) {
	tests := []struct {
		p1      string
		p2      string
		want    bool
		wantErr bool
	}{
		{"linux/amd64", "LINUX/AMD64", true, false},
		{"linux/arm/v7", "linux/arm/V7", true, false},
		{"windows/amd64", "linux/amd64", false, false},
		{"linux/arm/v7", "linux/arm/v6", false, false},
		{"linux/arm/v7", "linux/arm", false, false}, // one has variant, one doesn't
		{"linux", "linux/amd64", false, true},       // invalid platform1
		{"linux/amd64", "linux", false, true},       // invalid platform2
	}
	for _, tt := range tests {
		got, err := MatchesPlatform(tt.p1, tt.p2)
		if tt.wantErr {
			assert.Error(t, err, "expected error for %q vs %q", tt.p1, tt.p2)
			continue
		}
		assert.NoError(t, err, "unexpected error for %q vs %q", tt.p1, tt.p2)
		assert.Equal(t, tt.want, got, "match result for %q vs %q", tt.p1, tt.p2)
	}
}

func TestFilterPlatforms(t *testing.T) {
	input := []string{
		"linux/amd64",   // valid
		"foo/bar",       // invalid os
		"windows/arm64", // valid
		"linux",         // invalid format
		"/amd64",        // invalid os empty
		"linux/",        // invalid arch empty
		"darwin/amd64",  // valid
		"linux/arm",     // valid
		"linux/unknown", // invalid arch
	}
	wantValid := []string{
		"linux/amd64",
		"windows/arm64",
		"darwin/amd64",
		"linux/arm",
	}
	wantInvalid := []string{
		"foo/bar",
		"linux",
		"/amd64",
		"linux/",
		"linux/unknown",
	}

	valid, invalid := FilterPlatforms(input)
	assert.Equal(t, wantValid, valid)
	assert.Equal(t, wantInvalid, invalid)
}

func Test_isValidOS(t *testing.T) {
	assert.True(t, isValidOS("LINUX"))
	assert.True(t, isValidOS("darwin"))
	assert.True(t, isValidOS("windows"))
	assert.False(t, isValidOS("unknownos"))
}

func Test_isValidArch(t *testing.T) {
	assert.True(t, isValidArch("AMD64"))
	assert.True(t, isValidArch("arm64"))
	assert.True(t, isValidArch("arm"))
	assert.False(t, isValidArch("sparc"))
}
