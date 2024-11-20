package admindepinject

import (
	"cosmossdk.io/x/accounts/accountstd"
	"cosmossdk.io/x/accounts/defaults/admin"
)

func ProvideAccount() accountstd.DepinjectAccount {
	return accountstd.DIAccount(admin.Type, admin.NewAdmin)
}
