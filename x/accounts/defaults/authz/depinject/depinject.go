package depinject

import (
	"cosmossdk.io/x/accounts/accountstd"
	"cosmossdk.io/x/accounts/defaults/authz"
)

func ProvideAllLockupAccounts() []accountstd.DepinjectAccount {
	return []accountstd.DepinjectAccount{
		ProvideAuthzAccount(),
	}
}

func ProvideAuthzAccount() accountstd.DepinjectAccount {
	return accountstd.DIAccount(authz.AUTHZ_ACCOUNT, authz.NewAccount)
}
