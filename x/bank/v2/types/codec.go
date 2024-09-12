package types

import (
	"cosmossdk.io/core/registry"
	"cosmossdk.io/core/transaction"
)

func RegisterInterfaces(registrar registry.InterfaceRegistrar) {
	registrar.RegisterImplementations((*transaction.Msg)(nil),
		&MsgUpdateParams{},
	)
}
