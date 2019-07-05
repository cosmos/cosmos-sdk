package types

import (
	"fmt"
	"strings"

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

// HasPermission returns whether the PermAddr contains permission.
func (pa PermAddr) HasPermission(permission string) bool {
	for _, perm := range pa.permissions {
		if perm == permission {
			return true
		}
	}
	return false
}

// GetAddress returns the address of the PermAddr object
func (pa PermAddr) GetAddress() sdk.AccAddress {
	return pa.address
}

// GetPermissions returns the permissions granted to the address
func (pa PermAddr) GetPermissions() []string {
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
