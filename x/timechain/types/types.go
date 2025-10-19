// SPDX-License-Identifier: MPL-2.0
// Copyright © 2025 Timechain-Arweave-LunCoSim Contributors

package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ValidatorRank struct {
	Address    sdk.ValAddress
	RankScore  sdk.Int
	FixedStake sdk.Coin
}

type Slot struct {
	ID          uint64
	Epoch       uint64
	VDFOutput   []byte
	Proposer    sdk.ValAddress
	Confirmers  []sdk.ValAddress
	PayloadHash []byte
	Confirmed   bool
}

type CrossChainEvent struct {
	SrcChainID string
	DstChainID string
	EventBytes []byte
	TSSSig     []byte
}
