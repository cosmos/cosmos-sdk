package app

import (
	"context"
	"encoding/json"

	abci "github.com/tendermint/tendermint/abci/types"
	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/core/module/app"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/authn"
)

type Module struct {
	*authn.Module
	Key app.RootModuleKey
}

var _ app.Module = &Module{}
var _ app.HasTxMiddleware = &Module{}

func (m *Module) RegisterTypes(registry types.InterfaceRegistry) {
	panic("implement me")
}

func (m *Module) InitGenesis(context sdk.Context, codec codec.JSONCodec, message json.RawMessage) []abci.ValidatorUpdate {
	panic("implement me")
}

func (m *Module) ExportGenesis(context sdk.Context, codec codec.JSONCodec) json.RawMessage {
	panic("implement me")
}

func (m *Module) RegisterMsgServices(registrar grpc.ServiceRegistrar) {
	panic("implement me")
}

func (m *Module) RegisterQueryServices(registrar grpc.ServiceRegistrar) {
	panic("implement me")
}

func (m *Module) RegisterTxMiddleware(registrar app.TxMiddlewareRegistrar) {
	registrar.RegisterTxMiddlewareFactory(&authn.ValidateMemoMiddleware{}, func(i interface{}) app.TxMiddleware {
		cfg := i.(*authn.ValidateMemoMiddleware)
		return func(ctx context.Context, tx tx.Tx, next app.TxMiddleware) error {
			memoLength := len(tx.Body.Memo)
			if uint64(memoLength) > cfg.MaxMemoCharacters {
				return sdkerrors.Wrapf(sdkerrors.ErrMemoTooLarge,
					"maximum number of characters is %d but received %d characters",
					cfg.MaxMemoCharacters, memoLength,
				)
			}

			return nil
		}
	})
}
