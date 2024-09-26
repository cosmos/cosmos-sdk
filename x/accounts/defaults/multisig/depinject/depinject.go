package multisigdepinject

import (
	"cosmossdk.io/x/accounts/accountstd"
	"cosmossdk.io/x/accounts/defaults/multisig"
)

func ProvideAccount() accountstd.DepinjectAccount {
	return accountstd.DIAccount("multisig", multisig.NewAccount)
}
