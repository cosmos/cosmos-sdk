package slashing

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/ed25519"

	sdk "github.com/cosmos/cosmos-sdk/types"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/slashing/internal/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

func TestBeginBlocker(t *testing.T) {
	ctx, ck, sk, _, keeper := slashingkeeper.CreateTestInput(t, DefaultParams())
	power := int64(100)
	amt := sdk.TokensFromConsensusPower(power)
	addr, pk := slashingkeeper.Addrs[2], slashingkeeper.Pks[2]

	// bond the validator
	got := staking.NewHandler(sk)(ctx, slashingkeeper.NewTestMsgCreateValidator(addr, pk, amt))
	require.True(t, got.IsOK())
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

func benchmarkBeginBlocker(b *testing.B, num int) {
	params := DefaultParams()
	params.SignedBlocksWindow = 500
	ctx, _, _, _, keeper, store, _ := slashingkeeper.CreateTestInputStore(b, params)

	validators := []abci.Validator{}
	signInfo := types.ValidatorSigningInfo{}
	pubkey := [32]byte{}
	for i := 0; i < num; i++ {
		rand.Read(pubkey[:])
		edkey := ed25519.PubKeyEd25519(pubkey)
		addr := edkey.Address()
		validators = append(validators, abci.Validator{Address: addr})
		keeper.SetValidatorSigningInfo(ctx, sdk.ConsAddress(addr), signInfo)
		keeper.AddPubkey(ctx, edkey)
		// uncomment if MissedBlockBitArray is used
		/*
			for i := 0; i < 500; i++ {
				keeper.SetValidatorMissedBlockBitArray(ctx, sdk.ConsAddress(addr), int64(i), true)
			}
		*/

		keeper.StoreVoteArray(ctx, sdk.ConsAddress(addr), types.NewVoteArray(int(params.SignedBlocksWindow)))

	}
	store.Commit()
	require.NoError(b, store.LoadVersion(0))
	require.NoError(b, store.LoadVersion(1))
	req := abci.RequestBeginBlock{
		LastCommitInfo: abci.LastCommitInfo{
			Votes: []abci.VoteInfo{},
		},
	}
	op := abci.RequestBeginBlock{
		LastCommitInfo: abci.LastCommitInfo{
			Votes: []abci.VoteInfo{},
		},
	}

	// two requests with opposite votes to trigger changes in missed block array
	for i := range validators {
		req.LastCommitInfo.Votes = append(req.LastCommitInfo.Votes, abci.VoteInfo{Validator: validators[i], SignedLastBlock: true})
		op.LastCommitInfo.Votes = append(op.LastCommitInfo.Votes, abci.VoteInfo{Validator: validators[i]})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx = ctx.WithBlockHeight(int64(i))
		if i%2 == 0 {
			BeginBlocker(ctx, req, keeper)
		} else {
			BeginBlocker(ctx, op, keeper)
		}
		commit := store.Commit()
		// reset iavl tree with latest version
		require.NoError(b, store.LoadVersion(0))
		require.NoError(b, store.LoadVersion(commit.Version))
	}

}

func BenchmarkBeginBlocker100(b *testing.B) {
	benchmarkBeginBlocker(b, 100)
}
