package cometbft

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	abci "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v1"
	cryptoenc "github.com/cometbft/cometbft/crypto/encoding"
	protoio "github.com/cosmos/gogoproto/io"
	gogoproto "github.com/cosmos/gogoproto/proto"
	gogoany "github.com/cosmos/gogoproto/types/any"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"

	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/comet"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/server"
	"cosmossdk.io/core/transaction"
	errorsmod "cosmossdk.io/errors" // we aren't using errors/v2 as it doesn't support grpc status codes
	"cosmossdk.io/server/v2/cometbft/handlers"
	"cosmossdk.io/x/consensus/types"

	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func queryResponse(res transaction.Msg, height int64) (*abci.QueryResponse, error) {
	// this is a tied to protobuf due to client responses always being handled in protobuf
	bz, err := gogoproto.Marshal(res)
	if err != nil {
		return nil, err
	}

	return &abci.QueryResponse{
		Value:  bz,
		Height: height,
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
	in *server.BlockResponse,
	cp *cmtproto.ConsensusParams,
	appHash []byte,
	indexSet map[string]struct{},
	cfg *AppTomlConfig,
) (*abci.FinalizeBlockResponse, error) {
	events := make([]abci.Event, 0)

	if !cfg.DisableABCIEvents {
		var err error
		events, err = intoABCIEvents(
			append(in.BeginBlockEvents, in.EndBlockEvents...),
			indexSet,
			cfg.DisableIndexABCIEvents,
		)
		if err != nil {
			return nil, err
		}
	}

	txResults, err := intoABCITxResults(in.TxResults, indexSet, cfg)
	if err != nil {
		return nil, err
	}

	resp := &abci.FinalizeBlockResponse{
		Events:                events,
		TxResults:             txResults,
		ValidatorUpdates:      intoABCIValidatorUpdates(in.ValidatorUpdates),
		AppHash:               appHash,
		ConsensusParamUpdates: cp,
	}

	return resp, nil
}

func intoABCIValidatorUpdates(updates []appmodulev2.ValidatorUpdate) []abci.ValidatorUpdate {
	valsetUpdates := make([]abci.ValidatorUpdate, len(updates))

	for i, v := range updates {
		valsetUpdates[i] = abci.ValidatorUpdate{
			PubKeyBytes: v.PubKey,
			PubKeyType:  v.PubKeyType,
			Power:       v.Power,
		}
	}

	return valsetUpdates
}

func intoABCITxResults(
	results []server.TxResult,
	indexSet map[string]struct{},
	cfg *AppTomlConfig,
) ([]*abci.ExecTxResult, error) {
	res := make([]*abci.ExecTxResult, len(results))
	for i := range results {
		var err error
		events := make([]abci.Event, 0)

		if !cfg.DisableABCIEvents {
			events, err = intoABCIEvents(results[i].Events, indexSet, cfg.DisableIndexABCIEvents)
			if err != nil {
				return nil, err
			}
		}

		res[i] = responseExecTxResultWithEvents(
			results[i].Error,
			results[i].GasWanted,
			results[i].GasUsed,
			events,
			cfg.Trace,
		)
	}

	return res, nil
}

func intoABCIEvents(events []event.Event, indexSet map[string]struct{}, indexNone bool) ([]abci.Event, error) {
	indexAll := len(indexSet) == 0
	abciEvents := make([]abci.Event, len(events))
	for i, e := range events {
		attrs, err := e.Attributes()
		if err != nil {
			return nil, err
		}

		abciEvents[i] = abci.Event{
			Type:       e.Type,
			Attributes: make([]abci.EventAttribute, len(attrs)),
		}

		for j, attr := range attrs {
			_, index := indexSet[fmt.Sprintf("%s.%s", e.Type, attr.Key)]
			abciEvents[i].Attributes[j] = abci.EventAttribute{
				Key:   attr.Key,
				Value: attr.Value,
				Index: !indexNone && (index || indexAll),
			}
		}
	}
	return abciEvents, nil
}

func intoABCISimulationResponse(txRes server.TxResult, indexSet map[string]struct{}, indexNone bool) ([]byte, error) {
	abciEvents, err := intoABCIEvents(txRes.Events, indexSet, indexNone)
	if err != nil {
		return nil, err
	}

	msgResponses := make([]*gogoany.Any, len(txRes.Resp))
	for i, resp := range txRes.Resp {
		// use this hack to maintain the protov2 API here for now
		anyMsg, err := gogoany.NewAnyWithCacheWithValue(resp)
		if err != nil {
			return nil, err
		}
		msgResponses[i] = anyMsg
	}

	errMsg := ""
	if txRes.Error != nil {
		errMsg = txRes.Error.Error()
	}

	res := &sdk.SimulationResponse{
		GasInfo: sdk.GasInfo{
			GasWanted: txRes.GasWanted,
			GasUsed:   txRes.GasUsed,
		},
		Result: &sdk.Result{
			Data:         []byte{},
			Log:          errMsg,
			Events:       abciEvents,
			MsgResponses: msgResponses,
		},
	}

	return gogoproto.Marshal(res)
}

// ToSDKEvidence takes comet evidence and returns sdk evidence
func ToSDKEvidence(ev []abci.Misbehavior) []*comet.Evidence {
	evidence := make([]*comet.Evidence, len(ev))
	for i, e := range ev {
		evidence[i] = &comet.Evidence{
			Type:             comet.MisbehaviorType(e.Type),
			Height:           e.Height,
			Time:             e.Time,
			TotalVotingPower: e.TotalVotingPower,
			Validator: comet.Validator{
				Address: e.Validator.Address,
				Power:   e.Validator.Power,
			},
		}
	}
	return evidence
}

// ToSDKCommitInfo takes comet commit info and returns sdk commit info
func ToSDKCommitInfo(commit abci.CommitInfo) *comet.CommitInfo {
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

// queryResult returns a ResponseQuery from an error. It will try to parse ABCI info from the error.
func queryResult(err error, debug bool) *abci.QueryResponse {
	space, code, log := errorsmod.ABCIInfo(err, debug)
	return &abci.QueryResponse{
		Codespace: space,
		Code:      code,
		Log:       log,
	}
}

func gRPCErrorToSDKError(err error) *abci.QueryResponse {
	toQueryResp := func(sdkErr *errorsmod.Error, err error) *abci.QueryResponse {
		res := &abci.QueryResponse{
			Code:      sdkErr.ABCICode(),
			Codespace: sdkErr.Codespace(),
		}
		type grpcStatus interface{ GRPCStatus() *grpcstatus.Status }
		if grpcErr, ok := err.(grpcStatus); ok {
			res.Log = grpcErr.GRPCStatus().Message()
		} else {
			res.Log = err.Error()
		}
		return res
	}

	status, ok := grpcstatus.FromError(err)
	if !ok {
		return toQueryResp(sdkerrors.ErrInvalidRequest, err)
	}

	switch status.Code() {
	case codes.NotFound:
		return toQueryResp(sdkerrors.ErrKeyNotFound, err)
	case codes.InvalidArgument:
		return toQueryResp(sdkerrors.ErrInvalidRequest, err)
	case codes.FailedPrecondition:
		return toQueryResp(sdkerrors.ErrInvalidRequest, err)
	case codes.Unauthenticated:
		return toQueryResp(sdkerrors.ErrUnauthorized, err)
	default:
		return toQueryResp(sdkerrors.ErrUnknownRequest, err)
	}
}

func (c *consensus[T]) validateFinalizeBlockHeight(req *abci.FinalizeBlockRequest) error {
	if req.Height < 1 {
		return fmt.Errorf("invalid height: %d", req.Height)
	}

	lastBlockHeight, _, err := c.store.StateLatest()
	if err != nil {
		return err
	}

	// expectedHeight holds the expected height to validate
	var expectedHeight uint64
	if lastBlockHeight == 0 && c.initialHeight > 1 {
		// In this case, we're validating the first block of the chain, i.e no
		// previous commit. The height we're expecting is the initial height.
		expectedHeight = c.initialHeight
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
func GetConsensusParams[T transaction.Tx](ctx context.Context, app handlers.AppManager[T]) (*cmtproto.ConsensusParams, error) {
	res, err := app.Query(ctx, 0, &types.QueryParamsRequest{})
	if err != nil {
		return nil, err
	}

	if r, ok := res.(*types.QueryParamsResponse); !ok {
		return nil, errors.New("failed to query consensus params")
	} else {
		// convert our params to cometbft params
		return r.Params, nil
	}
}

func (c *consensus[T]) GetBlockRetentionHeight(cp *cmtproto.ConsensusParams, commitHeight int64) int64 {
	// pruning is disabled if minRetainBlocks is zero
	if c.cfg.AppTomlConfig.MinRetainBlocks == 0 {
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

	v := commitHeight - int64(c.cfg.AppTomlConfig.MinRetainBlocks)
	retentionHeight = minNonZero(retentionHeight, v)

	if retentionHeight <= 0 {
		// prune nothing in the case of a non-positive height
		return 0
	}

	return retentionHeight
}

// checkHalt checks if height or time exceeds halt-height or halt-time respectively.
func (c *consensus[T]) checkHalt(height int64, time time.Time) error {
	var halt bool
	switch {
	case c.cfg.AppTomlConfig.HaltHeight > 0 && uint64(height) >= c.cfg.AppTomlConfig.HaltHeight:
		halt = true

	case c.cfg.AppTomlConfig.HaltTime > 0 && time.Unix() >= int64(c.cfg.AppTomlConfig.HaltTime):
		halt = true
	}

	if halt {
		return fmt.Errorf("halt per configuration height %d time %d", c.cfg.AppTomlConfig.HaltHeight, c.cfg.AppTomlConfig.HaltTime)
	}

	return nil
}

// uint64ToInt64 converts a uint64 to an int64, returning math.MaxInt64 if the uint64 is too large.
func uint64ToInt64(u uint64) int64 {
	if u > uint64(math.MaxInt64) {
		return math.MaxInt64
	}
	return int64(u)
}

// RawTx allows access to the raw bytes of a transaction even if it failed
// to decode.
func RawTx(tx []byte) transaction.Tx {
	return InjectedTx(tx)
}

type InjectedTx []byte

var _ transaction.Tx = InjectedTx{}

func (tx InjectedTx) Bytes() []byte {
	return tx
}

func (tx InjectedTx) Hash() [32]byte {
	return sha256.Sum256(tx)
}

func (tx InjectedTx) GetGasLimit() (uint64, error) {
	return 0, nil
}

func (tx InjectedTx) GetMessages() ([]transaction.Msg, error) {
	return nil, nil
}

func (tx InjectedTx) GetSenders() ([]transaction.Identity, error) {
	return [][]byte{[]byte("cometbft")}, nil
}

// ValidateVoteExtensions defines a helper function for verifying vote extension
// signatures that may be passed or manually injected into a block proposal from
// a proposer in PrepareProposal. It returns an error if any signature is invalid
// or if unexpected vote extensions and/or signatures are found or less than 2/3
// power is received.
// If commitInfo is nil, this function can be used to check a set of vote extensions
// without comparing them to a commit.
func ValidateVoteExtensions[T transaction.Tx](
	ctx context.Context,
	app handlers.AppManager[T],
	chainID string,
	validatorStore func(context.Context, []byte) (cryptotypes.PubKey, error),
	extCommit abci.ExtendedCommitInfo,
	currentHeight int64,
	commitInfo *abci.CommitInfo,
) error {
	cp, err := GetConsensusParams(ctx, app)
	if err != nil {
		return err
	}

	if commitInfo != nil {
		// Check that both extCommit + commit are ordered in accordance with vp/address.
		if err := validateExtendedCommitAgainstLastCommit(extCommit, *commitInfo); err != nil {
			return err
		}
	}

	// Start checking vote extensions only **after** the vote extensions enable
	// height, because when `currentHeight == VoteExtensionsEnableHeight`
	// PrepareProposal doesn't get any vote extensions in its request.
	extsEnabled := cp.Feature != nil && cp.Feature.VoteExtensionsEnableHeight != nil && currentHeight > cp.Feature.VoteExtensionsEnableHeight.Value && cp.Feature.VoteExtensionsEnableHeight.Value != 0
	if !extsEnabled {
		extsEnabled = cp.Abci != nil && currentHeight > cp.Abci.VoteExtensionsEnableHeight && cp.Abci.VoteExtensionsEnableHeight != 0
	}
	if !extsEnabled {
		return nil
	}

	marshalDelimitedFn := func(msg gogoproto.Message) ([]byte, error) {
		var buf bytes.Buffer
		if err := protoio.NewDelimitedWriter(&buf).WriteMsg(msg); err != nil {
			return nil, err
		}

		return buf.Bytes(), nil
	}

	var (
		// Total voting power of all vote extensions.
		totalVP int64
		// Total voting power of all validators that submitted valid vote extensions.
		sumVP int64
	)

	for _, vote := range extCommit.Votes {
		totalVP += vote.Validator.Power

		// Only check + include power if the vote is a commit vote. There must be super-majority, otherwise the
		// previous block (the block the vote is for) could not have been committed.
		if vote.BlockIdFlag != cmtproto.BlockIDFlagCommit {
			continue
		}

		if !extsEnabled {
			if len(vote.VoteExtension) > 0 {
				return fmt.Errorf("vote extensions disabled; received non-empty vote extension at height %d", currentHeight)
			}
			if len(vote.ExtensionSignature) > 0 {
				return fmt.Errorf("vote extensions disabled; received non-empty vote extension signature at height %d", currentHeight)
			}

			continue
		}

		if len(vote.ExtensionSignature) == 0 {
			return fmt.Errorf("vote extensions enabled; received empty vote extension signature at height %d", currentHeight)
		}

		valConsAddr := sdk.ConsAddress(vote.Validator.Address)

		pubKeyProto, err := validatorStore(ctx, valConsAddr)
		if err != nil {
			return fmt.Errorf("failed to get validator %X public key: %w", valConsAddr, err)
		}

		cmtpk, err := cryptocodec.ToCmtProtoPublicKey(pubKeyProto)
		if err != nil {
			return fmt.Errorf("failed to convert validator %X public key: %w", valConsAddr, err)
		}

		cmtPubKey, err := cryptoenc.PubKeyFromProto(cmtpk)
		if err != nil {
			return fmt.Errorf("failed to convert validator %X public key: %w", valConsAddr, err)
		}

		cve := cmtproto.CanonicalVoteExtension{
			Extension: vote.VoteExtension,
			Height:    currentHeight - 1, // the vote extension was signed in the previous height
			Round:     int64(extCommit.Round),
			ChainId:   chainID,
		}

		extSignBytes, err := marshalDelimitedFn(&cve)
		if err != nil {
			return fmt.Errorf("failed to encode CanonicalVoteExtension: %w", err)
		}

		if !cmtPubKey.VerifySignature(extSignBytes, vote.ExtensionSignature) {
			return fmt.Errorf("failed to verify validator %X vote extension signature", valConsAddr)
		}

		sumVP += vote.Validator.Power
	}

	// This check is probably unnecessary, but better safe than sorry.
	if totalVP <= 0 {
		return fmt.Errorf("total voting power must be positive, got: %d", totalVP)
	}

	// If the sum of the voting power has not reached (2/3 + 1) we need to error.
	if requiredVP := ((totalVP * 2) / 3) + 1; sumVP < requiredVP {
		return fmt.Errorf(
			"insufficient cumulative voting power received to verify vote extensions; got: %d, expected: >=%d",
			sumVP, requiredVP,
		)
	}
	return nil
}

// validateExtendedCommitAgainstLastCommit validates an ExtendedCommitInfo against a LastCommit. Specifically,
// it checks that the ExtendedCommit + LastCommit (for the same height), are consistent with each other + that
// they are ordered correctly (by voting power) in accordance with
// [comet](https://github.com/cometbft/cometbft/blob/4ce0277b35f31985bbf2c25d3806a184a4510010/types/validator_set.go#L784).
func validateExtendedCommitAgainstLastCommit(ec abci.ExtendedCommitInfo, lc abci.CommitInfo) error {
	// check that the rounds are the same
	if ec.Round != lc.Round {
		return fmt.Errorf("extended commit round %d does not match last commit round %d", ec.Round, lc.Round)
	}

	// check that the # of votes are the same
	if len(ec.Votes) != len(lc.Votes) {
		return fmt.Errorf("extended commit votes length %d does not match last commit votes length %d", len(ec.Votes), len(lc.Votes))
	}

	// check sort order of extended commit votes
	if !slices.IsSortedFunc(ec.Votes, func(vote1, vote2 abci.ExtendedVoteInfo) int {
		if vote1.Validator.Power == vote2.Validator.Power {
			return bytes.Compare(vote1.Validator.Address, vote2.Validator.Address) // addresses sorted in ascending order (used to break vp conflicts)
		}
		return -int(vote1.Validator.Power - vote2.Validator.Power) // vp sorted in descending order
	}) {
		return errors.New("extended commit votes are not sorted by voting power")
	}

	addressCache := make(map[string]struct{}, len(ec.Votes))
	// check consistency between LastCommit and ExtendedCommit
	for i, vote := range ec.Votes {
		// cache addresses to check for duplicates
		if _, ok := addressCache[string(vote.Validator.Address)]; ok {
			return fmt.Errorf("extended commit vote address %X is duplicated", vote.Validator.Address)
		}
		addressCache[string(vote.Validator.Address)] = struct{}{}

		if !bytes.Equal(vote.Validator.Address, lc.Votes[i].Validator.Address) {
			return fmt.Errorf("extended commit vote address %X does not match last commit vote address %X", vote.Validator.Address, lc.Votes[i].Validator.Address)
		}
		if vote.Validator.Power != lc.Votes[i].Validator.Power {
			return fmt.Errorf("extended commit vote power %d does not match last commit vote power %d", vote.Validator.Power, lc.Votes[i].Validator.Power)
		}
	}

	return nil
}
