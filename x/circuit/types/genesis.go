package types

import fmt "fmt"

func DefaultGenesisState() *GenesisState {
	return &GenesisState{}
}

func (gs *GenesisState) Validate() error {

	for _, accounts := range gs.AccountPermissions {
		if accounts.Address == "" {
			return fmt.Errorf("invalid account address: %s", accounts.Address)
		}
		if accounts.Permissions == nil {
			return fmt.Errorf("account has empty permissions, account address: %s", accounts.Address)
		}

		if accounts.Permissions.Level != Permissions_LEVEL_ALL_MSGS && accounts.Permissions.Level != Permissions_LEVEL_SOME_MSGS && accounts.Permissions.Level != Permissions_LEVEL_SUPER_ADMIN {
			return fmt.Errorf("invalid permission level account address: %s, permission level: %s", accounts.Address, accounts.Permissions.Level)
		}
	}

	return nil
}
