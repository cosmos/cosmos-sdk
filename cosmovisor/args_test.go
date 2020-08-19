package cosmovisor

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigPaths(t *testing.T) {
	cases := map[string]struct {
		cfg           Config
		upgradeName   string
		expectRoot    string
		expectGenesis string
		expectUpgrade string
	}{
		"simple": {
			cfg:           Config{Home: "/foo", Name: "myd"},
			upgradeName:   "bar",
			expectRoot:    fmt.Sprintf("/foo/%s", rootName),
			expectGenesis: fmt.Sprintf("/foo/%s/genesis/bin/myd", rootName),
			expectUpgrade: fmt.Sprintf("/foo/%s/upgrades/bar/bin/myd", rootName),
		},
		"handle space": {
			cfg:           Config{Home: "/longer/prefix/", Name: "yourd"},
			upgradeName:   "some spaces",
			expectRoot:    fmt.Sprintf("/longer/prefix/%s", rootName),
			expectGenesis: fmt.Sprintf("/longer/prefix/%s/genesis/bin/yourd", rootName),
			expectUpgrade: "/longer/prefix/cosmovisor/upgrades/some%20spaces/bin/yourd",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.cfg.Root(), filepath.FromSlash(tc.expectRoot))
			assert.Equal(t, tc.cfg.GenesisBin(), filepath.FromSlash(tc.expectGenesis))
			assert.Equal(t, tc.cfg.UpgradeBin(tc.upgradeName), filepath.FromSlash(tc.expectUpgrade))
		})
	}
}

// Test validate
func TestValidate(t *testing.T) {
	relPath := filepath.Join("testdata", "validate")
	absPath, err := filepath.Abs(relPath)
	assert.NoError(t, err)

	testdata, err := filepath.Abs("testdata")
	assert.NoError(t, err)

	cases := map[string]struct {
		cfg   Config
		valid bool
	}{
		"happy": {
			cfg:   Config{Home: absPath, Name: "bind"},
			valid: true,
		},
		"happy with download": {
			cfg:   Config{Home: absPath, Name: "bind", AllowDownloadBinaries: true},
			valid: true,
		},
		"missing home": {
			cfg:   Config{Name: "bind"},
			valid: false,
		},
		"missing name": {
			cfg:   Config{Home: absPath},
			valid: false,
		},
		"relative path": {
			cfg:   Config{Home: relPath, Name: "bind"},
			valid: false,
		},
		"no upgrade manager subdir": {
			cfg:   Config{Home: testdata, Name: "bind"},
			valid: false,
		},
		"no such dir": {
			cfg:   Config{Home: filepath.FromSlash("/no/such/dir"), Name: "bind"},
			valid: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := tc.cfg.validate()
			if tc.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestEnsureBin(t *testing.T) {
	relPath := filepath.Join("testdata", "validate")
	absPath, err := filepath.Abs(relPath)
	assert.NoError(t, err)

	cfg := Config{Home: absPath, Name: "dummyd"}
	assert.NoError(t, cfg.validate())

	err = EnsureBinary(cfg.GenesisBin())
	assert.NoError(t, err)

	cases := map[string]struct {
		upgrade string
		hasBin  bool
	}{
		"proper":         {"chain2", true},
		"no binary":      {"nobin", false},
		"not executable": {"noexec", false},
		"no directory":   {"foobarbaz", false},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := EnsureBinary(cfg.UpgradeBin(tc.upgrade))
			if tc.hasBin {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
