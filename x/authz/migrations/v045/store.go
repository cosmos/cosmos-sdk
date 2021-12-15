package v045

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	v044 "github.com/cosmos/cosmos-sdk/x/authz/migrations/v044"
)

// MigrateStore performs in-place store migrations from v0.44 to v0.45. The
// migration includes:
//
// - pruning expired authorizations
// -
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
	store := ctx.KVStore(storeKey)
	err := addExpiredGrantsIndex(ctx, store, cdc)
	if err != nil {
		return err
	}

	return nil
}

func addExpiredGrantsIndex(ctx sdk.Context, store storetypes.KVStore, cdc codec.BinaryCodec) error {
	grantsStore := prefix.NewStore(store, v044.GrantPrefix)

	grantsIter := grantsStore.Iterator(nil, nil)
	defer grantsIter.Close()

	ggmTriples := make(map[time.Time][]*authz.GGMTriple)

	for ; grantsIter.Valid(); grantsIter.Next() {
		var grant authz.Grant
		bz := grantsIter.Value()
		if err := cdc.Unmarshal(bz, &grant); err != nil {
			return err
		}

		// delete expired authorization
		if grant.Expiration.Before(ctx.BlockTime()) {
			grantsStore.Delete(grantsIter.Key())
		} else {
			granter, grantee, msgType := v044.ParseGrantKey(grantsIter.Key())
			ggmTriple, ok := ggmTriples[grant.Expiration]

			if !ok {
				ggmTriples[grant.Expiration] = []*authz.GGMTriple{
					{
						Granter:    granter.String(),
						Grantee:    grantee.String(),
						MsgTypeUrl: msgType,
					},
				}
			} else {
				ggmTriple = append(ggmTriple, &authz.GGMTriple{
					Granter:    granter.String(),
					Grantee:    grantee.String(),
					MsgTypeUrl: msgType,
				})
				ggmTriples[grant.Expiration] = ggmTriple
			}
		}

	}

	for k, v := range ggmTriples {
		queueKey := GrantQueueKey(k)
		bz, err := cdc.Marshal(&authz.GrantQueueItem{
			GgmTriples: v,
		})
		if err != nil {
			return err
		}
		store.Set(queueKey, bz)
	}

	return nil
}
