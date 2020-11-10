package keeper_test

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
	set "github.com/deckarep/golang-set"
)

type OnRecvPacketTestCase  = struct {
	bankBefore []bank.Balance
	packet types.FungibleTokenPacketData
	bankAfter []bank.Balance
	pass bool
}

type OwnedCoin struct {
	Address string
	Coin string
}

func (suite *KeeperTestSuite) SetBankBalances(chain *ibctesting.TestChain, bank []bank.Balance) error {
	for balance := range bank {
		addr, err := sdk.AccAddressFromBech32(bank[balance].Address)
		if err != nil {
			return err
		}
		chain.App.BankKeeper.SetBalances(chain.GetContext(), addr, bank[balance].Coins)
	}
	return nil
}

func (suite *KeeperTestSuite) CheckBankBalances(chain *ibctesting.TestChain, bank []bank.Balance) error {
	existing := set.NewSet()
	chain.App.BankKeeper.IterateAllBalances(chain.GetContext(), func(address sdk.AccAddress, coin sdk.Coin) (stop bool){
		existing.Add(OwnedCoin{address.String(), coin.String()})
		return false
	})

	required := set.NewSet()
	for _, balance := range bank {
		address, err := sdk.AccAddressFromBech32(balance.Address)
		if err != nil {
			return err
		}
		for _, coin := range balance.Coins {
			required.Add(OwnedCoin{address.String(), coin.String()})
		}
	}

	existingNotRequired := existing.Difference(required)
	requiredNotExisting := required.Difference(existing)
    diff := ""
	if existingNotRequired.Cardinality() != 0 {
		diff += "Existing but not required: " + existingNotRequired.String() + "\n"
	}
	if requiredNotExisting.Cardinality() != 0 {
		diff += "Required but not existing: " + requiredNotExisting.String() + "\n"
	}
	if len(diff) != 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, diff)
	}
	return nil
}


func (suite *KeeperTestSuite) TestModelBasedOnRecvPacket() {
	suite.Run(fmt.Sprintf("Model based test for OnRecvPacket"), func() {
		suite.SetupTest() // reset
		_, _, connA, connB := suite.coordinator.SetupClientConnections(suite.chainA, suite.chainB, ibctesting.Tendermint)
		_, _ = suite.coordinator.CreateTransferChannels(suite.chainA, suite.chainB, connA, connB, channeltypes.UNORDERED)

		err := suite.CheckBankBalances(suite.chainA, []bank.Balance{})
		suite.Require().NoError(err)
	})
}
