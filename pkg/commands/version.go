// Copyright 2019 ko Build Authors All Rights Reserved.
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

package commands

import (
	"fmt"
	"regexp"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
)

// Version is provided by govvv at compile-time
var Version string

// addVersion augments our CLI surface with version.
func addVersion(topLevel *cobra.Command) {
	topLevel.AddCommand(&cobra.Command{
		Use:   "version",
		Short: `Print ko version.`,
		Run: func(_ *cobra.Command, _ []string) {
			v := version()
			if v == "" {
				fmt.Println("could not determine build information")
			} else {
				fmt.Println(v)
			}
		},
	})
}

func version() string {
	if Version == "" {
		i, ok := debug.ReadBuildInfo()
		if !ok {
			return ""
		}
		Version = i.Main.Version
	}
	return Version
}

// formatVersion formats the version string for display
func formatVersion(v string) string {
	if v == "" {
		return "unknown"
	}
	// Clean up version strings that might have build info
	if strings.Contains(v, "+") {
		parts := strings.Split(v, "+")
		return parts[0]
	}
	return v
}

// isValidVersion checks if the version string matches semantic versioning format
func isValidVersion(v string) bool {
	if v == "" || v == "unknown" {
		return false
	}
	// Basic semver regex pattern
	semverPattern := regexp.MustCompile(`^v?\d+\.\d+\.\d+(-[\w\.\-]+)?(\+[\w\.\-]+)?$`)
	return semverPattern.MatchString(v)
}

// getVersionInfo returns formatted version information
func getVersionInfo() string {
	v := version()
	formatted := formatVersion(v)

	var status string
	if isValidVersion(v) {
		status = "valid"
	} else {
		status = "non-standard"
	}

	return fmt.Sprintf("ko version %s (%s)", formatted, status)
}
