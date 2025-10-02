package plan

import (
	"archive/zip"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/go-getter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type DownloaderTestSuite struct {
	suite.Suite

	// Home is a temporary directory for use in these tests.
	// It will have a src/ for things to download.
	Home string
}

func (s *DownloaderTestSuite) SetupTest() {
	s.Home = s.T().TempDir()
	s.Assert().NoError(os.MkdirAll(filepath.Join(s.Home, "src"), 0o777), "creating src/ dir")
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
	copy(tz, testFiles)
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

// saveSrcTestZip saves a TestZip in this test's Home/src directory with the given name.
// The full path to the saved archive is returned.
func (s *DownloaderTestSuite) saveSrcTestZip(name string, z TestZip) string {
	fullName := filepath.Join(s.Home, "src", name)
	s.Require().NoError(z.SaveAs(fullName), "saving test zip %s", name)
	return fullName
}

// saveSrcTestFile saves a TestFile in this test's Home/src directory.
// The full path to the saved file is returned.
func (s *DownloaderTestSuite) saveSrcTestFile(f *TestFile) string {
	path := filepath.Join(s.Home, "src")
	fullName, err := f.SaveIn(path)
	s.Require().NoError(err, "saving test file %s", f.Name)
	return fullName
}

// requireFileExistsAndIsExecutable requires that the given file exists and is executable.
func requireFileExistsAndIsExecutable(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	require.NoError(t, err, "stat error")
	perm := info.Mode().Perm()
	// Checks if at least one executable bit is set (user, group, or other)
	isExe := perm&0o111 != 0
	require.True(t, isExe, "is executable: permissions = %s", perm)
}

// requireFileEquals requires that the contents of the file at the given path
// is equal to the contents of the given TestFile.
func requireFileEquals(t *testing.T, path string, tf *TestFile) {
	t.Helper()
	file, err := os.ReadFile(path)
	require.NoError(t, err, "reading file")
	require.Equal(t, string(tf.Contents), string(file), "file contents")
}

// makeFileURL converts the given path to a URL with the correct checksum query parameter.
func makeFileURL(t *testing.T, path string) string {
	t.Helper()
	f, err := os.Open(path)
	require.NoError(t, err, "opening file")
	defer f.Close()
	hasher := sha256.New()
	_, err = io.Copy(hasher, f)
	require.NoError(t, err, "copying file to hasher")
	return fmt.Sprintf("file://%s?checksum=sha256:%x", path, hasher.Sum(nil))
}

func (s *DownloaderTestSuite) TestDownloadUpgrade() {
	justAFile := NewTestFile("just-a-file", "#!/usr/bin\necho 'I am just a file'\n")
	someFileName := "some-file"
	someFileInBin := NewTestFile("bin"+someFileName, "#!/usr/bin\necho 'I am some file in bin'\n")
	anotherFile := NewTestFile("another-file", "#!/usr/bin\necho 'I am just another file'\n")
	justAFilePath := s.saveSrcTestFile(justAFile)
	justAFileZip := s.saveSrcTestZip(justAFile.Name+".zip", NewTestZip(justAFile))
	someFileInBinZip := s.saveSrcTestZip(someFileInBin.Name+".zip", NewTestZip(someFileInBin))
	allFilesZip := s.saveSrcTestZip(anotherFile.Name+".zip", NewTestZip(justAFile, someFileInBin, anotherFile))
	getDstDir := func(testName string) string {
		_, tName := filepath.Split(testName)
		return s.Home + "/dst/" + tName
	}

	s.T().Run("url does not exist", func(t *testing.T) {
		dstRoot := getDstDir(t.Name())
		url := "file:///never/gonna/be/a/thing.zip?checksum=sha256:2c22e34510bd1d4ad2343cdc54f7165bccf30caef73f39af7dd1db2795a3da48"
		err := DownloadUpgrade(dstRoot, url, "nothing")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no such file or directory")
	})

	s.T().Run("url has incorrect checksum", func(t *testing.T) {
		dstRoot := getDstDir(t.Name())
		badChecksum := "2c22e34510bd1d4ad2343cdc54f7165bccf30caef73f39af7dd1db2795a3da48"
		url := "file://" + justAFilePath + "?checksum=sha256:" + badChecksum
		err := DownloadUpgrade(dstRoot, url, justAFile.Name)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Checksums did not match")
		assert.Contains(t, err.Error(), "Expected: "+badChecksum)
	})

	s.T().Run("url returns single file", func(t *testing.T) {
		dstRoot := getDstDir(t.Name())
		url := makeFileURL(t, justAFilePath)
		err := DownloadUpgrade(dstRoot, url, justAFile.Name)
		require.NoError(t, err)
		expectedFile := filepath.Join(dstRoot, "bin", justAFile.Name)
		requireFileExistsAndIsExecutable(t, expectedFile)
		requireFileEquals(t, expectedFile, justAFile)
	})

	s.T().Run("url returns archive with file in bin", func(t *testing.T) {
		dstRoot := getDstDir(t.Name())
		url := makeFileURL(t, someFileInBinZip)
		err := DownloadUpgrade(dstRoot, url, someFileName)
		require.NoError(t, err)
		expectedFile := filepath.Join(dstRoot, "bin", someFileName)
		requireFileExistsAndIsExecutable(t, expectedFile)
		requireFileEquals(t, expectedFile, someFileInBin)
	})

	s.T().Run("url returns archive with just expected file", func(t *testing.T) {
		dstRoot := getDstDir(t.Name())
		url := makeFileURL(t, justAFileZip)
		err := DownloadUpgrade(dstRoot, url, justAFile.Name)
		require.NoError(t, err)
		expectedFile := filepath.Join(dstRoot, "bin", justAFile.Name)
		requireFileExistsAndIsExecutable(t, expectedFile)
		requireFileEquals(t, expectedFile, justAFile)
	})

	s.T().Run("url returns archive without expected file", func(t *testing.T) {
		dstRoot := getDstDir(t.Name())
		url := makeFileURL(t, allFilesZip)
		err := DownloadUpgrade(dstRoot, url, "not-expected")
		require.Error(t, err)
		require.Contains(t, err.Error(), "result does not contain a bin/not-expected or not-expected file")
	})
}

func (s *DownloaderTestSuite) TestDownloadPreUpgradeScript() {
	preUpgradeScript := NewTestFile("preupgrade.sh", "#!/usr/bin/env bash\necho 'pre-upgrade running'\n")
	invalidChecksumScript := NewTestFile("preupgrade-bad.sh", "#!/usr/bin/env bash\necho 'bad checksum'\n")

	scriptPath := s.saveSrcTestFile(preUpgradeScript)
	badScriptPath := s.saveSrcTestFile(invalidChecksumScript)

	getDstDir := func(testName string) string {
		_, tName := filepath.Split(testName)
		return s.Home + "/dst/" + tName
	}

	s.T().Run("url does not exist", func(t *testing.T) {
		dstRoot := getDstDir(t.Name())
		url := "file:///not/a/real/path/preupgrade.sh?checksum=sha256:2c22e34510bd1d4ad2343cdc54f7165bccf30caef73f39af7dd1db2795a3da48"
		err := DownloadPreUpgradeScript(dstRoot, url)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no such file or directory")
	})

	s.T().Run("url has incorrect checksum", func(t *testing.T) {
		dstRoot := getDstDir(t.Name())
		badChecksum := "2c22e34510bd1d4ad2343cdc54f7165bccf30caef73f39af7dd1db2795a3da48"
		url := "file://" + badScriptPath + "?checksum=sha256:" + badChecksum
		err := DownloadPreUpgradeScript(dstRoot, url)
		require.Error(t, err)
		var checksumError *getter.ChecksumError
		assert.True(t, errors.As(err, &checksumError))
	})

	s.T().Run("url returns single valid script file", func(t *testing.T) {
		dstRoot := getDstDir(t.Name())
		url := makeFileURL(t, scriptPath)
		err := DownloadPreUpgradeScript(dstRoot, url)
		require.NoError(t, err)

		target := filepath.Join(dstRoot, "preupgrade.sh")
		requireFileExistsAndIsExecutable(t, target)
		requireFileEquals(t, target, preUpgradeScript)
	})

	s.T().Run("script is downloaded and made executable via EnsurePreUpgrade", func(t *testing.T) {
		dstRoot := getDstDir(t.Name())
		// Create destination root
		require.NoError(t, os.MkdirAll(dstRoot, 0755))

		// Manually write script (simulate partial download)
		target := filepath.Join(dstRoot, "preupgrade.sh")
		err := os.WriteFile(target, []byte(preUpgradeScript.Contents), 0644) // not executable yet
		require.NoError(t, err)

		// Now call EnsurePreUpgrade (which is tested via DownloadPreUpgradeScript flow)
		err = EnsurePreUpgrade(target)
		require.NoError(t, err)

		// Check it's now executable
		info, err := os.Stat(target)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0755), info.Mode().Perm())
	})

}

func (s *DownloaderTestSuite) TestEnsurePreUpgrade() {
	// Setup test files
	nonExeScript := s.saveSrcTestFile(NewTestFile("preupgrade.sh", "#!/bin/bash\necho 'pre-upgrade'\n"))
	s.Require().NoError(os.Chmod(nonExeScript, 0o644), "failed to make non-executable")

	exeScript := s.saveSrcTestFile(NewTestFile("preupgrade.sh", "#!/bin/bash\necho 'already exec'\n"))
	s.Require().NoError(os.Chmod(exeScript, 0o755), "failed to make executable")

	dirPath := filepath.Join(s.Home, "test-dir")
	s.Require().NoError(os.MkdirAll(dirPath, 0o755), "failed to create test dir")

	symlinkPath := filepath.Join(s.Home, "preupgrade-link.sh")
	s.Require().NoError(os.Symlink("non-existent", symlinkPath), "failed to create symlink")

	// Run subtests
	s.T().Run("file does not exist", func(t *testing.T) {
		missing := filepath.Join(s.Home, "preupgrade.sh")
		err := EnsurePreUpgrade(missing)
		require.Error(t, err)
		assert.True(t, os.IsNotExist(err), "expected 'file does not exist' error")
	})

	s.T().Run("path is a directory", func(t *testing.T) {
		err := EnsurePreUpgrade(dirPath)
		require.Error(t, err)
		assert.EqualError(t, err, "test-dir is not a regular file")
	})

	s.T().Run("file exists but not executable, becomes executable", func(t *testing.T) {
		err := EnsurePreUpgrade(nonExeScript)
		require.NoError(t, err)

		requireFileExistsAndIsExecutable(t, nonExeScript)
		info, err := os.Stat(nonExeScript)
		require.NoError(t, err)
		mode := info.Mode().Perm()
		assert.True(t, (mode&0o111) == 0o111, "expected all execute bits to be set")
	})

	s.T().Run("file is already executable, no change", func(t *testing.T) {
		infoBefore, err := os.Stat(exeScript)
		require.NoError(t, err)
		modeBefore := infoBefore.Mode().Perm()

		err = EnsurePreUpgrade(exeScript)
		require.NoError(t, err)

		infoAfter, err := os.Stat(exeScript)
		require.NoError(t, err)
		modeAfter := infoAfter.Mode().Perm()

		assert.Equal(t, modeBefore, modeAfter, "mode should not change if already executable")
		requireFileExistsAndIsExecutable(t, exeScript)
	})
}

func (s *DownloaderTestSuite) TestEnsureBinary() {
	nonExeName := s.saveSrcTestFile(NewTestFile("non-exe.txt", "Not executable"))
	s.Require().NoError(os.Chmod(nonExeName, 0o600), "chmod error nonExeName")
	isExeName := s.saveSrcTestFile(NewTestFile("is-exe.sh", "#!/bin/bash\necho 'executing'\n"))
	s.Require().NoError(os.Chmod(isExeName, 0o777), "chmod error isExeName")

	s.T().Run("file does not exist", func(t *testing.T) {
		name := filepath.Join(s.Home, "does-not-exist.txt")
		actual := EnsureBinary(name)
		require.Error(t, actual)
	})

	s.T().Run("file is a directory", func(t *testing.T) {
		name := filepath.Join(s.Home, "src")
		actual := EnsureBinary(name)
		require.EqualError(t, actual, fmt.Sprintf("%s is not a regular file", "src"))
	})

	s.T().Run("file exists and becomes executable", func(t *testing.T) {
		name := nonExeName
		actual := EnsureBinary(name)
		require.NoError(t, actual, "EnsureBinary error")
		requireFileExistsAndIsExecutable(t, name)
	})

	s.T().Run("file is already executable", func(t *testing.T) {
		name := isExeName
		actual := EnsureBinary(name)
		require.NoError(t, actual, "EnsureBinary error")
		requireFileExistsAndIsExecutable(t, name)
	})
}

func (s *DownloaderTestSuite) TestDownloadURL() {
	planContents := `{"binaries":{"xxx/yyy":"url"}}`
	planFile := NewTestFile("plan-info.json", planContents)
	planPath := s.saveSrcTestFile(planFile)
	planChecksum := fmt.Sprintf("%x", sha256.Sum256(planFile.Contents))
	emptyFile := NewTestFile("empty-plan-info.json", "")
	emptyPlanPath := s.saveSrcTestFile(emptyFile)
	emptyChecksum := fmt.Sprintf("%x", sha256.Sum256(emptyFile.Contents))

	s.T().Run("url does not exist", func(t *testing.T) {
		url := "file:///never-gonna-be-a-thing?checksum=sha256:2c22e34510bd1d4ad2343cdc54f7165bccf30caef73f39af7dd1db2795a3da48"
		_, err := DownloadURL(url)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "could not download url")
	})

	s.T().Run("without checksum", func(t *testing.T) {
		url := "file://" + planPath
		actual, err := DownloadURL(url)
		require.NoError(t, err)
		require.Equal(t, planContents, actual)
	})

	s.T().Run("with correct checksum", func(t *testing.T) {
		url := "file://" + planPath + "?checksum=sha256:" + planChecksum
		actual, err := DownloadURL(url)
		require.NoError(t, err)
		require.Equal(t, planContents, actual)
	})

	s.T().Run("with incorrect checksum", func(t *testing.T) {
		badChecksum := "2c22e34510bd1d4ad2343cdc54f7165bccf30caef73f39af7dd1db2795a3da48"
		url := "file://" + planPath + "?checksum=sha256:" + badChecksum
		_, err := DownloadURL(url)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Checksums did not match")
		assert.Contains(t, err.Error(), "Expected: "+badChecksum)
		assert.Contains(t, err.Error(), "Got: "+planChecksum)
	})

	s.T().Run("plan is empty", func(t *testing.T) {
		url := "file://" + emptyPlanPath + "?checksum=sha256:" + emptyChecksum
		_, err := DownloadURL(url)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no content returned")
	})
}
