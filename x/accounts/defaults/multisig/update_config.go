package multisig

import (
	"context"

	v1 "cosmossdk.io/x/accounts/defaults/multisig/v1"
)

// Authenticate implements the authentication flow of an abstracted base account.
func (a Account) UpdateConfig(ctx context.Context, msg *v1.MsgUpdateConfigRequest) (*v1.MsgUpdateConfigResponse, error) {
	// set members
	for i := range msg.UpdatePubKeys {
		if msg.UpdateWeights[i] == 0 {
			if err := a.Members.Remove(ctx, msg.UpdatePubKeys[i]); err != nil {
				return nil, err
			}
			continue
		}
		if err := a.Members.Set(ctx, msg.UpdatePubKeys[i], msg.UpdateWeights[i]); err != nil {
			return nil, err
		}
	}

	// TODO: verify if this looks good to everyone from a UX perspective
	if msg.Config != nil {
		// set config
		if err := a.Config.Set(ctx, *msg.Config); err != nil {
			return nil, err
		}
	}

	// verify that the new set of members and config are valid
	// get the weight from the stored members
	totalWeight := uint64(0)
	err := a.Members.Walk(ctx, nil, func(_ []byte, value uint64) (stop bool, err error) {
		totalWeight += value
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	// get the config from state given that we might not have it in the message
	config, err := a.Config.Get(ctx)
	if err != nil {
		return nil, err
	}

	if err := validateConfig(config, totalWeight); err != nil {
		return nil, err
	}

	return &v1.MsgUpdateConfigResponse{}, nil
}
