package context

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/viper"

	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
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
	fromStr := viper.GetString(client.FlagName)
	var keyNames []string
	if fromStr != "" {
		keyNames = nil
		fmt.Println("** Note --name is deprecated and will be removed next release. Please use --from instead **")
	} else {
		fromStr = viper.GetString(client.FlagFrom)
		keyNames = stringToList(fromStr)
	}

	accNums := normalizeIntList(stringToIntList(viper.GetString(client.FlagAccountNumber)), len(keyNames))
	seqs := normalizeIntList(stringToIntList(viper.GetString(client.FlagSequence)), len(keyNames))

	return CoreContext{
		ChainID:          chainID,
		Height:           viper.GetInt64(client.FlagHeight),
		Gas:              viper.GetInt64(client.FlagGas),
		Fee:              viper.GetString(client.FlagFee),
		TrustNode:        viper.GetBool(client.FlagTrustNode),
		FromAddressNames: keyNames,
		NodeURI:          nodeURI,
		AccountNumbers:   accNums,
		Sequences:        seqs,
		Memo:             viper.GetString(client.FlagMemo),
		Client:           rpc,
		Decoder:          nil,
		AccountStore:     "acc",
		UseLedger:        viper.GetBool(client.FlagUseLedger),
		Async:            viper.GetBool(client.FlagAsync),
		JSON:             viper.GetBool(client.FlagJson),
		PrintResponse:    viper.GetBool(client.FlagPrintResponse),
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

// EnsureAccount - automatically set account number if none provided
func EnsureAccountNumbers(ctx CoreContext) (CoreContext, error) {
	addrs, err := ctx.GetFromAddresses()
	if err != nil {
		return ctx, err
	}

	var accNums []int64
	var err2 error
	for i, accnum := range ctx.AccountNumbers {
		if accnum != 0 {
			continue
		}
		accNums[i], err2 = ctx.GetAccountNumber(addrs[i])
		if err2 != nil {
			return ctx, err2
		}
		fmt.Printf("Defaulting to account number: %d for account: %s\n", accnum, ctx.FromAddressNames[i])
	}

	ctx = ctx.WithAccountNumbers(accNums)
	return ctx, nil
}

// EnsureSequence - automatically set sequence number if none provided
func EnsureSequences(ctx CoreContext) (CoreContext, error) {
	addrs, err := ctx.GetFromAddresses()
	if err != nil {
		return ctx, err
	}

	var seqs []int64
	var err2 error
	for i, seq := range ctx.Sequences {
		if seq != 0 {
			continue
		}
		seqs[i], err2 = ctx.NextSequence(addrs[i])
		if err2 != nil {
			return ctx, err2
		}
		fmt.Printf("Defaulting to next sequence: %d for account: %s\n", seq, ctx.FromAddressNames[i])
	}

	ctx = ctx.WithSequences(seqs)
	return ctx, nil
}

func stringToList(str string) []string {
	lst := strings.Split(str, ",")
	for i, s := range lst {
		lst[i] = strings.Trim(s, " ")
	}
	return lst
}

func stringToIntList(str string) []int64 {
	lst := stringToList(str)
	intList := make([]int64, len(lst))
	for i, s := range lst {
		// anything that is not an integer will be interpreted as 0.
		intList[i], _ = strconv.ParseInt(s, 10, 64)
	}
	return intList
}

// will enforce that accountNumber lists and sequence lists are same length as Accounts list
func normalizeIntList(nums []int64, length int) []int64 {
	newLst := make([]int64, length)
	for i, n := range nums {
		newLst[i] = n
	}
	return newLst
}
