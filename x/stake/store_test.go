package stake

import (
	"bytes"
	"testing"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dbm "github.com/tendermint/tmlibs/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	crypto "github.com/tendermint/go-crypto"
)

func initTestStore(t *testing.T) sdk.KVStore {
	// Capabilities key to access the main KVStore.
	db, err := dbm.NewGoLevelDB("stake", "data")
	require.Nil(t, err)
	stakeStoreKey := sdk.NewKVStoreKey("stake")
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(stakeStoreKey, sdk.StoreTypeIAVL, db)
	ms.LoadLatestVersion()
	return ms.GetKVStore(stakeStoreKey)
}

func TestState(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	store := initTestStore(t)
	cdc.RegisterInterface((*crypto.PubKey)(nil), nil)
	cdc.RegisterConcrete(crypto.PubKeyEd25519{}, "crypto/PubKeyEd25519", nil)

	//delegator := crypto.Address{[]byte("addressdelegator")}
	//validator := crypto.Address{[]byte("addressvalidator")}
	delegator := []byte("addressdelegator")
	validator := []byte("addressvalidator")

	pk := newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB57")

	//----------------------------------------------------------------------
	// Candidate checks

	// XXX expand to include both liabilities and assets use/test all candidate fields
	candidate := &Candidate{
		Owner:       validator,
		PubKey:      pk,
		Assets:      9, //rational.New(9),
		Liabilities: 9, // rational.New(9),
		VotingPower: 0, //rational.Zero,
	}

	candidatesEqual := func(c1, c2 *Candidate) bool {
		return c1.Status == c2.Status &&
			c1.PubKey.Equals(c2.PubKey) &&
			bytes.Equal(c1.Owner, c2.Owner) &&
			c1.Assets == c2.Assets &&
			c1.Liabilities == c2.Liabilities &&
			c1.VotingPower == c2.VotingPower &&
			c1.Description == c2.Description
	}

	// check the empty store first
	resCand := loadCandidate(store, pk)
	assert.Nil(resCand)
	resPks := loadCandidates(store)
	assert.Zero(len(resPks))

	// set and retrieve a record
	saveCandidate(store, candidate)
	resCand = loadCandidate(store, pk)
	assert.True(candidatesEqual(candidate, resCand))

	// modify a records, save, and retrieve
	candidate.Liabilities = 99 //rational.New(99)
	saveCandidate(store, candidate)
	resCand = loadCandidate(store, pk)
	assert.True(candidatesEqual(candidate, resCand))

	store.Write()
	// also test that the pubkey has been added to pubkey list
	resPks = loadCandidates(store)
	require.Equal(1, len(resPks))
	assert.Equal(pk, resPks[0].PubKey)

	//----------------------------------------------------------------------
	// Bond checks

	bond := &DelegatorBond{
		PubKey: pk,
		Shares: 9, // rational.New(9),
	}

	bondsEqual := func(b1, b2 *DelegatorBond) bool {
		return b1.PubKey.Equals(b2.PubKey) &&
			b1.Shares == b2.Shares
	}

	//check the empty store first
	resBond := loadDelegatorBond(store, delegator, pk)
	assert.Nil(resBond)

	//Set and retrieve a record
	saveDelegatorBond(store, delegator, bond)
	resBond = loadDelegatorBond(store, delegator, pk)
	assert.True(bondsEqual(bond, resBond))

	//modify a records, save, and retrieve
	bond.Shares = 99 //rational.New(99)
	saveDelegatorBond(store, delegator, bond)
	resBond = loadDelegatorBond(store, delegator, pk)
	assert.True(bondsEqual(bond, resBond))

	//----------------------------------------------------------------------
	// Param checks

	params := defaultParams()

	//check that the empty store loads the default
	resParams := loadParams(store)
	assert.Equal(params, resParams)

	//modify a params, save, and retrieve
	params.MaxVals = 777
	saveParams(store, params)
	resParams = loadParams(store)
	assert.Equal(params, resParams)
}

func TestGetValidators(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	store := initTestStore(t)
	N := 5
	addrs := newAddrs(N)
	candidatesFromActors(store, addrs, []int{400, 200, 0, 0, 0})

	validators := getValidators(store, 5)
	require.Equal(2, len(validators))
	assert.Equal(pks[0], validators[0].PubKey)
	assert.Equal(pks[1], validators[1].PubKey)
}
