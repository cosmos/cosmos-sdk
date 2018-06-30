package lcd

import (
	"github.com/tendermint/tendermint/types"

	lcdErr "github.com/cosmos/cosmos-sdk/lcd/errors"
)

var _ Certifier = (*InquiringCertifier)(nil)

// InquiringCertifier wraps a dynamic certifier and implements an auto-update strategy. If a call
// to Certify fails due to a change it validator set, InquiringCertifier will try and find a
// previous FullCommit which it can use to safely update the validator set. It uses a source
// provider to obtain the needed FullCommits. It stores properly validated data on the local system.
type InquiringCertifier struct {
	cert *DynamicCertifier
	// These are only properly validated data, from local system
	trusted Provider
	// This is a source of new info, like a node rpc, or other import method
	Source Provider
}

// NewInquiringCertifier returns a new Inquiring object. It uses the trusted provider to store
// validated data and the source provider to obtain missing FullCommits.
//
// Example: The trusted provider should a CacheProvider, MemProvider or files.Provider. The source
// provider should be a client.HTTPProvider.
func NewInquiringCertifier(chainID string, fc FullCommit, trusted Provider,
	source Provider) (*InquiringCertifier, error) {

	// store the data in trusted
	err := trusted.StoreCommit(fc)
	if err != nil {
		return nil, err
	}

	return &InquiringCertifier{
		cert:    NewDynamicCertifier(chainID, fc.Validators, fc.Height()),
		trusted: trusted,
		Source:  source,
	}, nil
}

// ChainID returns the chain id.
// Implements Certifier.
func (ic *InquiringCertifier) ChainID() string {
	return ic.cert.ChainID()
}

// Validators returns the validator set.
func (ic *InquiringCertifier) Validators() *types.ValidatorSet {
	return ic.cert.cert.vSet
}

// LastHeight returns the last height.
func (ic *InquiringCertifier) LastHeight() int64 {
	return ic.cert.lastHeight
}

// Certify makes sure this is checkpoint is valid.
//
// If the validators have changed since the last know time, it looks
// for a path to prove the new validators.
//
// On success, it will store the checkpoint in the store for later viewing
// Implements Certifier.
func (ic *InquiringCertifier) Certify(commit Commit) error {
	err := ic.useClosestTrust(commit.Height())
	if err != nil {
		return err
	}

	err = ic.cert.Certify(commit)
	if !lcdErr.IsValidatorsChangedErr(err) {
		return err
	}
	err = ic.updateToHash(commit.Header.ValidatorsHash)
	if err != nil {
		return err
	}

	err = ic.cert.Certify(commit)
	if err != nil {
		return err
	}

	// store the new checkpoint
	return ic.trusted.StoreCommit(NewFullCommit(commit, ic.Validators()))
}

// Update will verify if this is a valid change and update
// the certifying validator set if safe to do so.
func (ic *InquiringCertifier) Update(fc FullCommit) error {
	err := ic.useClosestTrust(fc.Height())
	if err != nil {
		return err
	}

	err = ic.cert.Update(fc)
	if err == nil {
		err = ic.trusted.StoreCommit(fc)
	}
	return err
}

func (ic *InquiringCertifier) useClosestTrust(h int64) error {
	closest, err := ic.trusted.GetByHeight(h)
	if err != nil {
		return err
	}

	// if the best seed is not the one we currently use,
	// let's just reset the dynamic validator
	if closest.Height() != ic.LastHeight() {
		ic.cert = NewDynamicCertifier(ic.ChainID(), closest.Validators, closest.Height())
	}
	return nil
}

// updateToHash gets the validator hash we want to update to
// if IsTooMuchChangeErr, we try to find a path by binary search over height
func (ic *InquiringCertifier) updateToHash(vhash []byte) error {
	// try to get the match, and update
	fc, err := ic.Source.GetByHash(vhash)
	if err != nil {
		return err
	}
	err = ic.cert.Update(fc)
	// handle IsTooMuchChangeErr by using divide and conquer
	if lcdErr.IsTooMuchChangeErr(err) {
		err = ic.updateToHeight(fc.Height())
	}
	return err
}

// updateToHeight will use divide-and-conquer to find a path to h
func (ic *InquiringCertifier) updateToHeight(h int64) error {
	// try to update to this height (with checks)
	fc, err := ic.Source.GetByHeight(h)
	if err != nil {
		return err
	}
	start, end := ic.LastHeight(), fc.Height()
	if end <= start {
		return lcdErr.ErrNoPathFound()
	}
	err = ic.Update(fc)

	// we can handle IsTooMuchChangeErr specially
	if !lcdErr.IsTooMuchChangeErr(err) {
		return err
	}

	// try to update to mid
	mid := (start + end) / 2
	err = ic.updateToHeight(mid)
	if err != nil {
		return err
	}

	// if we made it to mid, we recurse
	return ic.updateToHeight(h)
}
