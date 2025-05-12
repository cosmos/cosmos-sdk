package main

import migration "github.com/cosmos/cosmos-sdk/tools/migrate"

var typeReplacements = []migration.TypeReplacement{
	// initChain
	{
		ImportPath: "github.com/cometbft/cometbft/abci/types",
		OldType:    "RequestInitChain",
		NewType:    "InitChainRequest",
	},
	{
		ImportPath: "github.com/cometbft/cometbft/abci/types",
		OldType:    "ResponseInitChain",
		NewType:    "InitChainResponse",
	},

	// echo
	{
		ImportPath: "github.com/cometbft/cometbft/abci/types",
		OldType:    "RequestEcho",
		NewType:    "EchoRequest",
	},
	{
		ImportPath: "github.com/cometbft/cometbft/abci/types",
		OldType:    "ResponseEcho",
		NewType:    "EchoResponse",
	},

	// flush
	{
		ImportPath: "github.com/cometbft/cometbft/abci/types",
		OldType:    "RequestFlush",
		NewType:    "FlushRequest",
	},
	{
		ImportPath: "github.com/cometbft/cometbft/abci/types",
		OldType:    "ResponseFlush",
		NewType:    "FlushResponse",
	},

	// info
	{
		ImportPath: "github.com/cometbft/cometbft/abci/types",
		OldType:    "RequestInfo",
		NewType:    "InfoRequest",
	},
	{
		ImportPath: "github.com/cometbft/cometbft/abci/types",
		OldType:    "ResponseInfo",
		NewType:    "InfoResponse",
	},

	// extendVote
	{
		ImportPath: "github.com/cometbft/cometbft/abci/types",
		OldType:    "RequestExtendVote",
		NewType:    "ExtendVoteRequest",
	},
	{
		ImportPath: "github.com/cometbft/cometbft/abci/types",
		OldType:    "ResponseExtendVote",
		NewType:    "ExtendVoteResponse",
	},

	// verifyVote
	{
		ImportPath: "github.com/cometbft/cometbft/abci/types",
		OldType:    "RequestVerifyVoteExtension",
		NewType:    "VerifyVoteExtensionRequest",
	},
	{
		ImportPath: "github.com/cometbft/cometbft/abci/types",
		OldType:    "ResponseVerifyVoteExtension",
		NewType:    "VerifyVoteExtensionResponse",
	},

	// commit
	{
		ImportPath: "github.com/cometbft/cometbft/abci/types",
		OldType:    "RequestCommit",
		NewType:    "CommitRequest",
	},
	{
		ImportPath: "github.com/cometbft/cometbft/abci/types",
		OldType:    "ResponseCommit",
		NewType:    "CommitResponse",
	},

	// checkTx
	{
		ImportPath: "github.com/cometbft/cometbft/abci/types",
		OldType:    "RequestCheckTx",
		NewType:    "CheckTxRequest",
	},
	{
		ImportPath: "github.com/cometbft/cometbft/abci/types",
		OldType:    "ResponseCheckTx",
		NewType:    "CheckTxResponse",
	},

	// finalizeBlock
	{
		ImportPath: "github.com/cometbft/cometbft/abci/types",
		OldType:    "RequestFinalizeBlock",
		NewType:    "FinalizeBlockRequest",
	},
	{
		ImportPath: "github.com/cometbft/cometbft/abci/types",
		OldType:    "ResponseFinalizeBlock",
		NewType:    "FinalizeBlockResponse",
	},

	// processProposal
	{
		ImportPath: "github.com/cometbft/cometbft/abci/types",
		OldType:    "RequestProcessProposal",
		NewType:    "ProcessProposalRequest",
	},
	{
		ImportPath: "github.com/cometbft/cometbft/abci/types",
		OldType:    "ResponseProcessProposal",
		NewType:    "ProcessProposalResponse",
	},

	// prepareProposal
	{
		ImportPath: "github.com/cometbft/cometbft/abci/types",
		OldType:    "RequestPrepareProposal",
		NewType:    "PrepareProposalRequest",
	},
	{
		ImportPath: "github.com/cometbft/cometbft/abci/types",
		OldType:    "ResponsePrepareProposal",
		NewType:    "PrepareProposalResponse",
	},

	// listSnapshots
	{
		ImportPath: "github.com/cometbft/cometbft/abci/types",
		OldType:    "RequestListSnapshots",
		NewType:    "ListSnapshotsRequest",
	},
	{
		ImportPath: "github.com/cometbft/cometbft/abci/types",
		OldType:    "ResponseListSnapshots",
		NewType:    "ListSnapshotsResponse",
	},

	// applySnapshotChunk
	{
		ImportPath: "github.com/cometbft/cometbft/abci/types",
		OldType:    "RequestApplySnapshotChunk",
		NewType:    "ApplySnapshotChunkRequest",
	},
	{
		ImportPath: "github.com/cometbft/cometbft/abci/types",
		OldType:    "ResponseApplySnapshotChunk",
		NewType:    "ApplySnapshotChunkResponse",
	},

	// query
	{
		ImportPath: "github.com/cometbft/cometbft/abci/types",
		OldType:    "RequestQuery",
		NewType:    "QueryRequest",
	},
	{
		ImportPath: "github.com/cometbft/cometbft/abci/types",
		OldType:    "ResponseQuery",
		NewType:    "QueryResponse",
	},

	// loadSnapshotChunk
	{
		ImportPath: "github.com/cometbft/cometbft/abci/types",
		OldType:    "RequestLoadSnapshotChunk",
		NewType:    "LoadSnapshotChunkRequest",
	},
	{
		ImportPath: "github.com/cometbft/cometbft/abci/types",
		OldType:    "ResponseLoadSnapshotChunk",
		NewType:    "LoadSnapshotChunkResponse",
	},

	// offerSnapshot
	{
		ImportPath: "github.com/cometbft/cometbft/abci/types",
		OldType:    "RequestOfferSnapshot",
		NewType:    "OfferSnapshotRequest",
	},
	{
		ImportPath: "github.com/cometbft/cometbft/abci/types",
		OldType:    "ResponseOfferSnapshot",
		NewType:    "OfferSnapshotResponse",
	},

	// eNUMS...
	{
		ImportPath: "github.com/cometbft/cometbft/abci/types",
		OldType:    "ResponseOfferSnapshot_Result",
		NewType:    "OfferSnapshotResult",
	},
	{
		ImportPath: "github.com/cometbft/cometbft/abci/types",
		OldType:    "ResponseOfferSnapshot_ACCEPT",
		NewType:    "OFFER_SNAPSHOT_RESULT_ACCEPT",
	},
	{
		ImportPath: "github.com/cometbft/cometbft/abci/types",
		OldType:    "ResponseOfferSnapshot_REJECT",
		NewType:    "OFFER_SNAPSHOT_RESULT_REJECT",
	},
	{
		ImportPath: "github.com/cometbft/cometbft/abci/types",
		OldType:    "ResponseOfferSnapshot_REJECT_FORMAT",
		NewType:    "OFFER_SNAPSHOT_RESULT_REJECT_FORMAT",
	},
	{
		ImportPath: "github.com/cometbft/cometbft/abci/types",
		OldType:    "ResponseApplySnapshotChunk_RETRY",
		NewType:    "APPLY_SNAPSHOT_CHUNK_RESULT_RETRY",
	},
	{
		ImportPath: "github.com/cometbft/cometbft/abci/types",
		OldType:    "ResponseVerifyVoteExtension_REJECT",
		NewType:    "VERIFY_VOTE_EXTENSION_STATUS_REJECT",
	},
	{
		ImportPath: "github.com/cometbft/cometbft/abci/types",
		OldType:    "ResponseVerifyVoteExtension_ACCEPT",
		NewType:    "VERIFY_VOTE_EXTENSION_STATUS_ACCEPT",
	},
}
