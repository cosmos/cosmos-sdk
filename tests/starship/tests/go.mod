module github.com/cosmos-sdk/tests/starship/tests

go 1.22

toolchain go1.22.0

replace (
	github.com/99designs/keyring => github.com/cosmos/keyring v1.2.0
	// Update to rosetta-sdk-go temporarly to have `check:spec` passing. See https://github.com/coinbase/rosetta-sdk-go/issues/449
	github.com/coinbase/rosetta-sdk-go => github.com/coinbase/rosetta-sdk-go v0.8.2-0.20221007214527-e03849ba430a
	// We always want to test against the latest version of the SDK.
	github.com/cosmos/cosmos-sdk => ../../../.
)

// SimApp on main always tests the latest extracted SDK modules importing the sdk
replace (
	cosmossdk.io/api => ../../../api
	cosmossdk.io/client/v2 => ../../../client/v2
	cosmossdk.io/core => ../../../core
	cosmossdk.io/depinject => ../../../depinject
	cosmossdk.io/simapp => ../../../simapp
	cosmossdk.io/x/accounts => ../../../x/accounts
	cosmossdk.io/x/accounts/defaults/lockup => ../../../x/accounts/defaults/lockup
	cosmossdk.io/x/auth => ../../../x/auth
	cosmossdk.io/x/authz => ../../../x/authz
	cosmossdk.io/x/bank => ../../../x/bank
	cosmossdk.io/x/circuit => ../../../x/circuit
	cosmossdk.io/x/distribution => ../../../x/distribution
	cosmossdk.io/x/epochs => ../../../x/epochs
	cosmossdk.io/x/evidence => ../../../x/evidence
	cosmossdk.io/x/feegrant => ../../../x/feegrant
	cosmossdk.io/x/gov => ../../../x/gov
	cosmossdk.io/x/group => ../../../x/group
	cosmossdk.io/x/mint => ../../../x/mint
	cosmossdk.io/x/nft => ../../../x/nft
	cosmossdk.io/x/protocolpool => ../../../x/protocolpool
	cosmossdk.io/x/slashing => ../../../x/slashing
	cosmossdk.io/x/staking => ../../../x/staking
	cosmossdk.io/x/upgrade => ../../../x/upgrade
)
