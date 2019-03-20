package slashing

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

func TestBeginBlocker(t *testing.T) {
	ctx, ck, sk, _, keeper := createTestInput(t, DefaultParams())
	power := int64(100)
	amt := sdk.TokensFromTendermintPower(power)
	addr, pk := addrs[2], pks[2]

	// bond the validator
	got := staking.NewHandler(sk)(ctx, NewTestMsgCreateValidator(addr, pk, amt))
	require.True(t, got.IsOK())
	staking.EndBlocker(ctx, sk)
	require.Equal(
		t, ck.GetCoins(ctx, sdk.AccAddress(addr)),
		sdk.Coins{sdk.NewCoin(sk.GetParams(ctx).BondDenom, initCoins.Sub(amt))},
	)
	require.Equal(t, amt, sk.Validator(ctx, addr).GetBondedTokens())

	val := abci.Validator{
		Address: pk.Address(),
		Power:   amt.Int64(),
	}

	// mark the validator as having signed
	req := abci.RequestBeginBlock{
		LastCommitInfo: abci.LastCommitInfo{
			Votes: []abci.VoteInfo{{
				Validator:       val,
				SignedLastBlock: true,
			}},
		},
	}
	BeginBlocker(ctx, req, keeper)

	info, found := keeper.getValidatorSigningInfo(ctx, sdk.ConsAddress(pk.Address()))
	require.True(t, found)
	require.Equal(t, ctx.BlockHeight(), info.StartHeight)
	require.Equal(t, int64(1), info.IndexOffset)
	require.Equal(t, time.Unix(0, 0).UTC(), info.JailedUntil)
	require.Equal(t, int64(0), info.MissedBlocksCounter)

	height := int64(0)

	// for 1000 blocks, mark the validator as having signed
	for ; height < keeper.SignedBlocksWindow(ctx); height++ {
		ctx = ctx.WithBlockHeight(height)
		req = abci.RequestBeginBlock{
			LastCommitInfo: abci.LastCommitInfo{
				Votes: []abci.VoteInfo{{
					Validator:       val,
					SignedLastBlock: true,
				}},
			},
		}
		BeginBlocker(ctx, req, keeper)
	}

	// for 500 blocks, mark the validator as having not signed
	for ; height < ((keeper.SignedBlocksWindow(ctx) * 2) - keeper.MinSignedPerWindow(ctx) + 1); height++ {
		ctx = ctx.WithBlockHeight(height)
		req = abci.RequestBeginBlock{
			LastCommitInfo: abci.LastCommitInfo{
				Votes: []abci.VoteInfo{{
					Validator:       val,
					SignedLastBlock: false,
				}},
			},
		}
		BeginBlocker(ctx, req, keeper)
	}

	// end block
	staking.EndBlocker(ctx, sk)

	// validator should be jailed
	validator, found := sk.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(pk))
	require.True(t, found)
	require.Equal(t, sdk.Unbonding, validator.GetStatus())
}
