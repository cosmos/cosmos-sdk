package types

// ModuleAccount defines an account for modules that holds coins on a pool
type ModuleAccount struct {
	*MockBaseAccount
	name        string   `json:"name" yaml:"name"`              // name of the module
	permissions []string `json:"permissions" yaml"permissions"` // permissions of module account
}

// HasPermission returns whether or not the module account has permission.
func (ma ModuleAccount) HasPermission(permission string) bool {
	for _, perm := range ma.permissions {
		if perm == permission {
			return true
		}
	}
	return false
}

// GetName returns the the name of the holder's module
func (ma ModuleAccount) GetName() string {
	return ma.name
}

// GetPermissions returns permissions granted to the module account
func (ma ModuleAccount) GetPermissions() []string {
	return ma.permissions
}
