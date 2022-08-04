package teststaking

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/crypto"
	"testing"

	"github.com/stretchr/testify/require"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
)

// NewValidator is a testing helper method to create validators in tests
func NewValidator(t testing.TB, operator sdk.ValAddress, pubKey cryptotypes.PubKey, orchAddr sdk.AccAddress, ethAddr common.Address) types.Validator {
	v, err := types.NewValidator(operator, pubKey, types.Description{}, orchAddr, ethAddr)
	require.NoError(t, err)
	return v
}

func RandomEthAddress() (*common.Address, error) {
	ethPrivateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}

	orchEthPublicKey := ethPrivateKey.Public().(*ecdsa.PublicKey)
	ethAddr := crypto.PubkeyToAddress(*orchEthPublicKey)

	return &ethAddr, nil
}
