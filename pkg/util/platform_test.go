package util

import (
	"testing"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPlatformString_Valid_001 tests the String method for valid Platform struct values.
func TestPlatformString_Valid_001(t *testing.T) {
	platform := Platform{
		OS:           "linux",
		Architecture: "amd64",
		Variant:      "v8",
	}
	expected := "linux/amd64/v8"
	result := platform.String()
	assert.Equal(t, expected, result)
}

// TestParsePlatform_ValidAndInvalid_002 tests the ParsePlatform function for both valid and invalid inputs.
func TestParsePlatform_ValidAndInvalid_002(t *testing.T) {
	validPlatform := "linux/amd64/v8"
	invalidPlatform := "linux"

	// Test valid platform
	parsedPlatform, err := ParsePlatform(validPlatform)
	require.NoError(t, err)
	assert.Equal(t, "linux", parsedPlatform.OS)
	assert.Equal(t, "amd64", parsedPlatform.Architecture)
	assert.Equal(t, "v8", parsedPlatform.Variant)

	// Test invalid platform
	parsedPlatform, err = ParsePlatform(invalidPlatform)
	assert.Error(t, err)
	assert.Nil(t, parsedPlatform)
}

// TestToPlatform_Conversion_003 tests the ToPlatform function for correct conversion.
func TestToPlatform_Conversion_003(t *testing.T) {
	v1Platform := v1.Platform{
		OS:           "linux",
		Architecture: "amd64",
		Variant:      "v8",
	}
	expectedPlatform := Platform{
		OS:           "linux",
		Architecture: "amd64",
		Variant:      "v8",
	}
	result := ToPlatform(v1Platform)
	assert.Equal(t, expectedPlatform, result)
}

// TestIsValidPlatform_Validation_004 tests the IsValidPlatform function for valid and invalid inputs.
func TestIsValidPlatform_Validation_004(t *testing.T) {
	validPlatform := "linux/amd64"
	invalidPlatform := "unknown/unknown"

	assert.True(t, IsValidPlatform(validPlatform))
	assert.False(t, IsValidPlatform(invalidPlatform))
}

// TestNormalizePlatform_Lowercase_005 tests the NormalizePlatform function for correct normalization.
func TestNormalizePlatform_Lowercase_005(t *testing.T) {
	platform := "Linux/AMD64/V8"
	expected := "linux/amd64/v8"

	result, err := NormalizePlatform(platform)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

// TestFilterPlatforms_Filtering_006 tests the FilterPlatforms function for correct filtering.
func TestFilterPlatforms_Filtering_006(t *testing.T) {
	platforms := []string{"linux/amd64", "unknown/unknown", "darwin/arm64"}
	expectedValid := []string{"linux/amd64", "darwin/arm64"}
	expectedInvalid := []string{"unknown/unknown"}

	valid, invalid := FilterPlatforms(platforms)
	assert.Equal(t, expectedValid, valid)
	assert.Equal(t, expectedInvalid, invalid)
}

// TestMatchesPlatform_Matching_007 tests the MatchesPlatform function for correct matching.
func TestMatchesPlatform_Matching_007(t *testing.T) {
	platform1 := "linux/amd64/v8"
	platform2 := "linux/amd64/v8"
	platform3 := "linux/amd64"

	match, err := MatchesPlatform(platform1, platform2)
	require.NoError(t, err)
	assert.True(t, match)

	match, err = MatchesPlatform(platform1, platform3)
	require.NoError(t, err)
	assert.False(t, match)
}

// TestGetHostPlatform_Host_008 tests the GetHostPlatform function for correct host platform retrieval.
func TestGetHostPlatform_Host_008(t *testing.T) {
	hostPlatform := GetHostPlatform()
	assert.NotEmpty(t, hostPlatform.OS)
	assert.NotEmpty(t, hostPlatform.Architecture)
}

// TestToV1Platform_Conversion_011 tests the ToV1Platform method for correct conversion from Platform to v1.Platform.
func TestToV1Platform_Conversion_011(t *testing.T) {
	platform := Platform{
		OS:           "linux",
		Architecture: "amd64",
		Variant:      "v8",
	}
	expectedV1Platform := v1.Platform{
		OS:           "linux",
		Architecture: "amd64",
		Variant:      "v8",
	}
	result := platform.ToV1Platform()
	assert.Equal(t, expectedV1Platform, result)
}

// TestNormalizePlatform_InvalidInput_012 tests the NormalizePlatform function for invalid input.
func TestNormalizePlatform_InvalidInput_012(t *testing.T) {
	invalidPlatform := "invalidplatform"

	result, err := NormalizePlatform(invalidPlatform)
	assert.Error(t, err)
	assert.Empty(t, result)
}

// TestMatchesPlatform_InvalidInput_013 tests the MatchesPlatform function for invalid platform inputs.
func TestMatchesPlatform_InvalidInput_013(t *testing.T) {
	invalidPlatform1 := "invalidplatform1"
	invalidPlatform2 := "invalidplatform2"

	match, err := MatchesPlatform(invalidPlatform1, invalidPlatform2)
	assert.Error(t, err)
	assert.False(t, match)
}

// TestPlatformString_NoVariant_101 tests the String method for a Platform without a variant.
func TestPlatformString_NoVariant_101(t *testing.T) {
	p := Platform{
		OS:           "darwin",
		Architecture: "arm64",
	}
	expected := "darwin/arm64"
	result := p.String()
	assert.Equal(t, expected, result)
}

// TestParsePlatform_ErrorScenarios_202 covers various error scenarios for ParsePlatform.
func TestParsePlatform_ErrorScenarios_202(t *testing.T) {
	testCases := []struct {
		name          string
		platformStr   string
		expectedError string
	}{
		{
			name:          "empty platform string",
			platformStr:   "",
			expectedError: "platform cannot be empty",
		},
		{
			name:          "too many components",
			platformStr:   "linux/amd64/v8/extra",
			expectedError: "invalid platform format: linux/amd64/v8/extra (too many components)",
		},
		{
			name:          "empty OS",
			platformStr:   "/amd64",
			expectedError: "platform OS cannot be empty",
		},
		{
			name:          "empty architecture",
			platformStr:   "linux/",
			expectedError: "platform architecture cannot be empty",
		},
		{
			name:          "valid no variant",
			platformStr:   "linux/amd64",
			expectedError: "", // No error
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p, err := ParsePlatform(tc.platformStr)
			if tc.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
				assert.Nil(t, p)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, p)
			}
		})
	}
}

// TestIsValidPlatform_FailureCases_303 tests failure cases for IsValidPlatform.
func TestIsValidPlatform_FailureCases_303(t *testing.T) {
	assert.False(t, IsValidPlatform("linux"))                 // Parse error
	assert.False(t, IsValidPlatform("nonexistentos/amd64"))   // Invalid OS
	assert.False(t, IsValidPlatform("linux/nonexistentarch")) // Invalid arch
}

// TestMatchesPlatform_Comprehensive_404 provides comprehensive tests for MatchesPlatform.
func TestMatchesPlatform_Comprehensive_404(t *testing.T) {
	testCases := []struct {
		name      string
		p1        string
		p2        string
		wantMatch bool
		wantErr   bool
	}{
		{name: "os mismatch", p1: "linux/amd64", p2: "darwin/amd64", wantMatch: false, wantErr: false},
		{name: "arch mismatch", p1: "linux/amd64", p2: "linux/arm64", wantMatch: false, wantErr: false},
		{name: "case-insensitive match", p1: "LINUX/AMD64", p2: "linux/amd64", wantMatch: true, wantErr: false},
		{name: "variant mismatch", p1: "linux/arm/v7", p2: "linux/arm/v8", wantMatch: false, wantErr: false},
		{name: "case-insensitive variant match", p1: "linux/arm/V7", p2: "linux/arm/v7", wantMatch: true, wantErr: false},
		{name: "one with variant, one without (reversed)", p1: "linux/amd64", p2: "linux/amd64/v8", wantMatch: false, wantErr: false},
		{name: "invalid platform 1", p1: "invalid", p2: "linux/amd64", wantMatch: false, wantErr: true},
		{name: "invalid platform 2", p1: "linux/amd64", p2: "invalid", wantMatch: false, wantErr: true},
		{name: "both variants empty match", p1: "linux/amd64", p2: "linux/amd64", wantMatch: true, wantErr: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotMatch, err := MatchesPlatform(tc.p1, tc.p2)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.wantMatch, gotMatch)
		})
	}
}
