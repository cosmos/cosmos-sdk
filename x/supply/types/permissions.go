package types

import "errors"

// permissions
const (
	Holder = "holder"
	Minter = "minter"
	Burner = "burner"
)

// validate the input permission
func validatePermission(permission string) error {
	switch permission {
	case Holder, Minter, Burner:
		return nil
	default:
		return errors.New("invalid module permission string")
	}
}
