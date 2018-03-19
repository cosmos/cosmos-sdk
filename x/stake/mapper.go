package stake

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

//nolint
const (
	// Keys for store prefixes
	CandidatesAddrKey = []byte{0x01} // key for all candidates' addresses
	ParamKey          = []byte{0x02} // key for global parameters relating to staking
	GlobalStateKey    = []byte{0x03} // key for global parameters relating to staking

	// Key prefixes
	CandidateKeyPrefix        = []byte{0x04} // prefix for each key to a candidate
	ValidatorKeyPrefix        = []byte{0x05} // prefix for each key to a candidate
	ValidatorUpdatesKeyPrefix = []byte{0x06} // prefix for each key to a candidate
	DelegatorBondKeyPrefix    = []byte{0x07} // prefix for each key to a delegator's bond
	DelegatorBondsKeyPrefix   = []byte{0x08} // prefix for each key to a delegator's bond
)

// CandidateKey - get the key for the candidate with address
func CandidateKey(address sdk.Address) []byte {
	return append(CandidateKeyPrefix, address.Bytes()...)
}

// ValidatorKey - get the key for the validator used in the power-store
func ValidatorKey(address sdk.Address, power sdk.Rational, cdc *wire.Codec) []byte {
	b, _ := cdc.MarshalJSON(power)                                      // TODO need to handle error here?
	return append(ValidatorKeyPrefix, append(b, address.Bytes()...)...) // TODO does this need prefix if its in its own store
}

// ValidatorUpdatesKey - get the key for the validator used in the power-store
func ValidatorUpdatesKey(address sdk.Address) []byte {
	return append(ValidatorUpdatesKeyPrefix, address.Bytes()...) // TODO does this need prefix if its in its own store
}

// DelegatorBondKey - get the key for delegator bond with candidate
func DelegatorBondKey(delegator sdk.Address, delegatee sdk.Address) []byte {
	return append(append(DelegatorBondKeyPrefix, delegator), delegatee)
}

// // DelegatorBondKeyPrefix - get the prefix for a delegator for all candidates
// func DelegatorBondKeyPrefix(delegator sdk.Address, cdc *wire.Codec) []byte {
// 	return append(DelegatorBondKeyPrefix, res...)
// }

// DelegatorBondsKey - get the key for list of all the delegator's bonds
func DelegatorBondsKey(delegator sdk.Address, cdc *wire.Codec) []byte {
	res, err := cdc.MarshalJSON(&delegator)
	if err != nil {
		panic(err)
	}
	return append(DelegatorBondsKeyPrefix, res...)
}

//___________________________________________________________________________

// mapper of the staking store
type stakeMapper struct {

	// The reference to the CoinKeeper to modify balances
	ck bank.CoinKeeper

	// The (unexposed) keys used to access the stores from the Context.
	stakeStoreKey sdk.StoreKey

	// The wire codec for binary encoding/decoding.
	cdc *wire.Codec
}

func NewStakeMapper(key sdk.StoreKey, ck bank.CoinKeeper) StakeMapper {
	cdc := wire.NewCodec()
	return stakeMapper{
		ck:            ck,
		stakeStoreKey: key,
		cdc:           cdc,
	}
}

// Returns the go-wire codec.
func (sm stakingMapper) WireCodec() *wire.Codec {
	return gm.cdc
}

func (sm stakeMapper) getCandidate(ctx sdk.Context, address sdk.Address) Candidate, sdk.Error {
	
	candidate := new(Candidate)

	store := ctx.KVStore(sm.stakeStoreKey)
	bz := sm.store.Get(CandidateKey(address))
	if bz == nil {
		return *candidate, ErrCandidateEmpty()
	}

	err := sm.cdc.UnmarshalJSON(bz, candidate)
	if err != nil {
		panic(err) // This error should never occur big problem if does
	}
	return *candidate, nil
}

func (sm stakeMapper) setCandidate(candidate Candidate) {

	// XXX should only remove validator if we know candidate is a validator
	sm.removeValidator(candidate.Address)
	validator := &Validator{candidate.Address, candidate.VotingPower}
	sm.updateValidator(validator)

	bz, err := m.cdc.MarshalJSON(*candidate)
	if err != nil {
		panic(err)
	}
	sm.store.Set(CandidateKey(candidate.Address), b)
}

func (sm stakeMapper) removeCandidate(address sdk.Address) {

	// XXX should only remove validator if we know candidate is a validator
	m.removeValidator(address)
	m.store.Delete(GetCandidateKey(address))
}

//___________________________________________________________________________

//func loadValidator(m.store sdk.KVStore, address sdk.Address, votingPower sdk.Rational) *Validator {
//b := m.store.Get(GetValidatorKey(address, votingPower))
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
func (sm stakeMapper) updateValidator(validator *Validator) {

	b, err := m.cdc.MarshalJSON(*validator)
	if err != nil {
		panic(err)
	}

	// add to the validators to update list if necessary
	m.store.Set(GetValidatorUpdatesKey(validator.Address), b)

	// update the list ordered by voting power
	m.store.Set(GetValidatorKey(validator.Address, validator.VotingPower, m.cdc), b)
}

func (sm stakeMapper) removeValidator(address sdk.Address) {

	//add validator with zero power to the validator updates
	b, err := m.cdc.MarshalJSON(Validator{address, sdk.ZeroRat})
	if err != nil {
		panic(err)
	}
	m.store.Set(GetValidatorUpdatesKey(address), b)

	// now actually delete from the validator set
	candidate := m.loadCandidate(address)
	if candidate != nil {
		m.store.Delete(GetValidatorKey(address, candidate.VotingPower, m.cdc))
	}
}

// get the most recent updated validator set from the Candidates. These bonds
// are already sorted by VotingPower from the UpdateVotingPower function which
// is the only function which is to modify the VotingPower
func (sm stakeMapper) getValidators(maxVal uint16) (validators []Validator) {

	iterator := m.store.Iterator(subspace(ValidatorKeyPrefix)) //smallest to largest

	validators = make([]Validator, maxVal)
	for i := 0; ; i++ {
		if !iterator.Valid() || i > int(maxVal) {
			iterator.Close()
			break
		}
		valBytes := iterator.Value()
		var val Validator
		err := m.cdc.UnmarshalJSON(valBytes, &val)
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
func (sm stakeMapper) getValidatorUpdates() (updates []Validator) {

	iterator := m.store.Iterator(subspace(ValidatorUpdatesKeyPrefix)) //smallest to largest

	for ; iterator.Valid(); iterator.Next() {
		valBytes := iterator.Value()
		var val Validator
		err := m.cdc.UnmarshalJSON(valBytes, &val)
		if err != nil {
			panic(err)
		}
		updates = append(updates, val)
	}

	iterator.Close()
	return
}

// remove all validator update entries
func (sm stakeMapper) clearValidatorUpdates(maxVal int) {
	iterator := m.store.Iterator(subspace(ValidatorUpdatesKeyPrefix))
	for ; iterator.Valid(); iterator.Next() {
		m.store.Delete(iterator.Key()) // XXX write test for this, may need to be in a second loop
	}
	iterator.Close()
}

//---------------------------------------------------------------------

// loadCandidates - get the active list of all candidates TODO replace with  multistore
func (sm stakeMapper) loadCandidates() (candidates Candidates) {

	iterator := m.store.Iterator(subspace(CandidateKeyPrefix))
	//iterator := m.store.Iterator(CandidateKeyPrefix, []byte(nil))
	//iterator := m.store.Iterator([]byte{}, []byte(nil))

	for ; iterator.Valid(); iterator.Next() {
		candidateBytes := iterator.Value()
		var candidate Candidate
		err := m.cdc.UnmarshalJSON(candidateBytes, &candidate)
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
func (sm stakeMapper) getDelegators(delegator sdk.Address) (candidateAddrs []sdk.Address) {

	candidateBytes := m.store.Get(GetDelegatorBondsKey(delegator, m.cdc))
	if candidateBytes == nil {
		return nil
	}

	err := m.cdc.UnmarshalJSON(candidateBytes, &candidateAddrs)
	if err != nil {
		panic(err)
	}
	return
}


// load the pubkeys of all candidates a delegator is delegated too
func (sm stakeMapper) getDelegations(delegator sdk.Address) (candidateAddrs []sdk.Address) {

	candidateBytes := m.store.Get(GetDelegatorBondsKey(delegator, m.cdc))
	if candidateBytes == nil {
		return nil
	}

	err := m.cdc.UnmarshalJSON(candidateBytes, &candidateAddrs)
	if err != nil {
		panic(err)
	}
	return
}

//_____________________________________________________________________

func (sm stakeMapper) getDelegatorBond(delegator sdk.Address, delegatee sdk.Address) *DelegatorBond {
	store := ctx.KVStore(gm.proposalStoreKey)
	bz := store.Get(DelegatorBondKey(delegator, candidate))
	if bz == nil {
		return nil
	}

	bond := new(DelegatorBond)
	err := m.cdc.UnmarshalJSON(bz, bond)
	if err != nil {
		panic(err)
	}

	return bond
}

func (sm stakeMapper) setDelegatorBond(bond *DelegatorBond) {

	// if a new bond add to the list of bonds
	if sm.getDelegatorBond(bond.Delegator, bond.Delegatee) == nil {
		pks := m.loadDelegatorCandidates(delegator)
		pks = append(pks, (*bond).Address)
		b, err := m.cdc.MarshalJSON(pks)
		if err != nil {
			panic(err)
		}
		m.store.Set(GetDelegatorBondsKey(delegator, m.cdc), b)
	}

	// now actually save the bond
	b, err := m.cdc.MarshalJSON(*bond)
	if err != nil {
		panic(err)
	}
	m.store.Set(GetDelegatorBondKey(delegator, bond.Address, m.cdc), b)
	//updateDelegatorBonds(store, delegator) //XXX remove?
}

func (sm stakeMapper) removeDelegatorBond(delegator sdk.Address, candidateAddr sdk.Address) {
	// TODO use list queries on multistore to remove iterations here!
	// first remove from the list of bonds
	addrs := m.loadDelegatorCandidates(delegator)
	for i, addr := range addrs {
		if bytes.Equal(candidateAddr, addr) {
			addrs = append(addrs[:i], addrs[i+1:]...)
		}
	}
	b, err := m.cdc.MarshalJSON(addrs)
	if err != nil {
		panic(err)
	}
	m.store.Set(GetDelegatorBondsKey(delegator, m.cdc), b)

	// now remove the actual bond
	m.store.Delete(GetDelegatorBondKey(delegator, candidateAddr, m.cdc))
	//updateDelegatorBonds(store, delegator) //XXX remove?
}

//_______________________________________________________________________

// load/save the global staking params
func (sm stakeMapper) loadParams() (params Params) {
	b := m.store.Get(ParamKey)
	if b == nil {
		return defaultParams()
	}

	err := m.cdc.UnmarshalJSON(b, &params)
	if err != nil {
		panic(err) // This error should never occur big problem if does
	}
	return
}
func (m Mapper) saveParams(params Params) {
	b, err := m.cdc.MarshalJSON(params)
	if err != nil {
		panic(err)
	}
	m.store.Set(ParamKey, b)
}

//_______________________________________________________________________

// load/save the global staking state
func (sm stakeMap// load the pubkeys of all candidates a delegator is delegated too
func (sm stakeMapper) getDelegations(delegator sdk.Address) (candidateAddrs []sdk.Address) {

	candidateBytes := m.store.Get(GetDelegatorBondsKey(delegator, m.cdc))
	if candidateBytes == nil {
		return nil
	}

	err := m.cdc.UnmarshalJSON(candidateBytes, &candidateAddrs)
	if err != nil {
		panic(err)
	}
	return
}
per) loadGlobalState() (gs *GlobalState) {
	b := m.store.Get(GlobalStateKey)
	if b == nil {
		return initialGlobalState()
	}
	gs = new(GlobalState)
	err := m.cdc.UnmarshalJSON(b, gs)
	if err != nil {
		panic(err) // This error should never occur big problem if does
	}
	return
}

func (sm stakeMapper) saveGlobalState(gs *GlobalState) {
	b, err := m.cdc.MarshalJSON(*gs)
	if err != nil {
		panic(err)
	}
	m.store.Set(GlobalStateKey, b)
}

// Perform all the actions required to bond tokens to a delegator bond from their account
func (sm stakeMapper) BondCoins(bond *DelegatorBond, candidate *Candidate, tokens sdk.Coin) sdk.Error {

	_, err := tr.coinKeeper.SubtractCoins(tr.ctx, candidate.Address, sdk.Coins{tokens})
	if err != nil {
		return err
	}
	newShares := candidate.addTokens(tokens.Amount, tr.gs)
	bond.Shares = bond.Shares.Add(newShares)
	return nil
}

// Perform all the actions required to bond tokens to a delegator bond from their account
func (sm stakeMapper) UnbondCoins(bond *DelegatorBond, candidate *Candidate, shares sdk.Rat) sdk.Error {

	// subtract bond tokens from delegator bond
	if bond.Shares.LT(shares) {
		return sdk.ErrInsufficientFunds("") // TODO
	}
	bond.Shares = bond.Shares.Sub(shares)

	returnAmount := candidate.removeShares(shares, tr.gs)
	returnCoins := sdk.Coins{{tr.params.BondDenom, returnAmount}}

	_, err := tr.coinKeeper.AddCoins(tr.ctx, candidate.Address, returnCoins)
	if err != nil {
		return err
	}
	return nil
}
