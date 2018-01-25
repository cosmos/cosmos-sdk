package stake

import (
	"bytes"
	"encoding/hex"
	"testing"

	sdkstore "github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	crypto "github.com/tendermint/go-crypto"
	dbm "github.com/tendermint/tmlibs/db"
)

func newPubKey(pk string) (res crypto.PubKey, err error) {
	pkBytes, err := hex.DecodeString(pk)
	if err != nil {
		return
	}
	//res, err = crypto.PubKeyFromBytes(pkBytes)
	var pkEd crypto.PubKeyEd25519
	copy(pkEd[:], pkBytes[:])
	return pkEd, nil
}

func TestState(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	db, err := dbm.NewGoLevelDB("basecoin", "basecoin-data")
	require.Nil(err)
	cacheSize := 10000
	numHistory := int64(100)
	stakeLoader := sdkstore.NewIAVLStoreLoader(db, cacheSize, numHistory)
	var stakeStoreKey = types.NewKVStoreKey("stake")
	multiStore := sdkstore.NewCommitMultiStore(db)
	multiStore.SetSubstoreLoader(stakeStoreKey, stakeLoader)
	multiStore.LoadLatestVersion()
	store := multiStore.GetKVStore(stakeStoreKey)
	cdc.RegisterInterface((*crypto.PubKey)(nil), nil)
	cdc.RegisterConcrete(crypto.PubKeyEd25519{}, "crypto/PubKeyEd25519", nil)

	//delegator := crypto.Address{[]byte("addressdelegator")}
	//validator := crypto.Address{[]byte("addressvalidator")}
	delegator := []byte("addressdelegator")
	validator := []byte("addressvalidator")

	pk, err := newPubKey("0B485CFC0EECC619440448436F8FC9DF40566F2369E72400281454CB552AFB57")
	require.Nil(err)

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
	resPks := loadCandidatesPubKeys(store)
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

	// also test that the pubkey has been added to pubkey list
	resPks = loadCandidatesPubKeys(store)
	require.Equal(1, len(resPks))
	assert.Equal(pk, resPks[0])

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

func candidatesFromActors(actors []sdk.Actor, amts []int) (candidates Candidates) {
	for i := 0; i < len(actors); i++ {
		c := &Candidate{
			PubKey:      pks[i],
			Owner:       actors[i],
			Shares:      int64(amts[i]),
			VotingPower: int64(amts[i]),
		}
		candidates = append(candidates, c)
	}

	return
}

func TestGetValidators(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	N := 5
	actors := newActors(N)
	candidates := candidatesFromActors(actors, []int{400, 200, 0, 0, 0})

	validators := candidates.Validators()
	require.Equal(2, len(validators))
	assert.Equal(candidates[0].PubKey, validators[0].PubKey)
	assert.Equal(candidates[1].PubKey, validators[1].PubKey)
}
