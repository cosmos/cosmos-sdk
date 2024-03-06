package keeper_test

import (
	"bytes"
	"sort"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	protoio "github.com/cosmos/gogoproto/io"
	"github.com/cosmos/gogoproto/proto"
	"gotest.tools/v3/assert"

	"cosmossdk.io/core/comet"
	"cosmossdk.io/core/header"
	"cosmossdk.io/math"
	"cosmossdk.io/x/staking/testutil"
	stakingtypes "cosmossdk.io/x/staking/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	ed25519 "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const chainID = "chain-id-123"

// TestValidateVoteExtensions is a unit test function that tests the validation of vote extensions.
// It sets up the necessary fixtures and validators, generates vote extensions for each validator,
// and validates the vote extensions using the baseapp.ValidateVoteExtensions function.
func TestValidateVoteExtensions(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	// enable vote extensions
	cp := simtestutil.DefaultConsensusParams
	cp.Abci = &cmtproto.ABCIParams{VoteExtensionsEnableHeight: 1}
	f.sdkCtx = f.sdkCtx.WithConsensusParams(*cp).WithHeaderInfo(header.Info{Height: 2, ChainID: chainID})

	// setup the validators
	numVals := 1
	privKeys := []cryptotypes.PrivKey{}
	for i := 0; i < numVals; i++ {
		privKeys = append(privKeys, ed25519.GenPrivKey())
	}

	vals := []stakingtypes.Validator{}
	for _, v := range privKeys {
		valAddr := sdk.ValAddress(v.PubKey().Address())
		acc := f.accountKeeper.NewAccountWithAddress(f.sdkCtx, sdk.AccAddress(v.PubKey().Address()))
		f.accountKeeper.SetAccount(f.sdkCtx, acc)
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
			Height:    f.sdkCtx.HeaderInfo().Height - 1, // the vote extension was signed in the previous height
			Round:     0,
			ChainId:   chainID,
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

	eci, ci := extendedCommitToLastCommit(abci.ExtendedCommitInfo{Round: 0, Votes: votes})
	f.sdkCtx = f.sdkCtx.WithCometInfo(ci)

	err := baseapp.ValidateVoteExtensions(f.sdkCtx, f.stakingKeeper, eci)
	assert.NilError(t, err)
}

func mashalVoteExt(msg proto.Message) ([]byte, error) {
	var buf bytes.Buffer
	if err := protoio.NewDelimitedWriter(&buf).WriteMsg(msg); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func extendedCommitToLastCommit(ec abci.ExtendedCommitInfo) (abci.ExtendedCommitInfo, comet.Info) {
	// sort the extended commit info
	sort.Sort(extendedVoteInfos(ec.Votes))

	// convert the extended commit info to last commit info
	lastCommit := comet.CommitInfo{
		Round: ec.Round,
		Votes: make([]comet.VoteInfo, len(ec.Votes)),
	}

	for i, vote := range ec.Votes {
		lastCommit.Votes[i] = comet.VoteInfo{
			Validator: comet.Validator{
				Address: vote.Validator.Address,
				Power:   vote.Validator.Power,
			},
		}
	}

	return ec, comet.Info{
		LastCommit: lastCommit,
	}
}

type extendedVoteInfos []abci.ExtendedVoteInfo

func (v extendedVoteInfos) Len() int {
	return len(v)
}

func (v extendedVoteInfos) Less(i, j int) bool {
	if v[i].Validator.Power == v[j].Validator.Power {
		return bytes.Compare(v[i].Validator.Address, v[j].Validator.Address) == -1
	}
	return v[i].Validator.Power > v[j].Validator.Power
}

func (v extendedVoteInfos) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}
