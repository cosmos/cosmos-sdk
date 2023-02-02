package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// UpgradeInfoFileName file to store upgrade information
const UpgradeInfoFilename = "upgrade-info.json"

// ValidateBasic does basic validation of a Plan
func (p Plan) ValidateBasic() error {
	if !p.Time.IsZero() {
		return sdkerrors.ErrInvalidRequest.Wrap("time-based upgrades have been deprecated in the SDK")
	}
	if p.UpgradedClientState != nil {
		return sdkerrors.ErrInvalidRequest.Wrap("upgrade logic for IBC has been moved to the IBC module")
	}
	if len(p.Name) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "name cannot be empty")
	}
	if p.Height <= 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "height must be greater than 0")
	}
	if p.Info != "" && len(p.Artifacts) > 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "info and artifacts cannot both be set")
	}

	for _, artifact := range p.Artifacts {
		if err := artifact.ValidateBasic(); err != nil {
			return fmt.Errorf("artifact invalid: %w", err)
		}
	}

	return nil
}

// ShouldExecute returns true if the Plan is ready to execute given the current context
func (p Plan) ShouldExecute(ctx sdk.Context) bool {
	if p.Height > 0 {
		return p.Height <= ctx.BlockHeight()
	}
	return false
}

// DueAt is a string representation of when this plan is due to be executed
func (p Plan) DueAt() string {
	return fmt.Sprintf("height: %d", p.Height)
}

var SupportedChecksumAlgo = []string{"md5", "sha256", "sha512"}

// ValidateBasic does basic validation of an Artifact
func (a Artifact) ValidateBasic() error {
	if a.Platform == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "platform cannot be empty")
	}

	if a.Url == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "url cannot be empty")
	}

	if a.Checksum == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "checksum cannot be empty")
	}

	if a.ChecksumAlgo == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "checksum algo cannot be empty")
	}

	for _, algo := range SupportedChecksumAlgo {
		if a.ChecksumAlgo == algo {
			return nil
		}
	}

	return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "checksum algo %s is not supported", a.ChecksumAlgo)
}
