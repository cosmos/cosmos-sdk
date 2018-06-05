package gov

import (
	"github.com/cosmos/cosmos-sdk/wire"
)

// TODO: Delete?
func RegisterWire(cdc *wire.Codec) {
	// TODO: bring this back ...
	/*
		// TODO include option to always include prefix bytes.
		cdc.RegisterConcrete(SubmitProposalMsg{}, "cosmos-sdk/SubmitProposalMsg", nil)
		cdc.RegisterConcrete(DepositMsg{}, "cosmos-sdk/DepositMsg", nil)
		cdc.RegisterConcrete(VoteMsg{}, "cosmos-sdk/VoteMsg", nil)
	*/
}
