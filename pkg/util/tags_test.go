package util

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsAlphanumeric(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want bool
	}{
		{"lowercase letter", 'a', true},
		{"uppercase letter", 'Z', true},
		{"digit", '5', true},
		{"dash", '-', false},
		{"underscore", '_', false},
		{"dot", '.', false},
		{"space", ' ', false},
		{"non-ascii", 'ø', false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isAlphanumeric(tt.r)
			assert.Equal(t, tt.want, got)
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
		{"digit", '0', false},
		{"space", ' ', false},
		{"other punctuation", '#', false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSpecialChar(tt.r)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCollapseRepeatedChars(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"no specials", "abc123", "abc123"},
		{"single specials", "a-b.c_d", "a-b.c_d"},
		{"repeated dashes", "a--b---c", "a-b-c"},
		{"repeated dots", "a..b...c", "a.b.c"},
		{"repeated underscores", "a__b___c", "a_b_c"},
		{"mixed repeated specials", "a-._-._b", "a-._-._b"},
		{"consecutive different specials", "a-._._-b", "a-._._-b"},
		{"only specials", "----", "-"},
		{"starts with specials", "--a--", "-a-"},
		{"ends with specials", "a--..__", "a-._"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := collapseRepeatedChars(tt.in)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsValidTag(t *testing.T) {
	tests := []struct {
		name string
		tag  string
		want bool
	}{
		{"empty", "", false},
		{"simple valid", "v1", true},
		{"letters digits", "abc123", true},
		{"with separators", "release-1.0_0", true},
		{"starts with dash", "-abc", false},
		{"starts with dot", ".abc", false},
		{"starts with underscore", "_abc", false},
		{"invalid character space", "a b", false},
		{"invalid character slash", "a/b", false},
		{"invalid character at", "a@b", false},
		{"max length exact", strings.Repeat("a", MaxTagLength), true},
		{"over max length", strings.Repeat("a", MaxTagLength+1), false},
		{"single char valid", "a", true},
		{"single char invalid", "-", false},
		{"mixed case allowed", "AbC-1.2_3", true},
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
		{"empty input", "", "", true},
		{"valid input unchanged", "v1.0.0", "v1.0.0", false},
		{"replace slash with dash", "feature/foo", "feature-foo", false},
		{"replace spaces and special with dash", "feat new@branch", "feat-new-branch", false},
		{"trim leading trailing specials", "..-v1.0.0-__", "v1.0.0", false},
		{"non alnum start gets v prefix", "-beta", "v-beta", false},
		{"starts with digit ok", "1beta", "1beta", false},
		{"only invalid chars becomes error", "///@@@", "", true},
		{"multiple specials collapsed", "a--b..c__d", "a-b.c_d", false},
		{
			"over max length truncated and trimmed",
			strings.Repeat("a", MaxTagLength-1) + "-extra",
			strings.Repeat("a", MaxTagLength-1),
			false,
		},
		{
			"truncation removes trailing special",
			strings.Repeat("a", MaxTagLength-1) + "-",
			strings.Repeat("a", MaxTagLength-1),
			false,
		},
		{
			"leading special after mapping gets v prefix",
			"/abc",
			"abc",
			false,
		},
		{
			"non-ascii becomes dash",
			"feat-ø-branch",
			"feat-branch",
			false,
		},
		{
			"results must match IsValidTag",
			"Invalid Tag!*",
			"",
			false, // will be valid after sanitization
		},
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
			if tt.name == "results must match IsValidTag" {
				assert.True(t, IsValidTag(got))
			}
			if tt.want != "" {
				assert.Equal(t, tt.want, got)
			} else if !tt.wantErr {
				assert.NotEmpty(t, got)
				assert.True(t, IsValidTag(got))
			}
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
		{"empty ref", "", "", true},
		{"branch main", "refs/heads/main", "main", false},
		{"branch feature slash", "refs/heads/feature/foo", "feature-foo", false},
		{"tag ref", "refs/tags/v1.0.0", "v1.0.0", false},
		{"pull request head", "refs/pull/123/head", "pr-123", false},
		{"pull request minimal", "refs/pull/456", "pr-456", false},
		{"invalid pull ref format", "refs/pull/", "", true},
		{"unknown format as-is", "custom/ref", "custom-ref", false},
		{"unknown already valid", "valid-tag", "valid-tag", false},
		{"long branch name gets truncated", "refs/heads/" + strings.Repeat("a", MaxTagLength+10), strings.Repeat("a", MaxTagLength), false},
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
			assert.True(t, IsValidTag(got))
			if tt.want != "" {
				if len(tt.want) == MaxTagLength && len(got) == MaxTagLength && strings.HasPrefix(tt.name, "long branch name") {
					assert.Equal(t, tt.want, got)
				} else {
					assert.Equal(t, tt.want, got)
				}
			}
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
		{"maxLength zero", "tag", 0, "", true},
		{"maxLength negative", "tag", -1, "", true},
		{"invalid tag when shorter", "-bad", 10, "", true},
		{"valid tag no truncation", "valid-tag", 20, "valid-tag", false},
		{"truncate simple", "abcdefghij", 5, "abcde", false},
		{"truncate removes trailing special", "abc-.-", 5, "abc", false},
		{"truncate to MaxTagLength when larger", strings.Repeat("a", MaxTagLength+10), MaxTagLength + 5, strings.Repeat("a", MaxTagLength), false},
		{"maxLength greater than MaxTagLength capped", strings.Repeat("a", MaxTagLength+10), MaxTagLength + 50, strings.Repeat("a", MaxTagLength), false},
		{"unable to maintain validity", "-.-.-.-", 3, "", true},
		{"already exactly maxLength", strings.Repeat("a", 10), 10, strings.Repeat("a", 10), false},
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
			assert.LessOrEqual(t, len(got), MinInt(MaxTagLength, tt.maxLength))
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
		{"empty tag", "", "", true},
		{"already normalized", "v1.0.0", "v1.0.0", false},
		{"upper to lower", "V1.0.0", "v1.0.0", false},
		{"spaces and specials", " Feature/Branch@@ ", "feature-branch", false},
		{"non alnum start", "-BETA", "v-beta", false},
		{"long tag truncated", strings.ToUpper(strings.Repeat("a", MaxTagLength+5)), strings.Repeat("a", MaxTagLength), false},
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
			assert.True(t, IsValidTag(got))
			if tt.want != "" {
				assert.Equal(t, tt.want, got)
			}
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
		{"empty tag", "", "suffix", "", true},
		{"empty suffix returns tag", "base", "", "base", false},
		{"suffix with dash preserved", "base", "-sfx", "base-sfx", false},
		{"suffix without separator gets dash", "base", "sfx", "base-sfx", false},
		{"suffix with dot preserved", "base", ".sfx", "base.sfx", false},
		{"suffix with underscore preserved", "base", "_sfx", "base_sfx", false},
		{"combined within max length", "tag", "-123", "tag-123", false},
		{
			"needs truncation of base",
			strings.Repeat("a", MaxTagLength-2),
			"-x",
			strings.Repeat("a", MaxTagLength-2) + "-x",
			false,
		},
		{
			"base truncated to fit suffix",
			strings.Repeat("a", MaxTagLength),
			"-zz",
			strings.Repeat("a", MaxTagLength-3) + "-zz",
			false,
		},
		{
			"suffix too long for any base",
			"base",
			strings.Repeat("b", MaxTagLength),
			"",
			true,
		},
		{
			"invalid resulting tag",
			"-bad",
			"-suffix",
			"",
			true,
		},
		{
			"truncation keeps validity",
			strings.Repeat("a", MaxTagLength-1) + "-",
			"sfx",
			strings.Repeat("a", MaxTagLength-4) + "-sfx",
			false,
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
			assert.True(t, IsValidTag(got))
			assert.LessOrEqual(t, len(got), MaxTagLength)
			assert.Equal(t, tt.want, got)
		})
	}
}

// MinInt is a small helper for assertions, defined here to avoid importing math.
func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
