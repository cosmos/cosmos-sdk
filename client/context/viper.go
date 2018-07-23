package context

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/viper"

	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

// NewCoreContextFromViper - return a new context with parameters from the command line
func NewCoreContextFromViper() CoreContext {
	nodeURI := viper.GetString(client.FlagNode)
	var rpc rpcclient.Client
	if nodeURI != "" {
		rpc = rpcclient.NewHTTP(nodeURI, "/websocket")
	}
	chainID := viper.GetString(client.FlagChainID)
	// if chain ID is not specified manually, read default chain ID
	if chainID == "" {
		def, err := defaultChainID()
		if err != nil {
			chainID = def
		}
	}
	// TODO: Remove the following deprecation code after Gaia-7000 is launched
	keyName := viper.GetString(client.FlagName)
	if keyName != "" {
		fmt.Println("** Note --name is deprecated and will be removed next release. Please use --from instead **")
	} else {
		keyName = viper.GetString(client.FlagFrom)
	}
	return CoreContext{
		ChainID:         chainID,
		Height:          viper.GetInt64(client.FlagHeight),
		Gas:             viper.GetInt64(client.FlagGas),
		Fee:             viper.GetString(client.FlagFee),
		TrustNode:       viper.GetBool(client.FlagTrustNode),
		FromAddressName: keyName,
		NodeURI:         nodeURI,
		AccountNumber:   viper.GetInt64(client.FlagAccountNumber),
		Sequence:        viper.GetInt64(client.FlagSequence),
		Memo:            viper.GetString(client.FlagMemo),
		Client:          rpc,
		Decoder:         nil,
		AccountStore:    "acc",
		UseLedger:       viper.GetBool(client.FlagUseLedger),
		Async:           viper.GetBool(client.FlagAsync),
		JSON:            viper.GetBool(client.FlagJson),
		PrintResponse:   viper.GetBool(client.FlagPrintResponse),
	}
}

// read chain ID from genesis file, if present
func defaultChainID() (string, error) {
	cfg, err := tcmd.ParseConfig()
	if err != nil {
		return "", err
	}
	doc, err := tmtypes.GenesisDocFromFile(cfg.GenesisFile())
	if err != nil {
		return "", err
	}
	return doc.ChainID, nil
}

// EnsureAccountExists - Make sure account exists
func EnsureAccountExists(ctx CoreContext, name string) error {
	keybase, err := keys.GetKeyBase()
	if err != nil {
		return err
	}

	if name == "" {
		return errors.Errorf("must provide a from address name")
	}

	info, err := keybase.Get(name)
	if err != nil {
		return errors.Errorf("no key for: %s", name)
	}

	accAddr := sdk.AccAddress(info.GetPubKey().Address())

	Acc, err := ctx.QueryStore(auth.AddressStoreKey(accAddr), ctx.AccountStore)
	if err != nil {
		return err
	}

	// Check if account was found
	if Acc == nil {
		return errors.Errorf("No account with address %s was found in the state.\nAre you sure there has been a transaction involving it?", accAddr)
	}
	return nil
}

// EnsureAccount - automatically set account number if none provided
func EnsureAccountNumber(ctx CoreContext) (CoreContext, error) {
	// Should be viper.IsSet, but this does not work - https://github.com/spf13/viper/pull/331
	if viper.GetInt64(client.FlagAccountNumber) != 0 {
		return ctx, nil
	}
	from, err := ctx.GetFromAddress()
	if err != nil {
		return ctx, err
	}
	accnum, err := ctx.GetAccountNumber(from)
	if err != nil {
		return ctx, err
	}
	fmt.Printf("Defaulting to account number: %d\n", accnum)
	ctx = ctx.WithAccountNumber(accnum)
	return ctx, nil
}

// EnsureSequence - automatically set sequence number if none provided
func EnsureSequence(ctx CoreContext) (CoreContext, error) {
	// Should be viper.IsSet, but this does not work - https://github.com/spf13/viper/pull/331
	if viper.GetInt64(client.FlagSequence) != 0 {
		return ctx, nil
	}
	from, err := ctx.GetFromAddress()
	if err != nil {
		return ctx, err
	}
	seq, err := ctx.NextSequence(from)
	if err != nil {
		return ctx, err
	}
	fmt.Printf("Defaulting to next sequence number: %d\n", seq)
	ctx = ctx.WithSequence(seq)
	return ctx, nil
}
