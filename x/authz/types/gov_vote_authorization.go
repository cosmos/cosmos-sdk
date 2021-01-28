package types

import (
	"reflect"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
)

var (
	_ Authorization = &VoteAuthorization{}
)

// NewVoteAuthorization creates a new VoteAuthorization object.
func NewVoteAuthorization() *VoteAuthorization {
	return &VoteAuthorization{}
}

// MethodName implements Authorization.MethodName.
func (authorization VoteAuthorization) MethodName() string {
	return "/cosmos.gov.v1beta1.Msg/Vote"
}

// Accept implements Authorization.Accept.
func (authorization VoteAuthorization) Accept(msg sdk.ServiceMsg, block tmproto.Header) (allow bool, updated Authorization, delete bool, err error) {
	if reflect.TypeOf(msg.Request) == reflect.TypeOf(&gov.MsgVote{}) {
		_, ok := msg.Request.(*gov.MsgVote)
		if ok {
			return true, &authorization, false, nil
		}
	}
	return false, nil, false, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "type mismatch")
}
