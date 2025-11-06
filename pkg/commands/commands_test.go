package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatePrerequisites(t *testing.T) {
	tests := []struct {
		name  string
		setup func() []string
		want  []string
	}{
		{
			name: "all commands available",
			setup: func() []string {
				return []string{}
			},
			want: []string{},
		},
		{
			name: "some commands missing",
			setup: func() []string {
				return []string{"docker"}
			},
			want: []string{"docker"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mocking validatePrerequisites
			validatePrerequisites = func() []string {
				return tt.setup()
			}
			got := validatePrerequisites()
			assert.Equal(t, tt.want, got)
		})
	}
}
