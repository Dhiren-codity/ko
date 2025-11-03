package commands

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/spf13/cobra"
    "bytes"
    "context"
    "time"
)

func TestVersion(t *testing.T) {
    tests := []struct {
        name    string
        version string
        want    string
    }{
        {"empty version", "", ""},
        {"non-empty version", "v1.0.0", "v1.0.0"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            Version = tt.version
            got := version()
            assert.Equal(t, tt.want, got)
        })
    }
}

func TestAddVersion(t *testing.T) {
    var buf bytes.Buffer
    rootCmd := &cobra.Command{}
    addVersion(rootCmd)

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    tests := []struct {
        name    string
        version string
        want    string
    }{
        {"empty version", "", "could not determine build information\n"},
        {"non-empty version", "v1.0.0", "v1.0.0\n"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            Version = tt.version
            rootCmd.SetOut(&buf)
            rootCmd.SetArgs([]string{"version"})
            err := rootCmd.ExecuteContext(ctx)
            assert.NoError(t, err)
            assert.Equal(t, tt.want, buf.String())
            buf.Reset()
        })
    }
}

func TestVersionConcurrency(t *testing.T) {
    var wg sync.WaitGroup
    results := make(chan string, 10)

    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(val int) {
            defer wg.Done()
            Version = fmt.Sprintf("v1.0.%d", val)
            results <- version()
        }(i)
    }

    wg.Wait()
    close(results)

    count := 0
    for result := range results {
        assert.Contains(t, result, "v1.0.")
        count++
    }
    assert.Equal(t, 10, count)
}