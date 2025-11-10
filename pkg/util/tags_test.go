package util

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidTag_ValidCases(t *testing.T) {
	tests := []struct {
		name string
		tag  string
	}{
		{"simple", "v1.0.0"},
		{"with_underscore", "foo_bar"},
		{"with_dash_and_dot", "A-B.C"},
		{"single_char", "a"},
		{"max_length", strings.Repeat("a", MaxTagLength)},
		{"ends_with_separator_allowed", "A-._"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.True(t, IsValidTag(tt.tag))
		})
	}
}

func TestIsValidTag_InvalidCases(t *testing.T) {
	tests := []struct {
		name string
		tag  string
	}{
		{"empty", ""},
		{"starts_with_dot", ".tag"},
		{"starts_with_dash", "-tag"},
		{"starts_with_underscore", "_tag"},
		{"contains_space", "has space"},
		{"contains_invalid_char", "bad!tag"},
		{"too_long", strings.Repeat("a", MaxTagLength+1)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.False(t, IsValidTag(tt.tag))
		})
	}
}

func TestSanitizeTag_Success(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"branch_with_slash", "feature/foo", "feature-foo"},
		{"spaces_and_punct", "Hello World!", "Hello-World"},
		{"trim_leading_trailing_separators", "..--abc-.-", "abc"},
		{"starts_with_invalid_colon", ":foo", "foo"},
		{"numeric_ok", "123", "123"},
		{"collapse_separators", "a---...___b", "a-b"},
		{"unicode_to_dash", "tag🚀name", "tag-name"},
		{"only_valid_chars", "A_b-c.d", "A_b-c.d"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SanitizeTag(tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.True(t, IsValidTag(got))
		})
	}
}

func TestSanitizeTag_Errors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"only_separators", "---___..."},
		{"only_spaces", "     "},
		{"only_invalid_chars", "!!!"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SanitizeTag(tt.input)
			assert.Error(t, err)
			assert.Empty(t, got)
		})
	}
}

func TestSanitizeTag_LengthTruncationAndTrim(t *testing.T) {
	// Long input should be truncated to MaxTagLength and remain valid
	long := strings.Repeat("a", 200)
	got, err := SanitizeTag(long)
	assert.NoError(t, err)
	assert.Equal(t, MaxTagLength, len(got))
	assert.True(t, IsValidTag(got))

	// Ensure trailing separators are trimmed after truncation
	input := strings.Repeat("a", MaxTagLength-1) + "-" + strings.Repeat("b", 100) // guarantee truncation at '-'
	got2, err := SanitizeTag(input)
	assert.NoError(t, err)
	assert.False(t, strings.HasSuffix(got2, "-"))
	assert.True(t, IsValidTag(got2))
}

func TestCollapseRepeatedChars(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"no_collapse", "abc", "abc"},
		{"collapse_mixed", "a-._-._b", "a-b"},         // after first '-', following specials are dropped until 'b'
		{"collapse_run", "a---...___b", "a-b"},        // entire run collapsed to first '-'
		{"start_with_specials", "-.-a", "-a"},         // first '-' kept, following special '.' dropped
		{"only_specials", "---", "-"},                 // single remaining special
		{"alternating_specials", "._._._", "."},       // keep the first special only
		{"specials_between", "a__b--c..d", "a_b-c.d"}, // each run reduced to one
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := collapseRepeatedChars(tt.in)
			assert.Equal(t, tt.want, got)
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
		{'_', false},
		{'-', false},
		{'.', false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, isAlphanumeric(tt.r))
	}
}

func TestGenerateTagFromRef(t *testing.T) {
	tests := []struct {
		name string
		ref  string
		want string
	}{
		{"heads_feature", "refs/heads/feature/awesome", "feature-awesome"},
		{"tags_version", "refs/tags/v1.2.3", "v1.2.3"},
		{"pull_request", "refs/pull/123/head", "pr-123"},
		{"pull_request_missing_number_becomes_pr", "refs/pull//merge", "pr"},
		{"unknown_format", "weird/ref", "weird-ref"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateTagFromRef(tt.ref)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.True(t, IsValidTag(got))
		})
	}
}

func TestGenerateTagFromRef_Errors(t *testing.T) {
	_, err := GenerateTagFromRef("")
	assert.Error(t, err)

	// invalid pull request ref format: only "refs/pull/" with less than 3 parts?
	// "refs/pull/" splits to ["refs","pull",""], len >= 3 -> handled -> becomes "pr"
	got, err := GenerateTagFromRef("refs/pull/")
	assert.NoError(t, err)
	assert.Equal(t, "pr", got)
}

func TestTruncateTag_SuccessAndTrimming(t *testing.T) {
	tests := []struct {
		name      string
		tag       string
		maxLen    int
		want      string
		shouldErr bool
	}{
		{"short_valid_no_truncate", "abc", 5, "abc", false},
		{"truncate_simple", "abcdef", 3, "abc", false},
		{"truncate_and_trim_trailing", "abc----", 5, "abc", false}, // "abc--" -> trim -> "abc"
		{"truncate_and_trim_both_ends", "-abcdef", 3, "ab", false}, // "-ab" -> trim -> "ab"
		{"max_greater_than_cap", strings.Repeat("a", 200), 1000, strings.Repeat("a", MaxTagLength), false},
		{"invalid_when_not_truncated", "-abc", 10, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TruncateTag(tt.tag, tt.maxLen)
			if tt.shouldErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.True(t, IsValidTag(got))
			assert.LessOrEqual(t, len(got), tt.maxLen)
			assert.LessOrEqual(t, len(got), MaxTagLength)
		})
	}
}

func TestTruncateTag_Errors(t *testing.T) {
	_, err := TruncateTag("abc", 0)
	assert.Error(t, err)
	_, err = TruncateTag("abc", -5)
	assert.Error(t, err)
}

func TestNormalizeTag(t *testing.T) {
	tests := []struct {
		name    string
		tag     string
		want    string
		wantErr bool
	}{
		{"lowercase_and_sanitize", "FeaTURE/Foo", "feature-foo", false},
		{"already_valid_mixed", "V1.0.0", "v1.0.0", false},
		{"empty_error", "", "", true},
		{"collapse_separators", "HELLO___WORLD", "hello_world", false},
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
			assert.Equal(t, tt.want, got)
			assert.True(t, IsValidTag(got))
		})
	}
}

func TestAppendSuffix_SimpleAndTruncate(t *testing.T) {
	tests := []struct {
		name      string
		tag       string
		suffix    string
		want      string
		shouldErr bool
	}{
		{"simple_append_dash_added", "base", "rc1", "base-rc1", false},
		{"append_with_dot_prefix_kept", "base", ".suf", "base.suf", false},
		{"append_with_underscore_prefix_kept", "base", "_suf", "base_suf", false},
		{"truncate_base_to_fit_suffix", strings.Repeat("b", 127), "12345", strings.Repeat("b", 122) + "-12345", false}, // suffix gets dash added
		{"suffix_empty_returns_base_even_if_invalid", "-abc", "", "-abc", false},
		{"invalid_combined_due_to_suffix", "base", "!", "", true}, // '!' invalid char causes validation error
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AppendSuffix(tt.tag, tt.suffix)
			if tt.shouldErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
			if tt.suffix != "" {
				assert.True(t, IsValidTag(got))
			}
			assert.LessOrEqual(t, len(got), MaxTagLength)
		})
	}
}

func TestAppendSuffix_SuffixTooLong(t *testing.T) {
	base := "base"
	tooLongSuffix := strings.Repeat("x", MaxTagLength+100) // will have '-' prepended, ensuring > MaxTagLength
	_, err := AppendSuffix(base, tooLongSuffix)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "suffix too long")
}

func TestAppendSuffix_TruncateInvalidBaseToValid(t *testing.T) {
	// Base is invalid and very long; truncation will be applied and trimming removes leading dash.
	base := "-" + strings.Repeat("a", 200)
	suffix := "-x"
	got, err := AppendSuffix(base, suffix)
	assert.NoError(t, err)
	assert.True(t, IsValidTag(got))
	assert.True(t, strings.HasSuffix(got, "-x"))
}

func TestAppendSuffix_InvalidCombinedWhenNoTruncation(t *testing.T) {
	// Base invalid but short; since no truncation occurs, validation of combined should fail.
	_, err := AppendSuffix("-abc", "-suf")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "resulting tag is invalid")
}
