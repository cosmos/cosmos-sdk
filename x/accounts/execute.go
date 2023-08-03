package accounts

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"
)

func (a Accounts[H]) Execute(ctx context.Context, from []byte, to []byte, msg proto.Message) (proto.Message, error) {
	typ, err := a.AccountsByType.Get(ctx, to)
	if err != nil {
		return nil, err
	}

	accountImpl, exists := a.accounts[typ]
	if !exists {
		return nil, fmt.Errorf("the chain does not support account type %s", typ)
	}

	ctx = a.createContext(ctx, from, to)
	resp, err := accountImpl.Execute(ctx, msg)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
