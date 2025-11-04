package commands

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestVersion(t *testing.T) {
    tests := []struct {
        name       string
        version    string
        expected   string
    }{
        {
            name:       "Version set",
            version:    "v1.0.0",
            expected:   "v1.0.0",
        },
        {
            name:       "Version not set, build info not available",
            version:    "",
            expected:   "",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            Version = tt.version
            got := version()
            assert.Equal(t, tt.expected, got)
        })
    }
}