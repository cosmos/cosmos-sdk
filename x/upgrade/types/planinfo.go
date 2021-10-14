package types

import (
	"encoding/json"
	"errors"
	"fmt"
	neturl "net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hashicorp/go-getter"
)

var osArchRx *regexp.Regexp

func init() {
	osArchRx = regexp.MustCompile(`[a-zA-Z0-9]+/[a-zA-Z0-9]+`)
}

// PlanInfo is the special structure that the Plan.Info string can be (as json).
type PlanInfo struct {
	Binaries BinaryDownloadURLMap `json:"binaries"`
}

// BinaryDownloadURLMap is a map of os/architecture stings to a URL where the binary can be downloaded.
type BinaryDownloadURLMap map[string]string

// ParsePlanInfo parses an info string into a map of os/arch strings to URL string.
// If the infoStr is a url, an GET request will be made to it, and its response will be parsed instead.
func ParsePlanInfo(infoStr string) (*PlanInfo, error) {
	infoStr = strings.TrimSpace(infoStr)

	// If it's a url, download it and treat the result as the real info.
	if _, err := neturl.Parse(infoStr); err == nil {
		infoStr, err = getPlanInfoFromURL(infoStr)
		if err != nil {
			return nil, err
		}
	}

	// Now, try to parse it into the expected structure.
	var planInfo PlanInfo
	if err := json.Unmarshal([]byte(infoStr), &planInfo); err != nil {
		return nil, fmt.Errorf("could not parse plan info: %v", err)
	}

	return &planInfo, nil
}

// getPlanInfoFromURL gets the contents of the file at the given URL.
func getPlanInfoFromURL(url string) (string, error) {
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

// ValidateBasic does stateless validation of this PlanInfo.
func (m PlanInfo) ValidateBasic() error {
	return m.Binaries.ValidateBasic()
}

// ValidateBasic does stateless validation of this BinaryDownloadURLMap.
// It validates that:
//  * This has at least one entry.
//  * All entry keys have the format "os/arch" or are "any".
//  * All entry values are valid URLs.
func (m BinaryDownloadURLMap) ValidateBasic() error {
	// Make sure there's at least one.
	if len(m) == 0 {
		return errors.New("no \"binaries\" entries found")
	}

	for key, val := range m {
		if key != "any" && !osArchRx.MatchString(key) {
			return fmt.Errorf("invalid os/arch format in key \"%s\"", key)
		}
		if _, err := neturl.Parse(val); err != nil {
			return fmt.Errorf("invalid url \"%s\" in binaries[%s]: %v", val, key, err)
		}
	}

	return nil
}

// CheckURLs checks that all entries have valid URLs that return data.
// The provided daemonName is the name of the executable file expected in all downloaded directories.
// Warning: This is an expensive process.
// It will make an HTTP GET request to each URL and download the response.
func (m BinaryDownloadURLMap) CheckURLs(daemonName string) error {
	tempDir, err := os.MkdirTemp("", "os-arch-downloads")
	if err != nil {
		return fmt.Errorf("could not create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)
	for osArch, url := range m {
		dstRoot := filepath.Join(tempDir, strings.ReplaceAll(osArch, "/", "-"))
		if err = DownloadUpgrade(dstRoot, url, daemonName); err != nil {
			return fmt.Errorf("error downloading binary for os/arch %s: %v", osArch, err)
		}
	}
	return nil
}

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
	return ensureBinary(target)
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
	err = ensureBinary(dstDirBinFile)
	if err == nil {
		return nil
	}

	// Otherwise, check for a root {daemonName} file and move it to the bin/ directory if found.
	dstDirFile := filepath.Join(dstDir, daemonName)
	err = ensureBinary(dstDirFile)
	if err == nil {
		err = os.Rename(dstDirFile, dstDirBinFile)
		if err != nil {
			return fmt.Errorf("could not move %s to the bin directory: %w", daemonName, err)
		}
		return nil
	}

	return fmt.Errorf("url \"%s\" result does not contain a bin/%s or %s file", url, daemonName, daemonName)
}

// ensureBinary checks that the given file exists as a regular file.
// If it exists, this also sets all the executable bits.
func ensureBinary(path string) error {
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
