package testutil

import (
	
	"reflect"
	"testing"
	"unsafe"

	
	"github.com/stretchr/testify/require"

	runtimev1alpha1 "cosmossdk.io/api/cosmos/app/runtime/v1alpha1"
	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	"cosmossdk.io/core/address"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/mempool"
)



func MakeMinimalConfig() depinject.Config {
	var (
		mempoolOpt            = baseapp.SetMempool(mempool.NewSenderNonceMempool())
		addressCodec          = func() address.Codec { return addresscodec.NewBech32Codec("cosmos") }
		validatorAddressCodec = func() address.ValidatorAddressCodec { return addresscodec.NewBech32Codec("cosmosvaloper") }
		consensusAddressCodec = func() address.ConsensusAddressCodec { return addresscodec.NewBech32Codec("cosmosvalcons") }
	)

	return depinject.Configs(
		depinject.Supply(mempoolOpt, addressCodec, validatorAddressCodec, consensusAddressCodec),
		appconfig.Compose(&appv1alpha1.Config{
			Modules: []*appv1alpha1.ModuleConfig{
				{
					Name: "runtime",
					Config: appconfig.WrapAny(&runtimev1alpha1.Module{
						AppName: "BaseAppApp",
					}),
				},
			},
		}))
}

func TestLoadVersionHelper(t *testing.T, app *baseapp.BaseApp, expectedHeight int64, expectedID storetypes.CommitID) {
	t.Helper()
	lastHeight := app.LastBlockHeight()
	lastID := app.LastCommitID()
	require.Equal(t, expectedHeight, lastHeight)
	require.Equal(t, expectedID, lastID)
}

func GetCheckStateCtx(app *baseapp.BaseApp) sdk.Context {
	v := reflect.ValueOf(app).Elem()
	f := v.FieldByName("checkState")
	rf := reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()
	return rf.MethodByName("Context").Call(nil)[0].Interface().(sdk.Context)
}

func GetFinalizeBlockStateCtx(app *baseapp.BaseApp) sdk.Context {
	v := reflect.ValueOf(app).Elem()
	f := v.FieldByName("finalizeBlockState")
	rf := reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()
	return rf.MethodByName("Context").Call(nil)[0].Interface().(sdk.Context)
}