package v2_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/consensus"
	v1 "github.com/cosmos/cosmos-sdk/x/consensus/migrations/v1"
	v2 "github.com/cosmos/cosmos-sdk/x/consensus/migrations/v2"
	"github.com/stretchr/testify/require"
	tmtypes "github.com/tendermint/tendermint/types"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type mockParamStore struct {
	ps tmproto.ConsensusParams
}

func newMockSubspace(consensusParams tmproto.ConsensusParams) mockParamStore {
	return mockParamStore{ps: consensusParams}
}

func (ms mockParamStore) Get(ctx sdk.Context, key []byte, ps interface{}) {
	switch stringKey := string(key); stringKey {
	case string(v1.ParamStoreKeyBlockParams):
		*ps.(*tmproto.BlockParams) = *ms.ps.Block
	case string(v1.ParamStoreKeyValidatorParams):
		*ps.(*tmproto.ValidatorParams) = *ms.ps.Validator
	case string(v1.ParamStoreKeyEvidenceParams):
		*ps.(*tmproto.EvidenceParams) = *ms.ps.Evidence
	default:
		*ps.(*tmproto.VersionParams) = tmproto.VersionParams{App: 0}
	}
}

func TestMigrate(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig(consensus.AppModuleBasic{})
	cdc := encCfg.Codec

	storeKey := sdk.NewKVStoreKey(v2.ModuleName)
	tKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	store := ctx.KVStore(storeKey)

	legacyParamStore := newMockSubspace(tmtypes.DefaultConsensusParams().ToProto())
	require.NoError(t, v2.MigrateStore(ctx, storeKey, cdc, legacyParamStore))

	var res tmproto.ConsensusParams
	bz := store.Get(v2.ParamStoreKeyConsensusParams)
	require.NoError(t, cdc.Unmarshal(bz, &res))
	require.Equal(t, legacyParamStore.ps, res)
}
