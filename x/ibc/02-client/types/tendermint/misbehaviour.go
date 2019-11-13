package tendermint

import (
	evidenceexported "github.com/cosmos/cosmos-sdk/x/evidence/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
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
	if err := m.Evidence.ValidateBasic(); err != nil {
		return err
	}

	return host.DefaultClientIdentifierValidator(m.ClientID)
}

// CheckMisbehaviour checks if the evidence provided is a misbehaviour
func CheckMisbehaviour(m Misbehaviour) error {
	return m.Evidence.DuplicateVoteEvidence.Verify(
		m.Evidence.ChainID, m.Evidence.DuplicateVoteEvidence.PubKey,
	)
}
