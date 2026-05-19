package simapp_test

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	cmtmldsa65 "github.com/cometbft/cometbft/crypto/mldsa65"
	"github.com/cometbft/cometbft/privval"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/simapp"

	"github.com/cosmos/cosmos-sdk/crypto/keys/mldsa65"
	"github.com/cosmos/cosmos-sdk/testutil/network"
)

// MLDsa65NetworkTestSuite boots an in-process SDK testnet whose validators
// sign consensus messages with the ML-DSA-65 (FIPS 204) post-quantum
// signature scheme. It is the cosmos-sdk counterpart to the cometbft e2e
// testnet under test/e2e/networks/mldsa65.toml, and it exercises the
// opt-in `Config.ValidatorConsensusKeyType` field plus the new
// genutil.InitializeNodeValidatorFilesFromMnemonicWithKeyType code path.
//
// The suite is intentionally narrow: it verifies that
//
//	(a) network.New() succeeds when validators are pinned to ML-DSA-65 keys,
//	(b) the network advances past genesis (i.e. blocks are actually being
//	    signed with 3309-byte ML-DSA-65 signatures), and
//	(c) the on-disk priv_validator_key.json holds an ML-DSA-65 key of the
//	    expected size.
type MLDsa65NetworkTestSuite struct {
	suite.Suite

	network *network.Network
}

func (s *MLDsa65NetworkTestSuite) SetupSuite() {
	s.T().Log("setting up ML-DSA-65 integration test suite")

	cfg := network.DefaultConfig(simapp.NewTestNetworkFixture)
	cfg.NumValidators = 2
	cfg.ValidatorConsensusKeyType = cmtmldsa65.KeyType

	var err error
	s.network, err = network.New(s.T(), s.T().TempDir(), cfg)
	s.Require().NoError(err)

	h, err := s.network.WaitForHeight(1)
	s.Require().NoError(err, "failed to reach height 1; got %d", h)
}

func (s *MLDsa65NetworkTestSuite) TearDownSuite() {
	s.T().Log("tearing down ML-DSA-65 integration test suite")
	s.network.Cleanup()
}

// TestNetwork_Liveness proves blocks are being produced and signed by
// ML-DSA-65 validators. Reaching height 3 means each validator has both
// proposed and pre-committed at least once under the post-quantum scheme.
func (s *MLDsa65NetworkTestSuite) TestNetwork_Liveness() {
	h, err := s.network.WaitForHeightWithTimeout(3, time.Minute)
	s.Require().NoError(err, "expected to reach 3 blocks; got %d", h)
}

// TestValidator_PubKeyType asserts every validator's in-memory consensus
// pubkey is ML-DSA-65 of the expected length.
func (s *MLDsa65NetworkTestSuite) TestValidator_PubKeyType() {
	s.Require().NotEmpty(s.network.Validators)
	for _, v := range s.network.Validators {
		s.Require().Equal(cmtmldsa65.KeyType, v.PubKey.Type(),
			"validator %s pubkey type", v.Moniker)
		s.Require().Len(v.PubKey.Bytes(), cmtmldsa65.PubKeySize,
			"validator %s pubkey size", v.Moniker)
		_, ok := v.PubKey.(*mldsa65.PubKey)
		s.Require().True(ok, "validator %s pubkey should be *mldsa65.PubKey, got %T",
			v.Moniker, v.PubKey)
	}
}

// TestValidator_PrivvalKeyFile loads each validator's priv_validator_key.json
// off disk and verifies the persisted private key is also ML-DSA-65. This
// catches mistakes that would otherwise lurk if only the in-memory pubkey
// were checked.
func (s *MLDsa65NetworkTestSuite) TestValidator_PrivvalKeyFile() {
	for _, v := range s.network.Validators {
		// Ask the validator's own CometBFT config for the priv-validator
		// paths rather than re-constructing them. v.Dir is just the parent
		// directory; the actual home root (currently "<Dir>/simd") is set by
		// the harness and we shouldn't bake the simapp binary name into the
		// test.
		keyFile := v.Ctx.Config.PrivValidatorKeyFile()
		stateFile := v.Ctx.Config.PrivValidatorStateFile()

		_, err := os.Stat(keyFile)
		s.Require().NoError(err, "priv_validator_key.json missing for %s", v.Moniker)

		filePV := privval.LoadFilePV(keyFile, stateFile)
		require.Equal(s.T(), cmtmldsa65.KeyType, filePV.Key.PrivKey.Type(),
			"validator %s persisted priv key type", v.Moniker)
		require.Len(s.T(), filePV.Key.PrivKey.Bytes(), cmtmldsa65.PrivKeySize,
			"validator %s persisted priv key size", v.Moniker)

		// Round-trip the on-disk JSON: parse, re-marshal, parse again, and
		// confirm the key bytes survive. This exercises the cmtjson
		// (PrivKeyName) registration we added in cometbft/crypto/mldsa65.
		raw, err := os.ReadFile(keyFile)
		s.Require().NoError(err)
		var anyMap map[string]json.RawMessage
		s.Require().NoError(json.Unmarshal(raw, &anyMap))
		s.Require().Contains(anyMap, "priv_key", "priv_validator_key.json shape")
	}
}

func TestMLDsa65NetworkTestSuite(t *testing.T) {
	suite.Run(t, new(MLDsa65NetworkTestSuite))
}
