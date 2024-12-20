package simulation

import (
	"context"
	"github.com/cosmos/cosmos-sdk/simsx/common"
	"github.com/cosmos/cosmos-sdk/simsx/module"

	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func MsgUpdateParamsFactory() module.SimMsgFactoryFn[*types.MsgUpdateParams] {
	return func(_ context.Context, testData *common.ChainDataSource, reporter common.SimulationReporter) ([]common.SimAccount, *types.MsgUpdateParams) {
		r := testData.Rand()
		params := types.DefaultParams()
		params.MaxMemoCharacters = r.Uint64InRange(1, 1000)
		params.TxSigLimit = r.Uint64InRange(1, 1000)
		params.TxSizeCostPerByte = r.Uint64InRange(1, 1000)
		params.SigVerifyCostED25519 = r.Uint64InRange(1, 1000)
		params.SigVerifyCostSecp256k1 = r.Uint64InRange(1, 1000)

		return nil, &types.MsgUpdateParams{
			Authority: testData.ModuleAccountAddress(reporter, "gov"),
			Params:    params,
		}
	}
}
