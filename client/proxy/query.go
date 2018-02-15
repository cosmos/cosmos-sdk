package proxy

import (
	"github.com/pkg/errors"

	cmn "github.com/tendermint/tmlibs/common"

	"github.com/tendermint/tendermint/lite"
	"github.com/tendermint/tendermint/lite/client"
	certerr "github.com/tendermint/tendermint/lite/errors"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// KeyProof represents a proof of existence or absence of a single key.
// Copied from iavl repo. TODO
type KeyProof interface {
	// Verify verfies the proof is valid. To verify absence,
	// the value should be nil.
	Verify(key, value, root []byte) error

	// Root returns the root hash of the proof.
	Root() []byte

	// Serialize itself
	Bytes() []byte
}

// GetWithProof will query the key on the given node, and verify it has
// a valid proof, as defined by the certifier.
//
// If there is any error in checking, returns an error.
// If val is non-empty, proof should be KeyExistsProof
// If val is empty, proof should be KeyMissingProof
func GetWithProof(key []byte, reqHeight int64, node rpcclient.Client,
	cert lite.Certifier) (
	val cmn.HexBytes, height int64, proof KeyProof, err error) {

	if reqHeight < 0 {
		err = errors.Errorf("Height cannot be negative")
		return
	}

	_resp, proof, err := GetWithProofOptions("/key", key,
		rpcclient.ABCIQueryOptions{Height: int64(reqHeight)},
		node, cert)
	if _resp != nil {
		resp := _resp.Response
		val, height = resp.Value, resp.Height
	}
	return val, height, proof, err
}

// GetWithProofOptions is useful if you want full access to the ABCIQueryOptions
func GetWithProofOptions(path string, key []byte, opts rpcclient.ABCIQueryOptions,
	node rpcclient.Client, cert lite.Certifier) (
	*ctypes.ResultABCIQuery, KeyProof, error) {

	_resp, err := node.ABCIQueryWithOptions(path, key, opts)
	if err != nil {
		return nil, nil, err
	}
	resp := _resp.Response

	// make sure the proof is the proper height
	if resp.IsErr() {
		err = errors.Errorf("Query error for key %d: %d", key, resp.Code)
		return nil, nil, err
	}
	if len(resp.Key) == 0 || len(resp.Proof) == 0 {
		return nil, nil, ErrNoData()
	}
	if resp.Height == 0 {
		return nil, nil, errors.New("Height returned is zero")
	}

	// AppHash for height H is in header H+1
	commit, err := GetCertifiedCommit(resp.Height+1, node, cert)
	if err != nil {
		return nil, nil, err
	}

	_ = commit
	return &ctypes.ResultABCIQuery{Response: resp}, nil, nil

	/* // TODO refactor so iavl stuff is not in tendermint core
	   // https://github.com/tendermint/tendermint/issues/1183
	if len(resp.Value) > 0 {
		// The key was found, construct a proof of existence.
		proof, err := iavl.ReadKeyProof(resp.Proof)
		if err != nil {
			return nil, nil, errors.Wrap(err, "Error reading proof")
		}

		eproof, ok := proof.(*iavl.KeyExistsProof)
		if !ok {
			return nil, nil, errors.New("Expected KeyExistsProof for non-empty value")
		}

		// Validate the proof against the certified header to ensure data integrity.
		err = eproof.Verify(resp.Key, resp.Value, commit.Header.AppHash)
		if err != nil {
			return nil, nil, errors.Wrap(err, "Couldn't verify proof")
		}
		return &ctypes.ResultABCIQuery{Response: resp}, eproof, nil
	}

	// The key wasn't found, construct a proof of non-existence.
	proof, err := iavl.ReadKeyProof(resp.Proof)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Error reading proof")
	}

	aproof, ok := proof.(*iavl.KeyAbsentProof)
	if !ok {
		return nil, nil, errors.New("Expected KeyAbsentProof for empty Value")
	}

	// Validate the proof against the certified header to ensure data integrity.
	err = aproof.Verify(resp.Key, nil, commit.Header.AppHash)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Couldn't verify proof")
	}
	return &ctypes.ResultABCIQuery{Response: resp}, aproof, ErrNoData()
	*/
}

// GetCertifiedCommit gets the signed header for a given height
// and certifies it.  Returns error if unable to get a proven header.
func GetCertifiedCommit(h int64, node rpcclient.Client, cert lite.Certifier) (lite.Commit, error) {

	// FIXME: cannot use cert.GetByHeight for now, as it also requires
	// Validators and will fail on querying tendermint for non-current height.
	// When this is supported, we should use it instead...
	rpcclient.WaitForHeight(node, h, nil)
	cresp, err := node.Commit(&h)
	if err != nil {
		return lite.Commit{}, err
	}

	commit := client.CommitFromResult(cresp)
	// validate downloaded checkpoint with our request and trust store.
	if commit.Height() != h {
		return lite.Commit{}, certerr.ErrHeightMismatch(h, commit.Height())
	}

	if err = cert.Certify(commit); err != nil {
		return lite.Commit{}, err
	}

	return commit, nil
}
