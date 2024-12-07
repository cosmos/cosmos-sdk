package assetdepinject

import (
	"cosmossdk.io/x/accounts/accountstd"
	asset "cosmossdk.io/x/accounts/defaults/asset"
)

func ProvideAccount() accountstd.DepinjectAccount {
	return accountstd.DIAccount(asset.Type, asset.NewAssetAccount)
}
