package teststaking

import (
	"crypto/ecdsa"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/stretchr/testify/require"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
)

// NewValidator is a testing helper method to create validators in tests
func NewValidator(t testing.TB, operator sdk.ValAddress, pubKey cryptotypes.PubKey, evmAddr common.Address) types.Validator {
	v, err := types.NewValidator(operator, pubKey, types.Description{}, evmAddr)
	require.NoError(t, err)
	return v
}

func RandomEVMAddress() (*common.Address, error) {
	ethPrivateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}

	orchEthPublicKey := ethPrivateKey.Public().(*ecdsa.PublicKey)
	evmAddr := crypto.PubkeyToAddress(*orchEthPublicKey)

	return &evmAddr, nil
}
