package slashing

import (
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake"
)

func TestBeginBlocker(t *testing.T) {
	ctx, ck, sk, keeper := createTestInput(t)
	addr, pk, amt := addrs[2], pks[2], sdk.NewInt(100)

	// bond the validator
	got := stake.NewHandler(sk)(ctx, newTestMsgCreateValidator(addr, pk, amt))
	require.True(t, got.IsOK())
	stake.EndBlocker(ctx, sk)
	require.Equal(t, ck.GetCoins(ctx, addr), sdk.Coins{{sk.GetParams(ctx).BondDenom, initCoins.Sub(amt)}})
	require.True(t, sdk.NewRatFromInt(amt).Equal(sk.Validator(ctx, addr).GetPower()))

	val := abci.Validator{
		PubKey: tmtypes.TM2PB.PubKey(pk),
		Power:  amt.Int64(),
	}

	// mark the validator as having signed
	req := abci.RequestBeginBlock{
		Validators: []abci.SigningValidator{{
			Validator:       val,
			SignedLastBlock: true,
		}},
	}
	BeginBlocker(ctx, req, keeper)

	info, found := keeper.getValidatorSigningInfo(ctx, sdk.ValAddress(pk.Address()))
	require.True(t, found)
	require.Equal(t, ctx.BlockHeight(), info.StartHeight)
	require.Equal(t, int64(1), info.IndexOffset)
	require.Equal(t, int64(0), info.JailedUntil)
	require.Equal(t, int64(1), info.SignedBlocksCounter)

	height := int64(0)

	// for 1000 blocks, mark the validator as having signed
	for ; height < SignedBlocksWindow; height++ {
		ctx = ctx.WithBlockHeight(height)
		req = abci.RequestBeginBlock{
			Validators: []abci.SigningValidator{{
				Validator:       val,
				SignedLastBlock: true,
			}},
		}
		BeginBlocker(ctx, req, keeper)
	}

	// for 500 blocks, mark the validator as having not signed
	for ; height < ((SignedBlocksWindow * 2) - MinSignedPerWindow + 1); height++ {
		ctx = ctx.WithBlockHeight(height)
		req = abci.RequestBeginBlock{
			Validators: []abci.SigningValidator{{
				Validator:       val,
				SignedLastBlock: false,
			}},
		}
		BeginBlocker(ctx, req, keeper)
	}

	// validator should be revoked
	validator, found := sk.GetValidatorByPubKey(ctx, pk)
	require.True(t, found)
	require.Equal(t, sdk.Unbonded, validator.GetStatus())
}
