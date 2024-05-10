package ante

import (
	"fmt"
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
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
	sigTx, ok := tx.(authsigning.SigVerifiableTx)
	if !ok {
		return ctx, errorsmod.Wrap(sdkerrors.ErrTxDecode, "invalid tx type")
	}

	signers, err := sigTx.GetSigners()
	if err != nil {
		return ctx, err
	}

	grantee := signers[0]

	msgs := tx.GetMsgs()
	for _, msg := range msgs {
		// Check if the message is an authorization message
		if authzMsg, ok := msg.(*authztypes.MsgExec); ok {

			rulesKeys, err := azd.azk.GetAuthzRulesKeys(ctx)
			if err != nil {
				return ctx, err
			}

			msgs, err := authzMsg.GetMessages()
			if err != nil {
				return ctx, err
			}

			for _, innerMsg := range msgs {
				switch innerMsgConverted := innerMsg.(type) {
				case *banktypes.MsgSend:
					sendRuleKeysInterface, ok := rulesKeys["Send"]
					if !ok {
						fmt.Println("no rule keys")
						continue
					}

					granter, err := azd.azk.AddressCodec().StringToBytes(innerMsgConverted.FromAddress)
					if err != nil {
						return ctx, err
					}

					_, rules := azd.azk.GetAuthzWithRules(ctx, grantee, granter, sdk.MsgTypeURL(&banktypes.MsgSend{}))
					if rules != nil {
						sendRulesKeys := sendRuleKeysInterface.([]string)
						if checkSendAuthzRulesViolated(innerMsgConverted, rules, sendRulesKeys) {
							return ctx, fmt.Errorf("authz rules are not meeting")
						}
					}
				}
			}
		}
	}

	// Continue with the transaction if all checks pass
	return next(ctx, tx, simulate)
}

// checkSendAuthzRulesViolated returns true if the rules are voilated
func checkSendAuthzRulesViolated(msg *banktypes.MsgSend, sendAuthzRules map[string]interface{}, sendRulesKeys []string) bool {
	for _, key := range sendRulesKeys {

		fmt.Printf("\">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>\": %v\n", ">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")
		fmt.Printf("sendAuthzRules: %v\n", sendAuthzRules)
		if blockedAddrsStrInt, ok := sendAuthzRules["AllowRecipients"]; key == "AllowRecipients" && ok {
			blockedAddrsStr := blockedAddrsStrInt.(string)
			blockedAddrs := strings.Split(blockedAddrsStr, ",")
			for _, blockedRecipient := range blockedAddrs {
				if msg.ToAddress == blockedRecipient {
					return true
				}
			}
		}

		if spendLimitInt, ok := sendAuthzRules["SpendLImit"]; key == "SpendLImit" && ok {
			spendLimit := spendLimitInt.(string)
			limit, err := sdk.ParseCoinsNormalized(spendLimit)
			if err != nil {
				return true
			}
			if !limit.IsAllGTE(msg.Amount) {
				return true
			}

			return true
		}

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
