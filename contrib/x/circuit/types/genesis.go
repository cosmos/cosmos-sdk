package types

import "fmt"

func DefaultGenesisState() *GenesisState {
	return &GenesisState{}
}

func (gs *GenesisState) Validate() error {
	for _, account := range gs.AccountPermissions {
		if account.Address == "" {
			return fmt.Errorf("invalid account address: %s", account.Address)
		}
		if account.Permissions == nil {
			return fmt.Errorf("account has empty permissions, account address: %s", account.Address)
		}

		if err := CheckPermission(account); err != nil {
			return err
		}
	}

	return nil
}

func CheckPermission(account *GenesisAccountPermissions) error {
	if account.Permissions.Level != Permissions_LEVEL_ALL_MSGS && account.Permissions.Level != Permissions_LEVEL_SOME_MSGS && account.Permissions.Level != Permissions_LEVEL_SUPER_ADMIN {
		return fmt.Errorf("invalid permission level account address: %s, permission level: %s", account.Address, account.Permissions.Level)
	}
	return nil
}
