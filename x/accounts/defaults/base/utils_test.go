package base

import (
	"context"
	"testing"

	gogoproto "github.com/cosmos/gogoproto/proto"

	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	"cosmossdk.io/collections"
	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/codec"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/header"
	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/x/accounts/accountstd"
	accountsv1 "cosmossdk.io/x/accounts/v1"
	"cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

// mock statecodec
type mockStateCodec struct {
	codec.Codec
}

var _ codec.Codec = mockStateCodec{}

func (c mockStateCodec) Marshal(m gogoproto.Message) ([]byte, error) {
	// Size() check can catch the typed nil value.
	if m == nil || gogoproto.Size(m) == 0 {
		// return empty bytes instead of nil, because nil has special meaning in places like store.Set
		return []byte{}, nil
	}

	return gogoproto.Marshal(m)
}

func (c mockStateCodec) Unmarshal(bz []byte, ptr gogoproto.Message) error {
	err := gogoproto.Unmarshal(bz, ptr)

	return err
}

// mock address codec
type addressCodec struct{}

func (a addressCodec) StringToBytes(text string) ([]byte, error) { return []byte(text), nil }
func (a addressCodec) BytesToString(bz []byte) (string, error)   { return string(bz), nil }

func newMockContext(t *testing.T) (context.Context, store.KVStoreService) {
	t.Helper()
	return accountstd.NewMockContext(
		0, []byte("mock_base_account"), []byte("sender"), nil,
		func(ctx context.Context, sender []byte, msg transaction.Msg) (transaction.Msg, error) {
			return nil, nil
		}, func(ctx context.Context, req transaction.Msg) (transaction.Msg, error) {
			return &accountsv1.AccountNumberResponse{
				Number: 1,
			}, nil
		},
	)
}

type transactionService struct{}

func (t transactionService) ExecMode(ctx context.Context) transaction.ExecMode {
	return transaction.ExecModeFinalize
}

func makeMockDependencies(storeservice store.KVStoreService) accountstd.Dependencies {
	sb := collections.NewSchemaBuilder(storeservice)

	return accountstd.Dependencies{
		SchemaBuilder:    sb,
		AddressCodec:     addressCodec{},
		LegacyStateCodec: mockStateCodec{},
		Environment: appmodulev2.Environment{
			EventService:       eventService{},
			HeaderService:      headerService{},
			TransactionService: transactionService{},
		},
	}
}

type headerService struct{}

func (h headerService) HeaderInfo(context.Context) header.Info {
	return header.Info{
		ChainID: "test",
	}
}

type eventService struct{}

// EventManager implements event.Service.
func (eventService) EventManager(context.Context) event.Manager {
	return runtime.EventService{Events: runtime.Events{EventManagerI: sdk.NewEventManager()}}
}

var _ signing.SignModeHandler = directHandler{}

type directHandler struct{}

func (s directHandler) Mode() signingv1beta1.SignMode {
	return signingv1beta1.SignMode_SIGN_MODE_DIRECT
}

func (s directHandler) GetSignBytes(_ context.Context, signerData signing.SignerData, txData signing.TxData) ([]byte, error) {
	txDoc := tx.SignDoc{
		BodyBytes:     txData.BodyBytes,
		AuthInfoBytes: txData.AuthInfoBytes,
		ChainId:       signerData.ChainID,
		AccountNumber: signerData.AccountNumber,
	}

	return txDoc.Marshal()
}
