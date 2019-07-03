package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// permissions
const (
	Basic   = "basic"
	Minter  = "minter"
	Burner  = "burner"
	Staking = "staking"
)

// PermAddr defines the permissions for an address
type PermAddr struct {
	permissions []string
	address     sdk.AccAddress
}

// NewPermAddr creates a new PermAddr object
func NewPermAddr(name string, permissions []string) PermAddr {
	return PermAddr{
		permissions: permissions,
		address:     NewModuleAddress(name),
	}
}

// AddPermissions adds the permission to the list of granted permissions.
func (pa *PermAddr) AddPermissions(permissions ...string) {
	pa.permissions = append(pa.permissions, permissions...)
}

// RemovePermission removes the permission from the list of granted permissions
// or returns an error if the permission is has not been granted.
func (pa *PermAddr) RemovePermission(permission string) error {
	for i, perm := range pa.permissions {
		if perm == permission {
			pa.permissions = append(pa.permissions[:i], pa.permissions[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("cannot remove non granted permission %s", permission)
}

// GetAddress returns the address of the PermAddr object
func (pa PermAddr) GetAddress() sdk.AccAddress {
	return pa.address
}

// GetPermissions returns the permissions granted to the address
func (pa PermAddr) GetPermissions() []string {
	return pa.permissions
}

// validate the input permissions
func validatePermissions(permissions []string) error {
	for _, perm := range permissions {
		switch perm {
		case Basic, Minter, Burner, Staking:
			continue
		default:
			return fmt.Errorf("invalid module permission %s", perm)
		}
	}
	return nil
}
