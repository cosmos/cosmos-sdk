package stake

import (
	"encoding/json"

	crypto "github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/state"
)

// nolint
var (
	// Keys for store prefixes
	CandidatesPubKeysKey = []byte{0x01} // key for all candidates' pubkeys
	ParamKey             = []byte{0x02} // key for global parameters relating to staking
	GlobalStateKey       = []byte{0x03} // key for global parameters relating to staking

	// Key prefixes
	CandidateKeyPrefix      = []byte{0x04} // prefix for each key to a candidate
	DelegatorBondKeyPrefix  = []byte{0x05} // prefix for each key to a delegator's bond
	DelegatorBondsKeyPrefix = []byte{0x06} // prefix for each key to a delegator's bond
)

// GetCandidateKey - get the key for the candidate with pubKey
func GetCandidateKey(pubKey crypto.PubKey) []byte {
	return append(CandidateKeyPrefix, pubKey.Bytes()...)
}

// GetDelegatorBondKey - get the key for delegator bond with candidate
func GetDelegatorBondKey(delegator sdk.Actor, candidate crypto.PubKey) []byte {
	return append(GetDelegatorBondKeyPrefix(delegator), candidate.Bytes()...)
}

// GetDelegatorBondKeyPrefix - get the prefix for a delegator for all candidates
func GetDelegatorBondKeyPrefix(delegator sdk.Actor) []byte {
	return append(DelegatorBondKeyPrefix, wire.BinaryBytes(&delegator)...)
}

// GetDelegatorBondsKey - get the key for list of all the delegator's bonds
func GetDelegatorBondsKey(delegator sdk.Actor) []byte {
	return append(DelegatorBondsKeyPrefix, wire.BinaryBytes(&delegator)...)
}

//---------------------------------------------------------------------

// Get the active list of all the candidate pubKeys and owners
func loadCandidatesPubKeys(store state.SimpleDB) (pubKeys []crypto.PubKey) {
	bytes := store.Get(CandidatesPubKeysKey)
	if bytes == nil {
		return
	}
	err := wire.ReadBinaryBytes(bytes, &pubKeys)
	if err != nil {
		panic(err)
	}
	return
}
func saveCandidatesPubKeys(store state.SimpleDB, pubKeys []crypto.PubKey) {
	b := wire.BinaryBytes(pubKeys)
	store.Set(CandidatesPubKeysKey, b)
}

// loadCandidates - get the active list of all candidates TODO replace with  multistore
func loadCandidates(store state.SimpleDB) (candidates Candidates) {
	pks := loadCandidatesPubKeys(store)
	for _, pk := range pks {
		candidates = append(candidates, loadCandidate(store, pk))
	}
	return
}

//---------------------------------------------------------------------

// loadCandidate - loads the candidate object for the provided pubkey
func loadCandidate(store state.SimpleDB, pubKey crypto.PubKey) *Candidate {
	if pubKey.Empty() {
		return nil
	}
	b := store.Get(GetCandidateKey(pubKey))
	if b == nil {
		return nil
	}
	candidate := new(Candidate)
	err := json.Unmarshal(b, candidate)
	if err != nil {
		panic(err) // This error should never occure big problem if does
	}
	return candidate
}

func saveCandidate(store state.SimpleDB, candidate *Candidate) {

	if !store.Has(GetCandidateKey(candidate.PubKey)) {
		// TODO to be replaced with iteration in the multistore?
		pks := loadCandidatesPubKeys(store)
		saveCandidatesPubKeys(store, append(pks, candidate.PubKey))
	}

	b, err := json.Marshal(*candidate)
	if err != nil {
		panic(err)
	}
	store.Set(GetCandidateKey(candidate.PubKey), b)
}

func removeCandidate(store state.SimpleDB, pubKey crypto.PubKey) {
	store.Remove(GetCandidateKey(pubKey))

	// TODO to be replaced with iteration in the multistore?
	pks := loadCandidatesPubKeys(store)
	for i := range pks {
		if pks[i].Equals(pubKey) {
			saveCandidatesPubKeys(store,
				append(pks[:i], pks[i+1:]...))
			break
		}
	}
}

//---------------------------------------------------------------------

// load the pubkeys of all candidates a delegator is delegated too
func loadDelegatorCandidates(store state.SimpleDB,
	delegator sdk.Actor) (candidates []crypto.PubKey) {

	candidateBytes := store.Get(GetDelegatorBondsKey(delegator))
	if candidateBytes == nil {
		return nil
	}

	err := wire.ReadBinaryBytes(candidateBytes, &candidates)
	if err != nil {
		panic(err)
	}
	return
}

//---------------------------------------------------------------------

func loadDelegatorBond(store state.SimpleDB,
	delegator sdk.Actor, candidate crypto.PubKey) *DelegatorBond {

	delegatorBytes := store.Get(GetDelegatorBondKey(delegator, candidate))
	if delegatorBytes == nil {
		return nil
	}

	bond := new(DelegatorBond)
	err := json.Unmarshal(delegatorBytes, bond)
	if err != nil {
		panic(err)
	}
	return bond
}

func saveDelegatorBond(store state.SimpleDB, delegator sdk.Actor, bond *DelegatorBond) {

	// if a new bond add to the list of bonds
	if loadDelegatorBond(store, delegator, bond.PubKey) == nil {
		pks := loadDelegatorCandidates(store, delegator)
		pks = append(pks, (*bond).PubKey)
		b := wire.BinaryBytes(pks)
		store.Set(GetDelegatorBondsKey(delegator), b)
	}

	// now actually save the bond
	b, err := json.Marshal(*bond)
	if err != nil {
		panic(err)
	}
	store.Set(GetDelegatorBondKey(delegator, bond.PubKey), b)
	//updateDelegatorBonds(store, delegator)
}

func removeDelegatorBond(store state.SimpleDB, delegator sdk.Actor, candidate crypto.PubKey) {

	// TODO use list queries on multistore to remove iterations here!
	// first remove from the list of bonds
	pks := loadDelegatorCandidates(store, delegator)
	for i, pk := range pks {
		if candidate.Equals(pk) {
			pks = append(pks[:i], pks[i+1:]...)
		}
	}
	b := wire.BinaryBytes(pks)
	store.Set(GetDelegatorBondsKey(delegator), b)

	// now remove the actual bond
	store.Remove(GetDelegatorBondKey(delegator, candidate))
	//updateDelegatorBonds(store, delegator)
}

//func updateDelegatorBonds(store state.SimpleDB,
//delegator sdk.Actor) {

//var bonds []*DelegatorBond

//prefix := GetDelegatorBondKeyPrefix(delegator)
//l := len(prefix)
//delegatorsBytes := store.List(prefix,
//append(prefix[:l-1], (prefix[l-1]+1)), loadParams(store).MaxVals)

//for _, delegatorBytesModel := range delegatorsBytes {
//delegatorBytes := delegatorBytesModel.Value
//if delegatorBytes == nil {
//continue
//}

//bond := new(DelegatorBond)
//err := wire.ReadBinaryBytes(delegatorBytes, bond)
//if err != nil {
//panic(err)
//}
//bonds = append(bonds, bond)
//}

//if len(bonds) == 0 {
//store.Remove(GetDelegatorBondsKey(delegator))
//return
//}

//b := wire.BinaryBytes(bonds)
//store.Set(GetDelegatorBondsKey(delegator), b)
//}

//_______________________________________________________________________

// load/save the global staking params
func loadParams(store state.SimpleDB) (params Params) {
	b := store.Get(ParamKey)
	if b == nil {
		return defaultParams()
	}

	err := json.Unmarshal(b, &params)
	if err != nil {
		panic(err) // This error should never occure big problem if does
	}

	return
}
func saveParams(store state.SimpleDB, params Params) {
	b, err := json.Marshal(params)
	if err != nil {
		panic(err)
	}
	store.Set(ParamKey, b)
}

//_______________________________________________________________________

// load/save the global staking state
func loadGlobalState(store state.SimpleDB) (gs *GlobalState) {
	b := store.Get(GlobalStateKey)
	if b == nil {
		return initialGlobalState()
	}
	gs = new(GlobalState)
	err := json.Unmarshal(b, gs)
	if err != nil {
		panic(err) // This error should never occure big problem if does
	}
	return
}
func saveGlobalState(store state.SimpleDB, gs *GlobalState) {
	b, err := json.Marshal(*gs)
	if err != nil {
		panic(err)
	}
	store.Set(GlobalStateKey, b)
}
