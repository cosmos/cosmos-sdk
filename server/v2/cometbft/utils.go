package cometbft

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"cosmossdk.io/core/transaction"

	sdkabci "buf.build/gen/go/tendermint/tendermint/protocolbuffers/go/tendermint/abci"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	gogoproto "github.com/cosmos/gogoproto/proto"
	gogoany "github.com/cosmos/gogoproto/types/any"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/anypb"

	v1beta1 "cosmossdk.io/api/cosmos/base/abci/v1beta1"
	appmanager "cosmossdk.io/core/app"
	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/comet"
	"cosmossdk.io/core/event"
	errorsmod "cosmossdk.io/errors"

	consensus "github.com/cosmos/cosmos-sdk/x/consensus/types"
)

func queryResponse(res transaction.Type) (*abci.ResponseQuery, error) {
	// TODO(kocu): we are tightly coupled go gogoproto here, is this problem?
	bz, err := gogoproto.Marshal(res)
	if err != nil {
		return nil, err
	}

	// TODO: how do I reply? I suppose we need to different replies depending of the query
	return &abci.ResponseQuery{
		Code:  0,
		Log:   "",
		Info:  "",
		Index: 0,
		Key:   []byte{},
		Value: bz,
		// ProofOps:  &cmtcrypto.ProofOps{},
		Height:    0,
		Codespace: "",
	}, nil
}

// responseExecTxResultWithEvents returns an ABCI ExecTxResult object with fields
// filled in from the given error, gas values and events.
func responseExecTxResultWithEvents(err error, gw, gu uint64, events []abci.Event, debug bool) *abci.ExecTxResult {
	space, code, log := errorsmod.ABCIInfo(err, debug)
	return &abci.ExecTxResult{
		Codespace: space,
		Code:      code,
		Log:       log,
		GasWanted: int64(gw),
		GasUsed:   int64(gu),
		Events:    events,
	}
}

// splitABCIQueryPath splits a string path using the delimiter '/'.
//
// e.g. "this/is/funny" becomes []string{"this", "is", "funny"}
func splitABCIQueryPath(requestPath string) (path []string) {
	path = strings.Split(requestPath, "/")

	// first element is empty string
	if len(path) > 0 && path[0] == "" {
		path = path[1:]
	}

	return path
}

func finalizeBlockResponse(
	in *appmanager.BlockResponse,
	cp *cmtproto.ConsensusParams,
	appHash []byte,
	indexSet map[string]struct{},
) (*abci.ResponseFinalizeBlock, error) {
	allEvents := append(in.BeginBlockEvents, in.EndBlockEvents...)

	resp := &abci.ResponseFinalizeBlock{
		Events:                intoABCIEvents(allEvents, indexSet),
		TxResults:             intoABCITxResults(in.TxResults, indexSet),
		ValidatorUpdates:      intoABCIValidatorUpdates(in.ValidatorUpdates),
		AppHash:               appHash,
		ConsensusParamUpdates: cp,
	}
	return resp, nil
}

func intoABCIValidatorUpdates(updates []appmodulev2.ValidatorUpdate) []abci.ValidatorUpdate {
	valsetUpdates := make([]abci.ValidatorUpdate, len(updates))

	for i := range updates {
		valsetUpdates[i] = abci.ValidatorUpdate{
			PubKey: cmtprotocrypto.PublicKey{
				Sum: &cmtprotocrypto.PublicKey_Ed25519{ // by default we set ed25519
					Ed25519: updates[i].PubKey,
				},
			},
			Power: updates[i].Power,
		}

		if updates[i].PubKeyType == "secp256k1" {
			valsetUpdates[i].PubKey = cmtprotocrypto.PublicKey{
				Sum: &cmtprotocrypto.PublicKey_Secp256K1{
					Secp256K1: updates[i].PubKey,
				},
			}
		}
	}

	return valsetUpdates
}

func intoABCITxResults(results []appmanager.TxResult, indexSet map[string]struct{}) []*abci.ExecTxResult {
	res := make([]*abci.ExecTxResult, len(results))
	for i := range results {
		if results[i].Error == nil {
			res[i] = responseExecTxResultWithEvents(
				results[i].Error,
				results[i].GasWanted,
				results[i].GasUsed,
				intoABCIEvents(results[i].Events, indexSet),
				false,
			)
			continue
		}

		// TODO: handle properly once the we decide on the type of TxResult.Resp
	}

	return res
}

func intoABCIEvents(events []event.Event, indexSet map[string]struct{}) []abci.Event {
	indexAll := len(indexSet) == 0
	abciEvents := make([]abci.Event, len(events))
	for i, e := range events {
		abciEvents[i] = abci.Event{
			Type:       e.Type,
			Attributes: make([]abci.EventAttribute, len(e.Attributes)),
		}

		for j, attr := range e.Attributes {
			_, index := indexSet[fmt.Sprintf("%s.%s", e.Type, attr.Key)]
			abciEvents[i].Attributes[j] = abci.EventAttribute{
				Key:   attr.Key,
				Value: attr.Value,
				Index: index || indexAll,
			}
		}
	}
	return abciEvents
}

func intoABCISimulationResponse(txRes appmanager.TxResult, indexSet map[string]struct{}) ([]byte, error) {
	indexAll := len(indexSet) == 0
	abciEvents := make([]*sdkabci.Event, len(txRes.Events))
	for i, e := range txRes.Events {
		abciEvents[i] = &sdkabci.Event{
			Type:       e.Type,
			Attributes: make([]*sdkabci.EventAttribute, len(e.Attributes)),
		}

		for j, attr := range e.Attributes {
			_, index := indexSet[fmt.Sprintf("%s.%s", e.Type, attr.Key)]
			abciEvents[i].Attributes[j] = &sdkabci.EventAttribute{
				Key:   attr.Key,
				Value: attr.Value,
				Index: index || indexAll,
			}
		}
	}

	msgResponses := make([]*anypb.Any, len(txRes.Resp))
	for i, resp := range txRes.Resp {
		// use this hack to maintain the protov2 API here for now
		anyMsg, err := gogoany.NewAnyWithCacheWithValue(resp)
		if err != nil {
			return nil, err
		}
		msgResponses[i] = &anypb.Any{TypeUrl: anyMsg.TypeUrl, Value: anyMsg.Value}
	}

	res := &v1beta1.SimulationResponse{
		GasInfo: &v1beta1.GasInfo{
			GasWanted: txRes.GasWanted,
			GasUsed:   txRes.GasUsed,
		},
		Result: &v1beta1.Result{
			Data:         []byte{},
			Log:          txRes.Error.Error(),
			Events:       abciEvents,
			MsgResponses: msgResponses,
		},
	}

	return protojson.Marshal(res)
}

// ToSDKEvidence takes comet evidence and returns sdk evidence
func ToSDKEvidence(ev []abci.Misbehavior) []*consensus.Evidence {
	evidence := make([]*consensus.Evidence, len(ev))
	for i, e := range ev {
		evidence[i] = &consensus.Evidence{
			EvidenceType:     consensus.MisbehaviorType(e.Type),
			Height:           e.Height,
			Time:             &e.Time,
			TotalVotingPower: e.TotalVotingPower,
			Validator: &consensus.Validator{
				Address: e.Validator.Address,
				Power:   e.Validator.Power,
			},
		}
	}
	return evidence
}

// ToSDKDecidedCommitInfo takes comet commit info and returns sdk commit info
func ToSDKCommitInfo(commit abci.CommitInfo) *consensus.CommitInfo {
	ci := consensus.CommitInfo{
		Round: commit.Round,
	}

	for _, v := range commit.Votes {
		ci.Votes = append(ci.Votes, &consensus.VoteInfo{
			Validator: &consensus.Validator{
				Address: v.Validator.Address,
				Power:   v.Validator.Power,
			},
			BlockIdFlag: consensus.BlockIDFlag(v.BlockIdFlag),
		})
	}
	return &ci
}

// ToSDKExtendedCommitInfo takes comet extended commit info and returns sdk commit info
func ToSDKExtendedCommitInfo(commit abci.ExtendedCommitInfo) comet.CommitInfo {
	ci := comet.CommitInfo{
		Round: commit.Round,
	}

	for _, v := range commit.Votes {
		ci.Votes = append(ci.Votes, comet.VoteInfo{
			Validator: comet.Validator{
				Address: v.Validator.Address,
				Power:   v.Validator.Power,
			},
			BlockIDFlag: comet.BlockIDFlag(v.BlockIdFlag),
		})
	}

	return ci
}

// QueryResult returns a ResponseQuery from an error. It will try to parse ABCI
// info from the error.
func QueryResult(err error, debug bool) *abci.ResponseQuery {
	space, code, log := errorsmod.ABCIInfo(err, debug)
	return &abci.ResponseQuery{
		Codespace: space,
		Code:      code,
		Log:       log,
	}
}

func (c *Consensus[T]) validateFinalizeBlockHeight(req *abci.RequestFinalizeBlock) error {
	if req.Height < 1 {
		return fmt.Errorf("invalid height: %d", req.Height)
	}

	lastBlockHeight, _, err := c.store.StateLatest()
	if err != nil {
		return err
	}

	// expectedHeight holds the expected height to validate
	var expectedHeight uint64
	if lastBlockHeight == 0 && c.cfg.InitialHeight > 1 {
		// In this case, we're validating the first block of the chain, i.e no
		// previous commit. The height we're expecting is the initial height.
		expectedHeight = c.cfg.InitialHeight
	} else {
		// This case can mean two things:
		//
		// - Either there was already a previous commit in the store, in which
		// case we increment the version from there.
		// - Or there was no previous commit, in which case we start at version 1.
		expectedHeight = lastBlockHeight + 1
	}

	if req.Height != int64(expectedHeight) {
		return fmt.Errorf("invalid height: %d; expected: %d", req.Height, expectedHeight)
	}

	return nil
}

// GetConsensusParams makes a query to the consensus module in order to get the latest consensus
// parameters from committed state
func (c *Consensus[T]) GetConsensusParams(ctx context.Context) (*cmtproto.ConsensusParams, error) {
	cs := &cmtproto.ConsensusParams{}
	latestVersion, err := c.store.GetLatestVersion()

	res, err := c.app.Query(ctx, latestVersion, &consensus.QueryParamsRequest{})
	if err != nil {
		return nil, err
	}

	if r, ok := res.(*consensus.QueryParamsResponse); !ok {
		return nil, fmt.Errorf("failed to query consensus params")
	} else {
		// convert our params to cometbft params
		evidenceMaxDuration := r.Params.Evidence.MaxAgeDuration
		cs = &cmtproto.ConsensusParams{
			Block: &cmtproto.BlockParams{
				MaxBytes: r.Params.Block.MaxBytes,
				MaxGas:   r.Params.Block.MaxGas,
			},
			Evidence: &cmtproto.EvidenceParams{
				MaxAgeNumBlocks: r.Params.Evidence.MaxAgeNumBlocks,
				MaxAgeDuration:  evidenceMaxDuration,
			},
			Validator: &cmtproto.ValidatorParams{
				PubKeyTypes: r.Params.Validator.PubKeyTypes,
			},
			Version: &cmtproto.VersionParams{
				App: r.Params.Version.App,
			},
			Abci: &cmtproto.ABCIParams{
				VoteExtensionsEnableHeight: r.Params.Abci.VoteExtensionsEnableHeight,
			},
		}
	}
	return cs, nil
}

func (c *Consensus[T]) GetBlockRetentionHeight(cp *cmtproto.ConsensusParams, commitHeight int64) int64 {
	// pruning is disabled if minRetainBlocks is zero
	if c.cfg.MinRetainBlocks == 0 {
		return 0
	}

	minNonZero := func(x, y int64) int64 {
		switch {
		case x == 0:
			return y

		case y == 0:
			return x

		case x < y:
			return x

		default:
			return y
		}
	}

	// Define retentionHeight as the minimum value that satisfies all non-zero
	// constraints. All blocks below (commitHeight-retentionHeight) are pruned
	// from CometBFT.
	var retentionHeight int64

	// Define the number of blocks needed to protect against misbehaving validators
	// which allows light clients to operate safely. Note, we piggy back of the
	// evidence parameters instead of computing an estimated number of blocks based
	// on the unbonding period and block commitment time as the two should be
	// equivalent.
	if cp.Evidence != nil && cp.Evidence.MaxAgeNumBlocks > 0 {
		retentionHeight = commitHeight - cp.Evidence.MaxAgeNumBlocks
	}

	if c.snapshotManager != nil {
		snapshotRetentionHeights := c.snapshotManager.GetSnapshotBlockRetentionHeights()
		if snapshotRetentionHeights > 0 {
			retentionHeight = minNonZero(retentionHeight, commitHeight-snapshotRetentionHeights)
		}
	}

	v := commitHeight - int64(c.cfg.MinRetainBlocks)
	retentionHeight = minNonZero(retentionHeight, v)

	if retentionHeight <= 0 {
		// prune nothing in the case of a non-positive height
		return 0
	}

	return retentionHeight
}

// checkHalt checks if height or time exceeds halt-height or halt-time respectively.
func (c *Consensus[T]) checkHalt(height int64, time time.Time) error {
	var halt bool
	switch {
	case c.cfg.HaltHeight > 0 && uint64(height) > c.cfg.HaltHeight:
		halt = true

	case c.cfg.HaltTime > 0 && time.Unix() > int64(c.cfg.HaltTime):
		halt = true
	}

	if halt {
		return fmt.Errorf("halt per configuration height %d time %d", c.cfg.HaltHeight, c.cfg.HaltTime)
	}

	return nil
}

// int64ToUint64 converts an int64 to a uint64, returning 0 if the int64 is negative.
func int64ToUint64(i int64) uint64 {
	if i < 0 {
		return 0
	}
	return uint64(i)
}

// uint64ToInt64 converts a uint64 to an int64, returning math.MaxInt64 if the uint64 is too large.
func uint64ToInt64(u uint64) int64 {
	if u > uint64(math.MaxInt64) {
		return math.MaxInt64
	}
	return int64(u)
}
