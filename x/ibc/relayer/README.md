# Relayer

The `relayer` package contains some basic relayer implemenations that are meant to be used by users wishing to relay packets between IBC enalbed chains. It is also well documented and intended as a place where users who are interested in building their own relayer can come for working examples.

### Relayer Home Folder Layout 

```bash
~/.relayer
├── chains
│   ├── chain-id-1
│   │   ├── keys
│   │   │   └── keys.db
│   │   └── verifiers
│   │       └── chain-id-2.db
│   └── chain-id-2
│       ├── keys
│       │   └── keys.db
│       └── verifiers
│           └── chain-id-1.db
└── config
    └── config.yaml
```

### Configuring the Relayer

There are two major parts of `relayer` configuration:

```go
type Config struct {
    Global Global `yaml:"global"`
    Chains []Chain `yaml:"chains"`
}
```

#### Global Configuration

- Amount of time to sleep between relayer loops
- Which strategy to use for your relayer (`naieve` is the only one implemented)
- Home directory for the relayer (`~/.relayer`), to be passed in via flag

```go
type Global struct {
    RelayTimeout string `yaml:"relay-timeout"`
    Strategy     string `yaml:"strategy"`
}
```

#### Chains config

The `Chain` abstraction contains all the necessary data to connect to a given chain, query it's state, and send transactions to it. The config will contain an array of these chains (`[]Chain`). Using the `naieve` strategy it will attempt to connect all the chains to all their counterparties. The following data and tools will be needed by each `Chain`:

```go
type Chain struct {
    // Key is a reference to either the sdk.AccAddress or key name
    // which must exist in the keybase associated with this chain.
    Key   string `yaml:"key"`
    
    // The chainID for the chain
    ChainID       string `yaml:"chain-id"`

    // The sdk.AccAddress prefix for this specific chain
    AccountPrefix string `yaml:"account-prefix"`

    // HomeDir for this chainID will contain:
    // - Chain specific keybase
    // - Lite Client Verifier
	Verifier      tmlite.Verifier 
	Keybase       cryptokeys.Keybase 

    // NOTE: is there any weirdness around the codec? Should just the IBC codec suffice?
    Codec         *codec.Codec
    
    // A URL to connect to a node on this chain
    NodeRPC       string `yaml:"node-rpc"`

    // The list of counterparty chains that this chain should be connected to
    // NOTE: Relayer will connect and fwd packets for any chains listed here 
    // which also have an entry in the `[]Chain` 
    Counterparties []string `yaml:"counterparties"`

    // Gas/Fee settings, are persistent for each transaction on the chain 
	Gas                uint64       `yaml:"gas"`
	GasAdjustment      float64      `yaml:"gas-adjustment"`
	GasPrices          sdk.DecCoins `yaml:"gas-prices"`
    Fees               sdk.Coins    `yaml:"fees"`
    
    // Optional memo to add to each transaction
	Memo               string `yaml:"memo"`
}
```

### Notes:
- Relayer needs ability to do CRUD on each of the keybases it manages (e.g. `relayer keys --chain-id foo add`)
- Relayer should have ability to do CRUD on each of the lite clients it manages (e.g. `relayer lite --chain-id foo update`)
- When running (`relayer start`) Relayer needs to be able to use the keys each `Chain` configured in the `~/.relayer/config/config.yaml`
- Relayers can be kept off the open internet as there is no need to send them messages, this should help improve the scurity of the relayer even though it must keep "hot" keys.
- Relayers should gracefully handle `ErrInsufficentFunds` when/if accounts run out of funds
    * _Stretch_: Relayer should notify a configurable endpoint when it hits `ErrInsufficentFunds`