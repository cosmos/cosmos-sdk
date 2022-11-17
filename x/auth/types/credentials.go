package types

import (
	"fmt"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

// NewAccountWithModuleCredential creates an account with an unclaimable module credential.
func NewAccountWithModuleCredential(pubkey cryptotypes.PubKey) (*BaseAccount, error) {
	if pubkey == nil {
		return nil, fmt.Errorf("pubkey cannot be nil")
	}

	baseAccount := NewBaseAccountWithAddress(sdk.AccAddress(pubkey.Address()))
	if err := baseAccount.SetPubKey(pubkey); err != nil {
		return nil, fmt.Errorf("failed to create a valid account with credentials: %w", err)
	}

	if err := baseAccount.Validate(); err != nil {
		return nil, fmt.Errorf("failed to create a valid account with credentials: %w", err)
	}

	return baseAccount, nil
}

//nolint:gosec // this isn't an hardcoded credential
const ModuleCredentialType = "ModuleCredential"

var _ cryptotypes.PubKey = &ModuleCredential{}

func NewModuleCredential(moduleName string, derivationKeys [][]byte) *ModuleCredential {
	return &ModuleCredential{
		ModuleName:     moduleName,
		DerivationKeys: derivationKeys,
	}
}

func (m *ModuleCredential) Address() cryptotypes.Address {
	var module []byte
	for i, dk := range m.DerivationKeys {
		if i == 0 {
			module = address.Module(m.ModuleName, dk)
			continue
		}

		module = address.Derive(module, dk)
	}

	return module
}

func (m *ModuleCredential) Bytes() []byte {
	return nil
}

// VerifySignature returns always false, making the account unclaimable
func (m *ModuleCredential) VerifySignature(_ []byte, _ []byte) bool {
	return false
}

func (m *ModuleCredential) Equals(other cryptotypes.PubKey) bool {
	return false
}

func (m *ModuleCredential) Type() string {
	return ModuleCredentialType
}
