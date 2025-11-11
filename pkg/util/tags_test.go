package util

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func repeat(r rune, n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteRune(r)
	}
	return b.String()
}

func TestIsValidTag(t *testing.T) {
	tests := []struct {
		name string
		tag  string
		want bool
	}{
		{"simple", "v1.0.0", true},
		{"starts with digit", "1abc", true},
		{"underscore not allowed as first", "_foo", false},
		{"contains invalid char", "a@b", false},
		{"empty", "", false},
		{"longer than max", repeat('a', MaxTagLength+1), false},
		{"allowed mix", "A_b.C-1", true},
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
		{"valid unchanged", "v1.0.0", "v1.0.0", false},
		{"slashes replaced", "feature/foo", "feature-foo", false},
		{"spaces to dashes keep case", "Hello World", "Hello-World", false},
		{"leading slash trimmed", "/foo", "foo", false},
		{"only invalid becomes empty", "///", "", true},
		{"starts with underscore trimmed", "_foo", "foo", false},
		{"collapse specials", "a---b__c..d", "a-b_c.d", false},
		{"trim trailing specials", "abc---", "abc", false},
		{"uppercase preserved", "ABC_def", "ABC_def", false},
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
			assert.True(t, IsValidTag(got), "sanitized tag should be valid")
			assert.True(t, len(got) <= MaxTagLength)
		})
	}

	t.Run("max length enforced", func(t *testing.T) {
		input := repeat('a', 200)
		got, err := SanitizeTag(input)
		assert.NoError(t, err)
		assert.True(t, len(got) <= MaxTagLength)
		assert.True(t, IsValidTag(got))
	})

	t.Run("truncation removes trailing special", func(t *testing.T) {
		input := repeat('a', 150) + "-"
		got, err := SanitizeTag(input)
		assert.NoError(t, err)
		assert.True(t, len(got) <= MaxTagLength)
		assert.NotSuffix(t, got, "-")
	})

	t.Run("empty input error", func(t *testing.T) {
		_, err := SanitizeTag("")
		assert.Error(t, err)
	})
}

func TestGenerateTagFromRef(t *testing.T) {
	tests := []struct {
		name    string
		ref     string
		want    string
		wantErr bool
	}{
		{"branch with slash", "refs/heads/feature/foo", "feature-foo", false},
		{"tag ref", "refs/tags/v1.0.0", "v1.0.0", false},
		{"pull request head", "refs/pull/123/head", "pr-123", false},
		{"unknown format used as-is, sanitized", "weird/ref", "weird-ref", false},
		{"empty ref", "", "", true},
		{"invalid pull ref format", "refs/pull", "", true},
		{"invalid ref becomes unsanitizable", "///", "", true},
		{"custom refs prefix stays", "refs/custom/branch", "refs-custom-branch", false},
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
		{"short unchanged", "abcdef", 10, "abcdef", false},
		{"equal unchanged", "abcdef", 6, "abcdef", false},
		{"truncate simple", "abcdef", 3, "abc", false},
		{"avoid trailing sep after trim", "abc-", 3, "abc", false},
		{"invalid tag short returns error", "ab@cd", 10, "", true},
		{"maxLength <= 0 error", "abc", 0, "", true},
		{"trim both ends after truncation", "-abc-", 3, "abc", false},
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
			assert.True(t, len(got) <= tt.maxLength || tt.maxLength > MaxTagLength)
			assert.True(t, IsValidTag(got))
		})
	}

	t.Run("truncate too long invalid to empty -> error", func(t *testing.T) {
		// A long string of only dashes will truncate and trim to empty, causing error.
		_, err := TruncateTag(repeat('-', 200), 50)
		assert.Error(t, err)
	})

	t.Run("maxLength greater than MaxTagLength is capped", func(t *testing.T) {
		tag := repeat('a', MaxTagLength+10)
		got, err := TruncateTag(tag, MaxTagLength+10)
		assert.NoError(t, err)
		assert.True(t, len(got) <= MaxTagLength)
		assert.True(t, IsValidTag(got))
	})
}

func TestNormalizeTag(t *testing.T) {
	tests := []struct {
		name    string
		tag     string
		want    string
		wantErr bool
	}{
		{"lowercase and sanitize", "Hello World/FOO", "hello-world-foo", false},
		{"trim trailing specials", "A..B__C---", "a.b_c", false},
		{"empty error", "", "", true},
		{"unsanitizable error", "////", "", true},
		{"slashes replaced", "Feature/Bar", "feature-bar", false},
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
			assert.Equal(t, strings.ToLower(got), got)
		})
	}
}

func TestAppendSuffix(t *testing.T) {
	tests := []struct {
		name     string
		tag      string
		suffix   string
		expected string
		wantErr  bool
	}{
		{"basic hyphen added", "release", "123", "release-123", false},
		{"suffix with dot", "release", ".sha", "release.sha", false},
		{"empty suffix returns tag even if invalid", "bad tag", "", "bad tag", false},
		{"empty tag error", "", "x", "", true},
		{"invalid suffix makes combined invalid", "valid", "bad@", "", true},
		{"long suffix too long error", "base", repeat('x', MaxTagLength+100), "", true},
		{"already separated suffix", "base", "-s", "base-s", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AppendSuffix(tt.tag, tt.suffix)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, got)
			// only check validity when suffix not empty (since AppendSuffix doesn't sanitize when suffix is empty)
			if tt.suffix != "" {
				assert.True(t, IsValidTag(got))
				assert.True(t, len(got) <= MaxTagLength)
			}
		})
	}

	t.Run("combined exceeds max, base truncated appropriately", func(t *testing.T) {
		base := repeat('a', MaxTagLength-1) // 127
		got, err := AppendSuffix(base, "b")
		assert.NoError(t, err)
		assert.True(t, len(got) <= MaxTagLength)
		assert.True(t, IsValidTag(got))
		// suffix should be "-b" (since no leading sep), base truncated to MaxTagLength - len("-b") = 126
		assert.Equal(t, repeat('a', MaxTagLength-2)+"-b", got)
	})

	t.Run("base truncation trims trailing special", func(t *testing.T) {
		// Construct base so that truncation point is a trailing hyphen
		base := repeat('a', 125) + "-" + "zzzz"
		got, err := AppendSuffix(base, "x")
		assert.NoError(t, err)
		assert.True(t, len(got) <= MaxTagLength)
		assert.True(t, IsValidTag(got))
		assert.Equal(t, repeat('a', 125)+"-x", got)
	})

	t.Run("suffix already with separator, truncation respects it", func(t *testing.T) {
		base := repeat('a', 200)
		got, err := AppendSuffix(base, "-xyz")
		assert.NoError(t, err)
		assert.True(t, len(got) <= MaxTagLength)
		assert.True(t, IsValidTag(got))
		assert.Suffix(t, got, "-xyz")
	})
}

func TestCollapseRepeatedChars(t *testing.T) {
	tests := []struct {
		name string
		in   string
		out  string
	}{
		{"empty", "", ""},
		{"single", "a", "a"},
		{"all specials collapse to first", "-._-__..", "-"},
		{"mix with letters", "-.-__--..a--..__b", "-a-b"},
		{"no collapse needed", "abc.def_ghi-jkl", "abc.def_ghi-jkl"},
		{"leading specials collapse", "----abc", "-abc"},
		{"trailing specials collapse", "abc____", "abc_"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.out, collapseRepeatedChars(tt.in))
		})
	}
}

func TestIsSpecialCharAndIsAlphanumeric(t *testing.T) {
	testsSpecial := []struct {
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
	for _, tt := range testsSpecial {
		assert.Equal(t, tt.want, isSpecialChar(tt.r), "isSpecialChar(%q)", tt.r)
	}

	testsAlnum := []struct {
		r    rune
		want bool
	}{
		{'a', true},
		{'Z', true},
		{'0', true},
		{'-', false},
		{'.', false},
		{'_', false},
	}
	for _, tt := range testsAlnum {
		assert.Equal(t, tt.want, isAlphanumeric(tt.r), "isAlphanumeric(%q)", tt.r)
	}
}
