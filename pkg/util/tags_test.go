package util

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeTag(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{"simple valid", "v1.2.3", "v1.2.3", false},
		{"slashes and spaces replaced", "Feature/ABC 123", "Feature-ABC-123", false},
		{"leading and trailing specials trimmed", "-foo._", "foo", false},
		{"collapse repeated specials", "a-._-b", "a-b", false},
		{"starts with dot", ".abc", "abc", false},
		{"only invalid chars becomes error", "///", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SanitizeTag(tt.in)
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

func TestSanitizeTag_LengthAndTrim(t *testing.T) {
	// Exceeding length should be truncated and trailing specials trimmed after truncation
	long := strings.Repeat("a", MaxTagLength) + "--" // 128 'a's + "--"
	got, err := SanitizeTag(long)
	assert.NoError(t, err)
	assert.Equal(t, MaxTagLength, len(got))
	assert.True(t, IsValidTag(got))
	assert.Equal(t, strings.Repeat("a", MaxTagLength), got)
}

func TestIsValidTag(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"valid simple", "v1.2.3", true},
		{"valid uppercase underscore", "RC_1", true},
		{"valid trailing dash", "abc-", true},
		{"invalid empty", "", false},
		{"invalid starts with dash", "-bad", false},
		{"invalid starts with dot", ".bad", false},
		{"invalid starts with underscore", "_bad", false},
		{"invalid illegal char", "bad*tag", false},
		{"invalid too long", strings.Repeat("a", MaxTagLength+1), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsValidTag(tt.in))
		})
	}
}

func TestNormalizeTag(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{"lowercases and sanitizes", "Feature/ABC 123", "feature-abc-123", false},
		{"valid unchanged except case", "FOO_bar.baz", "foo_bar.baz", false},
		{"leading hyphen trimmed", "-Foo", "foo", false},
		{"empty error", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeTag(tt.in)
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
	tests := []struct {
		name    string
		tag     string
		suffix  string
		want    string
		wantErr bool
	}{
		{"adds dash when missing", "release", "2021.01", "release-2021.01", false},
		{"suffix already with dash", "release", "-rc1", "release-rc1", false},
		{"suffix starts with dot", "release", ".build", "release.build", false},
		{"empty suffix returns tag", "Tag", "", "Tag", false},
		{"invalid combined due to suffix chars", "abc", "*bad", "", true},
		{"empty tag error", "", "suffix", "", true},
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
		})
	}
}

func TestAppendSuffix_LengthAndErrors(t *testing.T) {
	// Combined length exceeds max -> base truncated
	base := strings.Repeat("a", MaxTagLength)
	suffix := "-b"
	got, err := AppendSuffix(base, suffix)
	assert.NoError(t, err)
	assert.Equal(t, MaxTagLength, len(got))
	assert.True(t, strings.HasSuffix(got, suffix))
	assert.True(t, IsValidTag(got))

	// Suffix too long error
	longSuffix := "-" + strings.Repeat("x", MaxTagLength)
	_, err = AppendSuffix("a", longSuffix)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "suffix too long")
}

func Test_collapseRepeatedChars(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"no repetition", "abc", "abc"},
		{"repeated dashes", "a---b", "a-b"},
		{"mixed specials collapse to one", "a-._-b", "a-b"},
		{"only specials", "___...---", "_"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := collapseRepeatedChars(tt.in)
			assert.Equal(t, tt.want, got)
		})
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
		{'0', false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, isSpecialChar(tt.r))
	}
}

func Test_isAlphanumeric(t *testing.T) {
	tests := []struct {
		r    rune
		want bool
	}{
		{'a', true},
		{'Z', true},
		{'0', true},
		{'9', true},
		{'-', false},
		{'.', false},
		{'_', false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, isAlphanumeric(tt.r))
	}
}
