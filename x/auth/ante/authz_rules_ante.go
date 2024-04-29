package ante

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type AuthzDecorator struct {
	azk AuthzKeeper
}

func NewAuthzDecorator(azk AuthzKeeper) AuthzDecorator {
	return AuthzDecorator{
		azk: azk,
	}
}

// AuthzDecorator checks the authorization message grants for some rules.
func (azd AuthzDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	msgs := tx.GetMsgs()
	for _, msg := range msgs {
		// Check if the message is an authorization message
		if authzMsg, ok := msg.(*authztypes.MsgGrant); ok {
			fmt.Println("coming here", ">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")

			authz, err := authzMsg.GetAuthorization()
			if err != nil {
				return ctx, err
			}
			rules, err := azd.azk.GetAuthzRules(ctx)
			fmt.Println("rules", rules, err, ">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")

			switch authzConverted := authz.(type) {
			case *banktypes.SendAuthorization:
				if err != nil && errors.Is(authztypes.ErrEmptyAuthzRules, err) {
					continue
				}

				if checkSendAuthzRulesVoilated(authzMsg, authzConverted, rules.Send) {
					return ctx, fmt.Errorf("authz rules are not meeting")
				}

			case *authztypes.GenericAuthorization:
				if err != nil && errors.Is(authztypes.ErrEmptyAuthzRules, err) {
					continue
				}

				if checkGenericAuthzRules(authzMsg, authzConverted, rules.Generic) {
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

func checkSendAuthzRulesVoilated(msgGrant *authztypes.MsgGrant, authz *banktypes.SendAuthorization, sendAuthzRules authztypes.SendAuthzRules) bool {
	fmt.Printf("\">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>\": %v\n", ">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")
	fmt.Printf("sendAuthzRules: %v\n", sendAuthzRules)
	if authz.SpendLimit.IsAllGT(sendAuthzRules.SpendLimit) {
		return true
	}

	for _, blockedRecipient := range sendAuthzRules.BlockedRecipients {
		if msgGrant.Grantee == blockedRecipient {
			return true
		}
	}

	return false
}

func checkGenericAuthzRules(msgGrant *authztypes.MsgGrant, authz *authztypes.GenericAuthorization, GenericAuthzRules authztypes.GenericAuthzRules) bool {
	for _, v := range GenericAuthzRules.BlockedMessages {
		if v == authz.Msg {
			return true
		}
	}

	return false
}
