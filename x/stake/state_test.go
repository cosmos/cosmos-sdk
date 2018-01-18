package stake

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/modules/auth"
	"github.com/cosmos/cosmos-sdk/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tmlibs/rational"
)

func TestState(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	db, err := dbm.NewGoLevelDB("basecoin", "basecoin-data")
	require.Nil(err)
	mainLoader := store.NewIAVLStoreLoader(int64(100), 10000, numHistory)
	var mainStoreKey = sdk.NewKVStoreKey("main")
	multiStore := store.NewCommitMultiStore(db)
	multiStore.SetSubstoreLoader(mainStoreKey, mainLoader)
	var store = auth.NewAccountStore(mainStoreKey, bcm.AppAccountCodec{})

	delegator := sdk.Actor{"testChain", "testapp", []byte("addressdelegator")}
	validator := sdk.Actor{"testChain", "testapp", []byte("addressvalidator")}

	pk := newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB57")

	//----------------------------------------------------------------------
	// Candidate checks

	// XXX expand to include both liabilities and assets use/test all candidate fields
	candidate := &Candidate{
		Owner:       validator,
		PubKey:      pk,
		Assets:      rational.New(9),
		Liabilities: rational.New(9),
		VotingPower: rational.Zero,
	}

	candidatesEqual := func(c1, c2 *Candidate) bool {
		return c1.Status == c2.Status &&
			c1.PubKey.Equals(c2.PubKey) &&
			c1.Owner.Equals(c2.Owner) &&
			c1.Assets.Equal(c2.Assets) &&
			c1.Liabilities.Equal(c2.Liabilities) &&
			c1.VotingPower.Equal(c2.VotingPower) &&
			c1.Description == c2.Description
	}

	// check the empty store first
	resCand := loadCandidate(store, pk)
	assert.Nil(resCand)
	resPks := loadCandidatesPubKeys(store)
	assert.Zero(len(resPks))

	// set and retrieve a record
	saveCandidate(store, candidate)
	resCand = loadCandidate(store, pk)
	assert.True(candidatesEqual(candidate, resCand))

	// modify a records, save, and retrieve
	candidate.Liabilities = rational.New(99)
	saveCandidate(store, candidate)
	resCand = loadCandidate(store, pk)
	assert.True(candidatesEqual(candidate, resCand))

	// also test that the pubkey has been added to pubkey list
	resPks = loadCandidatesPubKeys(store)
	require.Equal(1, len(resPks))
	assert.Equal(pk, resPks[0])

	//----------------------------------------------------------------------
	// Bond checks

	bond := &DelegatorBond{
		PubKey: pk,
		Shares: rational.New(9),
	}

	bondsEqual := func(b1, b2 *DelegatorBond) bool {
		return b1.PubKey.Equals(b2.PubKey) &&
			b1.Shares.Equal(b2.Shares)
	}

	//check the empty store first
	resBond := loadDelegatorBond(store, delegator, pk)
	assert.Nil(resBond)

	//Set and retrieve a record
	saveDelegatorBond(store, delegator, bond)
	resBond = loadDelegatorBond(store, delegator, pk)
	assert.True(bondsEqual(bond, resBond))

	//modify a records, save, and retrieve
	bond.Shares = rational.New(99)
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
