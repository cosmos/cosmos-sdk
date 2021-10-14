package planinfo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-getter"
)

// DownloadUpgrade downloads the given url into the provided directory and ensures it's valid.
// If this returns nil, the download was successful, and {dstRoot}/bin/{daemonName} is a regular executable file.
func DownloadUpgrade(dstRoot, url, daemonName string) error {
	target := filepath.Join(dstRoot, "bin", daemonName)
	// First try to download it as a single file. If there's no error, it's okay and we're done.
	if err := getter.GetFile(target, url); err != nil {
		// File download didn't work, try it as a directory.
		if err = downloadUpgradeAsDir(dstRoot, url, daemonName); err != nil {
			// Out of options, send back the error.
			return err
		}
	}
	return EnsureBinary(target)
}

// downloadUpgradeAsDir tries to download the given url as a directory, saving it as the given dstDir.
// If this returns nil, the download was successful, and {dstDir}/bin/{daemonName} is a regular executable file.
func downloadUpgradeAsDir(dstDir, url, daemonName string) error {
	err := getter.Get(dstDir, url)
	if err != nil {
		return err
	}

	// If bin/{daemonName} exists, we're done.
	dstDirBinFile := filepath.Join(dstDir, "bin", daemonName)
	err = EnsureBinary(dstDirBinFile)
	if err == nil {
		return nil
	}

	// Otherwise, check for a root {daemonName} file and move it to the bin/ directory if found.
	dstDirFile := filepath.Join(dstDir, daemonName)
	err = EnsureBinary(dstDirFile)
	if err == nil {
		err = os.Rename(dstDirFile, dstDirBinFile)
		if err != nil {
			return fmt.Errorf("could not move %s to the bin directory: %w", daemonName, err)
		}
		return nil
	}

	return fmt.Errorf("url \"%s\" result does not contain a bin/%s or %s file", url, daemonName, daemonName)
}

// EnsureBinary checks that the given file exists as a regular file.
// If it exists, this also sets all the executable bits.
func EnsureBinary(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		_, f := filepath.Split(path)
		return fmt.Errorf("%s is not a regular file", f)
	}
	// Make sure all executable bits are set.
	oldMode := info.Mode().Perm()
	newMode := oldMode | 0111
	if oldMode != newMode {
		return os.Chmod(path, newMode)
	}
	return nil
}

// DownloadPlanInfoFromURL gets the contents of the file at the given URL.
func DownloadPlanInfoFromURL(url string) (string, error) {
	tempDir, err := os.MkdirTemp("", "plan-info-reference")
	if err != nil {
		return "", fmt.Errorf("could not create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)
	planInfoPath := filepath.Join(tempDir, "plan-info.json")
	if err = getter.GetFile(planInfoPath, url); err != nil {
		return "", fmt.Errorf("could not download reference link \"%s\": %w", url, err)
	}
	planInfoBz, rerr := os.ReadFile(planInfoPath)
	if rerr != nil {
		return "", fmt.Errorf("could not read downloaded plan info file: %w", rerr)
	}
	planInfoStr := strings.TrimSpace(string(planInfoBz))
	if len(planInfoStr) == 0 {
		return "", fmt.Errorf("no content returned by \"%s\"", url)
	}
	return planInfoStr, nil
}
