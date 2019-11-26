package tendermint

import (
	"bytes"
	"fmt"

	evidenceexported "github.com/cosmos/cosmos-sdk/x/evidence/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/errors"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

var _ exported.Misbehaviour = Misbehaviour{}
var _ evidenceexported.Evidence = Misbehaviour{}

// Misbehaviour contains an evidence that a
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
		return errors.ErrInvalidEvidence(errors.DefaultCodespace, "empty evidence")
	}

	if err := m.Evidence.ValidateBasic(); err != nil {
		return err
	}

	return host.DefaultClientIdentifierValidator(m.ClientID)
}

// CheckMisbehaviour checks if the evidence provided is a misbehaviour
func CheckMisbehaviour(trustedCommitter Committer, m Misbehaviour) error {
	if err := m.ValidateBasic(); err != nil {
		return err
	}

	trustedValSet := trustedCommitter.ValidatorSet

	// Evidence is on same height as trustedCommiter. ValidatorSets must be the same
	if trustedCommitter.GetHeight() == uint64(m.GetHeight()) {
		if bytes.Equal(trustedValSet.Hash(), m.Evidence.Header1.ValidatorSet.Hash()) ||
			!bytes.Equal(trustedValSet.Hash(), m.Evidence.Header2.ValidatorSet.Hash()) {
			return errors.ErrInvalidEvidence(errors.DefaultCodespace, fmt.Sprintf("validator set is not valid for height: %d", m.GetHeight()))
		}
		return nil
	}

	// Evidence is on next height of trustedCommitter. Evidence Validator Set must match
	// trustedCommitter's NextValSetHash
	if trustedCommitter.GetHeight()+1 == uint64(m.GetHeight()) {
		if !bytes.Equal(trustedCommitter.NextValSetHash, m.Evidence.Header1.ValidatorSet.Hash()) ||
			!bytes.Equal(trustedCommitter.NextValSetHash, m.Evidence.Header2.ValidatorSet.Hash()) {
			return errors.ErrInvalidEvidence(errors.DefaultCodespace, fmt.Sprintf("validator set is not valid for height: %d", m.GetHeight()))
		}
		return nil
	}

	// Evidence is within trusting period. ValidatorSet must have 2/3 similarity with trustedCommitter ValidatorSet
	// check that the validator sets on both headers are valid given the last trusted validatorset
	// less than or equal to evidence height
	if err := trustedValSet.VerifyFutureCommit(m.Evidence.Header1.ValidatorSet, m.Evidence.ChainID,
		m.Evidence.Header1.Commit.BlockID, m.Evidence.Header1.Height, m.Evidence.Header1.Commit); err != nil {
		return err
	}
	if err := trustedValSet.VerifyFutureCommit(m.Evidence.Header2.ValidatorSet, m.Evidence.ChainID,
		m.Evidence.Header2.Commit.BlockID, m.Evidence.Header2.Height, m.Evidence.Header2.Commit); err != nil {
		return err
	}

	return nil
}
