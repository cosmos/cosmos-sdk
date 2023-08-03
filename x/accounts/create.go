package accounts

import (
	"context"
	"crypto/sha256"
	"fmt"

	internalaccounts "cosmossdk.io/x/accounts/internal/accounts"
	"google.golang.org/protobuf/proto"
)

func (a Accounts[H]) Create(ctx context.Context, typ string, from []byte, msg proto.Message) ([]byte, proto.Message, error) {
	accountImpl, exists := a.accounts[typ]
	if !exists {
		return nil, nil, fmt.Errorf("unknown account type %s", typ)
	}
	nextAccNum, err := a.GlobalAccountNumber.Next(ctx)
	if err != nil {
		return nil, nil, err
	}

	addr := sha256.Sum256([]byte(fmt.Sprintf("accounts/%d", nextAccNum))) // TODO: use a better address scheme.

	// deploy account
	accCtx := a.createContext(ctx, from, addr[:])
	resp, err := accountImpl.Init(accCtx, msg)
	if err != nil {
		return nil, nil, err
	}

	// if all went fine then save account info
	err = a.AccountsByType.Set(ctx, addr[:], typ)
	if err != nil {
		return nil, nil, err
	}

	return addr[:], resp, nil
}

func (a Accounts[H]) createContext(ctx context.Context, senderAddr []byte, accAddr []byte) context.Context {
	return internalaccounts.MakeContext(ctx, a.storeService, senderAddr, accAddr)
}
