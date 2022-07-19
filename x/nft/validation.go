package nft

import (
	"fmt"
	"regexp"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	// reClassIDString can be 3 ~ 100 characters long and support letters, followed by either
	// a letter, a number or a slash ('/') or a colon (':') or ('-').
	reClassIDString = `[a-zA-Z][a-zA-Z0-9/:-]{2,100}`
	reClassID       = regexp.MustCompile(fmt.Sprintf(`^%s$`, reClassIDString))

	// reNFTIDString can be 3 ~ 100 characters long and support letters, followed by either
	// a letter, a number or a slash ('/') or a colon (':') or ('-').
	reNFTID = reClassID
)

// ValidateClassID returns whether the class id is valid
func ValidateClassID(id string) error {
	if !reClassID.MatchString(id) {
		return sdkerrors.Wrapf(ErrInvalidClassID, "invalid class id: %s", id)
	}
	return nil
}

// ValidateNFTID returns whether the nft id is valid
func ValidateNFTID(id string) error {
	if !reNFTID.MatchString(id) {
		return sdkerrors.Wrapf(ErrInvalidID, "invalid nft id: %s", id)
	}
	return nil
}
