package util

import (
	"testing"

	"strings"

	"fmt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSanitizeTag_EmptyInput_001 tests the behavior of SanitizeTag when the input is empty.
func TestSanitizeTag_EmptyInput_001(t *testing.T) {
	result, err := SanitizeTag("")
	require.Error(t, err)
	assert.Equal(t, "", result)
}

// TestGenerateTagFromRef_ValidBranch_002 tests the behavior of GenerateTagFromRef with a valid branch reference.
func TestGenerateTagFromRef_ValidBranch_002(t *testing.T) {
	result, err := GenerateTagFromRef("refs/heads/main")
	require.NoError(t, err)
	assert.Equal(t, "main", result)
}

// TestTruncateTag_ExceedsMaxLength_003 tests the behavior of TruncateTag when the tag exceeds the maximum length.
func TestTruncateTag_ExceedsMaxLength_003(t *testing.T) {
	longTag := strings.Repeat("a", MaxTagLength+10)
	result, err := TruncateTag(longTag, MaxTagLength)
	require.NoError(t, err)
	assert.Equal(t, strings.Repeat("a", MaxTagLength), result)
}

// TestNormalizeTag_MixedCaseInput_004 tests the behavior of NormalizeTag with a mixed-case input.
func TestNormalizeTag_MixedCaseInput_004(t *testing.T) {
	result, err := NormalizeTag("MiXeDcAsE")
	require.NoError(t, err)
	assert.Equal(t, "mixedcase", result)
}

// TestAppendSuffix_SuffixTooLong_005 tests the behavior of AppendSuffix when the suffix is too long.
func TestAppendSuffix_SuffixTooLong_005(t *testing.T) {
	tag := "validtag"
	longSuffix := strings.Repeat("b", MaxTagLength+10)
	result, err := AppendSuffix(tag, longSuffix)
	require.Error(t, err)
	assert.Equal(t, "", result)
}

// TestSanitizeTag_InvalidCharacters_006 tests the behavior of SanitizeTag when the input contains invalid characters.
func TestSanitizeTag_InvalidCharacters_006(t *testing.T) {
	input := "invalid/tag@name"
	result, err := SanitizeTag(input)
	require.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.NotContains(t, result, "/")
	assert.NotContains(t, result, "@")
}

// TestTruncateTag_WithinMaxLength_009 tests the behavior of TruncateTag when the tag is already within the maximum length.
func TestTruncateTag_WithinMaxLength_009(t *testing.T) {
	tag := "validtag"
	result, err := TruncateTag(tag, MaxTagLength)
	require.NoError(t, err)
	assert.Equal(t, tag, result)
}

// TestAppendSuffix_ValidCombination_011 tests the behavior of AppendSuffix when the tag and suffix combination is valid.
func TestAppendSuffix_ValidCombination_011(t *testing.T) {
	tag := "validtag"
	suffix := "-suffix"
	result, err := AppendSuffix(tag, suffix)
	require.NoError(t, err)
	assert.Equal(t, "validtag-suffix", result)
}

// TestCollapseRepeatedChars_ConsecutiveSpecialChars_012 tests the behavior of collapseRepeatedChars when the input contains consecutive special characters.
func TestCollapseRepeatedChars_ConsecutiveSpecialChars_012(t *testing.T) {
	input := "a--b__c..d"
	result := collapseRepeatedChars(input)
	assert.Equal(t, "a-b_c.d", result)
}

// TestIsValidTag_Comprehensive_234 provides a comprehensive suite of tests for the IsValidTag function.
func TestIsValidTag_Comprehensive_234(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "empty tag", input: "", want: false},
		{name: "long tag", input: strings.Repeat("a", MaxTagLength+1), want: false},
		{name: "max length valid tag", input: strings.Repeat("a", MaxTagLength), want: true},
		{name: "valid tag with letters and numbers", input: "v1alpha2", want: true},
		{name: "valid tag with dot", input: "v1.0", want: true},
		{name: "valid tag with hyphen", input: "feature-branch", want: true},
		{name: "valid tag with underscore", input: "my_tag", want: true},
		{name: "starts with dot", input: ".v1", want: false},
		{name: "starts with hyphen", input: "-v1", want: false},
		{name: "starts with underscore", input: "_v1", want: false},
		{name: "contains invalid char", input: "v1!", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidTag(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestNormalizeTag_Comprehensive_567 provides a comprehensive suite of tests for the NormalizeTag function.
func TestNormalizeTag_Comprehensive_567(t *testing.T) {
	originalSanitizeTag := SanitizeTagMock
	defer func() { SanitizeTagMock = originalSanitizeTag }()

	tests := []struct {
		name          string
		tag           string
		mockSetup     func()
		want          string
		wantErr       bool
		expectedError string
	}{
		{name: "empty tag", tag: "", wantErr: true, expectedError: "tag cannot be empty"},
		{name: "tag with uppercase and invalid chars", tag: "My/Tag!", want: "my-tag"},
		{
			name: "sanitize tag fails",
			tag:  "SomeTag",
			mockSetup: func() {
				SanitizeTagMock = func(input string) (string, error) {
					return "", fmt.Errorf("sanitize failed")
				}
			},
			wantErr:       true,
			expectedError: "sanitize failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SanitizeTagMock = SanitizeTag // Reset mock
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			got, err := NormalizeTag(tt.tag)

			if tt.wantErr {
				require.Error(t, err)
				assert.EqualError(t, err, tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// TestAppendSuffix_Comprehensive_678 provides a comprehensive suite of tests for the AppendSuffix function.
func TestAppendSuffix_Comprehensive_678(t *testing.T) {
	originalTruncateTag := TruncateTagMock
	originalIsValidTag := IsValidTagMock
	defer func() {
		TruncateTagMock = originalTruncateTag
		IsValidTagMock = originalIsValidTag
	}()

	tests := []struct {
		name          string
		tag           string
		suffix        string
		mockSetup     func()
		want          string
		wantErr       bool
		expectedError string
	}{
		{name: "empty tag", tag: "", suffix: "foo", wantErr: true, expectedError: "tag cannot be empty"},
		{name: "empty suffix", tag: "base", suffix: "", want: "base"},
		{name: "suffix without separator", tag: "base", suffix: "foo", want: "base-foo"},
		{name: "suffix with dot separator", tag: "base", suffix: ".foo", want: "base.foo"},
		{name: "suffix with underscore separator", tag: "base", suffix: "_foo", want: "base_foo"},
		{
			name:   "combined length exceeds max, needs truncation",
			tag:    strings.Repeat("a", 120),
			suffix: "suffix12345", // 11 chars, becomes 12 with '-'
			want:   strings.Repeat("a", 116) + "-suffix12345",
		},
		{
			name:          "suffix is too long",
			tag:           "base",
			suffix:        strings.Repeat("b", MaxTagLength),
			wantErr:       true,
			expectedError: fmt.Sprintf("suffix too long: -%s", strings.Repeat("b", MaxTagLength)),
		},
		{
			name:   "truncate tag fails",
			tag:    strings.Repeat("a", 120),
			suffix: "longsuffix",
			mockSetup: func() {
				TruncateTagMock = func(tag string, maxLength int) (string, error) {
					return "", fmt.Errorf("truncate failed")
				}
			},
			wantErr:       true,
			expectedError: "failed to truncate tag for suffix: truncate failed",
		},
		{
			name:   "final tag is invalid",
			tag:    "base",
			suffix: "foo",
			mockSetup: func() {
				IsValidTagMock = func(tag string) bool { return tag != "base-foo" }
			},
			wantErr:       true,
			expectedError: "resulting tag is invalid: base-foo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			TruncateTagMock = TruncateTag
			IsValidTagMock = IsValidTag
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			got, err := AppendSuffix(tt.tag, tt.suffix)

			if tt.wantErr {
				require.Error(t, err)
				assert.EqualError(t, err, tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
