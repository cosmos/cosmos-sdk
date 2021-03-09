// DONTCOVER
package v036

import (
	"github.com/cosmos/cosmos-sdk/codec"
	v034auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v034"
)

const (
	ModuleName = "auth"
)

type (
	GenesisState struct {
		Params v034auth.Params `json:"params"`
	}
)

func NewGenesisState(params v034auth.Params) GenesisState {
	return GenesisState{params}
}

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	v034auth.RegisterCrypto(cdc)
}
