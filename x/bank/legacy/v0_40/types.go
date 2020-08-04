package v040

// DONTCOVER
// nolint

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	ModuleName = "bank"
)

// DenomUnits represents a struct that describes different
// denominations units of the basic token
type DenomUnits struct {
	Denom    string   `protobuf:"bytes,1,opt,name=denom,proto3" json:"denom,omitempty"`
	Exponent uint32   `protobuf:"varint,2,opt,name=exponent,proto3" json:"exponent,omitempty"`
	Aliases  []string `protobuf:"bytes,3,rep,name=aliases,proto3" json:"aliases,omitempty"`
}

// Metadata represents a struct that describes
// a basic token
type Metadata struct {
	Description string        `protobuf:"bytes,1,opt,name=description,proto3" json:"description,omitempty"`
	DenomUnits  []*DenomUnits `protobuf:"bytes,2,rep,name=denom_units,json=denomUnits,proto3" json:"denom_units,omitempty"`
	Base        string        `protobuf:"bytes,3,opt,name=base,proto3" json:"base,omitempty"`
	Display     string        `protobuf:"bytes,4,opt,name=display,proto3" json:"display,omitempty"`
}

// Send enabled configuration properties for each denomination
type SendEnabled struct {
	Denom   string `protobuf:"bytes,1,opt,name=denom,proto3" json:"denom,omitempty"`
	Enabled bool   `protobuf:"varint,2,opt,name=enabled,proto3" json:"enabled,omitempty"`
}

// Params defines the set of bank parameters.
type Params struct {
	SendEnabled        []*SendEnabled `protobuf:"bytes,1,rep,name=send_enabled,json=sendEnabled,proto3" json:"send_enabled,omitempty" yaml:"send_enabled,omitempty"`
	DefaultSendEnabled bool           `protobuf:"varint,2,opt,name=default_send_enabled,json=defaultSendEnabled,proto3" json:"default_send_enabled,omitempty" yaml:"default_send_enabled,omitempty"`
}

// GenesisState defines the bank module's genesis state.
type GenesisState struct {
	Params        Params     `protobuf:"bytes,1,opt,name=params,proto3,casttype=Params" json:"params"`
	Balances      []Balance  `protobuf:"bytes,2,rep,name=balances,proto3,casttype=Balance" json:"balances"`
	Supply        sdk.Coins  `protobuf:"bytes,3,rep,name=supply,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"supply"`
	DenomMetadata []Metadata `protobuf:"bytes,4,rep,name=denom_metadata,json=denomMetadata,proto3,casttype=Metadata" json:"denom_metadata" yaml:"denom_metadata"`
}

var _ GenesisBalance = (*Balance)(nil)

type (
	GenesisBalance interface {
		GetAddress() sdk.AccAddress
		GetCoins() sdk.Coins
	}

	Balance struct {
		Address sdk.AccAddress `json:"address" yaml:"address"`
		Coins   sdk.Coins      `json:"coins" yaml:"coins"`
	}
)

func (b Balance) GetAddress() sdk.AccAddress {
	return b.Address
}

func (b Balance) GetCoins() sdk.Coins {
	return b.Coins
}
