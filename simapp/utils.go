//nolint
package simapp

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"time"

	cmn "github.com/tendermint/tendermint/libs/common"

	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/tendermint/tendermint/crypto/secp256k1"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authsim "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	distributionsim "github.com/cosmos/cosmos-sdk/x/distribution/simulation"
	"github.com/cosmos/cosmos-sdk/x/genaccounts"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govsim "github.com/cosmos/cosmos-sdk/x/gov/simulation"
	"github.com/cosmos/cosmos-sdk/x/mint"
	mintsim "github.com/cosmos/cosmos-sdk/x/mint/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingsim "github.com/cosmos/cosmos-sdk/x/slashing/simulation"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingsim "github.com/cosmos/cosmos-sdk/x/staking/simulation"
	"github.com/cosmos/cosmos-sdk/x/supply"
	supplysim "github.com/cosmos/cosmos-sdk/x/supply/simulation"
)

var (
	genesisFile        string
	paramsFile         string
	exportParamsPath   string
	exportParamsHeight int
	exportStatePath    string
	exportStatsPath    string
	seed               int64
	numBlocks          int
	blockSize          int
	enabled            bool
	verbose            bool
	lean               bool
	commit             bool
	period             int
	onOperation        bool // TODO Remove in favor of binary search for invariant violation
	allInvariants      bool
	genesisTime        int64
)

// NewSimAppUNSAFE is used for debugging purposes only.
//
// NOTE: to not use this function with non-test code
func NewSimAppUNSAFE(logger log.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool,
	invCheckPeriod uint, baseAppOptions ...func(*baseapp.BaseApp),
) (gapp *SimApp, keyMain, keyStaking *sdk.KVStoreKey, stakingKeeper staking.Keeper) {

	gapp = NewSimApp(logger, db, traceStore, loadLatest, invCheckPeriod, baseAppOptions...)
	return gapp, gapp.keyMain, gapp.keyStaking, gapp.stakingKeeper
}

// AppStateFromGenesisFileFn util function to generate the genesis AppState
// from a genesis.json file
func AppStateFromGenesisFileFn(
	r *rand.Rand, _ []simulation.Account, _ time.Time,
) (json.RawMessage, []simulation.Account, string) {

	var genesis tmtypes.GenesisDoc
	cdc := MakeCodec()

	bytes, err := ioutil.ReadFile(genesisFile)
	if err != nil {
		panic(err)
	}

	cdc.MustUnmarshalJSON(bytes, &genesis)

	var appState GenesisState
	cdc.MustUnmarshalJSON(genesis.AppState, &appState)

	accounts := genaccounts.GetGenesisStateFromAppState(cdc, appState)

	var newAccs []simulation.Account
	for _, acc := range accounts {
		// Pick a random private key, since we don't know the actual key
		// This should be fine as it's only used for mock Tendermint validators
		// and these keys are never actually used to sign by mock Tendermint.
		privkeySeed := make([]byte, 15)
		r.Read(privkeySeed)

		privKey := secp256k1.GenPrivKeySecp256k1(privkeySeed)
		newAccs = append(newAccs, simulation.Account{privKey, privKey.PubKey(), acc.Address})
	}

	return genesis.AppState, newAccs, genesis.ChainID
}

// GetSimulationLog unmarshals the KVPair's Value to the corresponding type based on the
// each's module store key and the prefix bytes of the KVPair's key.
func GetSimulationLog(storeName string, cdcA, cdcB *codec.Codec, kvs []cmn.KVPair) (log string) {
	var kvA, kvB cmn.KVPair
	for i := 0; i < len(kvs); i += 2 {
		kvA = kvs[i]
		kvB = kvs[i+1]

		if len(kvA.Value) == 0 && len(kvB.Value) == 0 {
			// skip if the value doesn't have any bytes
			continue
		}

		switch storeName {
		case auth.StoreKey:
			log += authsim.DecodeStore(cdcA, cdcB, kvA, kvB)
		case mint.StoreKey:
			log += mintsim.DecodeStore(cdcA, cdcB, kvA, kvB)
		case staking.StoreKey:
			log += stakingsim.DecodeStore(cdcA, cdcB, kvA, kvB)
		case slashing.StoreKey:
			log += slashingsim.DecodeStore(cdcA, cdcB, kvA, kvB)
		case gov.StoreKey:
			log += govsim.DecodeStore(cdcA, cdcB, kvA, kvB)
		case distribution.StoreKey:
			log += distributionsim.DecodeStore(cdcA, cdcB, kvA, kvB)
		case supply.StoreKey:
			log += supplysim.DecodeStore(cdcA, cdcB, kvA, kvB)
		default:
			log += fmt.Sprintf("store A %X => %X\nstore B %X => %X\n", kvA.Key, kvA.Value, kvB.Key, kvB.Value)
		}
	}

	return
}
