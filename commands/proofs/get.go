package proofs

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/viper"

	wire "github.com/tendermint/go-wire"
	"github.com/tendermint/go-wire/data"

	"github.com/tendermint/tendermint/rpc/client"

	lc "github.com/tendermint/light-client"
	"github.com/tendermint/basecoin/commands"
	"github.com/tendermint/light-client/proofs"
)

// GetAndParseAppProof does most of the work of the query commands, but is quite
// opinionated, so if you want more control, set up the items and call GetProof
// directly.  Notably, it always uses go-wire.ReadBinaryBytes to deserialize,
// and Height and Node from standard flags.
//
// It will try to get the proof for the given key.  If it is successful,
// it will return the proof and also unserialize proof.Data into the data
// argument (so pass in a pointer to the appropriate struct)
func GetAndParseAppProof(key []byte, data interface{}) (lc.Proof, error) {
	height := GetHeight()
	node := commands.GetNode()
	prover := proofs.NewAppProver(node)

	proof, err := GetProof(node, prover, key, height)
	if err != nil {
		return proof, err
	}

	err = wire.ReadBinaryBytes(proof.Data(), data)
	return proof, err
}

// GetProof performs the get command directly from the proof (not from the CLI)
func GetProof(node client.Client, prover lc.Prover, key []byte, height int) (proof lc.Proof, err error) {
	proof, err = prover.Get(key, uint64(height))
	if err != nil {
		return
	}
	ph := int(proof.BlockHeight())
	// here is the certifier, root of all knowledge
	cert, err := commands.GetCertifier()
	if err != nil {
		return
	}

	// get and validate a signed header for this proof

	// FIXME: cannot use cert.GetByHeight for now, as it also requires
	// Validators and will fail on querying tendermint for non-current height.
	// When this is supported, we should use it instead...
	client.WaitForHeight(node, ph, nil)
	commit, err := node.Commit(ph)
	if err != nil {
		return
	}
	check := lc.Checkpoint{
		Header: commit.Header,
		Commit: commit.Commit,
	}
	err = cert.Certify(check)
	if err != nil {
		return
	}

	// validate the proof against the certified header to ensure data integrity
	err = proof.Validate(check)
	if err != nil {
		return
	}

	return proof, err
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
	return viper.GetInt(heightFlag)
}

type proof struct {
	Height uint64      `json:"height"`
	Data   interface{} `json:"data"`
}

// OutputProof prints the proof to stdout
// reuse this for printing proofs and we should enhance this for text/json,
// better presentation of height
func OutputProof(info interface{}, height uint64) error {
	wrap := proof{height, info}
	res, err := data.ToJSON(wrap)
	if err != nil {
		return err
	}
	fmt.Println(string(res))
	return nil
}
