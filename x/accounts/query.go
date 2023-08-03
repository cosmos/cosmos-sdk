package accounts

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"
)

func (a Accounts[H]) Query(ctx context.Context, addr []byte, msg proto.Message) (proto.Message, error) {
	typ, err := a.AccountsByType.Get(ctx, addr[:])
	if err != nil {
		return nil, err
	}

	accountImpl, exists := a.accounts[typ]
	if !exists {
		return nil, fmt.Errorf("the chain does not support account type %s", typ)
	}

	ctx = a.createContext(ctx, nil, addr)
	resp, err := accountImpl.Query(ctx, msg)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
