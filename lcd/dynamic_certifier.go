package lcd

import (
	"github.com/tendermint/tendermint/types"

	lcdErr "github.com/cosmos/cosmos-sdk/lcd/errors"
)

var _ Certifier = (*DynamicCertifier)(nil)

// DynamicCertifier uses a StaticCertifier for Certify, but adds an
// Update method to allow for a change of validators.
//
// You can pass in a FullCommit with another validator set,
// and if this is a provably secure transition (< 1/3 change,
// sufficient signatures), then it will update the
// validator set for the next Certify call.
// For security, it will only follow validator set changes
// going forward.
type DynamicCertifier struct {
	cert       *StaticCertifier
	lastHeight int64
}

// NewDynamic returns a new dynamic certifier.
func NewDynamicCertifier(chainID string, vals *types.ValidatorSet, height int64) *DynamicCertifier {
	return &DynamicCertifier{
		cert:       NewStaticCertifier(chainID, vals),
		lastHeight: height,
	}
}

// ChainID returns the chain id of this certifier.
// Implements Certifier.
func (dc *DynamicCertifier) ChainID() string {
	return dc.cert.ChainID()
}

// Validators returns the validators of this certifier.
func (dc *DynamicCertifier) Validators() *types.ValidatorSet {
	return dc.cert.vSet
}

// Hash returns the hash of this certifier.
func (dc *DynamicCertifier) Hash() []byte {
	return dc.cert.Hash()
}

// LastHeight returns the last height of this certifier.
func (dc *DynamicCertifier) LastHeight() int64 {
	return dc.lastHeight
}

// Certify will verify whether the commit is valid and will update the height if it is or return an
// error if it is not.
// Implements Certifier.
func (dc *DynamicCertifier) Certify(check Commit) error {
	err := dc.cert.Certify(check)
	if err == nil {
		// update last seen height if input is valid
		dc.lastHeight = check.Height()
	}
	return err
}

// Update will verify if this is a valid change and update
// the certifying validator set if safe to do so.
//
// Returns an error if update is impossible (invalid proof or IsTooMuchChangeErr)
func (dc *DynamicCertifier) Update(fc FullCommit) error {
	// ignore all checkpoints in the past -> only to the future
	h := fc.Height()
	if h <= dc.lastHeight {
		return lcdErr.ErrPastTime()
	}

	// first, verify if the input is self-consistent....
	err := fc.ValidateBasic(dc.ChainID())
	if err != nil {
		return err
	}

	// now, make sure not too much change... meaning this commit
	// would be approved by the currently known validator set
	// as well as the new set
	commit := fc.Commit.Commit
	err = dc.Validators().VerifyCommitAny(fc.Validators, dc.ChainID(), commit.BlockID, h, commit)
	if err != nil {
		return lcdErr.ErrTooMuchChange()
	}

	// looks good, we can update
	dc.cert = NewStaticCertifier(dc.ChainID(), fc.Validators)
	dc.lastHeight = h
	return nil
}
