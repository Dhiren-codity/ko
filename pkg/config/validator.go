// Copyright 2025 ko Build Authors All Rights Reserved.
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

package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Config represents a ko configuration file structure
type Config struct {
	BaseImage     string            `json:"baseImage"`
	Platforms     []string          `json:"platforms"`
	Images        []string          `json:"images"`
	Labels        map[string]string `json:"labels"`
	AllowInsecure bool              `json:"allowInsecure"`
}

// ValidateConfig validates a ko configuration file at the given path.
// Returns an error if the configuration is invalid.
func ValidateConfig(path string) error {
	if !FileExists(path) {
		return fmt.Errorf("configuration file not found: %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read configuration file: %w", err)
	}

	cfg, err := ParseConfig(data)
	if err != nil {
		return fmt.Errorf("failed to parse configuration: %w", err)
	}

	errors := ValidateConfigStructure(cfg)
	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(errors, "; "))
	}

	if err := ValidateImageReferences(cfg.Images); err != nil {
		return fmt.Errorf("invalid image references: %w", err)
	}

	return nil
}

// FileExists checks if a file exists at the given path.
func FileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// ParseConfig parses configuration data from JSON bytes.
func ParseConfig(data []byte) (*Config, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty configuration data")
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	return &cfg, nil
}

// ValidateConfigStructure validates the structure and required fields of a configuration.
// Returns a list of validation errors.
func ValidateConfigStructure(cfg *Config) []string {
	if cfg == nil {
		return []string{"configuration is nil"}
	}

	var errors []string

	requiredFieldErrors := CheckRequiredFields(cfg)
	errors = append(errors, requiredFieldErrors...)

	if len(cfg.Platforms) > 0 {
		for _, platform := range cfg.Platforms {
			if !IsValidPlatform(platform) {
				errors = append(errors, fmt.Sprintf("invalid platform format: %s", platform))
			}
		}
	}

	if cfg.BaseImage != "" && !IsValidImageReference(cfg.BaseImage) {
		errors = append(errors, fmt.Sprintf("invalid base image reference: %s", cfg.BaseImage))
	}

	return errors
}

// CheckRequiredFields checks if all required fields are present in the configuration.
func CheckRequiredFields(cfg *Config) []string {
	var errors []string

	if cfg.BaseImage == "" {
		errors = append(errors, "baseImage is required")
	}

	if len(cfg.Images) == 0 {
		errors = append(errors, "at least one image must be specified")
	}

	return errors
}

// ValidateImageReferences validates a list of image references.
func ValidateImageReferences(refs []string) error {
	if len(refs) == 0 {
		return fmt.Errorf("no image references provided")
	}

	var invalidRefs []string
	for _, ref := range refs {
		if !IsValidImageReference(ref) {
			invalidRefs = append(invalidRefs, ref)
		}
	}

	if len(invalidRefs) > 0 {
		return fmt.Errorf("invalid image references: %s", strings.Join(invalidRefs, ", "))
	}

	return nil
}

// IsValidImageReference checks if an image reference is valid.
func IsValidImageReference(ref string) bool {
	if ref == "" {
		return false
	}

	imagePattern := regexp.MustCompile(`^[a-z0-9]+([\.\-_][a-z0-9]+)*(/[a-z0-9]+([\.\-_][a-z0-9]+)*)*(:[a-zA-Z0-9\.\-_]+)?(@sha256:[a-f0-9]{64})?$`)
	return imagePattern.MatchString(ref)
}

// IsValidPlatform checks if a platform string is in valid format (os/arch).
func IsValidPlatform(platform string) bool {
	if platform == "" {
		return false
	}

	parts := strings.Split(platform, "/")
	if len(parts) != 2 {
		return false
	}

	validOS := map[string]bool{
		"linux":   true,
		"darwin":  true,
		"windows": true,
	}

	validArch := map[string]bool{
		"amd64": true,
		"arm64": true,
		"arm":   true,
		"386":   true,
	}

	return validOS[parts[0]] && validArch[parts[1]]
}

// GetConfigPath returns the default configuration file path.
func GetConfigPath() string {
	return filepath.Join(".", ".ko.json")
}

// LoadDefaultConfig loads the configuration from the default path.
func LoadDefaultConfig() (*Config, error) {
	path := GetConfigPath()
	if !FileExists(path) {
		return nil, fmt.Errorf("default configuration file not found")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read default configuration: %w", err)
	}

	return ParseConfig(data)
}
