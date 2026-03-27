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
		{"simple", "v1", true},
		{"numeric start", "1.2.3", true},
		{"underscore allowed after first", "a_b", true},
		{"underscore first invalid", "_abc", false},
		{"dot and hyphen allowed", "a.b-c", true},
		{"starts with hyphen invalid", "-abc", false},
		{"contains slash invalid", "feat/foo", false},
		{"max length ok", strings.Repeat("a", MaxTagLength), true},
		{"too long", strings.Repeat("a", MaxTagLength+1), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidTag(tt.tag)
			assert.Equal(t, tt.want, got)
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
		{"basic", "main", "main", false},
		{"spaces to dash", "abc def", "abc-def", false},
		{"slashes to dash", "feature/foo/bar", "feature-foo-bar", false},
		{"trim leading trailing specials", "-._abc-._", "abc", false},
		{"collapse specials", "a---..__b", "a-b", false},
		{"long truncated", strings.Repeat("a", MaxTagLength+5), strings.Repeat("a", MaxTagLength), false},
		{"invalid only specials", "!!!", "", true},
		{"empty error", "", "", true},
		{"underscore only", "_", "", true},
		{"leading dot numeric", ".1", "1", false},
		{"invalid leading char becomes dash then trimmed", "@abc", "abc", false},
		{"trailing specials trimmed", "v---", "v", false},
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
		{"feature branch with slash", "refs/heads/feature/foo", "feature-foo", false},
		{"tag ref", "refs/tags/v1.0.0", "v1.0.0", false},
		{"pull request head", "refs/pull/123/head", "pr-123", false},
		{"pull request no suffix", "refs/pull/123", "pr-123", false},
		{"pull request missing number part becomes pr", "refs/pull//head", "pr", false},
		{"unknown ref uses as-is sanitized", "weird/ref", "weird-ref", false},
		{"non pr ref sanitized", "refs/pull", "refs-pull", false},
		{"empty ref error", "", "", true},
		{"unsanitizable ref error", "!!!", "", true},
		{"simple other ref", "foo", "foo", false},
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
		{"max zero error", "abc", 0, "", true},
		{"valid no truncation", "abc", 5, "abc", false},
		{"invalid no truncation error", "-abc", 10, "", true},
		{"simple truncation", "abcdef", 3, "abc", false},
		{"truncation removes trailing specials", "abc--", 4, "abc", false},
		{"truncation removes invalid leading", "-abcdef", 3, "ab", false},
		{"clamp to MaxTagLength large max", strings.Repeat("a", 10), MaxTagLength + 1000, strings.Repeat("a", 10), false},
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
		{"empty error", "", "", true},
		{"lowercase simple", "ABC", "abc", false},
		{"spaces to dash via sanitize", "Feat Branch", "feat-branch", false},
		{"already valid", "v1.0.0", "v1.0.0", false},
		{"unsanitizable becomes error", "!!!", "", true},
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
	tests := []struct {
		name       string
		tag        string
		suffix     string
		wantPrefix string
		want       string
		wantErr    bool
	}{
		{"empty tag error", "", "rc1", "", "", true},
		{"empty suffix returns tag unvalidated", "-abc", "", "", "-abc", false},
		{"adds hyphen when missing", "v1", "rc1", "v1-rc1", "v1-rc1", false},
		{"keeps provided separator dot", "v1", ".nightly", "v1.nightly", "v1.nightly", false},
		{"valid result", "release", "-canary", "release-canary", "release-canary", false},
		{"invalid base tag yields invalid combined error", "-abc", "-x", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AppendSuffix(tt.tag, tt.suffix)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if tt.suffix != "" && tt.tag != "" {
				assert.Equal(t, tt.want, got)
				assert.True(t, IsValidTag(got))
				assert.LessOrEqual(t, len(got), MaxTagLength)
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}

	t.Run("truncates when exceeding max length", func(t *testing.T) {
		base := strings.Repeat("a", 127)
		suffix := "z" // will be prefixed with '-'
		got, err := AppendSuffix(base, suffix)
		assert.NoError(t, err)
		assert.True(t, IsValidTag(got))
		assert.Equal(t, MaxTagLength, len(got))
		assert.True(t, strings.HasSuffix(got, "-"+suffix))
	})

	t.Run("suffix too long error", func(t *testing.T) {
		base := "v1"
		suffix := strings.Repeat("b", MaxTagLength) // will get prefixed with '-'
		_, err := AppendSuffix(base, suffix)
		assert.Error(t, err)
	})

	t.Run("truncate path with invalid base succeeds by trimming", func(t *testing.T) {
		base := "-" + strings.Repeat("a", MaxTagLength-1) // invalid base, length == MaxTagLength
		suffix := "-x"
		got, err := AppendSuffix(base, suffix)
		assert.NoError(t, err)
		assert.True(t, IsValidTag(got))
		assert.LessOrEqual(t, len(got), MaxTagLength)
		assert.True(t, strings.HasSuffix(got, suffix))
	})
}

func TestIsAlphanumeric(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want bool
	}{
		{"lowercase", 'a', true},
		{"uppercase", 'Z', true},
		{"digit", '5', true},
		{"underscore", '_', false},
		{"dash", '-', false},
		{"dot", '.', false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isAlphanumeric(tt.r))
		})
	}
}

func TestCollapseRepeatedChars(t *testing.T) {
	tests := []struct {
		name string
		in   string
		out  string
	}{
		{"double dash", "a--b", "a-b"},
		{"mixed specials collapse", "a-._-b", "a-b"},
		{"all specials collapse to first", "-._._-", "-"},
		{"empty", "", ""},
		{"no specials", "ab", "ab"},
		{"underscore series", "__..--", "_"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.out, collapseRepeatedChars(tt.in))
		})
	}
}

func TestIsSpecialChar(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want bool
	}{
		{"dash", '-', true},
		{"dot", '.', true},
		{"underscore", '_', true},
		{"letter", 'a', false},
		{"digit", '9', false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isSpecialChar(tt.r))
		})
	}
}
