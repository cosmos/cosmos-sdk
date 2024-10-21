package multisigdepinject

import (
	"cosmossdk.io/x/accounts/accountstd"
	"cosmossdk.io/x/accounts/extensions/feegrant"
)

func ProvideAccountExtension() accountstd.DepinjectAccountExtension {
	return accountstd.DIAccountExtension("feegrant", feegrant.NewAccountExtension)
}
