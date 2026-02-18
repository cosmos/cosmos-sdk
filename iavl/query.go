package iavl

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"cosmossdk.io/store/kv"
	cmtprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	ics23 "github.com/cosmos/ics23/go"
	"google.golang.org/protobuf/encoding/protowire"

	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/iavl/internal"
)

// Query handles query paths for a single IAVL-backed store.
func (c *CommitTree) Query(req *storetypes.RequestQuery) (_ *storetypes.ResponseQuery, err error) {
	start := time.Now()
	defer func() {
		if err != nil {
			proofLatency.Record(context.Background(), time.Since(start).Milliseconds())
		}
	}()

	height, err := c.queryHeight(req.Height)
	if err != nil {
		return &storetypes.ResponseQuery{}, err
	}

	res := &storetypes.ResponseQuery{Height: height}

	tree, err := c.GetVersion(uint32(height))
	if err != nil {
		// Keep this as response metadata to match existing query behavior.
		res.Log = err.Error()
		return res, nil
	}

	switch req.Path {
	case "/key":
		if len(req.Data) == 0 {
			return &storetypes.ResponseQuery{}, errorsmod.Wrap(storetypes.ErrTxDecode, "query cannot be zero length")
		}

		key := req.Data
		res.Key = key

		value, err := tree.GetErr(key)
		if err != nil {
			return &storetypes.ResponseQuery{}, err
		}
		res.Value = value

		if req.Prove {
			res.ProofOps, err = iavlProofOps(&tree, key, value != nil)
			if err != nil {
				return &storetypes.ResponseQuery{}, errorsmod.Wrapf(storetypes.ErrInvalidRequest, "failed to create proof: %v", err)
			}
		}

		return res, nil

	case "/subspace":
		subspace := req.Data
		res.Key = subspace

		iterator := storetypes.KVStorePrefixIterator(tree, subspace)
		pairs := kv.Pairs{
			Pairs: make([]kv.Pair, 0),
		}
		for ; iterator.Valid(); iterator.Next() {
			pairs.Pairs = append(pairs.Pairs, kv.Pair{
				Key:   bytes.Clone(iterator.Key()),
				Value: bytes.Clone(iterator.Value()),
			})
		}
		if err := iterator.Close(); err != nil {
			return &storetypes.ResponseQuery{}, fmt.Errorf("failed to close iterator: %w", err)
		}

		bz, err := pairs.Marshal()
		if err != nil {
			panic(fmt.Errorf("failed to marshal KV pairs: %w", err))
		}

		res.Value = bz

		return res, nil

	default:
		return &storetypes.ResponseQuery{}, errorsmod.Wrapf(storetypes.ErrUnknownRequest, "unexpected query path: %v", req.Path)
	}
}

func (c *CommitTree) queryHeight(reqHeight int64) (int64, error) {
	if reqHeight == 0 {
		return int64(c.LatestVersion()), nil
	}
	if reqHeight < 0 {
		return 0, errorsmod.Wrapf(storetypes.ErrInvalidRequest, "invalid query height: %d", reqHeight)
	}
	if reqHeight > int64(^uint32(0)) {
		return 0, errorsmod.Wrapf(storetypes.ErrInvalidRequest, "query height %d exceeds max supported height", reqHeight)
	}

	return reqHeight, nil
}

func iavlProofOps(tree *internal.TreeReader, key []byte, exists bool) (*cmtprotocrypto.ProofOps, error) {
	var (
		commitmentProof *ics23.CommitmentProof
		err             error
	)

	if exists {
		commitmentProof, err = tree.GetMembershipProof(key)
	} else {
		commitmentProof, err = tree.GetNonMembershipProof(key)
	}
	if err != nil {
		return nil, err
	}

	op := storetypes.NewIavlCommitmentOp(key, commitmentProof)
	return &cmtprotocrypto.ProofOps{Ops: []cmtprotocrypto.ProofOp{op.ProofOp()}}, nil
}

type kvPair struct {
	key   []byte
	value []byte
}

// marshalLegacyKVPairs emits the same wire format as cosmos.store.internal.kv.v1beta1.Pairs.
func marshalLegacyKVPairs(pairs []kvPair) []byte {
	var out []byte
	for _, pair := range pairs {
		var pairMsg []byte
		pairMsg = protowire.AppendTag(pairMsg, 1, protowire.BytesType)
		pairMsg = protowire.AppendBytes(pairMsg, pair.key)
		pairMsg = protowire.AppendTag(pairMsg, 2, protowire.BytesType)
		pairMsg = protowire.AppendBytes(pairMsg, pair.value)

		out = protowire.AppendTag(out, 1, protowire.BytesType)
		out = protowire.AppendBytes(out, pairMsg)
	}

	return out
}

var _ storetypes.Queryable = (*CommitTree)(nil)
