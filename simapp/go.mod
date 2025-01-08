module cosmossdk.io/simapp

go 1.23.3

// Here are the short-lived replace from the SimApp
// Replace here are pending PRs, or version to be tagged
// replace (
// 	<temporary replace>
// )

// SimApp on main always tests the latest extracted SDK modules importing the sdk
replace (
	cosmossdk.io/client/v2 => ../client/v2
	cosmossdk.io/indexer/postgres => ../indexer/postgres
	cosmossdk.io/store => ../store
	cosmossdk.io/tools/benchmark => ../tools/benchmark
	cosmossdk.io/tools/confix => ../tools/confix
	cosmossdk.io/x/accounts => ../x/accounts
	cosmossdk.io/x/accounts/defaults/base => ../x/accounts/defaults/base
	cosmossdk.io/x/accounts/defaults/lockup => ../x/accounts/defaults/lockup
	cosmossdk.io/x/accounts/defaults/multisig => ../x/accounts/defaults/multisig
	cosmossdk.io/x/authz => ../x/authz
	cosmossdk.io/x/bank => ../x/bank
	cosmossdk.io/x/circuit => ../x/circuit
	cosmossdk.io/x/consensus => ../x/consensus
	cosmossdk.io/x/distribution => ../x/distribution
	cosmossdk.io/x/epochs => ../x/epochs
	cosmossdk.io/x/evidence => ../x/evidence
	cosmossdk.io/x/feegrant => ../x/feegrant
	cosmossdk.io/x/gov => ../x/gov
	cosmossdk.io/x/group => ../x/group
	cosmossdk.io/x/mint => ../x/mint
	cosmossdk.io/x/nft => ../x/nft
	cosmossdk.io/x/params => ../x/params
	cosmossdk.io/x/protocolpool => ../x/protocolpool
	cosmossdk.io/x/slashing => ../x/slashing
	cosmossdk.io/x/staking => ../x/staking
	cosmossdk.io/x/upgrade => ../x/upgrade
)

// Below are the long-lived replace of the SimApp
replace (
	// use cosmos fork of keyring
	github.com/99designs/keyring => github.com/cosmos/keyring v1.2.0
	// Simapp always use the latest version of the cosmos-sdk
	github.com/cosmos/cosmos-sdk => ../.
	// Fix upstream GHSA-h395-qcrw-5vmq and GHSA-3vp4-m3rf-835h vulnerabilities.
	// TODO Remove it: https://github.com/cosmos/cosmos-sdk/issues/10409
	github.com/gin-gonic/gin => github.com/gin-gonic/gin v1.9.1
	// replace broken goleveldb
	github.com/syndtr/goleveldb => github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7
)
