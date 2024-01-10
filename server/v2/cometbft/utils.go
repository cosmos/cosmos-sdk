package cometbft

import (
	"fmt"
	"strings"

	"cosmossdk.io/core/comet"
	errorsmod "cosmossdk.io/errors"
	coreappmgr "cosmossdk.io/server/v2/core/appmanager"
	"cosmossdk.io/server/v2/core/event"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtcrypto "github.com/cometbft/cometbft/api/cometbft/crypto/v1"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	gogoproto "github.com/cosmos/gogoproto/proto"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

func parseQueryRequest(req *abci.QueryRequest) (proto.Message, error) {
	desc, err := gogoproto.HybridResolver.FindDescriptorByName(protoreflect.FullName(req.Path))
	if err != nil {
		return nil, err
	}

	methodDesc, ok := desc.(protoreflect.MethodDescriptor)
	if !ok {
		return nil, fmt.Errorf("invalid method descriptor %s", desc.FullName())
	}

	queryReqType := dynamicpb.NewMessage(methodDesc.Input())
	err = proto.Unmarshal(req.Data, queryReqType)

	return queryReqType, err
}

// parseQueryResponse needs the request to get the path
func parseQueryResponse(req *abci.QueryRequest, res proto.Message) (*abci.QueryResponse, error) {
	desc, err := gogoproto.HybridResolver.FindDescriptorByName(protoreflect.FullName(req.Path))
	if err != nil {
		return nil, err
	}

	methodDesc, ok := desc.(protoreflect.MethodDescriptor)
	if !ok {
		return nil, fmt.Errorf("invalid method descriptor %s", desc.FullName())
	}

	queryRespType := dynamicpb.NewMessage(methodDesc.Output())
	proto.Merge(queryRespType, res)
	bz, err := proto.Marshal(res)
	if err != nil {
		return nil, err
	}

	//TODO: how do I reply? I suppose we need to different replies depending of the query

	return &abci.QueryResponse{ // TODO: fill all the fields
		Code:      0,
		Log:       "",
		Info:      "",
		Index:     0,
		Key:       []byte{},
		Value:     bz,
		ProofOps:  &cmtcrypto.ProofOps{},
		Height:    0,
		Codespace: "",
	}, nil
}

// SplitABCIQueryPath splits a string path using the delimiter '/'.
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

func parseFinalizeBlockResponse(in *coreappmgr.BlockResponse, appHash []byte) (*abci.FinalizeBlockResponse, error) {
	allEvents := append(in.BeginBlockEvents, in.EndBlockEvents...)

	resp := &abci.FinalizeBlockResponse{
		Events:                intoABCIEvents(allEvents),
		TxResults:             intoABCITxResults(in.TxResults),
		ValidatorUpdates:      intoABCIValidatorUpdates(in.ValidatorUpdates),
		AppHash:               appHash,
		ConsensusParamUpdates: nil, // TODO: figure out consensus params here, maybe parse the tx responses or events?
	}
	return resp, nil
}

func intoABCIValidatorUpdates(updates []coreappmgr.ValidatorUpdate) []abci.ValidatorUpdate {
	valsetUpdates := make([]abci.ValidatorUpdate, len(updates))

	for i := range updates {
		valsetUpdates[i] = abci.ValidatorUpdate{
			PubKey: cmtcrypto.PublicKey{
				Sum: &cmtcrypto.PublicKey_Ed25519{ // TODO: check if this is ok
					Ed25519: updates[i].PubKey,
				},
			},
			Power: updates[i].Power,
		}
	}

	return valsetUpdates
}

func intoABCITxResults(results []coreappmgr.TxResult) []*abci.ExecTxResult {
	res := make([]*abci.ExecTxResult, len(results))
	for i := range results {
		if results[i].Error == nil {
			res[i] = sdkerrors.ResponseExecTxResultWithEvents(
				results[i].Error,
				0, // TODO: gas wanted?
				results[i].GasUsed,
				intoABCIEvents(results[i].Events),
				false,
			)
			continue
		}

		// TODO: handle properly once the we decide on the type of TxResult.Resp
	}

	return res
}

func intoABCIEvents(events []event.Event) []abci.Event {
	abciEvents := make([]abci.Event, len(events))
	for i := range events {
		abciEvents[i] = abci.Event{
			Type:       events[i].Type,
			Attributes: intoABCIAttributes(events[i].Attributes),
		}
	}
	return abciEvents
}

func intoABCIAttributes(attributes []event.Attribute) []abci.EventAttribute {
	abciAttributes := make([]abci.EventAttribute, len(attributes))
	for i := range attributes {
		abciAttributes[i] = abci.EventAttribute{
			Key:   attributes[i].Key,
			Value: attributes[i].Value,
			Index: false, // TODO: who holds this config?
		}
	}
	return abciAttributes
}

// ToSDKEvidence takes comet evidence and returns sdk evidence
func ToSDKEvidence(ev []abci.Misbehavior) []comet.Evidence {
	evidence := make([]comet.Evidence, len(ev))
	for i, e := range ev {
		evidence[i] = comet.Evidence{
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

// ToSDKDecidedCommitInfo takes comet commit info and returns sdk commit info
func ToSDKCommitInfo(commit abci.CommitInfo) comet.CommitInfo {
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
func QueryResult(err error, debug bool) *abci.QueryResponse {
	space, code, log := errorsmod.ABCIInfo(err, debug)
	return &abci.QueryResponse{
		Codespace: space,
		Code:      code,
		Log:       log,
	}
}
