package types

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/authz"
)

func (g GenericAuthoriztion) MsgTypeURL() string {
	return g.MsgTypeUrl
}

func (g GenericAuthoriztion) Accept(ctx context.Context, msg sdk.Msg) (authz.AcceptResponse, error) {
	return authz.AcceptResponse{
		Accept: true,
	}, nil
}

func (g GenericAuthoriztion) ValidateBasic() error {
	if g.MsgTypeUrl == "" {
		return errors.New("msg type cannot be empty")
	}
	return nil
}
