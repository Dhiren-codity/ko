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
		{"valid tag unchanged", "v1.2.3", "v1.2.3", false},
		{"slashes and spaces replaced", "feature/foo bar", "feature-foo-bar", false},
		{"leading and trailing separators trimmed", "__abc_DEF-123__", "abc_DEF-123", false},
		{"collapse mixed special sequences", "a-._-b", "a-b", false},
		{"invalid becomes empty after sanitize", "///", "", true},
		{"leading invalid replaced and trimmed", "!abc", "abc", false},
		{"leading dot trimmed", ".abc", "abc", false},
		{"multiple specials collapsed", "hello----world", "hello-world", false},
		{"enforce max length and trim trailing", strings.Repeat("a", MaxTagLength-1) + "-" + "b" + "c", strings.Repeat("a", MaxTagLength-1), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SanitizeTag(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, got)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.True(t, IsValidTag(got))
			assert.LessOrEqual(t, len(got), MaxTagLength)
		})
	}
}

func TestIsValidTag(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid simple", "v1.2.3", true},
		{"empty", "", false},
		{"starts with dash", "-bad", false},
		{"contains space", "bad tag", false},
		{"trailing dash allowed", "abc-", true},
		{"only alnum single", "A", true},
		{"length 128 ok", strings.Repeat("a", MaxTagLength), true},
		{"length 129 not ok", strings.Repeat("a", MaxTagLength+1), false},
		{"underscore dot dash internal", "a_b.c-d", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsValidTag(tt.input))
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
		{"branch main", "refs/heads/main", "main", false},
		{"branch feature path", "refs/heads/feature/foo", "feature-foo", false},
		{"tag release", "refs/tags/v1.0.0", "v1.0.0", false},
		{"pull request head", "refs/pull/123/head", "pr-123", false},
		{"invalid pull format", "refs/pull", "", true},
		{"unknown format use as-is sanitized", "some/branch", "some-branch", false},
		{"sanitization error bubbles up", "refs/heads/---", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateTagFromRef(tt.ref)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, got)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.True(t, IsValidTag(got))
		})
	}
}

func TestTruncateTag(t *testing.T) {
	tests := []struct {
		name      string
		tag       string
		maxLength int
		want      string
		wantErr   bool
	}{
		{"valid unchanged", "abc", 10, "abc", false},
		{"invalid short fails", "-bad", 10, "", true},
		{"truncated and trimmed trailing", "abc---", 5, "abc", false},
		{"nonpositive max", "abc", 0, "", true},
		{"truncate to MaxTagLength", strings.Repeat("a", MaxTagLength+10), MaxTagLength + 50, strings.Repeat("a", MaxTagLength), false},
		{"truncation yields empty error", "----abcd", 3, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TruncateTag(tt.tag, tt.maxLength)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, got)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.True(t, IsValidTag(got))
			assert.LessOrEqual(t, len(got), MaxTagLength)
			if len(tt.tag) <= tt.maxLength && tt.maxLength <= MaxTagLength {
				assert.Equal(t, tt.tag, got)
			}
		})
	}
}

func TestNormalizeTag(t *testing.T) {
	tests := []struct {
		name    string
		tag     string
		want    string
		wantLen int
		wantErr bool
	}{
		{"lowercase and sanitize", "FeaTURE/ABC-123", "feature-abc-123", 0, false},
		{"empty error", "", "", 0, true},
		{"long truncated to max", "A" + strings.Repeat("a", 200) + "Z", "", MaxTagLength, false},
		{"invalid chars replaced and trimmed", "Hello World!", "hello-world", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeTag(tt.tag)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, got)
				return
			}
			assert.NoError(t, err)
			if tt.want != "" {
				assert.Equal(t, tt.want, got)
			}
			assert.True(t, IsValidTag(got))
			if tt.wantLen > 0 {
				assert.LessOrEqual(t, len(got), tt.wantLen)
			}
			assert.Equal(t, strings.ToLower(got), got)
		})
	}
}

func TestAppendSuffix(t *testing.T) {
	baseLong := strings.Repeat("a", 120)
	tests := []struct {
		name      string
		tag       string
		suffix    string
		want      string
		wantErr   bool
		checkFunc func(got string)
	}{
		{
			"basic with auto dash",
			"base",
			"sha123",
			"base-sha123",
			false,
			nil,
		},
		{
			"suffix empty returns base",
			"base",
			"",
			"base",
			false,
			nil,
		},
		{
			"base empty error",
			"",
			"suffix",
			"",
			true,
			nil,
		},
		{
			"combined too long triggers truncation",
			baseLong,
			"zzzzzzzzzz",
			"",
			false,
			func(got string) {
				assert.LessOrEqual(t, len(got), MaxTagLength)
				assert.True(t, strings.HasSuffix(got, "-zzzzzzzzzz"))
			},
		},
		{
			"suffix too long error",
			"base",
			strings.Repeat("x", MaxTagLength),
			"",
			true,
			nil,
		},
		{
			"invalid suffix makes combined invalid",
			"base",
			"bad space",
			"",
			true,
			nil,
		},
		{
			"invalid base makes combined invalid",
			"-bad",
			"ok",
			"",
			true,
			nil,
		},
		{
			"suffix with dot separator preserved",
			"base",
			".meta",
			"base.meta",
			false,
			nil,
		},
		{
			"suffix with underscore separator preserved",
			"base",
			"_meta",
			"base_meta",
			false,
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AppendSuffix(tt.tag, tt.suffix)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, got)
				return
			}
			assert.NoError(t, err)
			if tt.want != "" {
				assert.Equal(t, tt.want, got)
			}
			assert.True(t, IsValidTag(got))
			if tt.checkFunc != nil {
				tt.checkFunc(got)
			}
		})
	}
}

func TestIsAlphanumeric(t *testing.T) {
	tests := []struct {
		r    rune
		want bool
	}{
		{'a', true},
		{'Z', true},
		{'5', true},
		{'-', false},
		{'.', false},
		{'_', false},
	}
	for _, tt := range tests {
		t.Run(string(tt.r), func(t *testing.T) {
			assert.Equal(t, tt.want, isAlphanumeric(tt.r))
		})
	}
}

func TestIsSpecialChar(t *testing.T) {
	tests := []struct {
		r    rune
		want bool
	}{
		{'-', true},
		{'.', true},
		{'_', true},
		{'a', false},
		{'9', false},
	}
	for _, tt := range tests {
		t.Run(string(tt.r), func(t *testing.T) {
			assert.Equal(t, tt.want, isSpecialChar(tt.r))
		})
	}
}

func TestCollapseRepeatedChars(t *testing.T) {
	tests := []struct {
		name string
		in   string
		out  string
	}{
		{"no specials", "abc", "abc"},
		{"only specials collapse to first", "--..__", "-"},
		{"dash run collapsed", "a---b", "a-b"},
		{"mixed sequence collapsed", "a-._-b", "a-b"},
		{"leading specials kept first", "-._abc", "-abc"},
		{"empty string", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.out, collapseRepeatedChars(tt.in))
		})
	}
}
