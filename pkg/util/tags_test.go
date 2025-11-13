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
		{"simple", "main", true},
		{"mixed allowed chars", "abc-DEF_123.ghi", true},
		{"trailing hyphen allowed", "abc-", true},
		{"empty", "", false},
		{"starts with hyphen", "-abc", false},
		{"starts with dot", ".abc", false},
		{"starts with underscore", "_abc", false},
		{"contains invalid char", "ab$cd", false},
		{"too long", strings.Repeat("a", MaxTagLength+1), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidTag(tt.tag)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSanitizeTag_BasicAndEdge(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"already valid", "main", "main", false},
		{"replace slashes", "feature/foo", "feature-foo", false},
		{"double slashes collapse", "feature//foo", "feature-foo", false},
		{"spaces and punctuation", "Hello World!", "Hello-World", false},
		{"leading and trailing specials trimmed", "-._foo-._", "foo", false},
		{"internal specials collapsed", "a---..__b", "a-b", false},
		{"unicode replaced with dash", "feätürë", "fe-t-r", false},
		{"becomes empty after sanitization", "///", "", true},
		{"empty input", "", "", true},
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
		})
	}
}

func TestSanitizeTag_LongInputTruncation(t *testing.T) {
	long := strings.Repeat("a", 200)
	got, err := SanitizeTag(long)
	assert.NoError(t, err)
	assert.True(t, len(got) <= MaxTagLength)
	assert.Equal(t, strings.Repeat("a", MaxTagLength), got)
	assert.True(t, IsValidTag(got))
}

func TestGenerateTagFromRef(t *testing.T) {
	tests := []struct {
		name    string
		ref     string
		want    string
		wantErr bool
	}{
		{"branch main", "refs/heads/main", "main", false},
		{"branch with slash", "refs/heads/feature/foo", "feature-foo", false},
		{"tag version", "refs/tags/v1.0.0", "v1.0.0", false},
		{"pull request", "refs/pull/123/head", "pr-123", false},
		{"unknown ref uses as-is", "weird/ref", "weird-ref", false},
		{"empty", "", "", true},
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
		})
	}
}

func TestGenerateTagFromRef_ErrorCases(t *testing.T) {
	_, err := GenerateTagFromRef("refs/pull")
	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "invalid pull request ref format")
	}

	// Un-sanitizable (becomes empty) default case
	tag, err := GenerateTagFromRef("///")
	if err != nil {
		assert.Contains(t, err.Error(), "unable to sanitize tag")
	} else {
		assert.NotEmpty(t, tag)
		assert.True(t, IsValidTag(tag))
	}
}

func TestTruncateTag_Basic(t *testing.T) {
	tests := []struct {
		name      string
		tag       string
		maxLength int
		want      string
		wantErr   bool
	}{
		{"shorten", "abcdef", 3, "abc", false},
		{"no shorten valid", "abc", 10, "abc", false},
		{"trim trailing special after truncation", "abc-", 3, "abc", false},
		{"max greater than hard limit clamps", strings.Repeat("a", 130), 1000, strings.Repeat("a", MaxTagLength), false},
		{"invalid tag when shorter or equal max", "-abc", 10, "", true},
		{"unable to truncate to valid (all specials)", "-----", 3, "", true},
		{"non-positive max", "abc", 0, "", true},
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
		{"lowercase and sanitize", "FEATURE/New", "feature-new", false},
		{"empty", "", "", true},
		{"unsanitizable", "///", "", true},
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

func TestAppendSuffix_Basic(t *testing.T) {
	tests := []struct {
		name    string
		tag     string
		suffix  string
		want    string
		wantErr bool
	}{
		{"adds dash when needed", "feature", "abc", "feature-abc", false},
		{"uses provided separator", "feature", ".rc1", "feature.rc1", false},
		{"empty suffix returns tag as-is", "feature", "", "feature", false},
		{"empty tag error", "", "x", "", true},
		{"invalid combined due to bad suffix", "good", "bad@chars", "", true},
		{"invalid base allowed when suffix empty", "-abc", "", "-abc", false},
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
		})
	}
}

func TestAppendSuffix_TruncationAndEdge(t *testing.T) {
	// Truncation when combined exceeds limit
	base := strings.Repeat("a", 127)
	got, err := AppendSuffix(base, "x")
	assert.NoError(t, err)
	assert.Equal(t, strings.Repeat("a", 126)+"-x", got)
	assert.True(t, len(got) <= MaxTagLength)

	// Suffix too long
	_, err = AppendSuffix("base", strings.Repeat("x", MaxTagLength+10))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "suffix too long")

	// Truncation path where truncation fails due to invalid base
	_, err = AppendSuffix(strings.Repeat("_", 200), "x")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to truncate tag for suffix")
}

func Test_isAlphanumeric(t *testing.T) {
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
		assert.Equal(t, tt.want, isAlphanumeric(tt.r), string(tt.r))
	}
}

func Test_isSpecialChar(t *testing.T) {
	tests := []struct {
		r    rune
		want bool
	}{
		{'-', true},
		{'.', true},
		{'_', true},
		{'a', false},
		{'Z', false},
		{'0', false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, isSpecialChar(tt.r), string(tt.r))
	}
}

func Test_collapseRepeatedChars(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"hyphens collapse", "a---b", "a-b"},
		{"mixed specials collapse to first", "a-._b", "a-b"},
		{"leading specials preserved (not trimmed here)", "___abc", "_abc"},
		{"no specials", "abc", "abc"},
		{"alternating specials and letters", "a-b-c", "a-b-c"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := collapseRepeatedChars(tt.in)
			assert.Equal(t, tt.want, got)
		})
	}
}
