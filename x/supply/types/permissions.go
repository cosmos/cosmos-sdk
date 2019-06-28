package types

import "fmt"

// permissions
const (
	Basic  = "basic"
	Minter = "minter"
	Burner = "burner"
)

// validate the input permissions
func validatePermissions(permission string) error {
	switch permission {
	case Basic, Minter, Burner:
		return nil
	default:
		return fmt.Errorf("invalid module permission %s", permission)
	}
}
