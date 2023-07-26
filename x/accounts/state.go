package accounts

import "cosmossdk.io/collections"

var (
	GlobalAccountNumberPrefix = collections.NewPrefix(0)
	AccountsStatePrefix       = collections.NewPrefix(1)
	AccountsTypePrefix        = collections.NewPrefix(2)
)
