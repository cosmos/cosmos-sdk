package ante

import (
	"fmt"
	"strings"

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
			options := azd.azk.GetAuthzOptions()
			fmt.Println("rules", options, err, ">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")

			switch authzConverted := authz.(type) {
			case *banktypes.SendAuthorization:
				// if err != nil && errors.Is(authztypes.ErrEmptyAuthzRules, err) {
				// 	continue
				// }

				if sendRules, ok := options["send"]; !ok {
					if checkSendAuthzRulesViolated(authzMsg, authzConverted, sendRules) {
						return ctx, fmt.Errorf("authz rules are not meeting")
					}
				}

			case *authztypes.GenericAuthorization:
				// if err != nil && errors.Is(authztypes.ErrEmptyAuthzRules, err) {
				// 	continue
				// }

				if genericRules, ok := options["generic"]; !ok {
					if checkGenericAuthzRules(authzMsg, authzConverted, genericRules) {
						return ctx, fmt.Errorf("authz rules are not meeting")
					}
				}

			default:
				fmt.Println("default case reached here")
			}
		}
	}

	// Continue with the transaction if all checks pass
	return next(ctx, tx, simulate)
}

// checkSendAuthzRulesViolated returns true if the rules are voilated
func checkSendAuthzRulesViolated(msgGrant *authztypes.MsgGrant, authz *banktypes.SendAuthorization, sendAuthzRules map[string]string) bool {
	fmt.Printf("\">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>\": %v\n", ">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")
	fmt.Printf("sendAuthzRules: %v\n", sendAuthzRules)
	if blockedAddrsStr, ok := sendAuthzRules["blockedAddresses"]; ok {
		blockedAddrs := strings.Split(blockedAddrsStr, ",")
		for _, blockedRecipient := range blockedAddrs {
			if msgGrant.Grantee == blockedRecipient {
				return true
			}
		}
	}

	if spendLimit, ok := sendAuthzRules["spendLimit"]; ok {
		if len(spendLimit) > 1 {
			limit, err := sdk.ParseCoinsNormalized(spendLimit)
			if err != nil {
				return true
			}
			if !limit.IsAllGTE(authz.SpendLimit) {
				return true
			}
		}
		return true
	}

	return false
}

func checkGenericAuthzRules(_ *authztypes.MsgGrant, authz *authztypes.GenericAuthorization, genericRules map[string]string) bool {
	if msgsStr, ok := genericRules["blockedMessages"]; ok {
		msgs := strings.Split(msgsStr, ",")
		for _, v := range msgs {
			if v == authz.Msg {
				return true
			}
		}
	}

	return false
}
