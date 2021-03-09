// DONTCOVER
package v034

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"
)

const (
	ModuleName = "auth"
)

type (
	Params struct {
		MaxMemoCharacters      uint64 `json:"max_memo_characters"`
		TxSigLimit             uint64 `json:"tx_sig_limit"`
		TxSizeCostPerByte      uint64 `json:"tx_size_cost_per_byte"`
		SigVerifyCostED25519   uint64 `json:"sig_verify_cost_ed25519"`
		SigVerifyCostSecp256k1 uint64 `json:"sig_verify_cost_secp256k1"`
	}

	GenesisState struct {
		CollectedFees sdk.Coins `json:"collected_fees"`
		Params        Params    `json:"params"`
	}
)

// PubKeyMultisigThreshold implements a K of N threshold multisig.
// This struct is copy-pasted from:
// https://github.com/tendermint/tendermint/blob/v0.33.9/crypto/multisig/threshold_pubkey.go
type PubKeyMultisigThreshold struct {
	K       uint            `json:"threshold"`
	PubKeys []crypto.PubKey `json:"pubkeys"`
}

func RegisterCrypto(cdc *codec.LegacyAmino) {
	cdc.RegisterInterface((*cryptotypes.PubKey)(nil), nil)
	cdc.RegisterConcrete(&ed25519.PubKey{},
		ed25519.PubKeyName, nil)
	cdc.RegisterConcrete(&secp256k1.PubKey{},
		secp256k1.PubKeyName, nil)
	cdc.RegisterConcrete(&PubKeyMultisigThreshold{},
		kmultisig.PubKeyAminoRoute, nil)

	cdc.RegisterInterface((*cryptotypes.PrivKey)(nil), nil)
	cdc.RegisterConcrete(&ed25519.PrivKey{}, //nolint:staticcheck
		ed25519.PrivKeyName, nil)
	cdc.RegisterConcrete(&secp256k1.PrivKey{},
		secp256k1.PrivKeyName, nil)
}

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	RegisterCrypto(cdc)
}
