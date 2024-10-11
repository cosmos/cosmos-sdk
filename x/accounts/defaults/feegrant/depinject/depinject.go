package multisigdepinject

import (
	"cosmossdk.io/x/accounts/accountstd"
	"cosmossdk.io/x/accounts/defaults/feegrant"
)

func ProvideAccount() accountstd.DepinjectAccount {
	return accountstd.DIAccount("feegrant", feegrant.NewAccount)
}
