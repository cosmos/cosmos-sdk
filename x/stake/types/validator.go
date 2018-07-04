package types

import (
	"bytes"
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	tmtypes "github.com/tendermint/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
)

const doNotModifyDescVal = "[do-not-modify]"

// Validator defines the total amount of bond shares and their exchange rate to
// coins. Accumulation of interest is modelled as an in increase in the
// exchange rate, and slashing as a decrease.  When coins are delegated to this
// validator, the validator is credited with a Delegation whose number of
// bond shares is based on the amount of coins delegated divided by the current
// exchange rate. Voting power can be calculated as total bonds multiplied by
// exchange rate.
type Validator struct {
	Owner   sdk.Address   `json:"owner"`   // sender of BondTx - UnbondTx returns here
	PubKey  crypto.PubKey `json:"pub_key"` // pubkey of validator
	Revoked bool          `json:"revoked"` // has the validator been revoked from bonded status?

	PoolShares      PoolShares `json:"pool_shares"`      // total shares for tokens held in the pool
	DelegatorShares sdk.Rat    `json:"delegator_shares"` // total shares issued to a validator's delegators

	Description        Description `json:"description"`           // description terms for the validator
	BondHeight         int64       `json:"bond_height"`           // earliest height as a bonded validator
	BondIntraTxCounter int16       `json:"bond_intra_tx_counter"` // block-local tx index of validator change
	ProposerRewardPool sdk.Coins   `json:"proposer_reward_pool"`  // XXX reward pool collected from being the proposer

	Commission            sdk.Rat `json:"commission"`              // XXX the commission rate of fees charged to any delegators
	CommissionMax         sdk.Rat `json:"commission_max"`          // XXX maximum commission rate which this validator can ever charge
	CommissionChangeRate  sdk.Rat `json:"commission_change_rate"`  // XXX maximum daily increase of the validator commission
	CommissionChangeToday sdk.Rat `json:"commission_change_today"` // XXX commission rate change today, reset each day (UTC time)

	// fee related
	PrevBondedShares sdk.Rat `json:"prev_bonded_shares"` // total shares of a global hold pools
}

// NewValidator - initialize a new validator
func NewValidator(owner sdk.Address, pubKey crypto.PubKey, description Description) Validator {
	return Validator{
		Owner:                 owner,
		PubKey:                pubKey,
		Revoked:               false,
		PoolShares:            NewUnbondedShares(sdk.ZeroRat()),
		DelegatorShares:       sdk.ZeroRat(),
		Description:           description,
		BondHeight:            int64(0),
		BondIntraTxCounter:    int16(0),
		ProposerRewardPool:    sdk.Coins{},
		Commission:            sdk.ZeroRat(),
		CommissionMax:         sdk.ZeroRat(),
		CommissionChangeRate:  sdk.ZeroRat(),
		CommissionChangeToday: sdk.ZeroRat(),
		PrevBondedShares:      sdk.ZeroRat(),
	}
}

// what's kept in the store value
type validatorValue struct {
	PubKey                crypto.PubKey
	Revoked               bool
	PoolShares            PoolShares
	DelegatorShares       sdk.Rat
	Description           Description
	BondHeight            int64
	BondIntraTxCounter    int16
	ProposerRewardPool    sdk.Coins
	Commission            sdk.Rat
	CommissionMax         sdk.Rat
	CommissionChangeRate  sdk.Rat
	CommissionChangeToday sdk.Rat
	PrevBondedShares      sdk.Rat
}

// return the redelegation without fields contained within the key for the store
func MarshalValidator(cdc *wire.Codec, validator Validator) []byte {
	val := validatorValue{
		PubKey:                validator.PubKey,
		Revoked:               validator.Revoked,
		PoolShares:            validator.PoolShares,
		DelegatorShares:       validator.DelegatorShares,
		Description:           validator.Description,
		BondHeight:            validator.BondHeight,
		BondIntraTxCounter:    validator.BondIntraTxCounter,
		ProposerRewardPool:    validator.ProposerRewardPool,
		Commission:            validator.Commission,
		CommissionMax:         validator.CommissionMax,
		CommissionChangeRate:  validator.CommissionChangeRate,
		CommissionChangeToday: validator.CommissionChangeToday,
		PrevBondedShares:      validator.PrevBondedShares,
	}
	return cdc.MustMarshalBinary(val)
}

// unmarshal a redelegation from a store key and value
func UnmarshalValidator(cdc *wire.Codec, ownerAddr, value []byte) Validator {
	var storeValue validatorValue
	cdc.MustUnmarshalBinary(value, &storeValue)

	if len(ownerAddr) != 20 {
		panic("unexpected address length")
	}

	return Validator{
		Owner:                 ownerAddr,
		PubKey:                storeValue.PubKey,
		Revoked:               storeValue.Revoked,
		PoolShares:            storeValue.PoolShares,
		DelegatorShares:       storeValue.DelegatorShares,
		Description:           storeValue.Description,
		BondHeight:            storeValue.BondHeight,
		BondIntraTxCounter:    storeValue.BondIntraTxCounter,
		ProposerRewardPool:    storeValue.ProposerRewardPool,
		Commission:            storeValue.Commission,
		CommissionMax:         storeValue.CommissionMax,
		CommissionChangeRate:  storeValue.CommissionChangeRate,
		CommissionChangeToday: storeValue.CommissionChangeToday,
		PrevBondedShares:      storeValue.PrevBondedShares,
	}
}

// only the vitals - does not check bond height of IntraTxCounter
func (v Validator) Equal(c2 Validator) bool {
	return v.PubKey.Equals(c2.PubKey) &&
		bytes.Equal(v.Owner, c2.Owner) &&
		v.PoolShares.Equal(c2.PoolShares) &&
		v.DelegatorShares.Equal(c2.DelegatorShares) &&
		v.Description == c2.Description &&
		v.ProposerRewardPool.IsEqual(c2.ProposerRewardPool) &&
		v.Commission.Equal(c2.Commission) &&
		v.CommissionMax.Equal(c2.CommissionMax) &&
		v.CommissionChangeRate.Equal(c2.CommissionChangeRate) &&
		v.CommissionChangeToday.Equal(c2.CommissionChangeToday) &&
		v.PrevBondedShares.Equal(c2.PrevBondedShares)
}

// Description - description fields for a validator
type Description struct {
	Moniker  string `json:"moniker"`  // name
	Identity string `json:"identity"` // optional identity signature (ex. UPort or Keybase)
	Website  string `json:"website"`  // optional website link
	Details  string `json:"details"`  // optional details
}

// NewDescription returns a new Description with the provided values.
func NewDescription(moniker, identity, website, details string) Description {
	return Description{
		Moniker:  moniker,
		Identity: identity,
		Website:  website,
		Details:  details,
	}
}

// UpdateDescription updates the fields of a given description. An error is
// returned if the resulting description contains an invalid length.
func (d Description) UpdateDescription(d2 Description) (Description, sdk.Error) {
	if d.Moniker == doNotModifyDescVal {
		d2.Moniker = d.Moniker
	}
	if d.Identity == doNotModifyDescVal {
		d2.Identity = d.Identity
	}
	if d.Website == doNotModifyDescVal {
		d2.Website = d.Website
	}
	if d.Details == doNotModifyDescVal {
		d2.Details = d.Details
	}

	return Description{
		Moniker:  d2.Moniker,
		Identity: d2.Identity,
		Website:  d2.Website,
		Details:  d2.Details,
	}.EnsureLength()
}

// EnsureLength ensures the length of a validator's description.
func (d Description) EnsureLength() (Description, sdk.Error) {
	if len(d.Moniker) > 70 {
		return d, ErrDescriptionLength(DefaultCodespace, "moniker", len(d.Moniker), 70)
	}
	if len(d.Identity) > 3000 {
		return d, ErrDescriptionLength(DefaultCodespace, "identity", len(d.Identity), 3000)
	}
	if len(d.Website) > 140 {
		return d, ErrDescriptionLength(DefaultCodespace, "website", len(d.Website), 140)
	}
	if len(d.Details) > 280 {
		return d, ErrDescriptionLength(DefaultCodespace, "details", len(d.Details), 280)
	}

	return d, nil
}

// ABCIValidator returns an abci.Validator from a staked validator type.
func (v Validator) ABCIValidator() abci.Validator {
	return abci.Validator{
		PubKey: tmtypes.TM2PB.PubKey(v.PubKey),
		Power:  v.PoolShares.Bonded().RoundInt64(),
	}
}

// ABCIValidatorZero returns an abci.Validator from a staked validator type
// with with zero power used for validator updates.
func (v Validator) ABCIValidatorZero() abci.Validator {
	return abci.Validator{
		PubKey: tmtypes.TM2PB.PubKey(v.PubKey),
		Power:  0,
	}
}

// Status returns the validator's bond status inferred from the pool shares.
func (v Validator) Status() sdk.BondStatus {
	return v.PoolShares.Status
}

// UpdateStatus updates the location of the shares within a validator if it's
// bond status has changed.
func (v Validator) UpdateStatus(pool Pool, NewStatus sdk.BondStatus) (Validator, Pool) {
	var tokens int64

	switch v.Status() {
	case sdk.Unbonded:
		if NewStatus == sdk.Unbonded {
			return v, pool
		}
		pool, tokens = pool.removeSharesUnbonded(v.PoolShares.Amount)

	case sdk.Unbonding:
		if NewStatus == sdk.Unbonding {
			return v, pool
		}
		pool, tokens = pool.removeSharesUnbonding(v.PoolShares.Amount)

	case sdk.Bonded:
		if NewStatus == sdk.Bonded {
			// Return if nothing needs switching
			return v, pool
		}
		pool, tokens = pool.removeSharesBonded(v.PoolShares.Amount)
	}

	switch NewStatus {
	case sdk.Unbonded:
		pool, v.PoolShares = pool.addTokensUnbonded(tokens)
	case sdk.Unbonding:
		pool, v.PoolShares = pool.addTokensUnbonding(tokens)
	case sdk.Bonded:
		pool, v.PoolShares = pool.addTokensBonded(tokens)
	}

	return v, pool
}

// RemovePoolShares removes pool shares from a validator. It returns
// corresponding tokens, which could be burned (e.g. when slashing a validator)
// or redistributed elsewhere.
func (v Validator) RemovePoolShares(pool Pool, poolShares sdk.Rat) (Validator, Pool, int64) {
	var tokens int64

	switch v.Status() {
	case sdk.Unbonded:
		pool, tokens = pool.removeSharesUnbonded(poolShares)
	case sdk.Unbonding:
		pool, tokens = pool.removeSharesUnbonding(poolShares)
	case sdk.Bonded:
		pool, tokens = pool.removeSharesBonded(poolShares)
	}

	v.PoolShares.Amount = v.PoolShares.Amount.Sub(poolShares)
	return v, pool, tokens
}

// EquivalentBondedShares ...
//
// TODO: Remove should only be tokens get the power or potential power for a
// validator if bonded, the power is the BondedShares if not bonded, the power
// is the amount of bonded shares which the the validator would have it was
// bonded.
func (v Validator) EquivalentBondedShares(pool Pool) (eqBondedShares sdk.Rat) {
	return v.PoolShares.ToBonded(pool).Amount
}

//_________________________________________________________________________________________________________

// AddTokensFromDel adds tokens to a validator
func (v Validator) AddTokensFromDel(pool Pool, amount int64) (Validator, Pool, sdk.Rat) {
	var (
		poolShares             PoolShares
		equivalentBondedShares sdk.Rat
	)

	// bondedShare/delegatedShare
	exRate := v.DelegatorShareExRate(pool)

	switch v.Status() {
	case sdk.Unbonded:
		pool, poolShares = pool.addTokensUnbonded(amount)
	case sdk.Unbonding:
		pool, poolShares = pool.addTokensUnbonding(amount)
	case sdk.Bonded:
		pool, poolShares = pool.addTokensBonded(amount)
	}
	v.PoolShares.Amount = v.PoolShares.Amount.Add(poolShares.Amount)
	equivalentBondedShares = poolShares.ToBonded(pool).Amount

	// bondedShare/(bondedShare/delegatedShare) = delegatedShare
	issuedDelegatorShares := equivalentBondedShares.Quo(exRate)
	v.DelegatorShares = v.DelegatorShares.Add(issuedDelegatorShares)

	return v, pool, issuedDelegatorShares
}

// RemoveDelShares removes delegator shares from a validator.
//
// NOTE: This function assumes the shares have already been updated for the
// validator status.
func (v Validator) RemoveDelShares(pool Pool, delShares sdk.Rat) (Validator, Pool, int64) {
	amount := v.DelegatorShareExRate(pool).Mul(delShares)
	eqBondedSharesToRemove := NewBondedShares(amount)
	v.DelegatorShares = v.DelegatorShares.Sub(delShares)

	var createdCoins int64

	switch v.Status() {
	case sdk.Unbonded:
		unbondedShares := eqBondedSharesToRemove.ToUnbonded(pool).Amount
		pool, createdCoins = pool.removeSharesUnbonded(unbondedShares)
		v.PoolShares.Amount = v.PoolShares.Amount.Sub(unbondedShares)
	case sdk.Unbonding:
		unbondingShares := eqBondedSharesToRemove.ToUnbonding(pool).Amount
		pool, createdCoins = pool.removeSharesUnbonding(unbondingShares)
		v.PoolShares.Amount = v.PoolShares.Amount.Sub(unbondingShares)
	case sdk.Bonded:
		pool, createdCoins = pool.removeSharesBonded(eqBondedSharesToRemove.Amount)
		v.PoolShares.Amount = v.PoolShares.Amount.Sub(eqBondedSharesToRemove.Amount)
	}

	return v, pool, createdCoins
}

// DelegatorShareExRate gets the exchange rate of tokens over delegator shares.
// UNITS: eq-val-bonded-shares/delegator-shares
func (v Validator) DelegatorShareExRate(pool Pool) sdk.Rat {
	if v.DelegatorShares.IsZero() {
		return sdk.OneRat()
	}

	eqBondedShares := v.PoolShares.ToBonded(pool).Amount
	return eqBondedShares.Quo(v.DelegatorShares)
}

//______________________________________________________________________

// ensure fulfills the sdk validator types
var _ sdk.Validator = Validator{}

// nolint - for sdk.Validator
func (v Validator) GetRevoked() bool            { return v.Revoked }
func (v Validator) GetMoniker() string          { return v.Description.Moniker }
func (v Validator) GetStatus() sdk.BondStatus   { return v.Status() }
func (v Validator) GetOwner() sdk.Address       { return v.Owner }
func (v Validator) GetPubKey() crypto.PubKey    { return v.PubKey }
func (v Validator) GetPower() sdk.Rat           { return v.PoolShares.Bonded() }
func (v Validator) GetDelegatorShares() sdk.Rat { return v.DelegatorShares }
func (v Validator) GetBondHeight() int64        { return v.BondHeight }

// HumanReadableString returns a human readable string representation of a
// validator. An error is returned if the owner or the owner's public key
// cannot be converted to Bech32 format.
func (v Validator) HumanReadableString() (string, error) {
	bechOwner, err := sdk.Bech32ifyAcc(v.Owner)
	if err != nil {
		return "", err
	}

	bechVal, err := sdk.Bech32ifyValPub(v.PubKey)
	if err != nil {
		return "", err
	}

	resp := "Validator \n"
	resp += fmt.Sprintf("Owner: %s\n", bechOwner)
	resp += fmt.Sprintf("Validator: %s\n", bechVal)
	resp += fmt.Sprintf("Shares: Status %s,  Amount: %s\n", sdk.BondStatusToString(v.PoolShares.Status), v.PoolShares.Amount.FloatString())
	resp += fmt.Sprintf("Delegator Shares: %s\n", v.DelegatorShares.FloatString())
	resp += fmt.Sprintf("Description: %s\n", v.Description)
	resp += fmt.Sprintf("Bond Height: %d\n", v.BondHeight)
	resp += fmt.Sprintf("Proposer Reward Pool: %s\n", v.ProposerRewardPool.String())
	resp += fmt.Sprintf("Commission: %s\n", v.Commission.String())
	resp += fmt.Sprintf("Max Commission Rate: %s\n", v.CommissionMax.String())
	resp += fmt.Sprintf("Commission Change Rate: %s\n", v.CommissionChangeRate.String())
	resp += fmt.Sprintf("Commission Change Today: %s\n", v.CommissionChangeToday.String())
	resp += fmt.Sprintf("Previously Bonded Stares: %s\n", v.PrevBondedShares.String())

	return resp, nil
}
