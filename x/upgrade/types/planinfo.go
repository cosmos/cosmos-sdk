package types

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"strings"
)

// BinaryDownloadKeyAny is a special key used in a BinaryDownloadURLMap as a fallback when the desired os/arch key isn't available.
const BinaryDownloadKeyAny = "any"

// PlanInfo is the special structure that the Plan.Info string can be (as json).
type PlanInfo struct {
	Binaries BinaryDownloadURLMap `json:"binaries"`
}

// BinaryDownloadURLMap is a map of os/architecture stings to a URL where the binary can be downloaded.
type BinaryDownloadURLMap map[string]string

// ParsePlanInfo parses an info string into a map of os/arch strings to URL string.
// If the infoStr is a url, an GET request will be made to it, and its response will be parsed instead.
func ParsePlanInfo(infoStr string) (BinaryDownloadURLMap, error) {
	doc := strings.TrimSpace(infoStr)

	// If it's a url, download it and treat the result as the real info.
	if _, err := neturl.Parse(doc); err == nil {
		var newDocBuff bytes.Buffer
		err = doHttpGet(doc, &newDocBuff)
		if err != nil {
			return nil, fmt.Errorf("could not download reference link \"%s\": %v", doc, err)
		}
		doc = newDocBuff.String()
	}

	// Now, try to parse it into the expected structure.
	var planInfo PlanInfo
	if err := json.Unmarshal([]byte(doc), &planInfo); err != nil {
		return nil, fmt.Errorf("could not parse plan info: %v", err)
	}

	return planInfo.Binaries, nil
}

// GetURL looks for a URL for the given OS/ARCH. If not found, looks for a BinaryDownloadKeyAny URL.
// If neither are found, an empty string is returned.
func (m BinaryDownloadURLMap) GetURL(osArch string) string {
	if url, ok := m[osArch]; ok {
		return url
	}
	if url, ok := m[BinaryDownloadKeyAny]; ok {
		return url
	}
	return ""
}

// ValidateBasic validates that this BinaryDownloadURLMap is usable.
// It validates that:
//  * This has at least one entry.
//  * All entry values are valid URLs.
func (m BinaryDownloadURLMap) ValidateBasic() error {
	// Make sure there's at least one.
	if len(m) == 0 {
		return errors.New("no \"binaries\" entries found")
	}

	// Make sure all the values are URLs.
	for key, val := range m {
		if _, err := neturl.Parse(val); err != nil {
			return fmt.Errorf("invalid url \"%s\" in binaries[%s]: %v", val, key, err)
		}
	}

	return nil
}

// ValidateURLsExist checks that all entries have valid URLs that return data.
// Warning: This is an expensive process.
// It will actually make an HTTP request to each URL and download the response (if any).
func (m BinaryDownloadURLMap) ValidateURLsExist() error {
	for osArch, url := range m {
		err := doHttpGet(url, io.Discard)
		if err != nil {
			return fmt.Errorf("error downloading binary for os/arch %s: %v", osArch, err)
		}
	}
	return nil
}

// doHttpGet does an http GET request on the provided addr and copies the response body into the provided dst.
// If an error is encountered, it is returned, otherwise nil is returned.
func doHttpGet(addr string, dst io.Writer) error {
	resp, err := http.Get(addr)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(dst, resp.Body)
	if err != nil {
		return err
	}
	return nil
}
