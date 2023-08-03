package accounts

import (
	"context"
	"fmt"

	internalaccounts "cosmossdk.io/x/accounts/internal/accounts"
	"google.golang.org/protobuf/proto"
)

func (a Accounts[H]) Execute(ctx context.Context, from []byte, to []byte, msg proto.Message) (proto.Message, error) {
	accountImpl, err := a.getAccountImpl(ctx, to)
	if err != nil {
		return nil, err
	}

	ctx = a.createContext(ctx, from, to)
	resp, err := accountImpl.Execute(ctx, msg)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (a Accounts[H]) getAccountImpl(ctx context.Context, addr []byte) (internalaccounts.Implementation, error) {
	typ, err := a.AccountsByType.Get(ctx, addr)
	if err != nil {
		return internalaccounts.Implementation{}, err
	}

	accountImpl, exists := a.accounts[typ]
	if !exists {
		return internalaccounts.Implementation{}, fmt.Errorf("the chain does not support account type %s anymore, please migrate", typ)
	}

	return accountImpl, nil
}
