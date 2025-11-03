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

package commands

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

// AddKubeCommands augments our CLI surface with a passthru delete command, and an apply
// command that realizes the promise of ko, as outlined here:
//
//	https://github.com/google/go-containerregistry/issues/80
func AddKubeCommands(topLevel *cobra.Command) {
	addDelete(topLevel)
	addVersion(topLevel)
	addCreate(topLevel)
	addApply(topLevel)
	addResolve(topLevel)
	addBuild(topLevel)
	addRun(topLevel)
}

// check if kubectl is installed
func isKubectlAvailable() bool {
	if _, err := exec.LookPath("kubectl"); err != nil {
		return false
	}
	return true
}

// getKubectlVersion returns the kubectl version string
func getKubectlVersion() (string, error) {
	cmd := exec.Command("kubectl", "version", "--client=true", "--short=true")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get kubectl version: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// validateKubeCommands checks if all required tools for Kubernetes commands are available
func validateKubeCommands() error {
	if !isKubectlAvailable() {
		return fmt.Errorf("kubectl is required but not found in PATH")
	}

	version, err := getKubectlVersion()
	if err != nil {
		return fmt.Errorf("failed to validate kubectl: %w", err)
	}

	if version == "" {
		return fmt.Errorf("kubectl version could not be determined")
	}

	return nil
}

// isCommandAvailable checks if a given command is available in PATH
func isCommandAvailable(cmd string) bool {
	if _, err := exec.LookPath(cmd); err != nil {
		return false
	}
	return true
}

// getRequiredCommands returns a list of commands required for ko's full functionality
func getRequiredCommands() []string {
	return []string{"kubectl", "docker"}
}

// validatePrerequisites checks if all prerequisites for ko are met
func validatePrerequisites() []string {
	var missing []string
	required := getRequiredCommands()

	for _, cmd := range required {
		if !isCommandAvailable(cmd) {
			missing = append(missing, cmd)
		}
	}

	return missing
}
