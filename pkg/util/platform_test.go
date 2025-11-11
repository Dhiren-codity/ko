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
			name:     "preserve casing",
			p:        Platform{OS: "LINUX", Architecture: "ARM64", Variant: "V8"},
			expected: "LINUX/ARM64/V8",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.p.String())
		})
	}
}

func TestParsePlatform_Success(t *testing.T) {
	tests := []struct {
		in       string
		os       string
		arch     string
		variant  string
		expected string
	}{
		{"linux/amd64", "linux", "amd64", "", "linux/amd64"},
		{"linux/arm64/v8", "linux", "arm64", "v8", "linux/arm64/v8"},
		{"LINUX/ARM64/V8", "LINUX", "ARM64", "V8", "LINUX/ARM64/V8"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.in, func(t *testing.T) {
			p, err := ParsePlatform(tt.in)
			assert.NoError(t, err)
			assert.Equal(t, tt.os, p.OS)
			assert.Equal(t, tt.arch, p.Architecture)
			assert.Equal(t, tt.variant, p.Variant)
			assert.Equal(t, tt.expected, p.String())
		})
	}
}

func TestParsePlatform_Errors(t *testing.T) {
	tests := []struct {
		name        string
		in          string
		errContains string
	}{
		{"empty input", "", "platform cannot be empty"},
		{"not enough components", "linux", "invalid platform format: linux (expected os/arch or os/arch/variant)"},
		{"too many components", "linux/amd64/too/many", "invalid platform format: linux/amd64/too/many (too many components)"},
		{"empty arch", "linux/", "platform architecture cannot be empty"},
		{"empty os", "/amd64", "platform OS cannot be empty"},
		{"empty arch with variant", "linux//v7", "platform architecture cannot be empty"},
		{"empty os with variant", "/amd64/v8", "platform OS cannot be empty"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			p, err := ParsePlatform(tt.in)
			assert.Nil(t, p)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.errContains)
		})
	}
}

func TestToPlatform_And_ToV1Platform_RoundTrip(t *testing.T) {
	vp := v1.Platform{OS: "linux", Architecture: "arm64", Variant: "v8"}
	p := ToPlatform(vp)
	assert.Equal(t, "linux", p.OS)
	assert.Equal(t, "arm64", p.Architecture)
	assert.Equal(t, "v8", p.Variant)

	back := p.ToV1Platform()
	assert.Equal(t, vp, back)

	// Also test with no variant
	vp2 := v1.Platform{OS: "windows", Architecture: "amd64"}
	p2 := ToPlatform(vp2)
	assert.Equal(t, "", p2.Variant)
	assert.Equal(t, vp2, p2.ToV1Platform())
}

func TestIsValidPlatform(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"linux/amd64", true},
		{"LINUX/AMD64", true},
		{"darwin/amd64", true},
		{"windows/arm64", true},
		{"wasip1/wasm", true},
		{"linux/arm/v7", true},
		{"unknown/amd64", false},
		{"linux/unknown", false},
		{"linux", false},
		{"linux//v7", false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.in, func(t *testing.T) {
			assert.Equal(t, tt.want, IsValidPlatform(tt.in))
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
		in       string
		want     string
		wantErr  bool
		errMatch string
	}{
		{"LiNuX/AmD64", "linux/amd64", false, ""},
		{"LINUX/ARM/V7", "linux/arm/v7", false, ""},
		{"windows/ARM64", "windows/arm64", false, ""},
		{"bad", "", true, "invalid platform format"},
	}
	for _, tt := range tests {
		tt := tt
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
	tests := []struct {
		name        string
		p1          string
		p2          string
		want        bool
		wantErr     bool
		errContains string
	}{
		{"same os/arch case-insensitive", "LINUX/AMD64", "linux/amd64", true, false, ""},
		{"mismatched os", "linux/amd64", "windows/amd64", false, false, ""},
		{"mismatched arch", "linux/amd64", "linux/arm64", false, false, ""},
		{"both no variant", "linux/arm", "linux/arm", true, false, ""},
		{"same variant case-insensitive", "linux/arm/v7", "Linux/ARM/V7", true, false, ""},
		{"one has variant other not", "linux/arm", "linux/arm/v7", false, false, ""},
		{"different variants", "linux/arm/v7", "linux/arm/v8", false, false, ""},
		{"invalid platform1", "invalid", "linux/amd64", false, true, "invalid platform1"},
		{"invalid platform2", "linux/amd64", "bad", false, true, "invalid platform2"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := MatchesPlatform(tt.p1, tt.p2)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFilterPlatforms(t *testing.T) {
	input := []string{
		"linux/amd64",
		"LINUX/ARM64",
		"linux//v7",
		"unknown/amd64",
		"windows/amd64",
		"",
		"linux/arm/v7",
	}
	valid, invalid := FilterPlatforms(input)

	expectedValid := []string{
		"linux/amd64",
		"LINUX/ARM64",
		"windows/amd64",
		"linux/arm/v7",
	}
	expectedInvalid := []string{
		"linux//v7",
		"unknown/amd64",
		"",
	}

	assert.Equal(t, expectedValid, valid)
	assert.Equal(t, expectedInvalid, invalid)
}
