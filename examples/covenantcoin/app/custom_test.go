package app

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	cov "github.com/cosmos/cosmos-sdk/examples/covenantcoin/x/covenant"
	sdk "github.com/cosmos/cosmos-sdk/types"

	abci "github.com/tendermint/abci/types"
)

func TestCovenant(t *testing.T) {
	// Create a new covenant app for testing. Src of this is in app_test
	app := newCovenantApp()
	coinDenom := "foocoin"
	// Default state is addr1 has 10k foocoins
	genesisState := fmt.Sprintf(`{
      "accounts": [{
        "address": "%s",
        "coins": [
          {
            "denom": "%s",
            "amount": 10000
          }
        ]
      }]
    }`, addr1.String(), coinDenom)

	// no validators needed right now
	vals := []abci.Validator{}
	app.InitChain(abci.RequestInitChain{vals, []byte(genesisState)})
	app.Commit() // 'Finalize' this block.

	// Create a Coin object, and a create Covenant Message
	coinsToEscrow := sdk.Coin{Denom: coinDenom, Amount: 1000}
	createCov := cov.MsgCreateCovenant{Sender: addr1,
		Settlers:  []sdk.Address{addr1},
		Receivers: []sdk.Address{addr2, addr3},
		Amount:    []sdk.Coin{coinsToEscrow},
	}

	numCovenants := int64(5)
	// Send <numCovenants> create Covenant Messages, and check that the returned
	// covenant id is as expected for each of them, and that the transaction goes through.
	for i := int64(0); i < numCovenants; i++ {
		res := SignCheckDeliver(t, app, createCov, []int64{i}, true, priv1)
		var id int64
		app.cdc.UnmarshalBinary(res.Data, &id)
		require.Equal(t, i, id)
		app.Commit()
	}

	// Create a settle covenant message
	settleCov1 := cov.MsgSettleCovenant{CovID: int64(1),
		Settler:  addr1,
		Receiver: addr2,
	}
	// Claim the covenant
	SignCheckDeliver(t, app, settleCov1, []int64{5}, true, priv1)
	app.Commit()
	// Check that claiming a covenant twice fails
	SignCheckDeliver(t, app, settleCov1, []int64{6}, false, priv1)

	// Check that balances are as expected.
	CheckBalance(t, app, addr1, "5000foocoin")
	CheckBalance(t, app, addr2, "1000foocoin")
}
