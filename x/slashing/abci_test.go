package slashing

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

func TestBeginBlocker(t *testing.T) {
	ctx, ck, sk, _, keeper := slashingkeeper.CreateTestInput(t, DefaultParams())
	power := int64(100)
	amt := sdk.TokensFromConsensusPower(power)
	addr, pk := slashingkeeper.Addrs[2], slashingkeeper.Pks[2]

	// bond the validator
	res, err := staking.NewHandler(sk)(ctx, slashingkeeper.NewTestMsgCreateValidator(addr, pk, amt))
	require.NoError(t, err)
	require.NotNil(t, res)

	staking.EndBlocker(ctx, sk)
	require.Equal(
		t, ck.GetCoins(ctx, sdk.AccAddress(addr)),
		sdk.NewCoins(sdk.NewCoin(sk.GetParams(ctx).BondDenom, slashingkeeper.InitTokens.Sub(amt))),
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

	info, found := keeper.GetValidatorSigningInfo(ctx, sdk.ConsAddress(pk.Address()))
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
