package context

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/merkle"
	cmn "github.com/tendermint/tendermint/libs/common"
	tmliteErr "github.com/tendermint/tendermint/lite/errors"
	tmliteProxy "github.com/tendermint/tendermint/lite/proxy"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetNode returns an RPC client. If the context's client is not defined, an
// error is returned.
func (ctx CLIContext) GetNode() (rpcclient.Client, error) {
	if ctx.Client == nil {
		return nil, errors.New("no RPC client defined")
	}

	return ctx.Client, nil
}

// WaitForNBlocks blocks until the node defined on the context has advanced N blocks
func (ctx CLIContext) WaitForNBlocks(n int64) {
	node, err := ctx.GetNode()
	if err != nil {
		panic(err)
	}

	resBlock, err := node.Block(nil)
	var height int64
	if err != nil || resBlock.Block == nil {
		// wait for the first block to exist
		ctx.waitForHeight(1)
		height = 1 + n
	} else {
		height = resBlock.Block.Height + n
	}
	ctx.waitForHeight(height)
}

func (ctx CLIContext) waitForHeight(height int64) {
	node, err := ctx.GetNode()
	if err != nil {
		panic(err)
	}

	for {
		// get url, try a few times
		var resBlock *ctypes.ResultBlock
		var err error
	INNER:
		for i := 0; i < 5; i++ {
			resBlock, err = node.Block(nil)
			if err == nil {
				break INNER
			}
			time.Sleep(time.Millisecond * 200)
		}
		if err != nil {
			panic(err)
		}

		if resBlock.Block != nil && resBlock.Block.Height >= height {
			return
		}

		time.Sleep(time.Millisecond * 100)
	}
}

// Query performs a query to a Tendermint node with the provided path.
// It returns the result and height of the query upon success or an error if
// the query fails.
func (ctx CLIContext) Query(path string) ([]byte, int64, error) {
	return ctx.query(path, nil)
}

// QueryWithData performs a query to a Tendermint node with the provided path
// and a data payload. It returns the result and height of the query upon success
// or an error if the query fails.
func (ctx CLIContext) QueryWithData(path string, data []byte) ([]byte, int64, error) {
	return ctx.query(path, data)
}

// QueryStore performs a query to a Tendermint node with the provided key and
// store name. It returns the result and height of the query upon success
// or an error if the query fails.
func (ctx CLIContext) QueryStore(key cmn.HexBytes, storeName string) ([]byte, int64, error) {
	return ctx.queryStore(key, storeName, "key")
}

// QueryABCI performs a query to a Tendermint node with the provide RequestQuery.
// It returns the ResultQuery obtained from the query.
func (ctx CLIContext) QueryABCI(req abci.RequestQuery) (abci.ResponseQuery, error) {
	return ctx.queryABCI(req)
}

// QueryABCI performs a query to a Tendermint node with the provide RequestQuery.
// It returns the ResultQuery obtained from the query.
func (ctx CLIContext) QueryABCIWithProof(r *http.Request, module string, data []byte) (abci.ResponseQuery, error) {
	proveStr := r.FormValue("prove")
	prove := false
	if ok, err := strconv.ParseBool(proveStr); err == nil {
		prove = ok
	}
	req := abci.RequestQuery{
		Path:  fmt.Sprintf("store/%s/key", module),
		Data:  data,
		Prove: prove,
	}
	return ctx.queryABCI(req)
}

// QuerySubspace performs a query to a Tendermint node with the provided
// store name and subspace. It returns key value pair and height of the query
// upon success or an error if the query fails.
func (ctx CLIContext) QuerySubspace(subspace []byte, storeName string) (res []sdk.KVPair, height int64, err error) {
	resRaw, height, err := ctx.queryStore(subspace, storeName, "subspace")
	if err != nil {
		return res, height, err
	}

	ctx.Codec.MustUnmarshalBinaryLengthPrefixed(resRaw, &res)
	return
}

// GetFromAddress returns the from address from the context's name.
func (ctx CLIContext) GetFromAddress() sdk.AccAddress {
	return ctx.FromAddress
}

// GetFromName returns the key name for the current context.
func (ctx CLIContext) GetFromName() string {
	return ctx.FromName
}

func (ctx CLIContext) queryABCI(req abci.RequestQuery) (abci.ResponseQuery, error) {
	node, err := ctx.GetNode()
	if err != nil {
		return abci.ResponseQuery{}, err
	}

	// When a client did not provide a query height, manually query for it so it can
	// be injected downstream into responses.
	if ctx.Height == 0 {
		status, err := node.Status()
		if err != nil {
			return abci.ResponseQuery{}, err
		}
		ctx = ctx.WithHeight(status.SyncInfo.LatestBlockHeight)
	}

	opts := rpcclient.ABCIQueryOptions{
		Height: ctx.Height,
		Prove:  req.Prove || !ctx.TrustNode,
	}

	result, err := node.ABCIQueryWithOptions(req.Path, req.Data, opts)
	if err != nil {
		return abci.ResponseQuery{}, err
	}

	if !result.Response.IsOK() {
		err = errors.New(result.Response.Log)
		return abci.ResponseQuery{}, err
	}

	// data from trusted node or subspace query doesn't need verification
	if ctx.TrustNode || !isQueryStoreWithProof(req.Path) {
		return result.Response, nil
	}

	err = ctx.verifyProof(req.Path, result.Response)
	if err != nil {
		return abci.ResponseQuery{}, err
	}

	return result.Response, nil
}

// query performs a query to a Tendermint node with the provided store name
// and path. It returns the result and height of the query upon success
// or an error if the query fails. In addition, it will verify the returned
// proof if TrustNode is disabled. If proof verification fails or the query
// height is invalid, an error will be returned.
func (ctx CLIContext) query(path string, key cmn.HexBytes) (res []byte, height int64, err error) {
	resp, err := ctx.queryABCI(abci.RequestQuery{
		Path: path,
		Data: key,
	})
	if err != nil {
		return
	}

	return resp.Value, resp.Height, nil
}

// Verify verifies the consensus proof at given height.
func (ctx CLIContext) Verify(height int64) (tmtypes.SignedHeader, error) {
	check, err := tmliteProxy.GetCertifiedCommit(height, ctx.Client, ctx.Verifier)
	switch {
	case tmliteErr.IsErrCommitNotFound(err):
		return tmtypes.SignedHeader{}, ErrVerifyCommit(height)
	case err != nil:
		return tmtypes.SignedHeader{}, err
	}

	return check, nil
}

// verifyProof perform response proof verification.
func (ctx CLIContext) verifyProof(queryPath string, resp abci.ResponseQuery) error {
	if ctx.Verifier == nil {
		return fmt.Errorf("missing valid certifier to verify data from distrusted node")
	}

	// the AppHash for height H is in header H+1
	commit, err := ctx.Verify(resp.Height + 1)
	if err != nil {
		return err
	}

	// TODO: Instead of reconstructing, stash on CLIContext field?
	prt := rootmulti.DefaultProofRuntime()

	// TODO: Better convention for path?
	storeName, err := parseQueryStorePath(queryPath)
	if err != nil {
		return err
	}

	kp := merkle.KeyPath{}
	kp = kp.AppendKey([]byte(storeName), merkle.KeyEncodingURL)
	kp = kp.AppendKey(resp.Key, merkle.KeyEncodingURL)

	if resp.Value == nil {
		err = prt.VerifyAbsence(resp.Proof, commit.Header.AppHash, kp.String())
		if err != nil {
			return errors.Wrap(err, "failed to prove merkle proof")
		}
		return nil
	}
	err = prt.VerifyValue(resp.Proof, commit.Header.AppHash, kp.String(), resp.Value)
	if err != nil {
		return errors.Wrap(err, "failed to prove merkle proof")
	}

	return nil
}

// queryStore performs a query to a Tendermint node with the provided a store
// name and path. It returns the result and height of the query upon success
// or an error if the query fails.
func (ctx CLIContext) queryStore(key cmn.HexBytes, storeName, endPath string) ([]byte, int64, error) {
	path := fmt.Sprintf("/store/%s/%s", storeName, endPath)
	return ctx.query(path, key)
}

// isQueryStoreWithProof expects a format like /<queryType>/<storeName>/<subpath>
// queryType must be "store" and subpath must be "key" to require a proof.
func isQueryStoreWithProof(path string) bool {
	if !strings.HasPrefix(path, "/") {
		return false
	}

	paths := strings.SplitN(path[1:], "/", 3)
	switch {
	case len(paths) != 3:
		return false
	case paths[0] != "store":
		return false
	case rootmulti.RequireProof("/" + paths[2]):
		return true
	}

	return false
}

// parseQueryStorePath expects a format like /store/<storeName>/key.
func parseQueryStorePath(path string) (storeName string, err error) {
	if !strings.HasPrefix(path, "/") {
		return "", errors.New("expected path to start with /")
	}

	paths := strings.SplitN(path[1:], "/", 3)
	switch {
	case len(paths) != 3:
		return "", errors.New("expected format like /store/<storeName>/key")
	case paths[0] != "store":
		return "", errors.New("expected format like /store/<storeName>/key")
	case paths[2] != "key":
		return "", errors.New("expected format like /store/<storeName>/key")
	}

	return paths[1], nil
}
