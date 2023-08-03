package accounts

import (
	"context"

	"google.golang.org/protobuf/proto"
)

func (a Accounts[H]) Query(ctx context.Context, addr []byte, msg proto.Message) (proto.Message, error) {
	accountImpl, err := a.getAccountImpl(ctx, addr)
	if err != nil {
		return nil, err
	}
	ctx = a.createContext(ctx, nil, addr)
	resp, err := accountImpl.Query(ctx, msg)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
