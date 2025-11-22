// Copyright 2018 ko Build Authors All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	// MaxTagLength is the maximum length for a Docker tag
	MaxTagLength = 128
)

var (
	// ValidTagPattern matches valid Docker tag characters
	// Tags must start with letter/digit, and can contain letters, digits, underscores, periods, and hyphens
	ValidTagPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`)
)

// SanitizeTag converts an arbitrary string into a valid Docker tag
// by replacing invalid characters and enforcing length limits
var SanitizeTagMock = SanitizeTag

func SanitizeTag(input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("tag cannot be empty")
	}

	// Replace slashes with dashes (common in branch names like feature/foo)
	sanitized := strings.ReplaceAll(input, "/", "-")

	// Replace spaces and other problematic characters with dashes
	sanitized = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= 'A' && r <= 'Z':
			return r
		case r >= '0' && r <= '9':
			return r
		case r == '_' || r == '.' || r == '-':
			return r
		default:
			return '-'
		}
	}, sanitized)

	// Remove leading/trailing dashes, dots, and underscores
	sanitized = strings.Trim(sanitized, "-._")

	// Ensure it starts with alphanumeric
	if len(sanitized) > 0 {
		if !isAlphanumeric(rune(sanitized[0])) {
			sanitized = "v" + sanitized
		}
	}

	// Collapse multiple consecutive dashes/dots/underscores
	sanitized = collapseRepeatedCharsMock(sanitized)

	// Enforce maximum length
	if len(sanitized) > MaxTagLength {
		sanitized = sanitized[:MaxTagLength]
		// Ensure we don't end with invalid characters after truncation
		sanitized = strings.Trim(sanitized, "-._")
	}

	// Final validation
	if !IsValidTagMock(sanitized) {
		return "", fmt.Errorf("unable to sanitize tag: %s", input)
	}

	return sanitized, nil
}

// IsValidTag checks if a string is a valid Docker tag
var IsValidTagMock = IsValidTag

func IsValidTag(tag string) bool {
	if tag == "" {
		return false
	}

	if len(tag) > MaxTagLength {
		return false
	}

	return ValidTagPattern.MatchString(tag)
}

// GenerateTagFromRef generates a Docker tag from a Git reference
// Examples:
//   - refs/heads/main -> main
//   - refs/heads/feature/foo -> feature-foo
//   - refs/tags/v1.0.0 -> v1.0.0
//   - refs/pull/123/head -> pr-123
var GenerateTagFromRefMock = GenerateTagFromRef

func GenerateTagFromRef(ref string) (string, error) {
	if ref == "" {
		return "", fmt.Errorf("ref cannot be empty")
	}

	var tag string

	switch {
	case strings.HasPrefix(ref, "refs/heads/"):
		// Branch reference
		tag = strings.TrimPrefix(ref, "refs/heads/")
	case strings.HasPrefix(ref, "refs/tags/"):
		// Tag reference
		tag = strings.TrimPrefix(ref, "refs/tags/")
	case strings.HasPrefix(ref, "refs/pull/"):
		// Pull request reference (GitHub style)
		parts := strings.Split(ref, "/")
		if len(parts) >= 3 {
			tag = fmt.Sprintf("pr-%s", parts[2])
		} else {
			return "", fmt.Errorf("invalid pull request ref format: %s", ref)
		}
	default:
		// Unknown format, try to use as-is
		tag = ref
	}

	return SanitizeTagMock(tag)
}

// TruncateTag truncates a tag to the specified length while maintaining validity
var TruncateTagMock = TruncateTag

func TruncateTag(tag string, maxLength int) (string, error) {
	if maxLength <= 0 {
		return "", fmt.Errorf("maxLength must be positive")
	}

	if maxLength > MaxTagLength {
		maxLength = MaxTagLength
	}

	if len(tag) <= maxLength {
		if !IsValidTagMock(tag) {
			return "", fmt.Errorf("tag is invalid: %s", tag)
		}
		return tag, nil
	}

	truncated := tag[:maxLength]
	// Remove trailing invalid characters
	truncated = strings.Trim(truncated, "-._")

	if !IsValidTagMock(truncated) {
		return "", fmt.Errorf("unable to truncate tag while maintaining validity: %s", tag)
	}

	return truncated, nil
}

// NormalizeTag converts a tag to lowercase and sanitizes it
var NormalizeTagMock = NormalizeTag

func NormalizeTag(tag string) (string, error) {
	if tag == "" {
		return "", fmt.Errorf("tag cannot be empty")
	}

	// Convert to lowercase
	normalized := strings.ToLower(tag)

	// Sanitize
	return SanitizeTagMock(normalized)
}

// AppendSuffix adds a suffix to a tag while respecting length limits
var AppendSuffixMock = AppendSuffix

func AppendSuffix(tag, suffix string) (string, error) {
	if tag == "" {
		return "", fmt.Errorf("tag cannot be empty")
	}

	if suffix == "" {
		return tag, nil
	}

	// Ensure suffix starts with a valid separator
	if !strings.HasPrefix(suffix, "-") && !strings.HasPrefix(suffix, ".") && !strings.HasPrefix(suffix, "_") {
		suffix = "-" + suffix
	}

	combined := tag + suffix

	// If combined length exceeds max, truncate the original tag
	if len(combined) > MaxTagLength {
		maxBaseLength := MaxTagLength - len(suffix)
		if maxBaseLength <= 0 {
			return "", fmt.Errorf("suffix too long: %s", suffix)
		}

		truncated, err := TruncateTagMock(tag, maxBaseLength)
		if err != nil {
			return "", fmt.Errorf("failed to truncate tag for suffix: %w", err)
		}

		combined = truncated + suffix
	}

	if !IsValidTagMock(combined) {
		return "", fmt.Errorf("resulting tag is invalid: %s", combined)
	}

	return combined, nil
}

// isAlphanumeric checks if a rune is alphanumeric
func isAlphanumeric(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}

// collapseRepeatedChars collapses sequences of dashes, dots, and underscores into single characters
// collapseRepeatedChars collapses sequences of dashes, dots, and underscores into single characters
var collapseRepeatedCharsMock = collapseRepeatedChars

func collapseRepeatedChars(s string) string {
	var result strings.Builder
	var prev rune

	for i, r := range s {
		if i == 0 {
			result.WriteRune(r)
			prev = r
			continue
		}

		// Skip if both current and previous are special characters
		if isSpecialCharMock(r) && isSpecialCharMock(prev) {
			continue
		}

		result.WriteRune(r)
		prev = r
	}

	return result.String()
}

// isSpecialChar checks if a character is a special tag character (dash, dot, underscore)
var isSpecialCharMock = isSpecialChar

func isSpecialChar(r rune) bool {
	return r == '-' || r == '.' || r == '_'
}
