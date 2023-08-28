package keeper_test

import (
	"bytes"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	protoio "github.com/cosmos/gogoproto/io"
	"github.com/cosmos/gogoproto/proto"
	"gotest.tools/v3/assert"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/baseapp"
	ed25519 "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestValidateVoteExtensions(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	// enable vote extensions
	cp := simtestutil.DefaultConsensusParams
	cp.Abci = &cmtproto.ABCIParams{VoteExtensionsEnableHeight: 1}
	f.sdkCtx = f.sdkCtx.WithConsensusParams(*cp).WithBlockHeight(2)

	// setup the validators
	numVals := 3
	privKeys := []cryptotypes.PrivKey{}
	for i := 0; i < numVals; i++ {
		privKeys = append(privKeys, ed25519.GenPrivKey())
	}

	vals := []stakingtypes.Validator{}
	for _, v := range privKeys {
		valAddr := sdk.ValAddress(v.PubKey().Address())
		simtestutil.AddTestAddrsFromPubKeys(f.bankKeeper, f.stakingKeeper, f.sdkCtx, []cryptotypes.PubKey{v.PubKey()}, math.NewInt(100000000000))
		vals = append(vals, testutil.NewValidator(t, valAddr, v.PubKey()))
	}

	votes := []abci.ExtendedVoteInfo{}

	for i, v := range vals {
		v.Tokens = math.NewInt(1000000)
		v.Status = stakingtypes.Bonded
		assert.NilError(t, f.stakingKeeper.SetValidator(f.sdkCtx, v))
		assert.NilError(t, f.stakingKeeper.SetValidatorByConsAddr(f.sdkCtx, v))
		assert.NilError(t, f.stakingKeeper.SetNewValidatorByPowerIndex(f.sdkCtx, v))
		_, err := f.stakingKeeper.Delegate(f.sdkCtx, sdk.AccAddress(privKeys[i].PubKey().Address()), v.Tokens, stakingtypes.Unbonded, v, true)
		assert.NilError(t, err)

		// each val produces a vote
		voteExt := []byte("something" + v.OperatorAddress)
		cve := cmtproto.CanonicalVoteExtension{
			Extension: voteExt,
			Height:    f.sdkCtx.BlockHeight() - 1, // the vote extension was signed in the previous height
			Round:     0,
			ChainId:   "chain-id-123",
		}

		extSignBytes, err := mashalVoteExt(&cve)
		assert.NilError(t, err)

		sig, err := privKeys[i].Sign(extSignBytes)
		assert.NilError(t, err)

		valbz, err := f.stakingKeeper.ValidatorAddressCodec().StringToBytes(v.GetOperator())
		assert.NilError(t, err)
		ve := abci.ExtendedVoteInfo{
			Validator: abci.Validator{
				Address: valbz,
				Power:   1000,
			},
			VoteExtension:      voteExt,
			ExtensionSignature: sig,
			BlockIdFlag:        cmtproto.BlockIDFlagCommit,
		}
		votes = append(votes, ve)
	}

	err := baseapp.ValidateVoteExtensions(f.sdkCtx, f.stakingKeeper, f.sdkCtx.BlockHeight(), "chain-id-123", abci.ExtendedCommitInfo{Round: 0, Votes: votes})
	assert.NilError(t, err)
}

func mashalVoteExt(msg proto.Message) ([]byte, error) {
	var buf bytes.Buffer
	if err := protoio.NewDelimitedWriter(&buf).WriteMsg(msg); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
