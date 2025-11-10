package util

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeTag(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "empty input",
			input:   "",
			wantErr: true,
		},
		{
			name:  "slash replacement",
			input: "feature/foo",
			want:  "feature-foo",
		},
		{
			name:  "spaces and punctuation to dash",
			input: "Feature Foo!",
			want:  "Feature-Foo",
		},
		{
			name:  "trim leading specials",
			input: "._-Leading",
			want:  "Leading",
		},
		{
			name:  "collapse mixed consecutive specials",
			input: "a---b..__c-._-d",
			want:  "a-b.c-d",
		},
		{
			name:  "trim trailing special",
			input: "foo.",
			want:  "foo",
		},
		{
			name:    "only specials becomes invalid",
			input:   "___",
			wantErr: true,
		},
		{
			name:  "unicode characters replaced with dash",
			input: "naïve/branch",
			want:  "na-ve-branch",
		},
		{
			name:  "long input truncated to MaxTagLength",
			input: strings.Repeat("a", MaxTagLength+50),
			want:  strings.Repeat("a", MaxTagLength),
		},
		{
			name:  "uppercase preserved",
			input: "Hello-World",
			want:  "Hello-World",
		},
		{
			name:  "starts with digit allowed",
			input: "123abc",
			want:  "123abc",
		},
		{
			name:    "all invalid after replacement",
			input:   "///",
			wantErr: true,
		},
		{
			name:  "multiple slashes collapsed to single dash",
			input: "a///b",
			want:  "a-b",
		},
		{
			name:  "truncation removes trailing separator",
			input: strings.Repeat("a", MaxTagLength-1) + "-zzz",
			// After truncation the trailing '-' is trimmed, resulting in 127 'a's.
			want: strings.Repeat("a", MaxTagLength-1),
		},
		{
			name:  "mix of allowed special characters retained",
			input: "A_B.C-D",
			want:  "A_B.C-D",
		},
		{
			name:  "specials around get trimmed and collapsed",
			input: "__a..b--c__",
			want:  "a.b-c",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SanitizeTag(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err, "unexpected error: %v", err)
			assert.Equal(t, tt.want, got)
			assert.True(t, IsValidTag(got), "sanitized tag should be valid")
			assert.LessOrEqual(t, len(got), MaxTagLength)
		})
	}
}

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

func TestGenerateTagFromRef(t *testing.T) {
	tests := []struct {
		name    string
		ref     string
		want    string
		wantErr bool
	}{
		{
			name:    "empty ref error",
			ref:     "",
			wantErr: true,
		},
		{
			name: "branch main",
			ref:  "refs/heads/main",
			want: "main",
		},
		{
			name: "branch with slashes collapsed to dashes",
			ref:  "refs/heads/feature/foo",
			want: "feature-foo",
		},
		{
			name: "tag ref preserved",
			ref:  "refs/tags/v1.0.0",
			want: "v1.0.0",
		},
		{
			name: "pull request head",
			ref:  "refs/pull/123/head",
			want: "pr-123",
		},
		{
			name: "pull request no head",
			ref:  "refs/pull/123",
			want: "pr-123",
		},
		{
			name: "pull request empty number becomes pr after sanitize",
			ref:  "refs/pull/",
			want: "pr",
		},
		{
			name: "invalid pull ref format treated as unknown and sanitized",
			ref:  "refs/pull",
			want: "refs-pull",
		},
		{
			name: "unknown ref format sanitized",
			ref:  "unknown/format/ref",
			want: "unknown-format-ref",
		},
		{
			name: "long branch name truncated",
			ref:  "refs/heads/" + strings.Repeat("a", MaxTagLength+50),
			want: strings.Repeat("a", MaxTagLength),
		},
		{
			name:    "non-sanitizable unknown ref",
			ref:     "///",
			wantErr: true,
		},
		{
			name: "pull request non-numeric id",
			ref:  "refs/pull/abc/merge",
			want: "pr-abc",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateTagFromRef(tt.ref)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.True(t, IsValidTag(got), "generated tag should be valid")
			assert.LessOrEqual(t, len(got), MaxTagLength)
		})
	}
}

func TestTruncateTag(t *testing.T) {
	tests := []struct {
		name       string
		tag        string
		max        int
		want       string
		wantErr    bool
		errContain string
	}{
		{
			name:       "invalid max length",
			tag:        "abc",
			max:        0,
			wantErr:    true,
			errContain: "maxLength must be positive",
		},
		{
			name:       "short but invalid tag error",
			tag:        "-abc",
			max:        10,
			wantErr:    true,
			errContain: "tag is invalid",
		},
		{
			name: "no truncation returns same",
			tag:  "Valid_Tag-1.2",
			max:  MaxTagLength,
			want: "Valid_Tag-1.2",
		},
		{
			name: "no truncation when at max length, trailing separator allowed",
			tag:  strings.Repeat("a", MaxTagLength-1) + "-",
			max:  MaxTagLength,
			want: strings.Repeat("a", MaxTagLength-1) + "-",
		},
		{
			name: "truncate when provided max > MaxTagLength uses MaxTagLength",
			tag:  strings.Repeat("a", MaxTagLength+20),
			max:  MaxTagLength + 1000,
			want: strings.Repeat("a", MaxTagLength),
		},
		{
			name:       "truncate results in invalid empty tag error",
			tag:        strings.Repeat("-", MaxTagLength+20),
			max:        50,
			wantErr:    true,
			errContain: "unable to truncate tag while maintaining validity",
		},
		{
			name: "truncate basic",
			tag:  "abcdef",
			max:  4,
			want: "abcd",
		},
		{
			name: "no truncation or trimming when within max and valid",
			tag:  "abcd.",
			max:  5,
			want: "abcd.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TruncateTag(tt.tag, tt.max)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContain != "" {
					assert.Contains(t, err.Error(), tt.errContain)
				}
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.True(t, IsValidTag(got), "truncated tag should be valid")
			assert.LessOrEqual(t, len(got), MaxTagLength)
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
