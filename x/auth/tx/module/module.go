package module

import (
	"github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/container"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
)

var (
	_ app.Provisioner = Module{}
)

func (m Module) Provision(registrar container.Registrar) error {
	return registrar.Provide(func(marshaler codec.ProtoCodecMarshaler) client.TxConfig {
		signModes := m.EnabledSignModes
		if signModes == nil {
			signModes = tx.DefaultSignModes
		}

		return tx.NewTxConfig(marshaler, signModes)
	})
}
