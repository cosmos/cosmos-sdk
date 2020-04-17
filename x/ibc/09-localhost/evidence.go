package localhost

import (
	yaml "gopkg.in/yaml.v2"

	"github.com/tendermint/tendermint/crypto/tmhash"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"

	evidenceexported "github.com/cosmos/cosmos-sdk/x/evidence/exported"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
)

var (
	_ evidenceexported.Evidence   = Evidence{}
	_ clientexported.Misbehaviour = Evidence{}
)

// Evidence is not required for a loop-back client
type Evidence struct {
}

// ClientType is Localhost light client
func (ev Evidence) ClientType() clientexported.ClientType {
	return clientexported.Localhost
}

// GetClientID returns the ID of the client that committed a misbehaviour.
func (ev Evidence) GetClientID() string {
	return clientexported.Localhost.String()
}

// Route implements Evidence interface
func (ev Evidence) Route() string {
	return clienttypes.SubModuleName
}

// Type implements Evidence interface
func (ev Evidence) Type() string {
	return "client_misbehaviour"
}

// String implements Evidence interface
func (ev Evidence) String() string {
	// FIXME: implement custom marshaller
	bz, err := yaml.Marshal(ev)
	if err != nil {
		panic(err)
	}
	return string(bz)
}

// Hash implements Evidence interface.
func (ev Evidence) Hash() tmbytes.HexBytes {
	bz := SubModuleCdc.MustMarshalBinaryBare(ev)
	return tmhash.Sum(bz)
}

// GetHeight returns 0.
func (ev Evidence) GetHeight() int64 {
	return 0
}

// ValidateBasic implements Evidence interface.
func (ev Evidence) ValidateBasic() error {
	return nil
}
