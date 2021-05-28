package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/core/store"

	"github.com/cosmos/cosmos-sdk/x/authn"

	"github.com/cosmos/cosmos-sdk/core/tx/module"

	abci "github.com/tendermint/tendermint/abci/types"
	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/core/module/app"
)

type handler struct {
	*authn.Module
	cdc        codec.BinaryCodec
	kvStoreKey store.KVStoreKey
}

var _ app.Handler = &handler{}
var _ authn.MsgServer = &handler{}
var _ authn.QueryServer = &handler{}

const (
	CredentialPrefix = 0x0
	AccNumPrefix     = 0x1
	SeqPrefix        = 0x2
	NextAccNumKey    = 0x3
)

func CredentialKey(addrBz []byte) []byte {
	return append([]byte{CredentialPrefix}, addrBz...)
}

func AccNumKey(addrBz []byte) []byte {
	return append([]byte{AccNumPrefix}, addrBz...)
}

func SeqKey(addrBz []byte) []byte {
	return append([]byte{SeqPrefix}, addrBz...)
}

func (m *handler) addrStrToBz(addr string) ([]byte, error) {
	if len(strings.TrimSpace(addr)) == 0 {
		return nil, fmt.Errorf("empty address string is not allowed")
	}

	return types.GetFromBech32(addr, m.Bech32AddressPrefix)
}

func (m *handler) SetCredential(ctx context.Context, request *authn.MsgSetCredentialRequest) (*authn.MsgSetCredentialResponse, error) {
	addrBz, err := m.addrStrToBz(request.Address)
	if err != nil {
		return nil, err
	}

	var cred authn.Credential
	if request.Credential != nil {
		err = m.cdc.UnpackAny(request.Credential, &cred)
		if err != nil {
			return nil, err
		}
	}

	kvStore := m.kvStoreKey.Open(ctx)
	credKey := CredentialKey(addrBz)
	if !kvStore.Has(credKey) {
		if cred != nil {
			if !bytes.Equal(addrBz, cred.Address()) {
				return nil, fmt.Errorf("address must equal credential address when initializing a new account explicitly")
			}
		}

		// initialize new account
	} else {
		if cred == nil {
			return nil, fmt.Errorf("credential cannot be nil when initializing a new account")
		}
		// replace key
	}

	return &authn.MsgSetCredentialResponse{}, nil
}

func (m *handler) Account(ctx context.Context, request *authn.QueryAccountRequest) (*authn.QueryAccountResponse, error) {
	panic("implement me")
}

func (m *handler) Accounts(ctx context.Context, request *authn.QueryAccountsRequest) (*authn.QueryAccountsResponse, error) {
	panic("implement me")
}

func (m *handler) InitGenesis(ctx context.Context, jsonCodec codec.JSONCodec, message json.RawMessage) []abci.ValidatorUpdate {
	panic("implement me")
}

func (m *handler) ExportGenesis(ctx context.Context, jsonCodec codec.JSONCodec) json.RawMessage {
	panic("implement me")
}

func (m *handler) RegisterMsgServices(registrar grpc.ServiceRegistrar) {
	panic("implement me")
}

func (m *handler) RegisterQueryServices(registrar grpc.ServiceRegistrar) {
	panic("implement me")
}

func (m *handler) RegisterTxMiddleware(registrar module.MiddlewareRegistrar) {
	//registrar.RegisterTxMiddlewareFactory(&authn.ValidateMemoMiddleware{}, func(config interface{}) app.TxMiddleware {
	//	return validateMemoMiddlewareHandler{config.(*authn.ValidateMemoMiddleware)}
	//})
}
