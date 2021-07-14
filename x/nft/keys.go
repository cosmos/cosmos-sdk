package nft

import (
	fmt "fmt"
	"regexp"
)

const (
	// module name
	ModuleName = "nft"

	// StoreKey is the default store key for nft
	StoreKey = ModuleName

	// RouterKey is the message route for nft
	RouterKey = ModuleName
)

var (
	// reClassIDString can be 3 ~ 100 characters long and support letters, followed by either
	// a letter, a number or a separator ('/') or a separator (':') .
	reClassIDString = `[a-zA-Z][a-zA-Z0-9/-:]{2,100}`
	reClassID       = regexp.MustCompile(fmt.Sprintf(`^%s$`, reClassIDString))

	// reNFTIDString can be 3 ~ 100 characters long and support letters, followed by either
	// a letter, a number or a separator ('/') or a separator (':') .
	reNFTIDString = `[a-zA-Z][a-zA-Z0-9/-:]{2,100}`
	reNFTID       = regexp.MustCompile(fmt.Sprintf(`^%s$`, reNFTIDString))
)

func ValidateClassID(id string) error {
	if !reClassID.MatchString(id) {
		return fmt.Errorf("invalid class id: %s", id)
	}
	return nil
}

func ValidateNFTID(id string) error {
	if !reNFTID.MatchString(id) {
		return fmt.Errorf("invalid nft id: %s", id)
	}
	return nil
}
