package multisig

import (
	"context"
	"errors"

	v1 "cosmossdk.io/x/accounts/defaults/multisig/v1"
)

// UpdateConfig updates the configuration of the multisig account.
func (a Account) UpdateConfig(ctx context.Context, msg *v1.MsgUpdateConfig) (*v1.MsgUpdateConfigResponse, error) {
	// set members
	for i := range msg.UpdateMembers {
		addrBz, err := a.addrCodec.StringToBytes(msg.UpdateMembers[i].Address)
		if err != nil {
			return nil, err
		}

		if msg.UpdateMembers[i].Weight == 0 {
			if err := a.Members.Remove(ctx, addrBz); err != nil {
				return nil, err
			}
			continue
		}
		if err := a.Members.Set(ctx, addrBz, msg.UpdateMembers[i].Weight); err != nil {
			return nil, err
		}
	}

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

func validateConfig(cfg v1.Config, totalWeight uint64) error {
	// check for zero values
	if cfg.Threshold <= 0 || cfg.Quorum <= 0 || cfg.VotingPeriod <= 0 {
		return errors.New("threshold, quorum and voting period must be greater than zero")
	}

	// threshold must be less than or equal to the total weight
	if totalWeight < uint64(cfg.Threshold) {
		return errors.New("threshold must be less than or equal to the total weight")
	}

	// quota must be less than or equal to the total weight
	if totalWeight < uint64(cfg.Quorum) {
		return errors.New("quorum must be less than or equal to the total weight")
	}

	return nil
}
