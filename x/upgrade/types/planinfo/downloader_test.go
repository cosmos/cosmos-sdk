package planinfo

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type DownloaderTestSuite struct {
	suite.Suite
}

func TestDownloaderTestSuite(t *testing.T) {
	suite.Run(t, new(DownloaderTestSuite))
}

func (s DownloaderTestSuite) TestDownloadUpgrade() {
	// TODO: Unit tests on DownloadUpgrade(dstRoot, url, daemonName)
	// Tests:
	// * Url returns a file: No error.
	// * Url returns an archive with a bin/ dir with the file: No Error.
	// * Url returns an archive with a bin/ dir but no file: Error
	// * Url returns an archive with just the expected file: No Error.
	// * Url returns an archive without the expected file: Error.
	// For each that isn't supposed to return an error, make sure:
	// * The expected file exists.
	// * The expected file is executable.
}

func (s DownloaderTestSuite) TestEnsureBinary() {
	// TODO: Unit tests on EnsureBinary(path)
	// Tests:
	// * path doesn't exist: error
	// * path is a directory: error
	// * path is a non-executable file: executable after (no error)
	// * path is an executable file: executable after (no error)
}

func (s DownloaderTestSuite) TestDownloadPlanInfoFromURL() {
	// TODO: Unit tests on DownloadPlanInfoFromURL(url)
	// Tests:
	// * Url doesn't exist: Error.
	// * Url returns nothing?: Error.
	// * Url returns stuff: in returned string.
}

