package main

import (
	"go/ast"
	"strings"

	"github.com/rs/zerolog/log"
)

type TypeReplacement struct {
	ImportPath string // import path for the package containing the type
	OldType    string // old type name (without package prefix)
	NewType    string // new type name (without package prefix)
}

var (
	typeReplacements = []TypeReplacement{
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
)

// updateStructs finds and replaces all references to the specified struct types
func updateStructs(node *ast.File, typeReplacements []TypeReplacement) (bool, error) {
	modified := false
	// first, build a map of import paths to their aliases in this file
	importAliases := make(map[string]string) // maps import path to its alias
	for _, imp := range node.Imports {
		importPath := strings.Trim(imp.Path.Value, "\"")

		// determine the package alias
		var alias string
		if imp.Name != nil {
			// explicit alias
			alias = imp.Name.Name
		} else {
			// default alias is the last part of the import path
			parts := strings.Split(importPath, "/")
			alias = parts[len(parts)-1]
		}

		importAliases[importPath] = alias
	}

	// now walk the AST and find all selector expressions to replace
	ast.Inspect(node, func(n ast.Node) bool {
		// check if this is a selector expression (e.g., abci.RequestInitChain)
		selectorExpr, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		// get the identifier (package name/alias) - e.g., "abci" part in abci.RequestInitChain
		ident, ok := selectorExpr.X.(*ast.Ident)
		if !ok {
			return true
		}

		for _, replacement := range typeReplacements {
			alias, exists := importAliases[replacement.ImportPath]
			if !exists {
				// this file doesn't import the package we're interested in
				continue
			}

			// check if this selector matches our target type
			if ident.Name == alias && selectorExpr.Sel.Name == replacement.OldType {
				// we found a match, replace the type name
				selectorExpr.Sel.Name = replacement.NewType
				modified = true
				log.Debug().Msgf("Replaced %s.%s with %s.%s",
					ident.Name, replacement.OldType,
					ident.Name, replacement.NewType)
			}
		}

		return true
	})

	return modified, nil
}
