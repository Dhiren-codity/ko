package util

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidTag(t *testing.T) {
	tests := []struct {
		name string
		tag  string
		want bool
	}{
		{"simple lowercase", "abc", true},
		{"mixed case and specials", "A_b.C-d", true},
		{"semver style", "v1.2.3", true},
		{"empty", "", false},
		{"too long", strings.Repeat("a", MaxTagLength+1), false},
		{"starts with dash invalid", "-abc", false},
		{"starts with underscore invalid", "_abc", false},
		{"starts with dot invalid", ".abc", false},
		{"trailing separator allowed by regex", "abc-", true},
		{"contains invalid space", "ab c", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsValidTag(tt.tag))
		})
	}
}

func TestNormalizeTag(t *testing.T) {
	tests := []struct {
		name       string
		tag        string
		want       string
		wantErr    bool
		errContain string
	}{
		{
			name:       "empty input",
			tag:        "",
			wantErr:    true,
			errContain: "tag cannot be empty",
		},
		{
			name: "uppercase to lowercase and keep specials",
			tag:  "HELLO.World-01",
			want: "hello.world-01",
		},
		{
			name: "leading dot removed",
			tag:  ".DotStart",
			want: "dotstart",
		},
		{
			name:       "only specials error after sanitize",
			tag:        "...",
			wantErr:    true,
			errContain: "unable to sanitize tag",
		},
		{
			name: "slashes and unicode handled",
			tag:  "FeaTure/naïve",
			want: "feature-na-ve",
		},
		{
			name: "long lowers then truncates",
			tag:  strings.ToUpper(strings.Repeat("a", MaxTagLength+10)),
			want: strings.Repeat("a", MaxTagLength),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeTag(tt.tag)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContain != "" {
					assert.Contains(t, err.Error(), tt.errContain)
				}
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.True(t, IsValidTag(got), "normalized tag should be valid")
			assert.LessOrEqual(t, len(got), MaxTagLength)
			assert.Equal(t, strings.ToLower(got), got, "normalized must be lowercase")
		})
	}
}

func TestAppendSuffix(t *testing.T) {
	tests := []struct {
		name       string
		tag        string
		suffix     string
		want       string
		wantErr    bool
		errContain string
	}{
		{
			name:       "empty tag error",
			tag:        "",
			suffix:     "sfx",
			wantErr:    true,
			errContain: "tag cannot be empty",
		},
		{
			name:   "empty suffix returns tag",
			tag:    "base",
			suffix: "",
			want:   "base",
		},
		{
			name:   "suffix without separator gets dash",
			tag:    "base",
			suffix: "sfx",
			want:   "base-sfx",
		},
		{
			name:   "suffix with dot kept",
			tag:    "tag",
			suffix: ".meta",
			want:   "tag.meta",
		},
		{
			name:    "combined invalid due to invalid suffix character",
			tag:     "tag",
			suffix:  " invalid",
			wantErr: true,
		},
		{
			name:   "truncates base to fit suffix",
			tag:    strings.Repeat("a", MaxTagLength-1),
			suffix: "b",
			want:   strings.Repeat("a", MaxTagLength-2) + "-b",
		},
		{
			name:       "suffix too long produces error",
			tag:        "base",
			suffix:     "-" + strings.Repeat("x", MaxTagLength+10),
			wantErr:    true,
			errContain: "suffix too long",
		},
		{
			name:       "truncate base fails when base cannot be made valid",
			tag:        strings.Repeat("-", MaxTagLength+20),
			suffix:     "-sfx",
			wantErr:    true,
			errContain: "failed to truncate tag for suffix",
		},
		{
			name:       "resulting combined tag invalid",
			tag:        "-abc",
			suffix:     "sfx",
			wantErr:    true,
			errContain: "resulting tag is invalid",
		},
		{
			name:   "combined trailing dash is allowed",
			tag:    "abc",
			suffix: "-",
			want:   "abc-",
		},
		{
			name:   "uppercase in suffix allowed",
			tag:    "Tag",
			suffix: "SHA123",
			want:   "Tag-SHA123",
		},
		{
			name:   "combined exactly at max length",
			tag:    strings.Repeat("a", MaxTagLength-2),
			suffix: "-b",
			want:   strings.Repeat("a", MaxTagLength-2) + "-b",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AppendSuffix(tt.tag, tt.suffix)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContain != "" {
					assert.Contains(t, err.Error(), tt.errContain)
				}
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.LessOrEqual(t, len(got), MaxTagLength)
			assert.True(t, IsValidTag(got), "combined tag should be valid")
		})
	}
}
