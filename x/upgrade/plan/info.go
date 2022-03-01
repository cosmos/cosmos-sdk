package plan

import (
	"encoding/json"
	"errors"
	"fmt"
	neturl "net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cosmos/cosmos-sdk/internal/conv"
)

// Info is the special structure that the Plan.Info string can be (as json).
type Info struct {
	Binaries BinaryDownloadURLMap `json:"binaries"`
}

// BinaryDownloadURLMap is a map of os/architecture stings to a URL where the binary can be downloaded.
type BinaryDownloadURLMap map[string]string

// ParseInfo parses an info string into a map of os/arch strings to URL string.
// If the infoStr is a url, a GET request will be made to it, and its response will be parsed instead.
// If requireChecksum is true and the infoStr is a url, it must have a checksum query parameter.
// If the infoStr is not a url, requireChecksum is ignored.
func ParseInfo(infoStr string, requireChecksum bool) (*Info, error) {
	if len(strings.TrimSpace(infoStr)) == 0 {
		return nil, errors.New("plan info must not be blank")
	}
	infoJSON, err := ResolveInfo(infoStr, requireChecksum)
	if err != nil {
		return nil, err
	}
	return ParseInfoJSON(infoJSON)
}

// ResolveInfo returns the body of a url if the provided string is a url.
// Otherwise, the string is returned with leading and trailing spaces trimmed.
// If requireChecksum is true and the infoStr is a url, it must have a checksum query parameter.
// If the infoStr is not a url, requireChecksum is ignored.
func ResolveInfo(infoStr string, requireChecksum bool) (string, error) {
	infoStr = strings.TrimSpace(infoStr)
	if len(infoStr) == 0 {
		return infoStr, nil
	}
	if _, err := neturl.Parse(infoStr); err == nil {
		if requireChecksum {
			if err = ValidateIsURLWithChecksum(infoStr); err != nil {
				return "", err
			}
		}
		return DownloadURLWithChecksum(infoStr)
	}
	return infoStr, nil
}

// ParseInfoJSON parses an info JSON string into a map of os/arch strings to URL string.
func ParseInfoJSON(infoJSON string) (*Info, error) {
	var planInfo Info
	if err := json.Unmarshal(conv.UnsafeStrToBytes(infoJSON), &planInfo); err != nil {
		return nil, fmt.Errorf("could not parse plan info json: %w", err)
	}
	return &planInfo, nil
}

// ValidateFull does all possible validation of this Info.
// The provided daemonName is the name of the executable file expected in all downloaded directories.
// It checks that:
//  * Binaries.ValidateBasic() doesn't return an error.
//  * requireChecksums == false or Binaries.ValidateAllURLsHaveChecksum() doesn't return an error.
//  * Binaries.CheckURLs(daemonName) doesn't return an error.
// Warning: This is an expensive process. See BinaryDownloadURLMap.CheckURLs for more info.
func (m Info) ValidateFull(daemonName string, requireChecksums bool) error {
	if err := m.Binaries.ValidateBasic(); err != nil {
		return err
	}
	if requireChecksums {
		if err := m.Binaries.ValidateAllURLsHaveChecksum(); err != nil {
			return err
		}
	}
	if err := m.Binaries.CheckURLs(daemonName); err != nil {
		return err
	}
	return nil
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

	osArchRx := regexp.MustCompile(`[a-zA-Z0-9]+/[a-zA-Z0-9]+`)
	for key, val := range m {
		if key != "any" && !osArchRx.MatchString(key) {
			return fmt.Errorf("invalid os/arch format in key \"%s\"", key)
		}
		if _, err := neturl.Parse(val); err != nil {
			return fmt.Errorf("invalid url \"%s\" in binaries[%s]: %w", val, key, err)
		}
	}

	return nil
}

// ValidateAllURLsHaveChecksum makes sure that all values in this are urls that have a checksum query parameter.
func (m BinaryDownloadURLMap) ValidateAllURLsHaveChecksum() error {
	for key, val := range m {
		if err := ValidateIsURLWithChecksum(val); err != nil {
			return fmt.Errorf("invalid url \"%s\" in binaries[%s]: %w", val, key, err)
		}
	}
	return nil
}

// CheckURLs checks that all entries have valid URLs that return expected data.
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
			return fmt.Errorf("error downloading binary for os/arch %s: %w", osArch, err)
		}
	}
	return nil
}
