package tendermint

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	evidenceexported "github.com/cosmos/cosmos-sdk/x/evidence/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/errors"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

var _ exported.Misbehaviour = Misbehaviour{}
var _ evidenceexported.Evidence = Misbehaviour{}

// Misbehaviour contains evidence that a light client submitted a different header from
// a full node at the same height.
type Misbehaviour struct {
	*Evidence
	ClientID string `json:"client_id" yaml:"client_id"`
}

// ClientType is Tendermint light client
func (m Misbehaviour) ClientType() exported.ClientType {
	return exported.Tendermint
}

// GetEvidence returns the evidence to handle a light client misbehaviour
func (m Misbehaviour) GetEvidence() evidenceexported.Evidence {
	return m.Evidence
}

// ValidateBasic performs the basic validity checks for the evidence and the
// client ID.
func (m Misbehaviour) ValidateBasic() error {
	if m.Evidence == nil {
		return sdkerrors.Wrap(errors.ErrInvalidEvidence, "evidence is empty")
	}
	if err := m.Evidence.ValidateBasic(); err != nil {
		return sdkerrors.Wrap(errors.ErrInvalidEvidence, err.Error())
	}
	if err := host.DefaultClientIdentifierValidator(m.ClientID); err != nil {
		return sdkerrors.Wrap(errors.ErrInvalidEvidence, err.Error())
	}
	return nil
}

// CheckMisbehaviour checks if the evidence provided is a valid light client misbehaviour
func CheckMisbehaviour(trustedCommitter Committer, m Misbehaviour) error {
	if err := m.ValidateBasic(); err != nil {
		return err
	}

	trustedValSet := trustedCommitter.ValidatorSet

	// Evidence is within trusting period. ValidatorSet must have 2/3 similarity with trustedCommitter ValidatorSet
	// check that the validator sets on both headers are valid given the last trusted validatorset
	// less than or equal to evidence height
	if err := trustedValSet.VerifyFutureCommit(
		m.Evidence.Header1.ValidatorSet, m.Evidence.ChainID,
		m.Evidence.Header1.Commit.BlockID, m.Evidence.Header1.Height, m.Evidence.Header1.Commit,
	); err != nil {
		return sdkerrors.Wrapf(errors.ErrInvalidEvidence, "validator set in header 1 has too much change from last known committer: %v", err)
	}
	if err := trustedValSet.VerifyFutureCommit(
		m.Evidence.Header2.ValidatorSet, m.Evidence.ChainID,
		m.Evidence.Header2.Commit.BlockID, m.Evidence.Header2.Height, m.Evidence.Header2.Commit,
	); err != nil {
		return sdkerrors.Wrapf(errors.ErrInvalidEvidence, "validator set in header 2 has too much change from last known committer: %v", err)
	}

	return nil
}
