package keeper

import (
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	gov "github.com/cosmos/cosmos-sdk/x/gov/exported"
)

// dCERT Keeper
type Keeper struct {
	cdc codec.Codec
}

// ensure Msg interface compliance at compile time
var _ gov.Electionator = Keeper{}

// Active - TODO
func (k Keeper) Active() bool {
	return true
}

// VotePortion vote for portion
type VotePortion struct {
	Amount sdk.Dec
	Member sdk.AccAddress
}

// VoteForMemeber -TODO
type VoteForMemeber struct {
	sender sdk.AccAddress
	votes  []VotePortion
}

// Validate - TODO
func (v VoteForMember) Validate() error {
	totalAmount := sdk.ZeroDec()
	for _, vote := range votes {
		totalAmount := totalAmount.Add(vote.Amount)
	}
	if totalAmount.GT(sdk.OneDec()) {
		return errors.New("total amount of votes > 100%")
	}

	return nil
}

// cast a vote
func (k Keeper) Vote(vote []byte) error {

	var vfm VoteForMember
	k.Cdc.MustUnmarshal(vote, &vfm)

	// TODO add entries to store based on votes

}

// AcceptElection - TODO
func (k Keeper) AcceptElection(sdk.AccAddress) {

}

// RegisterRevoker - TODO
func (k Keeper) Revoke(revoker gov.Revoker) {
	switch revoker := revoker.(type) {
	case types.MsgVerifyInvariant:
		return handleMsgVerifyInvariant(ctx, msg, k)

	default:
		errMsg := fmt.Sprintf("unrecognized dcert message type: %T", msg)
		return sdk.ErrUnknownRequest(errMsg).Result()
	}
}

// RegisterHooks - TODO
func (k Keeper) RegisterHooks(ElectionatorHooks) {

}

// QueryWinners - TODO
func (k Keeper) QueryWinners() []sdk.AccAddress {

}

// QueryMetadata - TODO
func (k Keeper) QueryMetadata(sdk.AccAddress) []byte {

}
