package cosmovisor

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

type argsTestSuite struct {
	suite.Suite
}

func TestArgsTestSuite(t *testing.T) {
	suite.Run(t, new(argsTestSuite))
}

func (s *argsTestSuite) TestConfigPaths() {
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

	for _, tc := range cases {
		s.Require().Equal(tc.cfg.Root(), filepath.FromSlash(tc.expectRoot))
		s.Require().Equal(tc.cfg.GenesisBin(), filepath.FromSlash(tc.expectGenesis))
		s.Require().Equal(tc.cfg.UpgradeBin(tc.upgradeName), filepath.FromSlash(tc.expectUpgrade))
	}
}

// Test validate
func (s *argsTestSuite) TestValidate() {
	relPath := filepath.Join("testdata", "validate")
	absPath, err := filepath.Abs(relPath)
	s.Require().NoError(err)

	testdata, err := filepath.Abs("testdata")
	s.Require().NoError(err)

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

	for _, tc := range cases {
		err := tc.cfg.validate()
		if tc.valid {
			s.Require().NoError(err)
		} else {
			s.Require().Error(err)
		}
	}
}

func (s *argsTestSuite) TestEnsureBin() {
	relPath := filepath.Join("testdata", "validate")
	absPath, err := filepath.Abs(relPath)
	s.Require().NoError(err)

	cfg := Config{Home: absPath, Name: "dummyd"}
	s.Require().NoError(cfg.validate())

	s.Require().NoError(EnsureBinary(cfg.GenesisBin()))

	cases := map[string]struct {
		upgrade string
		hasBin  bool
	}{
		"proper":         {"chain2", true},
		"no binary":      {"nobin", false},
		"not executable": {"noexec", false},
		"no directory":   {"foobarbaz", false},
	}

	for _, tc := range cases {
		err := EnsureBinary(cfg.UpgradeBin(tc.upgrade))
		if tc.hasBin {
			s.Require().NoError(err)
		} else {
			s.Require().Error(err)
		}
	}
}

func (s *argsTestSuite) TestBooleanOption() {
	require := s.Require()
	name := "COSMOVISOR_TEST_VAL"

	check := func(def, expected, isErr bool, msg string) {
		v, err := booleanOption(name, def)
		if isErr {
			require.Error(err)
			return
		}
		require.NoError(err)
		require.Equal(expected, v, msg)
	}

	os.Setenv(name, "")
	check(true, true, false, "should correctly set default value")
	check(false, false, false, "should correctly set default value")

	os.Setenv(name, "wrong")
	check(true, true, true, "should error on wrong value")
	os.Setenv(name, "TRUE")
	check(true, true, true, "should error on wrong value")

	os.Setenv(name, "false")
	check(true, false, false, "should handle false value")
	check(false, false, false, "should handle false value")

	os.Setenv(name, "true")
	check(true, true, false, "should handle true value")
	check(false, true, false, "should handle true value")
}
