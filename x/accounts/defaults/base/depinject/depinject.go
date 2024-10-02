package basedepinject

import (
	"cosmossdk.io/depinject"
	"cosmossdk.io/x/accounts/accountstd"
	"cosmossdk.io/x/accounts/defaults/base"
	"cosmossdk.io/x/tx/signing"
)

type Inputs struct {
	depinject.In

	SignHandlersMap *signing.HandlerMap
	Options         []base.Option
}

func ProvideAccount(in Inputs) accountstd.DepinjectAccount {
	return accountstd.DepinjectAccount{MakeAccount: base.NewAccount("base", in.SignHandlersMap, in.Options...)}
}

func ProvideSecp256K1PubKey() base.Option {
	return base.WithSecp256K1PubKey()
}

func ProvideCustomPubkey[T any, PT base.PubKeyG[T]]() base.Option {
	return base.WithPubKey[T, PT]()
}

func ProvideCustomPubKeyAndValidationFunc[T any, PT base.PubKeyG[T]](validateFn func(PT) error) base.Option {
	return base.WithPubKeyWithValidationFunc(validateFn)
}
