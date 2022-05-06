package teststaking

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/crypto"
	"testing"

	"github.com/stretchr/testify/require"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// NewValidator is a testing helper method to create validators in tests
func NewValidator(t testing.TB, operator sdk.ValAddress, pubKey cryptotypes.PubKey, orchAddr sdk.AccAddress, ethAddr types.EthAddress) types.Validator {
	v, err := types.NewValidator(operator, pubKey, types.Description{}, orchAddr, ethAddr)
	require.NoError(t, err)
	return v
}

func RandomEthAddress() (*types.EthAddress, error) {
	ethPrivateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}

	orchEthPublicKey := ethPrivateKey.Public().(*ecdsa.PublicKey)
	ethAddr, err := types.NewEthAddress(crypto.PubkeyToAddress(*orchEthPublicKey).Hex())
	if err != nil {
		return nil, err
	}

	return ethAddr, nil
}
