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
		{"simple valid", "v1.2.3", true},
		{"allowed chars mix", "release_2025.11-rc1", true},
		{"empty", "", false},
		{"starts with dash", "-bad", false},
		{"only underscore", "_", false},
		{"max length exact", strings.Repeat("a", MaxTagLength), true},
		{"exceeds max length", strings.Repeat("a", MaxTagLength+1), false},
		{"trailing dash allowed", "abc-", true},
		{"contains hyphen dot underscore", "a-.-_b", true},
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
		{"branch with slash and spaces", "Feature/Foo Bar@2025!", "Feature-Foo-Bar-2025", false},
		{"trims leading trailing specials", "--abc--", "abc", false},
		{"multiple consecutive specials collapsed", "a__b..c---", "a_b.c", false},
		{"starts with dot", "...abc", "abc", false},
		{"empty input error", "", "", true},
		{"all invalid becomes empty", "///", "", true},
		{"enforce max length", strings.Repeat("a", MaxTagLength+10), strings.Repeat("a", MaxTagLength), false},
		{"ensure alphanumeric start via trim, not prefix", "-abc", "abc", false},
		{"spaces mapped and trimmed", "   foo   ", "foo", false},
		{"unicode and punctuation replaced", "rélease🚀/v1.0!", "r-lease-v1.0", false}, // unicode -> '-', then trim/collapse
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

func TestTruncateTag(t *testing.T) {
	tests := []struct {
		name      string
		tag       string
		maxLength int
		want      string
		wantErr   bool
	}{
		{"returns same when under max and valid", "abc", 5, "abc", false},
		{"errors when under max but invalid", "-abc", 10, "", true},
		{"truncate and trim trailing specials", "abc--", 4, "abc", false},
		{"truncate to MaxTag when requested too large", strings.Repeat("a", 200), 2000, strings.Repeat("a", MaxTagLength), false},
		{"maxLength <=0 error", "abc", 0, "", true},
		{"truncate results empty error", "--", 1, "", true},
		{"exact boundary", strings.Repeat("b", 10), 10, strings.Repeat("b", 10), false},
		{"trim leading specials after truncation", "--abc", 2, "", true}, // truncated "--" -> trimmed "" -> invalid
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
		{"lowercase and sanitize", "My/Tag", "my-tag", false},
		{"error empty", "", "", true},
		{"sanitize cannot create non-empty", "///", "", true},
		{"leading dash corrected via trim", "-ABC", "abc", false},
		{"already valid lowercased", "v1.2.3", "v1.2.3", false},
		{"long reduced to max", strings.Repeat("A", MaxTagLength+5), strings.Repeat("a", MaxTagLength), false},
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
		{"branch", "refs/heads/feature/foo", "feature-foo", false},
		{"tag", "refs/tags/v1.0.0", "v1.0.0", false},
		{"pull request", "refs/pull/123/head", "pr-123", false},
		{"pull request missing id but len>=3 (empty)", "refs/pull//head", "pr", false},
		{"pull ref invalid format", "refs/pull", "", true},
		{"unknown default sanitized", "custom/ref value", "custom-ref-value", false},
		{"empty error", "", "", true},
		{"branch with spaces and unicode", "refs/heads/feat/mañana 🚀", "feat-ma-ana", false},
		{"pull request just number", "refs/pull/456", "pr-456", false},
		{"refs/tags/with/slash", "refs/tags/release/1.2.3", "release-1.2.3", false},
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
			assert.LessOrEqual(t, len(got), MaxTagLength)
		})
	}
}

func TestAppendSuffix(t *testing.T) {
	tests := []struct {
		name      string
		tag       string
		suffix    string
		want      string
		wantErr   bool
		checkFunc func(string)
	}{
		{"simple add dashless suffix", "abc", "123", "abc-123", false, nil},
		{"suffix starting with dot kept", "abc", ".sfx", "abc.sfx", false, nil},
		{"suffix starting with underscore kept", "abc", "_sfx", "abc_sfx", false, nil},
		{"empty suffix returns base", "abc", "", "abc", false, nil},
		{"empty base returns error", "", "sfx", "", true, nil},
		{"combined invalid due to suffix chars", "abc", "#bad", "", true, nil},
		{"base invalid leading dash returns error", "-abc", "-ok", "", true, nil},
		{
			name:    "combined requires truncation of base",
			tag:     strings.Repeat("a", MaxTagLength),
			suffix:  "Z",
			want:    strings.Repeat("a", MaxTagLength-2) + "-Z",
			wantErr: false,
			checkFunc: func(res string) {
				assert.True(t, IsValidTag(res))
				assert.Equal(t, MaxTagLength, len(res))
				assert.True(t, strings.HasSuffix(res, "-Z"))
			},
		},
		{
			name:    "suffix too long error",
			tag:     "abc",
			suffix:  strings.Repeat("x", MaxTagLength+1),
			want:    "",
			wantErr: true,
		},
		{
			name:    "truncation uses TruncateTag rules (trims trailing specials)",
			tag:     strings.Repeat("b", MaxTagLength-1) + "-", // valid until trailing dash; still valid overall
			suffix:  "c",
			want:    strings.Repeat("b", MaxTagLength-2) + "-c",
			wantErr: false,
		},
		{
			name:    "no extra separator added when suffix already has '-'",
			tag:     "base",
			suffix:  "-tail",
			want:    "base-tail",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := AppendSuffix(tt.tag, tt.suffix)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
			if tt.checkFunc != nil {
				tt.checkFunc(got)
			}
			assert.True(t, IsValidTag(got))
			assert.LessOrEqual(t, len(got), MaxTagLength)
		})
	}
}

func TestIsAlphanumeric(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want bool
	}{
		{"lowercase letter", 'a', true},
		{"uppercase letter", 'Z', true},
		{"digit", '0', true},
		{"dash", '-', false},
		{"underscore", '_', false},
		{"dot", '.', false},
		{"slash", '/', false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isAlphanumeric(tt.r))
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
		{"digit", '1', false},
		{"slash", '/', false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isSpecialChar(tt.r))
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
		{"only specials", "-._", "-"},
		{"triple dash", "---", "-"},
		{"underscores collapsed", "a__b", "a_b"},
		{"dots collapsed", "a..b", "a.b"},
		{"mixed specials collapsed to first", "a-._b", "a-b"},
		{"already distinct", "a-b-c", "a-b-c"},
		{"no specials", "abc", "abc"},
		{"ends with special run", "a__b..c---", "a_b.c-"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := collapseRepeatedChars(tt.in)
			assert.Equal(t, tt.want, got)
		})
	}
}
