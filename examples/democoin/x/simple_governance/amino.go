package gov

import (
	"github.com/tendermint/go-amino"
)

func RegisterWire(cdc *amino.Codec) {
	// TODO include option to always include prefix bytes.
	cdc.RegisterConcrete(SubmitProposalMsg{}, "cosmos-sdk/SubmitProposalMsg", nil)
	cdc.RegisterConcrete(VoteMsg{}, "cosmos-sdk/VoteMsg", nil)
}
