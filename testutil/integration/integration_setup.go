package integration

import (
	"fmt"
	"testing"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"gotest.tools/v3/assert"

	abcitypes "github.com/cometbft/cometbft/abci/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
)

type IntegrationTestApp struct {
	*baseapp.BaseApp
	t                  *testing.T
	InterfaceRegistry  codectypes.InterfaceRegistry
	Ctx                sdk.Context
	QueryServiceHelper *baseapp.QueryServiceTestHelper
}

// func createIntegrationTestRegistry(msgs ...proto.Message) codectypes.InterfaceRegistry {
// 	interfaceRegistry := codectypes.NewInterfaceRegistry()
// 	interfaceRegistry.RegisterInterface("sdk.Msg",
// 		(*sdk.Msg)(nil),
// 		msgs...,
// 	)
// 	interfaceRegistry.RegisterImplementations((*sdk.Msg)(nil), msgs...)
// 	fmt.Println("msgs: ", msgs)
// 	fmt.Println("interface registry: ", interfaceRegistry.ListAllInterfaces())

// 	return interfaceRegistry
// }

func SetupTestApp(t *testing.T, keys map[string]*storetypes.KVStoreKey, modules ...module.AppModuleBasic) *IntegrationTestApp {
	logger := log.NewTestLogger(t)
	db := dbm.NewMemDB()

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	for _, module := range modules {
		module.RegisterInterfaces(interfaceRegistry)
	}

	txConfig := authtx.NewTxConfig(codec.NewProtoCodec(interfaceRegistry), authtx.DefaultSignModes)
	// testStore := storetypes.NewKVStoreKey("test")

	var initChainer sdk.InitChainer = func(ctx sdk.Context, req abcitypes.RequestInitChain) (abcitypes.ResponseInitChain, error) {
		return abcitypes.ResponseInitChain{}, nil
	}

	bApp := baseapp.NewBaseApp(t.Name(), logger, db, txConfig.TxDecoder())
	bApp.MountKVStores(keys)
	bApp.SetInitChainer(initChainer)

	router := baseapp.NewMsgServiceRouter()
	router.SetInterfaceRegistry(interfaceRegistry)
	bApp.SetMsgServiceRouter(router)

	assert.NilError(t, bApp.LoadLatestVersion())
	// testdata.RegisterMsgServer(bApp.MsgServiceRouter(), )

	ctx := bApp.NewContext(true, cmtproto.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, interfaceRegistry)

	return &IntegrationTestApp{
		BaseApp:            bApp,
		t:                  t,
		InterfaceRegistry:  interfaceRegistry,
		Ctx:                ctx,
		QueryServiceHelper: queryHelper,
	}
}

func (app *IntegrationTestApp) ExecMsgs(msgs ...sdk.Msg) ([]*codectypes.Any, error) {
	results := make([]*codectypes.Any, len(msgs))

	for i, msg := range msgs {
		handler := app.MsgServiceRouter().Handler(msg)
		if handler == nil {
			return nil, fmt.Errorf("no message handler found for %q", sdk.MsgTypeURL(msg))
		}
		r, err := handler(app.Ctx, msg)
		if err != nil {
			return nil, errorsmod.Wrapf(err, "message %s at position %d", sdk.MsgTypeURL(msg), i)
		}
		// Handler should always return non-nil sdk.Result.
		if r == nil {
			return nil, fmt.Errorf("got nil sdk.Result for message %q at position %d", msg, i)
		}

		var result *codectypes.Any
		if len(r.MsgResponses) != 0 {
			fmt.Printf("r.MsgResponses[0].Value: %v\n", r.MsgResponses[0].Value)
			msgResponse := r.MsgResponses[0]
			if msgResponse == nil {
				return nil, fmt.Errorf("got nil Msg response for msg %s", msg)
			}
			result = msgResponse
		}
		results[i] = result
	}

	return results, nil
}
