package auth

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Cache struct {
	updatedAccAddress map[string]struct{}
}

func NewCache() *Cache {
	return &Cache{
		updatedAccAddress: make(map[string]struct{}),
	}
}

func (c *Cache) AddUpdatedAccount(acc Account) {
	if _, ok := c.updatedAccAddress[acc.GetAddress().String()]; !ok {
		c.updatedAccAddress[acc.GetAddress().String()] = struct{}{}
	}
}

func (c *Cache) Flush() {
	c.updatedAccAddress = make(map[string]struct{})
}

// GetAllAccounts returns all accounts in the accountKeeper.
func (ak AccountKeeper) GetUpdatedAccAddress(ctx sdk.Context) (accs []sdk.AccAddress) {
	for acc := range ak.cache.updatedAccAddress {
		addr, err := sdk.AccAddressFromBech32(acc)
		if err == nil {
			accs = append(accs, addr)
		}
	}

	ak.cache.Flush()
	return accs
}
