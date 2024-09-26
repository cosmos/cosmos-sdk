package multisigdepinject

import (
	"cosmossdk.io/depinject"
	"cosmossdk.io/x/accounts/accountstd"
	"cosmossdk.io/x/accounts/defaults/multisig"
)

func init() {
	depinject.Provide(ProvideAccount)
}

func ProvideAccount() accountstd.DepinjectAccount {
	return accountstd.DIAccount("multisig", multisig.NewAccount)
}
