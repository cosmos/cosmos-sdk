package baseapp_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	baseapptestutil "github.com/cosmos/cosmos-sdk/baseapp/testutil"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
)

var (
	ParamStoreKey = []byte("paramstore")
)

func defaultLogger() log.Logger {
	if testing.Verbose() {
		return log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "baseapp/test")
	}

	return log.NewNopLogger()
}

type MsgKeyValueImpl struct{}

func (m MsgKeyValueImpl) Set(ctx context.Context, msg *baseapptestutil.MsgKeyValue) (*baseapptestutil.MsgCreateKeyValueResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.KVStore(capKey2).Set(msg.Key, msg.Value)
	return &baseapptestutil.MsgCreateKeyValueResponse{}, nil
}

type paramStore struct {
	db *dbm.MemDB
}

func (ps *paramStore) Set(_ sdk.Context, value *tmproto.ConsensusParams) {
	bz, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}

	ps.db.Set(ParamStoreKey, bz)
}

func (ps *paramStore) Has(_ sdk.Context) bool {
	ok, err := ps.db.Has(ParamStoreKey)
	if err != nil {
		panic(err)
	}

	return ok
}

func (ps paramStore) Get(ctx sdk.Context) (*tmproto.ConsensusParams, error) {
	bz, err := ps.db.Get(ParamStoreKey)
	if err != nil {
		panic(err)
	}

	if len(bz) == 0 {
		return nil, errors.New("params not found")
	}

	var params tmproto.ConsensusParams
	if err := json.Unmarshal(bz, &params); err != nil {
		panic(err)
	}

	return &params, nil
}

func setTxSignature(t *testing.T, builder client.TxBuilder, nonce uint64) {
	privKey := secp256k1.GenPrivKeyFromSecret([]byte("test"))
	pubKey := privKey.PubKey()
	err := builder.SetSignatures(
		signingtypes.SignatureV2{
			PubKey:   pubKey,
			Sequence: nonce,
			Data:     &signingtypes.SingleSignatureData{},
		},
	)
	require.NoError(t, err)
}

func testLoadVersionHelper(t *testing.T, app *baseapp.BaseApp, expectedHeight int64, expectedID storetypes.CommitID) {
	lastHeight := app.LastBlockHeight()
	lastID := app.LastCommitID()
	require.Equal(t, expectedHeight, lastHeight)
	require.Equal(t, expectedID, lastID)
}

func getCheckStateCtx(app *baseapp.BaseApp) sdk.Context {
	v := reflect.ValueOf(app).Elem()
	f := v.FieldByName("checkState")
	rf := reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()
	return rf.MethodByName("Context").Call(nil)[0].Interface().(sdk.Context)
}

func getDeliverStateCtx(app *baseapp.BaseApp) sdk.Context {
	v := reflect.ValueOf(app).Elem()
	f := v.FieldByName("deliverState")
	rf := reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()
	return rf.MethodByName("Context").Call(nil)[0].Interface().(sdk.Context)
}
