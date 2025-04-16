package types

import (
	"fmt"
	"slices"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// permissions
const (
	Minter  = "minter"
	Burner  = "burner"
	Staking = "staking"
)

// PermissionsForAddress defines all the registered permissions for an address
type PermissionsForAddress struct {
	permissions []string
	address     sdk.AccAddress
}

// NewPermissionsForAddress creates a new PermissionsForAddress object
func NewPermissionsForAddress(name string, permissions []string) PermissionsForAddress {
	return PermissionsForAddress{
		permissions: permissions,
		address:     NewModuleAddress(name),
	}
}

// HasPermission returns whether the PermissionsForAddress contains permission.
func (pa PermissionsForAddress) HasPermission(permission string) bool {
	return slices.Contains(pa.permissions, permission)
}

// GetAddress returns the address of the PermissionsForAddress object
func (pa PermissionsForAddress) GetAddress() sdk.AccAddress {
	return pa.address
}

// GetPermissions returns the permissions granted to the address
func (pa PermissionsForAddress) GetPermissions() []string {
	return pa.permissions
}

// performs basic permission validation
func validatePermissions(permissions ...string) error {
	for _, perm := range permissions {
		if strings.TrimSpace(perm) == "" {
			return fmt.Errorf("module permission is empty")
		}
	}
	return nil
}
