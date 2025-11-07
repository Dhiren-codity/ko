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
		{"empty", "", false},
		{"simple valid", "a", true},
		{"valid mixed case and symbols", "AbC.1_2-3", true},
		{"starts with dash", "-abc", false},
		{"starts with dot", ".abc", false},
		{"starts with underscore", "_abc", false},
		{"invalid char", "abc$def", false},
		{"max length exactly", strings.Repeat("a", MaxTagLength), true},
		{"exceeds max length", strings.Repeat("a", MaxTagLength+1), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsValidTag(tt.tag))
		})
	}
}

func TestSanitizeTag(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"simple branch with slash", "feature/foo", "feature-foo", false},
		{"spaces replaced and trimmed", "Hello World", "Hello-World", false},
		{"invalid chars replaced", "a@b#c", "a-b-c", false},
		{"leading and trailing separators trimmed", "-.abc-.", "abc", false},
		{"emoji leading replaced then trimmed", "💥boom", "boom", false},
		{"empty input error", "", "", true},
		{"only slashes become empty then error", "/", "", true},
		{"long string truncated to max", strings.Repeat("a", 200), strings.Repeat("a", MaxTagLength), false},
		{"trailing sep removed after truncation", strings.Repeat("a", 127) + "-zzz", strings.Repeat("a", 127), false},
		{"collapse repeated specials", "a---b..c__d-._e", "a-b.c_d-e", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SanitizeTag(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.True(t, IsValidTag(got))
			assert.LessOrEqual(t, len(got), MaxTagLength)
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
		{"branch with path", "refs/heads/feature/foo", "feature-foo", false},
		{"tag ref", "refs/tags/v1.2.3", "v1.2.3", false},
		{"pull request head", "refs/pull/123/head", "pr-123", false},
		{"pull request no subpath", "refs/pull/123", "pr-123", false},
		{"pull request missing id but trailing slash", "refs/pull/", "pr", false},
		{"invalid pull format error", "refs/pull", "", true},
		{"unknown format uses as-is sanitized", "weird/ref", "weird-ref", false},
		{"unknown with invalid chars", "weird@@ref", "weird-ref", false},
		{"empty ref error", "", "", true},
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
		{"error on non-positive max", "abc", 0, "", true},
		{"invalid tag when shorter", "-abc", 10, "", true},
		{"valid short tag unchanged", "abc", 10, "abc", false},
		{"truncates and trims trailing specials", "abc---", 5, "abc", false},
		{"truncate causing empty invalid", "--abc", 2, "", true},
		{"cap maxLength to MaxTagLength", strings.Repeat("a", 200), 1000, strings.Repeat("a", MaxTagLength), false},
		{"valid exact length", strings.Repeat("A", 5), 5, strings.Repeat("A", 5), false},
		{"truncate but still valid trailing char", "abcdEF", 4, "abcd", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TruncateTag(tt.tag, tt.maxLength)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.True(t, IsValidTag(got))
			assert.LessOrEqual(t, len(got), MaxTagLength)
			assert.LessOrEqual(t, len(got), tt.maxLength)
		})
	}
}

func TestNormalizeTag(t *testing.T) {
	tests := []struct {
		name    string
		tag     string
		want    string
		wantErr bool
	}{
		{"lowercase and replace spaces", "HEllo World", "hello-world", false},
		{"lowercase and replace slash", "Feature/Foo", "feature-foo", false},
		{"valid stays valid with case lowered", "FOO_bar", "foo_bar", false},
		{"empty error", "", "", true},
		{"only slashes error via sanitize", "///", "", true},
		{"version tag", "V1.2.3", "v1.2.3", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeTag(tt.tag)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.True(t, IsValidTag(got))
		})
	}
}

func TestAppendSuffix(t *testing.T) {
	base128 := strings.Repeat("a", MaxTagLength)
	base127 := strings.Repeat("a", MaxTagLength-1)
	tests := []struct {
		name      string
		tag       string
		suffix    string
		want      string
		wantErr   bool
	}{
		{"error on empty tag", "", "suf", "", true},
		{"no suffix returns tag", "base", "", "base", false},
		{"adds hyphen separator", "base", "suf", "base-suf", false},
		{"keeps existing dash", "base", "-suf", "base-suf", false},
		{"dot separator allowed", "base", ".meta", "base.meta", false},
		{"invalid base causes invalid combined", "in valid", "ok", "", true},
		{"invalid suffix causes invalid combined", "valid", "#$", "", true},
		{"truncate base to fit suffix", base128, "x", base127 + "-x", false},
		{"suffix too long error", "base", strings.Repeat("a", MaxTagLength), "", true},
		{"combined within limit no truncation", "AbC", "_XYZ", "AbC_XYZ", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AppendSuffix(tt.tag, tt.suffix)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.True(t, IsValidTag(got))
			assert.LessOrEqual(t, len(got), MaxTagLength)
		})
	}
}

func TestCollapseRepeatedChars(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"mixed specials collapse", "a---b..c__d-._e", "a-b.c_d-e"},
		{"all specials collapse to first", "...___---", "."},
		{"single special unchanged", "-", "-"},
		{"empty", "", ""},
		{"no specials", "abc123", "abc123"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, collapseRepeatedChars(tt.in))
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
		{'/', false},
		{' ', false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, isSpecialChar(tt.r))
	}
}

func TestIsAlphanumeric(t *testing.T) {
	tests := []struct {
		r    rune
		want bool
	}{
		{'a', true},
		{'Z', true},
		{'0', true},
		{'-', false},
		{'.', false},
		{'é', false}, // non-ASCII letter not matched by simple range checks
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, isAlphanumeric(tt.r))
	}
}
