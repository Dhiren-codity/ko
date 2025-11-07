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
			p:    Platform{OS: "linux", Architecture: "arm64", Variant: "v8"},
			want: "linux/arm64/v8",
		},
		{
			name: "with uppercase fields preserved",
			p:    Platform{OS: "LINUX", Architecture: "AMD64", Variant: "V8"},
			want: "LINUX/AMD64/V8",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.p.String())
		})
	}
}

func TestParsePlatform_Valid(t *testing.T) {
	tests := []struct {
		name        string
		in          string
		wantOS      string
		wantArch    string
		wantVariant string
	}{
		{
			name:     "simple two parts",
			in:       "linux/amd64",
			wantOS:   "linux",
			wantArch: "amd64",
		},
		{
			name:        "three parts with variant",
			in:          "linux/arm64/v8",
			wantOS:      "linux",
			wantArch:    "arm64",
			wantVariant: "v8",
		},
		{
			name:     "uppercase components allowed (no normalization here)",
			in:       "LINUX/AMD64",
			wantOS:   "LINUX",
			wantArch: "AMD64",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePlatform(tt.in)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantOS, got.OS)
			assert.Equal(t, tt.wantArch, got.Architecture)
			assert.Equal(t, tt.wantVariant, got.Variant)
		})
	}
}

func TestParsePlatform_Invalid(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{name: "empty", in: ""},
		{name: "not enough parts", in: "linux"},
		{name: "too many parts", in: "linux/amd64/too/many"},
		{name: "empty OS", in: "/amd64"},
		{name: "empty arch", in: "linux/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := ParsePlatform(tt.in)
			assert.Error(t, err)
			assert.Nil(t, p)
		})
	}
}

func TestToPlatform(t *testing.T) {
	vp := v1.Platform{OS: "linux", Architecture: "arm", Variant: "v7"}
	got := ToPlatform(vp)
	assert.Equal(t, "linux", got.OS)
	assert.Equal(t, "arm", got.Architecture)
	assert.Equal(t, "v7", got.Variant)
}

func TestPlatform_ToV1Platform(t *testing.T) {
	p := Platform{OS: "darwin", Architecture: "arm64", Variant: "v8"}
	vp := p.ToV1Platform()
	assert.Equal(t, "darwin", vp.OS)
	assert.Equal(t, "arm64", vp.Architecture)
	assert.Equal(t, "v8", vp.Variant)
}

func TestIsValidPlatform(t *testing.T) {
	tests := []struct {
		in    string
		valid bool
	}{
		{"linux/amd64", true},
		{"LINUX/AMD64", true}, // case-insensitive check
		{"darwin/arm64", true},
		{"linux/wasm", true},            // recognized arch even if unusual combo
		{"linux", false},                // invalid format
		{"/amd64", false},               // empty OS
		{"linux/", false},               // empty arch
		{"linux/amd64/too/many", false}, // too many parts
		{"foo/amd64", false},            // unknown OS
		{"linux/foo", false},            // unknown arch
	}
	for _, tt := range tests {
		assert.Equal(t, tt.valid, IsValidPlatform(tt.in), tt.in)
	}
}

func TestIsValidPlatform_AllRecognizedOSes(t *testing.T) {
	recognizedOS := []string{
		"linux", "darwin", "windows", "freebsd", "openbsd", "netbsd", "plan9",
		"solaris", "aix", "android", "ios", "js", "wasip1",
	}
	for _, os := range recognizedOS {
		s := os + "/amd64"
		assert.True(t, IsValidPlatform(s), "expected valid OS: %s", s)
	}
}

func TestIsValidPlatform_AllRecognizedArches(t *testing.T) {
	recognizedArch := []string{
		"386", "amd64", "arm", "arm64", "ppc64", "ppc64le", "mips", "mipsle",
		"mips64", "mips64le", "s390x", "riscv64", "wasm", "loong64",
	}
	for _, arch := range recognizedArch {
		s := "linux/" + arch
		assert.True(t, IsValidPlatform(s), "expected valid Arch: %s", s)
	}
}

func TestGetHostPlatform(t *testing.T) {
	p := GetHostPlatform()
	assert.Equal(t, runtime.GOOS, p.OS)
	assert.Equal(t, runtime.GOARCH, p.Architecture)
	// Variant is not set by GetHostPlatform; allow empty
}

func TestNormalizePlatform(t *testing.T) {
	tests := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"Linux/Arm64/V8", "linux/arm64/v8", false},
		{"linux/amd64", "linux/amd64", false},
		{"", "", true},
		{"/amd64", "", true},
		{"linux/", "", true},
	}
	for _, tt := range tests {
		got, err := NormalizePlatform(tt.in)
		if tt.wantErr {
			assert.Error(t, err, tt.in)
		} else {
			assert.NoError(t, err, tt.in)
			assert.Equal(t, tt.want, got, tt.in)
		}
	}
}

func TestMatchesPlatform(t *testing.T) {
	tests := []struct {
		name    string
		p1      string
		p2      string
		match   bool
		wantErr bool
	}{
		{
			name:  "identical",
			p1:    "linux/amd64",
			p2:    "linux/amd64",
			match: true,
		},
		{
			name:  "case-insensitive",
			p1:    "LINUX/AMD64",
			p2:    "linux/amd64",
			match: true,
		},
		{
			name:  "different arch",
			p1:    "linux/amd64",
			p2:    "linux/arm64",
			match: false,
		},
		{
			name:  "different OS",
			p1:    "linux/amd64",
			p2:    "darwin/amd64",
			match: false,
		},
		{
			name:  "both no variant",
			p1:    "linux/arm64",
			p2:    "linux/arm64",
			match: true,
		},
		{
			name:  "both with same variant (case-insensitive)",
			p1:    "linux/arm64/v8",
			p2:    "linux/arm64/V8",
			match: true,
		},
		{
			name:  "one variant missing",
			p1:    "linux/arm64/v8",
			p2:    "linux/arm64",
			match: false,
		},
		{
			name:  "different variants",
			p1:    "linux/arm64/v8",
			p2:    "linux/arm64/v7",
			match: false,
		},
		{
			name:    "invalid platform1",
			p1:      "",
			p2:      "linux/amd64",
			wantErr: true,
		},
		{
			name:    "invalid platform2",
			p1:      "linux/amd64",
			p2:      "",
			wantErr: true,
		},
		{
			name:  "arm without variants but same arch",
			p1:    "linux/arm",
			p2:    "linux/arm",
			match: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := MatchesPlatform(tt.p1, tt.p2)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.match, match)
		})
	}
}

func TestFilterPlatforms(t *testing.T) {
	input := []string{
		"linux/amd64",
		"darwin/arm64",
		"LINUX/AMD64",
		"linux/",
		"/amd64",
		"linux/arm64/v8",
		"foo/bar",
		"linux/wasm",
	}
	valid, invalid := FilterPlatforms(input)

	assert.ElementsMatch(t, []string{
		"linux/amd64",
		"darwin/arm64",
		"LINUX/AMD64",
		"linux/arm64/v8",
		"linux/wasm",
	}, valid)

	assert.ElementsMatch(t, []string{
		"linux/",
		"/amd64",
		"foo/bar",
	}, invalid)
}
