package v040

import (
	"gopkg.in/yaml.v2"

	"github.com/cosmos/cosmos-sdk/codec/types"
	v038evidence "github.com/cosmos/cosmos-sdk/x/evidence/legacy/v0_38"
)

// Default parameter values
const (
	ModuleName = "evidence"
)

var _ types.UnpackInterfacesMessage = &GenesisState{}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (gs *GenesisState) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	for _, any := range gs.Evidence {
		var evi v038evidence.Evidence
		err := unpacker.UnpackAny(any, &evi)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *Equivocation) String() string {
	bz, _ := yaml.Marshal(e)
	return string(bz)
}
