package context

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/store"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/merkle"
	"github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/log"
	tmlite "github.com/tendermint/tendermint/lite"
	tmliteErr "github.com/tendermint/tendermint/lite/errors"
	tmliteProxy "github.com/tendermint/tendermint/lite/proxy"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	tmtypes "github.com/tendermint/tendermint/types"
)

var (
	verifier tmlite.Verifier
)

// Verify verifies the consensus proof at given height.
func (ctx CLIContext) Verify(height int64) (tmtypes.SignedHeader, error) {
	check, err := tmliteProxy.GetCertifiedCommit(height, ctx.RPCClient, ctx.Verifier)
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
	prt := store.DefaultProofRuntime()

	// TODO: Better convention for path?
	storeName, err := parseQueryStorePath(queryPath)
	if err != nil {
		return err
	}

	kp := merkle.KeyPath{}
	kp = kp.AppendKey([]byte(storeName), merkle.KeyEncodingURL)
	kp = kp.AppendKey(resp.Key, merkle.KeyEncodingURL)

	err = prt.VerifyValue(resp.Proof, commit.Header.AppHash, kp.String(), resp.Value)
	if err != nil {
		return errors.Wrap(err, "failed to prove merkle proof")
	}

	return nil
}

func createVerifier() tmlite.Verifier {
	if !viper.IsSet(client.FlagTrustNode) || viper.GetBool(client.FlagTrustNode) {
		return nil
	}

	chainID := viper.GetString(client.FlagChainID)
	home := viper.GetString(cli.HomeFlag)
	nodeURI := viper.GetString(client.FlagNode)

	var errMsg bytes.Buffer

	if chainID == "" {
		errMsg.WriteString("--chain-id ")
	}

	if home == "" {
		errMsg.WriteString("--home ")
	}

	if nodeURI == "" {
		errMsg.WriteString("--node ")
	}

	if errMsg.Len() != 0 {
		fmt.Printf("Must specify these options: %s when --trust-node is false\n", errMsg.String())
		os.Exit(1)
	}

	node := rpcclient.NewHTTP(nodeURI, "/websocket")
	cacheSize := 10 // TODO: determine appropriate cache size1
	verifier, err := tmliteProxy.NewVerifier(
		chainID, filepath.Join(home, ".gaialite"),
		node, log.NewNopLogger(), cacheSize)

	if err != nil {
		fmt.Printf("Create verifier failed: %s\n", err.Error())
		fmt.Printf("Please check network connection and verify the address of the node to connect to\n")
		os.Exit(1)
	}

	return verifier
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
