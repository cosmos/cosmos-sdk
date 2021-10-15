package planinfo

import (
	"archive/zip"
	"crypto/sha256"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

type DownloaderTestSuite struct {
	suite.Suite

	// Home is a temporary directory for use in these tests.
	// It will have src/ and dst/ directories.
	// src/ is for things to "download"
	// dst/ is where things get downloaded to.
	Home string
}

func (s *DownloaderTestSuite) SetupTest() {
	s.Home = s.T().TempDir()
	s.Assert().NoError(os.MkdirAll(filepath.Join(s.Home, "src"), 0777), "creating src/ dir")
	s.Assert().NoError(os.MkdirAll(filepath.Join(s.Home, "dst"), 0777), "creating dst/ dir")
	s.T().Logf("Home: [%s]", s.Home)
}

func TestDownloaderTestSuite(t *testing.T) {
	suite.Run(t, new(DownloaderTestSuite))
}

// TestFile represents a file that will be used for a test.
type TestFile struct {
	// Name is the relative path and name of the file.
	Name string
	// Contents is the contents of the file.
	Contents []byte
}

func NewTestFile(name, contents string) *TestFile {
	return &TestFile{
		Name:     name,
		Contents: []byte(contents),
	}
}

// SaveIn saves this TestFile in the given path.
// The full path to the file is returned.
func (f TestFile) SaveIn(path string) (string, error) {
	name := filepath.Join(path, f.Name)
	file, err := os.Create(name)
	if err != nil {
		return name, err
	}
	defer file.Close()
	_, err = file.Write(f.Contents)
	return name, err
}

// TestZip represents a collection of TestFile objects to be zipped into an archive.
type TestZip []*TestFile

func NewTestZip(testFiles ...*TestFile) TestZip {
	tz := make([]*TestFile, len(testFiles))
	for i, tf := range testFiles {
		tz[i] = tf
	}
	return tz
}

// SaveAs saves this TestZip at the given path.
func (z TestZip) SaveAs(path string) error {
	archive, err := os.Create(path)
	if err != nil {
		return err
	}
	defer archive.Close()
	zipper := zip.NewWriter(archive)
	for _, tf := range z {
		zfw, zfwerr := zipper.Create(tf.Name)
		if zfwerr != nil {
			return zfwerr
		}
		_, err = zfw.Write(tf.Contents)
		if err != nil {
			return err
		}
	}
	return zipper.Close()
}

// saveTestZip saves a TestZip in this test's Home/src directory with the given name.
// The full path to the saved archive is returned.
func (s DownloaderTestSuite) saveSrcTestZip(name string, z TestZip) string {
	fullName := filepath.Join(s.Home, "src", name)
	s.Require().NoError(z.SaveAs(fullName), "saving test zip %s", name)
	return fullName
}

// saveSrcTestFile saves a TestFile in this test's Home/src directory.
// The full path to the saved file is returned.
func (s DownloaderTestSuite) saveSrcTestFile(f *TestFile) string {
	path := filepath.Join(s.Home, "src")
	fullName, err := f.SaveIn(path)
	s.Require().NoError(err, "saving test file %s", f.Name)
	return fullName
}

func (s *DownloaderTestSuite) TestDownloadUpgrade() {
	justAFile := NewTestFile("just-a-file", "#!/usr/bin\necho 'I am just a file'\n")
	someFileName := "some-file"
	someFileInBin := NewTestFile("bin" + someFileName, "#!/usr/bin\necho 'I am some file in bin'\n")
	anotherFile := NewTestFile("another-file",  "#!/usr/bin\necho 'I am just another file'\n")
	justAFilePath := s.saveSrcTestFile(justAFile)
	justAFileZip := s.saveSrcTestZip(justAFile.Name + ".zip", NewTestZip(justAFile))
	someFileInBinZip := s.saveSrcTestZip(someFileInBin.Name + ".zip", NewTestZip(someFileInBin))
	allFilesZip := s.saveSrcTestZip(anotherFile.Name + ".zip", NewTestZip(justAFile, someFileInBin, anotherFile))
	getDstDir := func(testName string) string {
		_, tName := filepath.Split(testName)
		return s.Home + "/dst/" + tName
	}
	
	s.T().Run("url does not exist returns error", func(t *testing.T) {
		dstRoot := getDstDir(t.Name())
		url := "file:///never-gonna-be-a-thing"
		err := DownloadUpgrade(dstRoot, url, "nothing")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "could not download reference link")
	})
	
	s.T().Run("url returns single file", func(t *testing.T) {
		dstRoot := getDstDir(t.Name())
		url := "file://" + justAFilePath
		err := DownloadUpgrade(dstRoot, url, justAFile.Name)
	})
	// TODO: Unit tests on DownloadUpgrade(dstRoot, url, daemonName)
	// Tests:
	// * Url does not exist: Error.
	// * Url returns a file: No error.
	// * Url returns an archive with a bin/ dir with the file: No Error.
	// * Url returns an archive with just the expected file: No Error.
	// * Url returns an archive without the expected file: Error.
	// For each that isn't supposed to return an error, make sure:
	// * The expected file exists.
	// * The expected file is executable.
}

func (s *DownloaderTestSuite) TestEnsureBinary() {
	nonExeName := s.saveSrcTestFile(NewTestFile("non-exe.txt", "Not executable"))
	s.Require().NoError(os.Chmod(nonExeName, 0600), "chmod error nonExeName")
	isExeName := s.saveSrcTestFile(NewTestFile("is-exe.sh", "#!/bin/bash\necho 'executing'\n"))
	s.Require().NoError(os.Chmod(isExeName, 0777), "chmod error isExeName")

	s.T().Run("file does not exist gives error", func(t *testing.T) {
		name := filepath.Join(s.Home, "does-not-exist.txt")
		actual := EnsureBinary(name)
		require.Error(t, actual)
	})

	s.T().Run("file is a directory gives error", func(t *testing.T) {
		name := filepath.Join(s.Home, "src")
		actual := EnsureBinary(name)
		require.EqualError(t, actual, fmt.Sprintf("%s is not a regular file", "src"))
	})

	s.T().Run("file exists and becomes executable", func(t *testing.T) {
		name := nonExeName
		actual := EnsureBinary(name)
		require.NoError(t, actual, "EnsureBinary error")
		info, err := os.Stat(name)
		require.NoError(t, err, "stat error")
		actualExec := info.Mode().Perm() & 0001
		require.Equal(t, fs.FileMode(1), actualExec, "permissions")
	})

	s.T().Run("file is already executable no error", func(t *testing.T) {
		name := isExeName
		actual := EnsureBinary(name)
		require.NoError(t, actual, "EnsureBinary error")
		info, err := os.Stat(name)
		require.NoError(t, err, "stat error")
		actualExec := info.Mode().Perm() & 0001
		require.Equal(t, fs.FileMode(1), actualExec, "permissions")
	})
}

func (s *DownloaderTestSuite) TestDownloadPlanInfoFromURL() {
	planContents := `{"binaries":{"xxx/yyy":"url"}}`
	planFile := NewTestFile("plan-info.json", planContents)
	planPath := s.saveSrcTestFile(planFile)
	planChecksum := fmt.Sprintf("%x", sha256.Sum256(planFile.Contents))
	emptyPlanPath := s.saveSrcTestFile(NewTestFile("empty-plan-info.json", ""))

	s.T().Run("url does not exist returns error", func(t *testing.T) {
		url := "file:///never-gonna-be-a-thing"
		_, err := DownloadPlanInfoFromURL(url)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "could not download reference link")
	})

	s.T().Run("without checksum returns content", func(t *testing.T) {
		url := "file://" + planPath
		actual, err := DownloadPlanInfoFromURL(url)
		require.NoError(t, err)
		require.Equal(t, planContents, actual)
	})

	s.T().Run("with correct checksum returns content", func(t *testing.T) {
		url := "file://" + planPath + "?checksum=sha256:" + planChecksum
		actual, err := DownloadPlanInfoFromURL(url)
		require.NoError(t, err)
		require.Equal(t, planContents, actual)
	})

	s.T().Run("with incorrect checksum returns error", func(t *testing.T) {
		badChecksum := "2c22e34510bd1d4ad2343cdc54f7165bccf30caef73f39af7dd1db2795a3da48"
		url := "file://" + planPath + "?checksum=sha256:" + badChecksum
		_, err := DownloadPlanInfoFromURL(url)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Checksums did not match")
		assert.Contains(t, err.Error(), "Expected: " + badChecksum)
		assert.Contains(t, err.Error(), "Got: " +planChecksum)
	})

	s.T().Run("plan is empty returns error", func(t *testing.T) {
		url := "file://" + emptyPlanPath
		_, err := DownloadPlanInfoFromURL(url)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no content returned")
	})
}

