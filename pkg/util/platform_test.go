package util

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"

	v1 "github.com/google/go-containerregistry/pkg/v1"
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
			name: "preserves case",
			p:    Platform{OS: "Linux", Architecture: "AMD64"},
			want: "Linux/AMD64",
		},
		{
			name: "preserves variant case",
			p:    Platform{OS: "Linux", Architecture: "ARM", Variant: "V7"},
			want: "Linux/ARM/V7",
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
		name   string
		in     string
		wantOS string
		wantAr string
		wantV  string
		wantS  string
	}{
		{
			name:   "simple",
			in:     "linux/amd64",
			wantOS: "linux",
			wantAr: "amd64",
			wantV:  "",
			wantS:  "linux/amd64",
		},
		{
			name:   "with variant",
			in:     "linux/arm/v7",
			wantOS: "linux",
			wantAr: "arm",
			wantV:  "v7",
			wantS:  "linux/arm/v7",
		},
		{
			name:   "preserve case",
			in:     "Linux/AMD64",
			wantOS: "Linux",
			wantAr: "AMD64",
			wantV:  "",
			wantS:  "Linux/AMD64",
		},
		{
			name:   "preserve variant case",
			in:     "Linux/ARM/V7",
			wantOS: "Linux",
			wantAr: "ARM",
			wantV:  "V7",
			wantS:  "Linux/ARM/V7",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePlatform(tt.in)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantOS, got.OS)
			assert.Equal(t, tt.wantAr, got.Architecture)
			assert.Equal(t, tt.wantV, got.Variant)
			assert.Equal(t, tt.wantS, got.String())
		})
	}
}

func TestParsePlatform_Invalid(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		errPart string
	}{
		{
			name:    "empty",
			in:      "",
			errPart: "platform cannot be empty",
		},
		{
			name:    "missing arch",
			in:      "linux",
			errPart: "invalid platform format",
		},
		{
			name:    "too many parts",
			in:      "linux/amd64/extra/more",
			errPart: "too many components",
		},
		{
			name:    "empty os",
			in:      "/amd64",
			errPart: "platform OS cannot be empty",
		},
		{
			name:    "empty arch",
			in:      "linux/",
			errPart: "platform architecture cannot be empty",
		},
		{
			name:    "empty arch in three parts",
			in:      "linux//arm",
			errPart: "platform architecture cannot be empty",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePlatform(tt.in)
			assert.Error(t, err)
			assert.Nil(t, got)
			assert.Contains(t, err.Error(), tt.errPart)
		})
	}
}

func TestToPlatform(t *testing.T) {
	vp := v1.Platform{
		OS:           "Linux",
		Architecture: "AMD64",
		Variant:      "V8",
	}
	got := ToPlatform(vp)
	assert.Equal(t, "Linux", got.OS)
	assert.Equal(t, "AMD64", got.Architecture)
	assert.Equal(t, "V8", got.Variant)
}

func TestToV1Platform(t *testing.T) {
	p := Platform{
		OS:           "Linux",
		Architecture: "AMD64",
		Variant:      "V7",
	}
	got := p.ToV1Platform()
	assert.Equal(t, "Linux", got.OS)
	assert.Equal(t, "AMD64", got.Architecture)
	assert.Equal(t, "V7", got.Variant)
}

func TestIsValidPlatform(t *testing.T) {
	valid := []string{
		"linux/amd64",
		"darwin/arm64",
		"windows/386",
		"freebsd/arm",
		"openbsd/mipsle",
		"netbsd/ppc64",
		"plan9/amd64",
		"solaris/s390x",
		"aix/ppc64le",
		"android/arm64",
		"ios/arm64",
		"js/wasm",
		"wasip1/wasm",
		"LINUX/AMD64",
		"linux/arm/v7",
		"linux/arm/v9",
		"linux/amd64/foo",
	}
	for _, s := range valid {
		assert.Truef(t, IsValidPlatform(s), "expected valid: %s", s)
	}

	invalid := []string{
		"beos/amd64",
		"linux/x86_64",
		"linux//arm",
		"/amd64",
		"linux/",
		"",
		"linux",
		"linux/amd64/too/many",
	}
	for _, s := range invalid {
		assert.Falsef(t, IsValidPlatform(s), "expected invalid: %s", s)
	}
}

func TestGetHostPlatform(t *testing.T) {
	h := GetHostPlatform()
	assert.Equal(t, runtime.GOOS, h.OS)
	assert.Equal(t, runtime.GOARCH, h.Architecture)
}

func TestNormalizePlatform(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
		errPart string
	}{
		{
			name: "lowercase simple",
			in:   "LINUX/AMD64",
			want: "linux/amd64",
		},
		{
			name: "lowercase with variant",
			in:   "Linux/Arm/V7",
			want: "linux/arm/v7",
		},
		{
			name:    "empty",
			in:      "",
			wantErr: true,
			errPart: "platform cannot be empty",
		},
		{
			name:    "missing arch",
			in:      "linux",
			wantErr: true,
			errPart: "invalid platform format",
		},
		{
			name:    "too many parts",
			in:      "linux/amd64/extra/more",
			wantErr: true,
			errPart: "too many components",
		},
		{
			name:    "empty arch",
			in:      "linux/",
			wantErr: true,
			errPart: "platform architecture cannot be empty",
		},
		{
			name:    "empty arch in 3 parts",
			in:      "linux//arm",
			wantErr: true,
			errPart: "platform architecture cannot be empty",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizePlatform(tt.in)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errPart)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestMatchesPlatform(t *testing.T) {
	tests := []struct {
		name    string
		p1      string
		p2      string
		want    bool
		wantErr bool
		errPart string
	}{
		{
			name: "exact match no variant",
			p1:   "linux/amd64",
			p2:   "linux/amd64",
			want: true,
		},
		{
			name: "case-insensitive match",
			p1:   "Linux/AMD64",
			p2:   "linux/amd64",
			want: true,
		},
		{
			name: "different os",
			p1:   "linux/amd64",
			p2:   "windows/amd64",
			want: false,
		},
		{
			name: "different arch",
			p1:   "linux/arm",
			p2:   "linux/amd64",
			want: false,
		},
		{
			name: "both variants equal (case-insensitive)",
			p1:   "linux/arm/v7",
			p2:   "linux/arm/V7",
			want: true,
		},
		{
			name: "one variant missing",
			p1:   "linux/arm/v7",
			p2:   "linux/arm",
			want: false,
		},
		{
			name: "different variants",
			p1:   "linux/arm/v6",
			p2:   "linux/arm/v7",
			want: false,
		},
		{
			name:    "invalid platform1",
			p1:      "linux",
			p2:      "linux/amd64",
			want:    false,
			wantErr: true,
			errPart: "invalid platform1",
		},
		{
			name:    "invalid platform2",
			p1:      "linux/amd64",
			p2:      "linux",
			want:    false,
			wantErr: true,
			errPart: "invalid platform2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MatchesPlatform(tt.p1, tt.p2)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errPart)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFilterPlatforms(t *testing.T) {
	input := []string{
		"linux/amd64",
		"linux/x86_64",
		"beos/amd64",
		"darwin/arm64",
		"Linux/ARM/v7",
		"linux//arm",
	}
	valid, invalid := FilterPlatforms(input)

	// Expect original strings preserved and order maintained within each category
	wantValid := []string{"linux/amd64", "darwin/arm64", "Linux/ARM/v7"}
	wantInvalid := []string{"linux/x86_64", "beos/amd64", "linux//arm"}

	assert.Equal(t, wantValid, valid)
	assert.Equal(t, wantInvalid, invalid)
}
