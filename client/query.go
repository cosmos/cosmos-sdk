package client

import (
	"github.com/pkg/errors"

	"github.com/tendermint/go-wire/data"
	"github.com/tendermint/iavl"

	"github.com/tendermint/tendermint/certifiers"
	"github.com/tendermint/tendermint/certifiers/client"
	certerr "github.com/tendermint/tendermint/certifiers/errors"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// GetWithProof will query the key on the given node, and verify it has
// a valid proof, as defined by the certifier.
//
// If there is any error in checking, returns an error.
// If val is non-empty, proof should be KeyExistsProof
// If val is empty, proof should be KeyMissingProof
func GetWithProof(key []byte, reqHeight int, node rpcclient.Client,
	cert certifiers.Certifier) (
	val data.Bytes, height uint64, proof iavl.KeyProof, err error) {

	if reqHeight < 0 {
		err = errors.Errorf("Height cannot be negative")
		return
	}

	resp, proof, err := GetWithProofOptions("/key", key,
		rpcclient.ABCIQueryOptions{Height: uint64(reqHeight)},
		node, cert)
	if resp != nil {
		val, height = resp.Value, resp.Height
	}
	return val, height, proof, err
}

// GetWithProofOptions is useful if you want full access to the ABCIQueryOptions
func GetWithProofOptions(path string, key []byte, opts rpcclient.ABCIQueryOptions,
	node rpcclient.Client, cert certifiers.Certifier) (
	*ctypes.ResultABCIQuery, iavl.KeyProof, error) {

	resp, err := node.ABCIQueryWithOptions(path, key, opts)
	if err != nil {
		return nil, nil, err
	}

	// make sure the proof is the proper height
	if !resp.Code.IsOK() {
		err = errors.Errorf("Query error %d: %s", resp.Code, resp.Code.String())
		return nil, nil, err
	}
	if len(resp.Key) == 0 || len(resp.Proof) == 0 {
		return nil, nil, ErrNoData()
	}
	if resp.Height == 0 {
		return nil, nil, errors.New("Height returned is zero")
	}

	// AppHash for height H is in header H+1
	commit, err := GetCertifiedCommit(int(resp.Height+1), node, cert)
	if err != nil {
		return nil, nil, err
	}

	if len(resp.Value) > 0 {
		// The key was found, construct a proof of existence.
		eproof, err := iavl.ReadKeyExistsProof(resp.Proof)
		if err != nil {
			return nil, nil, errors.Wrap(err, "Error reading proof")
		}

		// Validate the proof against the certified header to ensure data integrity.
		err = eproof.Verify(resp.Key, resp.Value, commit.Header.AppHash)
		if err != nil {
			return nil, nil, errors.Wrap(err, "Couldn't verify proof")
		}
		return resp, eproof, nil
	}

	// The key wasn't found, construct a proof of non-existence.
	var aproof *iavl.KeyAbsentProof
	aproof, err = iavl.ReadKeyAbsentProof(resp.Proof)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Error reading proof")
	}
	// Validate the proof against the certified header to ensure data integrity.
	err = aproof.Verify(resp.Key, nil, commit.Header.AppHash)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Couldn't verify proof")
	}
	return resp, aproof, ErrNoData()
}

// GetCertifiedCommit gets the signed header for a given height
// and certifies it.  Returns error if unable to get a proven header.
func GetCertifiedCommit(h int, node rpcclient.Client,
	cert certifiers.Certifier) (empty certifiers.Commit, err error) {

	// FIXME: cannot use cert.GetByHeight for now, as it also requires
	// Validators and will fail on querying tendermint for non-current height.
	// When this is supported, we should use it instead...
	rpcclient.WaitForHeight(node, h, nil)
	cresp, err := node.Commit(&h)
	if err != nil {
		return
	}
	commit := client.CommitFromResult(cresp)

	// validate downloaded checkpoint with our request and trust store.
	if commit.Height() != h {
		return empty, certerr.ErrHeightMismatch(h, commit.Height())
	}
	err = cert.Certify(commit)
	return commit, nil
}
