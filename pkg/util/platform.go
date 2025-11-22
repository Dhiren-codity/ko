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
	"runtime"
	"strings"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

// Platform represents an OS/architecture combination
type Platform struct {
	OS           string
	Architecture string
	Variant      string
}

// String returns the platform as a string in the format "os/arch[/variant]"
func (p Platform) String() string {
	if p.Variant != "" {
		return fmt.Sprintf("%s/%s/%s", p.OS, p.Architecture, p.Variant)
	}
	return fmt.Sprintf("%s/%s", p.OS, p.Architecture)
}

// ParsePlatform parses a platform string into its components
// Supports formats: "os/arch" and "os/arch/variant"
func ParsePlatform(platform string) (*Platform, error) {
	if platform == "" {
		return nil, fmt.Errorf("platform cannot be empty")
	}

	parts := strings.Split(platform, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid platform format: %s (expected os/arch or os/arch/variant)", platform)
	}

	if len(parts) > 3 {
		return nil, fmt.Errorf("invalid platform format: %s (too many components)", platform)
	}

	p := &Platform{
		OS:           parts[0],
		Architecture: parts[1],
	}

	if len(parts) == 3 {
		p.Variant = parts[2]
	}

	// Validate components
	if p.OS == "" {
		return nil, fmt.Errorf("platform OS cannot be empty")
	}

	if p.Architecture == "" {
		return nil, fmt.Errorf("platform architecture cannot be empty")
	}

	return p, nil
}

// ToPlatform converts a v1.Platform to our Platform type
func ToPlatform(vp v1.Platform) Platform {
	return Platform{
		OS:           vp.OS,
		Architecture: vp.Architecture,
		Variant:      vp.Variant,
	}
}

// ToV1Platform converts our Platform type to v1.Platform
func (p Platform) ToV1Platform() v1.Platform {
	return v1.Platform{
		OS:           p.OS,
		Architecture: p.Architecture,
		Variant:      p.Variant,
	}
}

// IsValidPlatform checks if a platform string is valid
func IsValidPlatform(platform string) bool {
	p, err := ParsePlatform(platform)
	if err != nil {
		return false
	}

	// Check if OS is recognized
	if !isValidOS(p.OS) {
		return false
	}

	// Check if arch is recognized
	if !isValidArch(p.Architecture) {
		return false
	}

	return true
}

// GetHostPlatform returns the current host platform
func GetHostPlatform() Platform {
	return Platform{
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
	}
}

// NormalizePlatform normalizes a platform string to lowercase
func NormalizePlatform(platform string) (string, error) {
	p, err := ParsePlatform(platform)
	if err != nil {
		return "", err
	}

	p.OS = strings.ToLower(p.OS)
	p.Architecture = strings.ToLower(p.Architecture)
	p.Variant = strings.ToLower(p.Variant)

	return p.String(), nil
}

// MatchesPlatform checks if two platforms match, considering variants
func MatchesPlatform(p1, p2 string) (bool, error) {
	platform1, err := ParsePlatform(p1)
	if err != nil {
		return false, fmt.Errorf("invalid platform1: %w", err)
	}

	platform2, err := ParsePlatform(p2)
	if err != nil {
		return false, fmt.Errorf("invalid platform2: %w", err)
	}

	// OS and Architecture must match
	if !strings.EqualFold(platform1.OS, platform2.OS) {
		return false, nil
	}

	if !strings.EqualFold(platform1.Architecture, platform2.Architecture) {
		return false, nil
	}

	// Variants match if both are empty or both are equal
	if platform1.Variant == "" && platform2.Variant == "" {
		return true, nil
	}

	if platform1.Variant != "" && platform2.Variant != "" {
		return strings.EqualFold(platform1.Variant, platform2.Variant), nil
	}

	// One has variant, one doesn't - not a match
	return false, nil
}

// FilterPlatforms filters a list of platforms keeping only valid ones
func FilterPlatforms(platforms []string) ([]string, []string) {
	var valid []string
	var invalid []string

	for _, p := range platforms {
		if IsValidPlatform(p) {
			valid = append(valid, p)
		} else {
			invalid = append(invalid, p)
		}
	}

	return valid, invalid
}

// isValidOS checks if an OS string is recognized
func isValidOS(os string) bool {
	validOSes := map[string]bool{
		"linux":   true,
		"darwin":  true,
		"windows": true,
		"freebsd": true,
		"openbsd": true,
		"netbsd":  true,
		"plan9":   true,
		"solaris": true,
		"aix":     true,
		"android": true,
		"ios":     true,
		"js":      true,
		"wasip1":  true,
	}

	return validOSes[strings.ToLower(os)]
}

// isValidArch checks if an architecture string is recognized
func isValidArch(arch string) bool {
	validArches := map[string]bool{
		"386":      true,
		"amd64":    true,
		"arm":      true,
		"arm64":    true,
		"ppc64":    true,
		"ppc64le":  true,
		"mips":     true,
		"mipsle":   true,
		"mips64":   true,
		"mips64le": true,
		"s390x":    true,
		"riscv64":  true,
		"wasm":     true,
		"loong64":  true,
	}

	return validArches[strings.ToLower(arch)]
}
