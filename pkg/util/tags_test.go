package util

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidTag_Table(t *testing.T) {
	tests := []struct {
		name string
		tag  string
		want bool
	}{
		{"empty", "", false},
		{"simple valid", "v1", true},
		{"starts with underscore", "_abc", false},
		{"contains slash", "feat/foo", false},
		{"allowed mix", "A_b.c-1", true},
		{"max length exact", strings.Repeat("a", MaxTagLength), true},
		{"exceeds max length", strings.Repeat("a", MaxTagLength+1), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidTag(tt.tag)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSanitizeTag_Table(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{"empty input", "", "", true},
		{"feature slash", "feature/foo", "feature-foo", false},
		{"release with space", "release candidate", "release-candidate", false},
		{"weird and special chars", " .weird//tag??", "weird-tag", false},
		{"collapse mixed specials", "a-._-._b", "a-b", false},
		{"only punctuation", "///...___", "", true},
		{"trim leading trailing", "--leading--", "leading", false},
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
			assert.True(t, IsValidTag(got), "sanitized tag should be valid")
		})
	}
}

func TestSanitizeTag_TruncationTrimsTrailing(t *testing.T) {
	// Create a string where character at position MaxTagLength is a special char ('.')
	// so that after truncation it ends with '.' and should be trimmed.
	in := strings.Repeat("a", MaxTagLength-1) + "." + "bbbb"
	got, err := SanitizeTag(in)
	assert.NoError(t, err)
	assert.Equal(t, strings.Repeat("a", MaxTagLength-1), got)
	assert.True(t, IsValidTag(got))
}

func TestTruncateTag_Table(t *testing.T) {
	tests := []struct {
		name      string
		tag       string
		max       int
		want      string
		wantErr   bool
		assertLen int
	}{
		{"non-positive max", "abc", 0, "", true, 0},
		{"valid no truncation", "abc", 5, "abc", false, 3},
		{"invalid no truncation", "-abc", 10, "", true, 0},
		{"truncate simple", "abcde", 3, "abc", false, 3},
		{"truncate then trim trailing special", strings.Repeat("a", 5) + "." + "tail", 6, strings.Repeat("a", 5), false, 5},
		{"exceed hard max enforces", strings.Repeat("a", MaxTagLength+2), MaxTagLength + 100, strings.Repeat("a", MaxTagLength), false, MaxTagLength},
		{"truncate cannot maintain validity", "-_-_-_-", 1, "", true, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TruncateTag(tt.tag, tt.max)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.Len(t, got, tt.assertLen)
			assert.True(t, IsValidTag(got))
		})
	}
}

func TestAppendSuffix_Table(t *testing.T) {
	tests := []struct {
		name       string
		tag        string
		suffix     string
		want       string
		wantErr    bool
		checkValid bool
	}{
		{"empty tag", "", "sfx", "", true, false},
		{"empty suffix returns base", "base", "", "base", false, true},
		{"adds dash when missing", "base", "sfx", "base-sfx", false, true},
		{"keeps provided dash", "base", "-sfx", "base-sfx", false, true},
		{"invalid characters in suffix", "base", "bad@chars", "", true, false},
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
			if tt.checkValid {
				assert.True(t, IsValidTag(got))
				assert.LessOrEqual(t, len(got), MaxTagLength)
			}
		})
	}

	t.Run("combined length exceeds max truncates base", func(t *testing.T) {
		base := strings.Repeat("a", 120)
		suffix := "verylongsuffix" // length 15, will be prefixed with '-' = 16
		got, err := AppendSuffix(base, suffix)
		assert.NoError(t, err)
		assert.Len(t, got, MaxTagLength)
		assert.True(t, IsValidTag(got))
		assert.True(t, strings.HasSuffix(got, "-"+suffix))
		// base portion should be truncated to MaxTagLength - len("-"+suffix)
		expectedBaseLen := MaxTagLength - (1 + len(suffix))
		assert.Equal(t, strings.Repeat("a", expectedBaseLen), strings.TrimSuffix(got, "-"+suffix))
	})

	t.Run("suffix too long", func(t *testing.T) {
		suffix := "-" + strings.Repeat("x", MaxTagLength-1) // total length 128
		_, err := AppendSuffix("base", suffix)
		assert.Error(t, err)
	})

	t.Run("failed to truncate base for suffix (invalid base after truncation)", func(t *testing.T) {
		// suffix length 127 so base allowed max length = 1
		suffix := "-" + strings.Repeat("x", MaxTagLength-2)
		_, err := AppendSuffix("-invalid", suffix)
		assert.Error(t, err)
	})
}

func TestIsAlphanumeric(t *testing.T) {
	tests := []struct {
		r    rune
		want bool
	}{
		{'a', true},
		{'Z', true},
		{'5', true},
		{'_', false},
		{'-', false},
		{'.', false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, isAlphanumeric(tt.r), string(tt.r))
	}
}

func TestCollapseRepeatedChars(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"no specials", "abc", "abc"},
		{"collapse hyphens", "a---b", "a-b"},
		{"collapse dots", "a...b", "a.b"},
		{"collapse underscores", "a___b", "a_b"},
		{"mixed sequences keep first", "a-._-._b", "a-b"},
		{"leading specials kept once", "--a", "-a"},
		{"complex mix", "a---b...c___d-._-._e", "a-b.c_d-e"},
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
		{'Z', false},
		{'0', false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, isSpecialChar(tt.r), string(tt.r))
	}
}
