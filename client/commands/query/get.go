package query

import (
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/viper"

	wire "github.com/tendermint/go-wire"
	"github.com/tendermint/go-wire/data"
	lc "github.com/tendermint/light-client"
	"github.com/tendermint/light-client/certifiers"
	"github.com/tendermint/light-client/proofs"
	"github.com/tendermint/merkleeyes/iavl"
	"github.com/tendermint/tendermint/rpc/client"

	"github.com/cosmos/cosmos-sdk/client/commands"
)

// GetParsed does most of the work of the query commands, but is quite
// opinionated, so if you want more control about parsing, call Get
// directly.
//
// It will try to get the proof for the given key.  If it is successful,
// it will return the height and also unserialize proof.Data into the data
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

// Get queries the given key and returns the value stored there and the
// height we checked at.
//
// If prove is true (and why shouldn't it be?),
// the data is fully verified before returning.  If prove is false,
// we just repeat whatever any (potentially malicious) node gives us.
// Only use that if you are running the full node yourself,
// and it is localhost or you have a secure connection (not HTTP)
func Get(key []byte, prove bool) (data.Bytes, uint64, error) {
	if !prove {
		node := commands.GetNode()
		resp, err := node.ABCIQuery("/key", key, false)
		return data.Bytes(resp.Value), resp.Height, err
	}
	val, h, _, _, err := GetWithProof(key)
	return val, h, err
}

// GetWithProof returns the values stored under a given key at the named
// height as in Get.  Additionally, it will return a validated merkle
// proof for the key-value pair if it exists, and all checks pass.
func GetWithProof(key []byte) (data.Bytes, uint64,
	*iavl.KeyExistsProof, *iavl.KeyNotExistsProof, error) {

	node := commands.GetNode()
	cert, err := commands.GetCertifier()
	if err != nil {
		return nil, 0, nil, nil, err
	}
	return getWithProof(key, node, cert)
}

func getWithProof(key []byte, node client.Client, cert certifiers.Certifier) (
	val data.Bytes, height uint64, eproof *iavl.KeyExistsProof, neproof *iavl.KeyNotExistsProof, err error) {

	resp, err := node.ABCIQuery("/key", key, true)
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

	check, err := getCertifiedCheckpoint(int(resp.Height), node, cert)
	if err != nil {
		return
	}

	if len(resp.Value) > 0 {
		// The key was found, construct a proof of existence.
		eproof = new(iavl.KeyExistsProof)
		err = wire.ReadBinaryBytes(resp.Proof, &eproof)
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
	} else {
		// The key wasn't found, construct a proof of non-existence.
		neproof = new(iavl.KeyNotExistsProof)
		err = wire.ReadBinaryBytes(resp.Proof, &neproof)
		if err != nil {
			err = errors.Wrap(err, "Error reading proof")
			return
		}

		// Validate the proof against the certified header to ensure data integrity.
		err = neproof.Verify(resp.Key, check.Header.AppHash)
		if err != nil {
			err = errors.Wrap(err, "Couldn't verify proof")
			return
		}
		err = lc.ErrNoData()
	}

	height = resp.Height
	return
}

// getCertifiedCheckpoint gets the signed header for a given height
// and certifies it.  Returns error if unable to get a proven header.
func getCertifiedCheckpoint(h int, node client.Client,
	cert certifiers.Certifier) (empty lc.Checkpoint, err error) {

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

	// validate downloaded checkpoint with our request and trust store.
	if check.Height() != h {
		return empty, lc.ErrHeightMismatch(h, check.Height())
	}
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

// GetHeight reads the viper config for the query height
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
