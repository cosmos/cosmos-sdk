package stake

import (
	crypto "github.com/tendermint/go-crypto"
	wire "github.com/tendermint/go-wire"

	"github.com/cosmos/cosmos-sdk/types"
)

//nolint
var (
	// Keys for store prefixes
	CandidatesPubKeysKey = []byte{0x01} // key for all candidates' pubkeys
	ParamKey             = []byte{0x02} // key for global parameters relating to staking
	GlobalStateKey       = []byte{0x03} // key for global parameters relating to staking

	// Key prefixes
	CandidateKeyPrefix        = []byte{0x04} // prefix for each key to a candidate
	ValidatorKeyPrefix        = []byte{0x05} // prefix for each key to a candidate
	ValidatorUpdatesKeyPrefix = []byte{0x06} // prefix for each key to a candidate
	DelegatorBondKeyPrefix    = []byte{0x07} // prefix for each key to a delegator's bond
	DelegatorBondsKeyPrefix   = []byte{0x08} // prefix for each key to a delegator's bond
)

// GetCandidateKey - get the key for the candidate with pubKey
func GetCandidateKey(pubKey crypto.PubKey) []byte {
	return append(CandidateKeyPrefix, pubKey.Bytes()...)
}

// GetValidatorKey - get the key for the validator used in the power-store
func GetValidatorKey(pubKey crypto.PubKey, power types.Rational) []byte {
	b, _ := cdc.MarshalJSON(power)                                     // TODO need to handle error here?
	return append(ValidatorKeyPrefix, append(b, pubKey.Bytes()...)...) // TODO does this need prefix if its in its own store
}

// GetValidatorUpdatesKey - get the key for the validator used in the power-store
func GetValidatorUpdatesKey(pubKey crypto.PubKey) []byte {
	return append(ValidatorUpdatesKeyPrefix, pubKey.Bytes()...) // TODO does this need prefix if its in its own store
}

// GetDelegatorBondKey - get the key for delegator bond with candidate
func GetDelegatorBondKey(delegator crypto.Address, candidate crypto.PubKey) []byte {
	return append(GetDelegatorBondKeyPrefix(delegator), candidate.Bytes()...)
}

// GetDelegatorBondKeyPrefix - get the prefix for a delegator for all candidates
func GetDelegatorBondKeyPrefix(delegator crypto.Address) []byte {
	res, err := cdc.MarshalJSON(&delegator)
	if err != nil {
		panic(err)
	}
	return append(DelegatorBondKeyPrefix, res...)
}

// GetDelegatorBondsKey - get the key for list of all the delegator's bonds
func GetDelegatorBondsKey(delegator crypto.Address) []byte {
	res, err := cdc.MarshalJSON(&delegator)
	if err != nil {
		panic(err)
	}
	return append(DelegatorBondsKeyPrefix, res...)
}

//___________________________________________________________________________

// keeper of the staking store
type Keeper struct {
	store types.KVStore
	cdc   *wire.Codec
}

func NewKeeper(ctx sdk.Context, key sdk.StoreKey) Keeper {
	cdc := wire.NewCodec()
	cdc.RegisterInterface((*types.Rational)(nil), nil) // XXX make like crypto.RegisterWire()
	cdc.RegisterConcrete(types.Rat{}, "rat", nil)
	crypto.RegisterWire(cdc)

	return StakeKeeper{
		store: ctx.KVStore(k.key),
		cdc:   cdc,
	}
}

func (k Keeper) loadCandidate(pubKey crypto.PubKey) *Candidate {
	b := k.store.Get(GetCandidateKey(pubKey))
	if b == nil {
		return nil
	}
	candidate := new(Candidate)
	err := cdc.UnmarshalJSON(b, candidate)
	if err != nil {
		panic(err) // This error should never occur big problem if does
	}
	return candidate
}

func (k Keeper) saveCandidate(candidate *Candidate) {

	// XXX should only remove validator if we know candidate is a validator
	removeValidator(k.store, candidate.PubKey)
	validator := &Validator{candidate.PubKey, candidate.VotingPower}
	updateValidator(k.store, validator)

	b, err := cdc.MarshalJSON(*candidate)
	if err != nil {
		panic(err)
	}
	k.store.Set(GetCandidateKey(candidate.PubKey), b)
}

func (k Keeper) removeCandidate(pubKey crypto.PubKey) {

	// XXX should only remove validator if we know candidate is a validator
	removeValidator(k.store, pubKey)
	k.store.Delete(GetCandidateKey(pubKey))
}

//___________________________________________________________________________

//func loadValidator(k.store types.KVStore, pubKey crypto.PubKey, votingPower types.Rational) *Validator {
//b := k.store.Get(GetValidatorKey(pubKey, votingPower))
//if b == nil {
//return nil
//}
//validator := new(Validator)
//err := cdc.UnmarshalJSON(b, validator)
//if err != nil {
//panic(err) // This error should never occur big problem if does
//}
//return validator
//}

// updateValidator - update a validator and create accumulate any changes
// in the changed validator substore
func (k Keeper) updateValidator(validator *Validator) {

	b, err := cdc.MarshalJSON(*validator)
	if err != nil {
		panic(err)
	}

	// add to the validators to update list if necessary
	k.store.Set(GetValidatorUpdatesKey(validator.PubKey), b)

	// update the list ordered by voting power
	k.store.Set(GetValidatorKey(validator.PubKey, validator.VotingPower), b)
}

func (k Keeper) removeValidator(pubKey crypto.PubKey) {

	//add validator with zero power to the validator updates
	b, err := cdc.MarshalJSON(Validator{pubKey, types.ZeroRat})
	if err != nil {
		panic(err)
	}
	k.store.Set(GetValidatorUpdatesKey(pubKey), b)

	// now actually delete from the validator set
	candidate := loadCandidate(k.store, pubKey)
	if candidate != nil {
		k.store.Delete(GetValidatorKey(pubKey, candidate.VotingPower))
	}
}

// get the most recent updated validator set from the Candidates. These bonds
// are already sorted by VotingPower from the UpdateVotingPower function which
// is the only function which is to modify the VotingPower
func (k Keeper) getValidators(maxVal int) (validators []Validator) {

	iterator := k.store.Iterator(subspace(ValidatorKeyPrefix)) //smallest to largest

	validators = make([]Validator, maxVal)
	for i := 0; ; i++ {
		if !iterator.Valid() || i > maxVal {
			iterator.Close()
			break
		}
		valBytes := iterator.Value()
		var val Validator
		err := cdc.UnmarshalJSON(valBytes, &val)
		if err != nil {
			panic(err)
		}
		validators[i] = val
		iterator.Next()
	}

	return
}

//_________________________________________________________________________

// get the most updated validators
func (k Keeper) getValidatorUpdates() (updates []Validator) {

	iterator := k.store.Iterator(subspace(ValidatorUpdatesKeyPrefix)) //smallest to largest

	for ; iterator.Valid(); iterator.Next() {
		valBytes := iterator.Value()
		var val Validator
		err := cdc.UnmarshalJSON(valBytes, &val)
		if err != nil {
			panic(err)
		}
		updates = append(updates, val)
	}

	iterator.Close()
	return
}

// remove all validator update entries
func (k Keeper) clearValidatorUpdates(maxVal int) {
	iterator := k.store.Iterator(subspace(ValidatorUpdatesKeyPrefix))
	for ; iterator.Valid(); iterator.Next() {
		k.store.Delete(iterator.Key()) // XXX write test for this, may need to be in a second loop
	}
	iterator.Close()
}

//---------------------------------------------------------------------

// loadCandidates - get the active list of all candidates TODO replace with  multistore
func (k Keeper) loadCandidates() (candidates Candidates) {

	iterator := k.store.Iterator(subspace(CandidateKeyPrefix))
	//iterator := k.store.Iterator(CandidateKeyPrefix, []byte(nil))
	//iterator := k.store.Iterator([]byte{}, []byte(nil))

	for ; iterator.Valid(); iterator.Next() {
		candidateBytes := iterator.Value()
		var candidate Candidate
		err := cdc.UnmarshalJSON(candidateBytes, &candidate)
		if err != nil {
			panic(err)
		}
		candidates = append(candidates, &candidate)
	}
	iterator.Close()
	return candidates
}

//_____________________________________________________________________

// load the pubkeys of all candidates a delegator is delegated too
func (k Keeper) loadDelegatorCandidates(delegator crypto.Address) (candidates []crypto.PubKey) {

	candidateBytes := k.store.Get(GetDelegatorBondsKey(delegator))
	if candidateBytes == nil {
		return nil
	}

	err := cdc.UnmarshalJSON(candidateBytes, &candidates)
	if err != nil {
		panic(err)
	}
	return
}

//_____________________________________________________________________

func (k Keeper) loadDelegatorBond(delegator crypto.Address,
	candidate crypto.PubKey) *DelegatorBond {

	delegatorBytes := k.store.Get(GetDelegatorBondKey(delegator, candidate))
	if delegatorBytes == nil {
		return nil
	}

	bond := new(DelegatorBond)
	err := cdc.UnmarshalJSON(delegatorBytes, bond)
	if err != nil {
		panic(err)
	}
	return bond
}

func (k Keeper) saveDelegatorBond(delegator crypto.Address,
	bond *DelegatorBond) {

	// if a new bond add to the list of bonds
	if loadDelegatorBond(k.store, delegator, bond.PubKey) == nil {
		pks := loadDelegatorCandidates(k.store, delegator)
		pks = append(pks, (*bond).PubKey)
		b, err := cdc.MarshalJSON(pks)
		if err != nil {
			panic(err)
		}
		k.store.Set(GetDelegatorBondsKey(delegator), b)
	}

	// now actually save the bond
	b, err := cdc.MarshalJSON(*bond)
	if err != nil {
		panic(err)
	}
	k.store.Set(GetDelegatorBondKey(delegator, bond.PubKey), b)
	//updateDelegatorBonds(store, delegator)
}

func (k Keeper) removeDelegatorBond(delegator crypto.Address, candidate crypto.PubKey) {
	// TODO use list queries on multistore to remove iterations here!
	// first remove from the list of bonds
	pks := loadDelegatorCandidates(k.store, delegator)
	for i, pk := range pks {
		if candidate.Equals(pk) {
			pks = append(pks[:i], pks[i+1:]...)
		}
	}
	b, err := cdc.MarshalJSON(pks)
	if err != nil {
		panic(err)
	}
	k.store.Set(GetDelegatorBondsKey(delegator), b)

	// now remove the actual bond
	k.store.Delete(GetDelegatorBondKey(delegator, candidate))
	//updateDelegatorBonds(store, delegator)
}

//_______________________________________________________________________

// load/save the global staking params
func (k Keeper) loadParams() (params Params) {
	b := k.store.Get(ParamKey)
	if b == nil {
		return defaultParams()
	}

	err := cdc.UnmarshalJSON(b, &params)
	if err != nil {
		panic(err) // This error should never occur big problem if does
	}
	return
}
func (k Keeper) saveParams(params Params) {
	b, err := cdc.MarshalJSON(params)
	if err != nil {
		panic(err)
	}
	k.store.Set(ParamKey, b)
}

//_______________________________________________________________________

// load/save the global staking state
func (k Keeper) loadGlobalState() (gs *GlobalState) {
	b := k.store.Get(GlobalStateKey)
	if b == nil {
		return initialGlobalState()
	}
	gs = new(GlobalState)
	err := cdc.UnmarshalJSON(b, gs)
	if err != nil {
		panic(err) // This error should never occur big problem if does
	}
	return
}

func (k Keeper) saveGlobalState(gs *GlobalState) {
	b, err := cdc.MarshalJSON(*gs)
	if err != nil {
		panic(err)
	}
	k.store.Set(GlobalStateKey, b)
}
