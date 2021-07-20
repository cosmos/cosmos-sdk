// Package v034 is used for legacy migration scripts. Actual migration scripts
// for v034 have been removed, but the v039->v042 migration script still
// references types from this file, so we're keeping it for now.
// DONTCOVER
package v034

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
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
