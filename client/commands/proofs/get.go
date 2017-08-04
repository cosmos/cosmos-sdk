package proofs

import (
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/viper"

	wire "github.com/tendermint/go-wire"
	"github.com/tendermint/go-wire/data"
	lc "github.com/tendermint/light-client"
	"github.com/tendermint/light-client/proofs"
	"github.com/tendermint/merkleeyes/iavl"
	"github.com/tendermint/tendermint/rpc/client"

	"github.com/tendermint/basecoin/client/commands"
)

// Get does most of the work of the query commands, but is quite
// opinionated, so if you want more control, set up the items and call Get
// directly.  Notably, it always uses go-wire.ReadBinaryBytes to deserialize,
// and Height and Node from standard flags.
//
// It will try to get the proof for the given key.  If it is successful,
// it will return the proof and also unserialize proof.Data into the data
// argument (so pass in a pointer to the appropriate struct)
func GetParsed(key []byte, data interface{}, prove bool) (uint64, error) {
	bs, h, err := Get(key, prove)
	if err != nil {
		return 0, err
	}
	err = wire.ReadBinaryBytes(bs, data)
	if err != nil {
		return 0, err
	}
	return h, nil
}

func Get(key []byte, prove bool) (data.Bytes, uint64, error) {
	if !prove {
		node := commands.GetNode()
		resp, err := node.ABCIQuery("/key", key, false)
		return data.Bytes(resp.Value), resp.Height, err
	}
	val, h, _, err := GetWithProof(key)
	return val, h, err
}

func GetWithProof(key []byte) (data.Bytes, uint64, *iavl.KeyExistsProof, error) {
	node := commands.GetNode()

	resp, err := node.ABCIQuery("/key", key, true)
	if err != nil {
		return nil, 0, nil, err
	}
	ph := int(resp.Height)

	// make sure the proof is the proper height
	if !resp.Code.IsOK() {
		return nil, 0, nil, errors.Errorf("Query error %d: %s", resp.Code, resp.Code.String())
	}
	// TODO: Handle null proofs
	if len(resp.Key) == 0 || len(resp.Value) == 0 || len(resp.Proof) == 0 {
		return nil, 0, nil, lc.ErrNoData()
	}
	if ph != 0 && ph != int(resp.Height) {
		return nil, 0, nil, lc.ErrHeightMismatch(ph, int(resp.Height))
	}

	check, err := GetCertifiedCheckpoint(ph)
	if err != nil {
		return nil, 0, nil, err
	}

	proof := new(iavl.KeyExistsProof)
	err = wire.ReadBinaryBytes(resp.Proof, &proof)
	if err != nil {
		return nil, 0, nil, err
	}

	// validate the proof against the certified header to ensure data integrity
	err = proof.Verify(resp.Key, resp.Value, check.Header.AppHash)
	if err != nil {
		return nil, 0, nil, err
	}

	return data.Bytes(resp.Value), resp.Height, proof, nil
}

// GetCertifiedCheckpoint gets the signed header for a given height
// and certifies it.  Returns error if unable to get a proven header.
func GetCertifiedCheckpoint(h int) (empty lc.Checkpoint, err error) {
	// here is the certifier, root of all knowledge
	node := commands.GetNode()
	cert, err := commands.GetCertifier()
	if err != nil {
		return
	}

	// get and validate a signed header for this proof

	// FIXME: cannot use cert.GetByHeight for now, as it also requires
	// Validators and will fail on querying tendermint for non-current height.
	// When this is supported, we should use it instead...
	client.WaitForHeight(node, h, nil)
	commit, err := node.Commit(h)
	if err != nil {
		return
	}
	check := lc.Checkpoint{
		Header: commit.Header,
		Commit: commit.Commit,
	}
	// double check we got the same height
	if check.Height() != h {
		return empty, lc.ErrHeightMismatch(h, check.Height())
	}

	// and now verify it matches our validators
	err = cert.Certify(check)
	return check, nil
}

// ParseHexKey parses the key flag as hex and converts to bytes or returns error
// argname is used to customize the error message
func ParseHexKey(args []string, argname string) ([]byte, error) {
	if len(args) == 0 {
		return nil, errors.Errorf("Missing required argument [%s]", argname)
	}
	if len(args) > 1 {
		return nil, errors.Errorf("Only accepts one argument [%s]", argname)
	}
	rawkey := args[0]
	if rawkey == "" {
		return nil, errors.Errorf("[%s] argument must be non-empty ", argname)
	}
	// with tx, we always just parse key as hex and use to lookup
	return proofs.ParseHexKey(rawkey)
}

func GetHeight() int {
	return viper.GetInt(FlagHeight)
}

type proof struct {
	Height uint64      `json:"height"`
	Data   interface{} `json:"data"`
}

// FoutputProof writes the output of wrapping height and info
// in the form {"data": <the_data>, "height": <the_height>}
// to the provider io.Writer
func FoutputProof(w io.Writer, v interface{}, height uint64) error {
	wrap := &proof{height, v}
	blob, err := data.ToJSON(wrap)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "%s\n", blob)
	return err
}

// OutputProof prints the proof to stdout
// reuse this for printing proofs and we should enhance this for text/json,
// better presentation of height
func OutputProof(data interface{}, height uint64) error {
	return FoutputProof(os.Stdout, data, height)
}
