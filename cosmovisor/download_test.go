// +build linux

package cosmovisor_test

import (
	"path/filepath"
	"testing"

	"github.com/cosmos/cosmos-sdk/cosmovisor"
	"github.com/stretchr/testify/suite"
)

type downloadTestSuite struct {
	suite.Suite
}

func TestDownloadTestSuite(t *testing.T) {
	suite.Run(t, new(downloadTestSuite))
}

func (s *upgradeTestSuite) TestGetDownloadURL() {
	// all download tests will fail if we are not on linux...
	ref, err := filepath.Abs(filepath.FromSlash("./testdata/repo/ref_to_chain3-zip_dir.json"))
	s.Require().NoError(err)
	badref, err := filepath.Abs(filepath.FromSlash("./testdata/repo/chain2-zip_bin/autod.zip")) // "./testdata/repo/zip_binary/autod.zip"))
	s.Require().NoError(err)

	cases := map[string]struct {
		info string
		url  string
		err  string
	}{
		"missing": {
			err: "downloading reference link : invalid source string:",
		},
		"follow reference": {
			info: ref,
			url:  "https://github.com/cosmos/cosmos-sdk/raw/master/cosmovisor/testdata/repo/chain3-zip_dir/autod.zip?checksum=sha256:8951f52a0aea8617de0ae459a20daf704c29d259c425e60d520e363df0f166b4",
		},
		"malformated reference target": {
			info: badref,
			err:  "upgrade info doesn't contain binary map",
		},
		"missing link": {
			info: "https://no.such.domain/exists.txt",
			err:  "dial tcp: lookup no.such.domain: no such host",
		},
		"proper binary": {
			info: `{"binaries": {"linux/amd64": "https://foo.bar/", "windows/amd64": "https://something.else"}}`,
			url:  "https://foo.bar/",
		},
		"any architecture not used": {
			info: `{"binaries": {"linux/amd64": "https://foo.bar/", "*": "https://something.else"}}`,
			url:  "https://foo.bar/",
		},
		"any architecture used": {
			info: `{"binaries": {"linux/arm": "https://foo.bar/arm-only", "any": "https://foo.bar/portable"}}`,
			url:  "https://foo.bar/portable",
		},
		"missing binary": {
			info: `{"binaries": {"linux/arm": "https://foo.bar/"}}`,
			err:  "cannot find binary for",
		},
	}

	for name, tc := range cases {
		s.Run(name, func() {
			url, err := cosmovisor.GetBinaryDownloadURL(cosmovisor.Plan{Info: tc.info})
			if tc.err != "" {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(tc.url, url)
			}
		})
	}
}
