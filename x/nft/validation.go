package nft

import (
	fmt "fmt"
	"regexp"
)

var (
	// reClassIDString can be 3 ~ 100 characters long and support letters, followed by either
	// a letter, a number or a slash ('/') or a colon (':').
	reClassIDString = `[a-zA-Z][a-zA-Z0-9/-:]{2,100}`
	reClassID       = regexp.MustCompile(fmt.Sprintf(`^%s$`, reClassIDString))

	// reNFTIDString can be 3 ~ 100 characters long and support letters, followed by either
	// a letter, a number or a slash ('/') or a colon (':').
	reNFTID = reClassID
)

// ValidateClassID returns whether the class id is valid
func ValidateClassID(id string) error {
	if !reClassID.MatchString(id) {
		return fmt.Errorf("invalid class id: %s", id)
	}
	return nil
}

// ValidateNFTID returns whether the nft id is valid
func ValidateNFTID(id string) error {
	if !reNFTID.MatchString(id) {
		return fmt.Errorf("invalid nft id: %s", id)
	}
	return nil
}
