module cosmossdk.io/x/feemarket

go 1.23.1

replace github.com/cosmos/cosmos-sdk => ../../.

// TODO remove post spinning out all modules
replace (
	cosmossdk.io/api => ../../api
	cosmossdk.io/collections => ../../collections
	cosmossdk.io/core/testing => ../../core/testing
	cosmossdk.io/store => ../../store
	cosmossdk.io/x/staking => ../staking
	cosmossdk.io/x/bank => ../bank
	cosmossdk.io/x/tx => ../tx
)
