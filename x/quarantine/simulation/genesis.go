package simulation

import (
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/quarantine"
)

const (
	QuarantineOptIn    = "quarantine-opt-in"
	QuarantineAutoResp = "quarantine-auto-resp"
	QuarantineFunds    = "quarantine-funds"
)

// RandomQuarantinedAddresses randomly selects accounts from the ones provided to be quarantined.
func RandomQuarantinedAddresses(r *rand.Rand, accounts []simtypes.Account) []string {
	// Max number of addresses:
	// 15% each: 1, 2, 3, 4, 5
	// 5% each: 6, 7, 8, 9, 0
	// Each provided account has a 25% chance to be quarantined.
	max := 0
	maxR := r.Intn(20)
	switch {
	case maxR <= 14:
		max = maxR/3 + 1
	case maxR >= 15 && maxR <= 18:
		max = maxR - 9
	case maxR == 19:
		max = 0
	default:
		panic(sdkerrors.ErrLogic.Wrapf("address max count random number case %d not present in switch", maxR))
	}

	if max == 0 {
		return nil
	}

	rv := make([]string, 0)
	for _, acct := range accounts {
		if r.Intn(4) == 0 {
			rv = append(rv, acct.Address.String())
		}
		if len(rv) >= max {
			break
		}
	}

	if len(rv) == 0 {
		return nil
	}

	return rv
}

// RandomQuarantineAutoResponses randomly defines some auto-responses for some of the provided addresses (and maybe others).
func RandomQuarantineAutoResponses(r *rand.Rand, quarantinedAddrs []string) []*quarantine.AutoResponseEntry {
	addrs := make([]string, 0)
	// First, identify the address that will have some auto-responses.
	// Each quarantined address has a 50% chance of having entries.
	for _, addr := range quarantinedAddrs {
		if r.Intn(2) == 0 {
			addrs = append(addrs, addr)
		}
	}

	// Then, maybe add some new ones. 25% each for 0, 1, 2, or 3 more.
	for _, acct := range simtypes.RandomAccounts(r, r.Intn(4)) {
		addrs = append(addrs, acct.Address.String())
	}

	if len(addrs) == 0 {
		return nil
	}

	rv := make([]*quarantine.AutoResponseEntry, 0)
	// For each address:
	// Number of entries: 50% 1, 25% 2, 25% 3
	// For each entry:
	// Response: 5% unspecified, 25% decline, 70% accept
	// From: 75% a brand-new address, 25% a quarantined address
	// Each quarantined address can be used only once for a given toAddr.
	// Once all quarantined address (other than the toAddr) get used, only brand-new addresses are used.
	for _, toAddr := range addrs {
		unusedQAddrs := append(make([]string, 0, len(quarantinedAddrs)), quarantinedAddrs...)

		entryCount := 0
		entryCountR := r.Intn(4)
		switch entryCountR {
		case 0, 1:
			entryCount = 1
		case 2:
			entryCount = 2
		case 3:
			entryCount = 3
		default:
			panic(sdkerrors.ErrLogic.Wrapf("entry count random number case %d not present in switch", entryCountR))
		}

		for i := 0; i < entryCount; i++ {
			entry := &quarantine.AutoResponseEntry{ToAddress: toAddr}

			respR := r.Intn(20)
			switch respR {
			case 0:
				entry.Response = quarantine.AUTO_RESPONSE_UNSPECIFIED
			case 1, 2, 3, 4, 5:
				entry.Response = quarantine.AUTO_RESPONSE_DECLINE
			case 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19:
				entry.Response = quarantine.AUTO_RESPONSE_ACCEPT
			default:
				panic(sdkerrors.ErrLogic.Wrapf("response type random number case %d not present in switch", respR))
			}

			fromR := 0
			if len(unusedQAddrs) > 0 {
				fromR = r.Intn(4)
			}
			switch fromR {
			case 0, 1, 2:
				acct := simtypes.RandomAccounts(r, 1)
				entry.FromAddress = acct[0].Address.String()
			case 3:
				indR := r.Intn(len(unusedQAddrs))
				entry.FromAddress = unusedQAddrs[indR]
				unusedQAddrs[indR] = unusedQAddrs[len(unusedQAddrs)-1]
				unusedQAddrs = unusedQAddrs[:len(unusedQAddrs)-1]
			default:
				panic(sdkerrors.ErrLogic.Wrapf("address from random number case %d not present in switch", fromR))
			}

			rv = append(rv, entry)
		}
	}

	return rv
}

// RandomQuarantinedFunds randomly generates some quarantined funds for some of the provided addresses.
func RandomQuarantinedFunds(r *rand.Rand, quarantinedAddrs []string) []*quarantine.QuarantinedFunds {
	addrs := make([]string, 0)
	// Each quarantined address has a 75% chance of having entries.
	for _, addr := range quarantinedAddrs {
		if r.Intn(4) != 0 {
			addrs = append(addrs, addr)
		}
	}

	if len(addrs) == 0 {
		return nil
	}

	rv := make([]*quarantine.QuarantinedFunds, 0)
	// For each address:
	// Number of entries: 50% 1, 25% 2, 25% 3
	// For each entry:
	// Number of from addresses: 75% 1, 20% 2, 5% 3
	// Coins: 1 to 1000 (inclusive) of the bond denom.
	// Declined: 80% false, 20% true

	for _, toAddr := range addrs {
		entryCount := 0
		entryCountR := r.Intn(4)
		switch entryCountR {
		case 0, 1:
			entryCount = 1
		case 2:
			entryCount = 2
		case 3:
			entryCount = 3
		default:
			panic(sdkerrors.ErrLogic.Wrapf("entry count random number case %d not present in switch", entryCountR))
		}

		for i := 0; i < entryCount; i++ {
			addrCount := 0
			addrCountR := r.Intn(20)
			switch addrCountR {
			case 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14:
				addrCount = 1
			case 15, 16, 17, 18:
				addrCount = 2
			case 19:
				addrCount = 3
			default:
				panic(sdkerrors.ErrLogic.Wrapf("address count random number case %d not present in switch", addrCountR))
			}

			entry := &quarantine.QuarantinedFunds{
				ToAddress:               toAddr,
				Coins:                   sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, r.Int63n(1000)+1)),
				Declined:                r.Intn(5) == 0,
				UnacceptedFromAddresses: make([]string, addrCount),
			}

			for a, acct := range simtypes.RandomAccounts(r, addrCount) {
				entry.UnacceptedFromAddresses[a] = acct.Address.String()
			}

			rv = append(rv, entry)
		}
	}

	return rv
}

// RandomizedGenState generates a random GenesisState for the quarantine module.
func RandomizedGenState(simState *module.SimulationState, fundsHolder sdk.AccAddress) {
	gen := &quarantine.GenesisState{}

	// QuarantinedAddresses
	simState.AppParams.GetOrGenerate(
		simState.Cdc, QuarantineOptIn, &gen.QuarantinedAddresses, simState.Rand,
		func(r *rand.Rand) { gen.QuarantinedAddresses = RandomQuarantinedAddresses(r, simState.Accounts) },
	)

	// AutoResponses
	simState.AppParams.GetOrGenerate(
		simState.Cdc, QuarantineAutoResp, &gen.AutoResponses, simState.Rand,
		func(r *rand.Rand) { gen.AutoResponses = RandomQuarantineAutoResponses(r, gen.QuarantinedAddresses) },
	)

	// QuarantinedFunds
	simState.AppParams.GetOrGenerate(
		simState.Cdc, QuarantineFunds, &gen.QuarantinedFunds, simState.Rand,
		func(r *rand.Rand) { gen.QuarantinedFunds = RandomQuarantinedFunds(r, gen.QuarantinedAddresses) },
	)

	simState.GenState[quarantine.ModuleName] = simState.Cdc.MustMarshalJSON(gen)

	totalQuarantined := sdk.Coins{}
	for _, qf := range gen.QuarantinedFunds {
		totalQuarantined = totalQuarantined.Add(qf.Coins...)
	}

	if !totalQuarantined.IsZero() {
		bankGenRaw := simState.GenState[banktypes.ModuleName]
		bankGen := banktypes.GenesisState{}
		simState.Cdc.MustUnmarshalJSON(bankGenRaw, &bankGen)
		bankGen.Balances = append(bankGen.Balances, banktypes.Balance{
			Address: fundsHolder.String(),
			Coins:   totalQuarantined,
		})
		bankGen.Supply = bankGen.Supply.Add(totalQuarantined...)

		simState.GenState[banktypes.ModuleName] = simState.Cdc.MustMarshalJSON(&bankGen)
	}
}
