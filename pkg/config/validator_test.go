package config

import (
    "os"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestValidateConfig(t *testing.T) {
    tests := []struct {
        name    string
        path    string
        setup   func()
        wantErr bool
        errMsg  string
    }{
        {
            name: "file does not exist",
            path: "nonexistent.json",
            setup: func() {},
            wantErr: true,
            errMsg: "configuration file not found",
        },
        {
            name: "invalid JSON",
            path: "invalid.json",
            setup: func() {
                os.WriteFile("invalid.json", []byte("{invalid json}"), 0644)
            },
            wantErr: true,
            errMsg: "invalid JSON",
        },
        {
            name: "valid configuration",
            path: "valid.json",
            setup: func() {
                os.WriteFile("valid.json", []byte(`{"baseImage": "valid/image", "images": ["valid/image"]}`), 0644)
            },
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tt.setup()
            err := ValidateConfig(tt.path)
            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                assert.NoError(t, err)
            }
            os.Remove(tt.path)
        })
    }
}

func TestFileExists(t *testing.T) {
    tests := []struct {
        name     string
        path     string
        setup    func()
        expected bool
    }{
        {
            name: "file exists",
            path: "exists.txt",
            setup: func() {
                os.WriteFile("exists.txt", []byte("content"), 0644)
            },
            expected: true,
        },
        {
            name: "file does not exist",
            path: "nonexistent.txt",
            setup: func() {},
            expected: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tt.setup()
            result := FileExists(tt.path)
            assert.Equal(t, tt.expected, result)
            os.Remove(tt.path)
        })
    }
}

func TestParseConfig(t *testing.T) {
    tests := []struct {
        name    string
        data    []byte
        wantErr bool
        errMsg  string
    }{
        {
            name: "empty data",
            data: []byte(""),
            wantErr: true,
            errMsg: "empty configuration data",
        },
        {
            name: "invalid JSON",
            data: []byte("{invalid json}"),
            wantErr: true,
            errMsg: "invalid JSON",
        },
        {
            name: "valid JSON",
            data: []byte(`{"baseImage": "valid/image", "images": ["valid/image"]}`),
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := ParseConfig(tt.data)
            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}

func TestValidateConfigStructure(t *testing.T) {
    tests := []struct {
        name     string
        cfg      *Config
        expected []string
    }{
        {
            name: "nil configuration",
            cfg: nil,
            expected: []string{"configuration is nil"},
        },
        {
            name: "missing baseImage",
            cfg: &Config{Images: []string{"valid/image"}},
            expected: []string{"baseImage is required"},
        },
        {
            name: "valid configuration",
            cfg: &Config{BaseImage: "valid/image", Images: []string{"valid/image"}},
            expected: []string{},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := ValidateConfigStructure(tt.cfg)
            assert.Equal(t, tt.expected, result)
        })
    }
}

func TestCheckRequiredFields(t *testing.T) {
    tests := []struct {
        name     string
        cfg      *Config
        expected []string
    }{
        {
            name: "missing baseImage",
            cfg: &Config{Images: []string{"valid/image"}},
            expected: []string{"baseImage is required"},
        },
        {
            name: "missing images",
            cfg: &Config{BaseImage: "valid/image"},
            expected: []string{"at least one image must be specified"},
        },
        {
            name: "all fields present",
            cfg: &Config{BaseImage: "valid/image", Images: []string{"valid/image"}},
            expected: []string{},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := CheckRequiredFields(tt.cfg)
            assert.Equal(t, tt.expected, result)
        })
    }
}

func TestValidateImageReferences(t *testing.T) {
    tests := []struct {
        name    string
        refs    []string
        wantErr bool
        errMsg  string
    }{
        {
            name: "no image references",
            refs: []string{},
            wantErr: true,
            errMsg: "no image references provided",
        },
        {
            name: "invalid image reference",
            refs: []string{"invalid/image"},
            wantErr: true,
            errMsg: "invalid image references",
        },
        {
            name: "valid image references",
            refs: []string{"valid/image"},
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateImageReferences(tt.refs)
            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}

func TestIsValidImageReference(t *testing.T) {
    tests := []struct {
        name     string
        ref      string
        expected bool
    }{
        {
            name: "empty reference",
            ref: "",
            expected: false,
        },
        {
            name: "invalid reference",
            ref: "invalid/image",
            expected: false,
        },
        {
            name: "valid reference",
            ref: "valid/image",
            expected: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := IsValidImageReference(tt.ref)
            assert.Equal(t, tt.expected, result)
        })
    }
}

func TestIsValidPlatform(t *testing.T) {
    tests := []struct {
        name     string
        platform string
        expected bool
    }{
        {
            name: "empty platform",
            platform: "",
            expected: false,
        },
        {
            name: "invalid platform",
            platform: "invalid/platform",
            expected: false,
        },
        {
            name: "valid platform",
            platform: "linux/amd64",
            expected: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := IsValidPlatform(tt.platform)
            assert.Equal(t, tt.expected, result)
        })
    }
}

func TestGetConfigPath(t *testing.T) {
    expected := "./.ko.json"
    result := GetConfigPath()
    assert.Equal(t, expected, result)
}

func TestLoadDefaultConfig(t *testing.T) {
    tests := []struct {
        name    string
        setup   func()
        wantErr bool
        errMsg  string
    }{
        {
            name: "default config not found",
            setup: func() {},
            wantErr: true,
            errMsg: "default configuration file not found",
        },
        {
            name: "valid default config",
            setup: func() {
                os.WriteFile("./.ko.json", []byte(`{"baseImage": "valid/image", "images": ["valid/image"]}`), 0644)
            },
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tt.setup()
            _, err := LoadDefaultConfig()
            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                assert.NoError(t, err)
            }
            os.Remove("./.ko.json")
        })
    }
}
