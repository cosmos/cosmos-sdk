package client

import (
	"github.com/pkg/errors"

	"github.com/tendermint/go-wire/data"
	"github.com/tendermint/iavl"
	lc "github.com/tendermint/light-client"
	"github.com/tendermint/light-client/certifiers"

	"github.com/tendermint/tendermint/rpc/client"
)

// GetWithProof will query the key on the given node, and verify it has
// a valid proof, as defined by the certifier.
//
// If there is any error in checking, returns an error.
// If val is non-empty, proof should be KeyExistsProof
// If val is empty, proof should be KeyMissingProof
func GetWithProof(key []byte, node client.Client, cert certifiers.Certifier) (
	val data.Bytes, height uint64, proof iavl.KeyProof, err error) {

	resp, err := node.ABCIQuery("/key", key)
	if err != nil {
		return
	}

	// make sure the proof is the proper height
	if !resp.Code.IsOK() {
		err = errors.Errorf("Query error %d: %s", resp.Code, resp.Code.String())
		return
	}
	if len(resp.Key) == 0 || len(resp.Proof) == 0 {
		err = lc.ErrNoData()
		return
	}
	if resp.Height == 0 {
		err = errors.New("Height returned is zero")
		return
	}

	// AppHash for height H is in header H+1
	var check lc.Checkpoint
	check, err = GetCertifiedCheckpoint(int(resp.Height+1), node, cert)
	if err != nil {
		return
	}

	if len(resp.Value) > 0 {
		// The key was found, construct a proof of existence.
		var eproof *iavl.KeyExistsProof
		eproof, err = iavl.ReadKeyExistsProof(resp.Proof)
		if err != nil {
			err = errors.Wrap(err, "Error reading proof")
			return
		}

		// Validate the proof against the certified header to ensure data integrity.
		err = eproof.Verify(resp.Key, resp.Value, check.Header.AppHash)
		if err != nil {
			err = errors.Wrap(err, "Couldn't verify proof")
			return
		}
		val = data.Bytes(resp.Value)
		proof = eproof
	} else {
		// The key wasn't found, construct a proof of non-existence.
		var aproof *iavl.KeyAbsentProof
		aproof, err = iavl.ReadKeyAbsentProof(resp.Proof)
		if err != nil {
			err = errors.Wrap(err, "Error reading proof")
			return
		}
		// Validate the proof against the certified header to ensure data integrity.
		err = aproof.Verify(resp.Key, nil, check.Header.AppHash)
		if err != nil {
			err = errors.Wrap(err, "Couldn't verify proof")
			return
		}
		err = lc.ErrNoData()
		proof = aproof
	}

	height = resp.Height
	return
}

// GetCertifiedCheckpoint gets the signed header for a given height
// and certifies it.  Returns error if unable to get a proven header.
func GetCertifiedCheckpoint(h int, node client.Client,
	cert certifiers.Certifier) (empty lc.Checkpoint, err error) {

	// FIXME: cannot use cert.GetByHeight for now, as it also requires
	// Validators and will fail on querying tendermint for non-current height.
	// When this is supported, we should use it instead...
	client.WaitForHeight(node, h, nil)
	commit, err := node.Commit(&h)
	if err != nil {
		return
	}
	check := lc.Checkpoint{
		Header: commit.Header,
		Commit: commit.Commit,
	}

	// validate downloaded checkpoint with our request and trust store.
	if check.Height() != h {
		return empty, lc.ErrHeightMismatch(h, check.Height())
	}
	err = cert.Certify(check)
	return check, nil
}
