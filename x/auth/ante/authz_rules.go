package ante

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type AuthzDecorator struct {
	accountKeeper AccountKeeper
	authzKeeper   AuthzKeeper
}

// AuthzDecorator checks the authorization message grants for some rules.
func (az AuthzDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	msgs := tx.GetMsgs()
	for _, msg := range msgs {
		// Check if the message is an authorization message
		if authzMsg, ok := msg.(*authztypes.MsgGrant); ok {
			authz, err := authzMsg.Grant.GetAuthorization()
			if err != nil {
				return ctx, err
			}

			switch authzConverted := authz.(type) {
			case *banktypes.SendAuthorization:
				if checkSendAuthzRulesVoilated(authzConverted) {
					return ctx, fmt.Errorf("authz rules are not meeting")
				}

			default:
				fmt.Println("default case reached here")
			}
		}
	}

	// Continue with the transaction if all checks pass
	return next(ctx, tx, simulate)
}

func checkSendAuthzRulesVoilated(authz *banktypes.SendAuthorization) bool {
	// more rules can be added here.
	return authz.SpendLimit.IsAllGT(sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000)))
}
