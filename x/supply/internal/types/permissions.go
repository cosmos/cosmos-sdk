package types

import "fmt"

// permissions
const (
	Basic  = "basic"
	Minter = "minter"
	Burner = "burner"
)

// validate the input permissions
func validatePermissions(permissions []string) error {
	for _, perm := range permissions {
		switch perm {
		case Basic, Minter, Burner:
			continue
		default:
			return fmt.Errorf("invalid module permission %s", perm)
		}
	}
	return nil
}
