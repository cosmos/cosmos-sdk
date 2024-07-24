package banktypes

import sdk "github.com/cosmos/cosmos-sdk/types"

type GenesisState struct {
	// params defines all the parameters of the module.
	Params Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
	// balances is an array containing the balances of all the accounts.
	Balances []Balance `protobuf:"bytes,2,rep,name=balances,proto3" json:"balances"`
	// supply represents the total supply. If it is left empty, then supply will be calculated based on the provided
	// balances. Otherwise, it will be used to validate that the sum of the balances equals this amount.
	Supply sdk.Coins `protobuf:"bytes,3,rep,name=supply,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"supply"`
	// denom_metadata defines the metadata of the different coins.
	DenomMetadata []Metadata `protobuf:"bytes,4,rep,name=denom_metadata,json=denomMetadata,proto3" json:"denom_metadata"`
	// send_enabled defines the denoms where send is enabled or disabled.
	SendEnabled []SendEnabled `protobuf:"bytes,5,rep,name=send_enabled,json=sendEnabled,proto3" json:"send_enabled"`
}

// Params defines the parameters for the bank module.
type Params struct {
	// Deprecated: Use of SendEnabled in params is deprecated.
	// For genesis, use the newly added send_enabled field in the genesis object.
	// Storage, lookup, and manipulation of this information is now in the keeper.
	//
	// As of cosmos-sdk 0.47, this only exists for backwards compatibility of genesis files.
	SendEnabled        []*SendEnabled `protobuf:"bytes,1,rep,name=send_enabled,json=sendEnabled,proto3" json:"send_enabled,omitempty"` // Deprecated: Do not use.
	DefaultSendEnabled bool           `protobuf:"varint,2,opt,name=default_send_enabled,json=defaultSendEnabled,proto3" json:"default_send_enabled,omitempty"`
}

// SendEnabled maps coin denom to a send_enabled status (whether a denom is
// sendable).
type SendEnabled struct {
	Denom   string `protobuf:"bytes,1,opt,name=denom,proto3" json:"denom,omitempty"`
	Enabled bool   `protobuf:"varint,2,opt,name=enabled,proto3" json:"enabled,omitempty"`
}

// Metadata represents a struct that describes
// a basic token.
type Metadata struct {
	Description string `protobuf:"bytes,1,opt,name=description,proto3" json:"description,omitempty"`
	// denom_units represents the list of DenomUnit's for a given coin
	DenomUnits []*DenomUnit `protobuf:"bytes,2,rep,name=denom_units,json=denomUnits,proto3" json:"denom_units,omitempty"`
	// base represents the base denom (should be the DenomUnit with exponent = 0).
	Base string `protobuf:"bytes,3,opt,name=base,proto3" json:"base,omitempty"`
	// display indicates the suggested denom that should be
	// displayed in clients.
	Display string `protobuf:"bytes,4,opt,name=display,proto3" json:"display,omitempty"`
	// name defines the name of the token (eg: Cosmos Atom)
	Name string `protobuf:"bytes,5,opt,name=name,proto3" json:"name,omitempty"`
	// symbol is the token symbol usually shown on exchanges (eg: ATOM). This can
	// be the same as the display.
	Symbol string `protobuf:"bytes,6,opt,name=symbol,proto3" json:"symbol,omitempty"`
	// URI to a document (on or off-chain) that contains additional information. Optional.
	URI string `protobuf:"bytes,7,opt,name=uri,proto3" json:"uri,omitempty"`
	// URIHash is a sha256 hash of a document pointed by URI. It's used to verify that
	// the document didn't change. Optional.
	URIHash string `protobuf:"bytes,8,opt,name=uri_hash,json=uriHash,proto3" json:"uri_hash,omitempty"`
}

// DenomUnit represents a struct that describes a given
// denomination unit of the basic token.
type DenomUnit struct {
	// denom represents the string name of the given denom unit (e.g uatom).
	Denom string `protobuf:"bytes,1,opt,name=denom,proto3" json:"denom,omitempty"`
	// exponent represents power of 10 exponent that one must
	// raise the base_denom to in order to equal the given DenomUnit's denom
	// 1 denom = 10^exponent base_denom
	// (e.g. with a base_denom of uatom, one can create a DenomUnit of 'atom' with
	// exponent = 6, thus: 1 atom = 10^6 uatom).
	Exponent uint32 `protobuf:"varint,2,opt,name=exponent,proto3" json:"exponent,omitempty"`
	// aliases is a list of string aliases for the given denom
	Aliases []string `protobuf:"bytes,3,rep,name=aliases,proto3" json:"aliases,omitempty"`
}

type Balance struct {
	// address is the address of the balance holder.
	Address string `json:"address,omitempty"`
	// coins defines the different coins this balance holds.
	Coins sdk.Coins `json:"coins"`
}

// DefaultGenesisState returns a default bank module genesis state.
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(DefaultParams(), []Balance{}, sdk.Coins{}, []Metadata{}, []SendEnabled{})
}

// NewGenesisState creates a new genesis state.
func NewGenesisState(params Params, balances []Balance, supply sdk.Coins, denomMetaData []Metadata, sendEnabled []SendEnabled) *GenesisState {
	rv := &GenesisState{
		Params:        params,
		Balances:      balances,
		Supply:        supply,
		DenomMetadata: denomMetaData,
		SendEnabled:   sendEnabled,
	}
	return rv
}

// DefaultParams is the default parameter configuration for the bank module
func DefaultParams() Params {
	return Params{
		SendEnabled:        nil,
		DefaultSendEnabled: true,
	}
}
