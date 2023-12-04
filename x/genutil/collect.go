package genutil

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	cfg "github.com/cometbft/cometbft/config"

	bankexported "cosmossdk.io/x/bank/exported"
	stakingtypes "cosmossdk.io/x/staking/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdkruntime "github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
)

// GenAppStateFromConfig gets the genesis app state from the config
func GenAppStateFromConfig(cdc codec.JSONCodec, txEncodingConfig client.TxEncodingConfig,
	config *cfg.Config, initCfg types.InitConfig, genesis *types.AppGenesis, genBalIterator types.GenesisBalancesIterator,
	validator types.MessageValidator, valAddrCodec sdkruntime.ValidatorAddressCodec,
) (appState json.RawMessage, err error) {
	// process genesis transactions, else create default genesis.json
	appGenTxs, persistentPeers, err := CollectTxs(
		cdc, txEncodingConfig.TxJSONDecoder(), config.Moniker, initCfg.GenTxsDir, genesis, genBalIterator, validator, valAddrCodec)
	if err != nil {
		return appState, err
	}

	config.P2P.PersistentPeers = persistentPeers
	cfg.WriteConfigFile(filepath.Join(config.RootDir, "config", "config.toml"), config)

	// if there are no gen txs to be processed, return the default empty state
	if len(appGenTxs) == 0 {
		return appState, errors.New("there must be at least one genesis tx")
	}

	// create the app state
	appGenesisState, err := types.GenesisStateFromAppGenesis(genesis)
	if err != nil {
		return appState, err
	}

	appGenesisState, err = SetGenTxsInAppGenesisState(cdc, txEncodingConfig.TxJSONEncoder(), appGenesisState, appGenTxs)
	if err != nil {
		return appState, err
	}

	appState, err = json.MarshalIndent(appGenesisState, "", "  ")
	if err != nil {
		return appState, err
	}

	genesis.AppState = appState
	err = ExportGenesisFile(genesis, config.GenesisFile())

	return appState, err
}

// CollectTxs processes and validates application's genesis Txs and returns
// the list of appGenTxs, and persistent peers required to generate genesis.json.
func CollectTxs(cdc codec.JSONCodec, txJSONDecoder sdk.TxDecoder, moniker, genTxsDir string,
	genesis *types.AppGenesis, genBalIterator types.GenesisBalancesIterator,
	validator types.MessageValidator, valAddrCodec sdkruntime.ValidatorAddressCodec,
) (appGenTxs []sdk.Tx, persistentPeers string, err error) {
	// prepare a map of all balances in genesis state to then validate
	// against the validators addresses
	var appState map[string]json.RawMessage
	if err := json.Unmarshal(genesis.AppState, &appState); err != nil {
		return appGenTxs, persistentPeers, err
	}

	fos, err := os.ReadDir(genTxsDir)
	if err != nil {
		return appGenTxs, persistentPeers, err
	}

	balancesMap := make(map[string]bankexported.GenesisBalance)

	genBalIterator.IterateGenesisBalances(
		cdc, appState,
		func(balance bankexported.GenesisBalance) (stop bool) {
			addr := balance.GetAddress()
			balancesMap[addr] = balance
			return false
		},
	)

	var wg sync.WaitGroup
	type result struct {
		tx     sdk.Tx
		peerIP string
		err    error
	}

	resultsCh := make(chan *result, runtime.NumCPU())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, fo := range fos {
		if fo.IsDir() {
			continue
		}
		if !strings.HasSuffix(fo.Name(), ".json") {
			continue
		}

		wg.Add(1)
		go func(path string) {
			defer wg.Done()

			genTx, peerIP, err := parseTxsAndPeers(ctx, txJSONDecoder, moniker, path, valAddrCodec, validator, balancesMap)
			if err != nil {
				// Any error should cancel the context ASAP to prevent
				// any further processing.
				cancel()
			}
			resultsCh <- &result{tx: genTx, peerIP: peerIP, err: err}
		}(filepath.Join(genTxsDir, fo.Name()))
	}

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	// addresses and IPs (and port) validator server info
	var addressesIPs []string

	for res := range resultsCh {
		genTx, peerIP, err := res.tx, res.peerIP, res.err
		if err != nil {
			cancel()
			return appGenTxs, persistentPeers, err
		}

		addressesIPs = append(addressesIPs, peerIP)
		appGenTxs = append(appGenTxs, genTx)
	}

	sort.Strings(addressesIPs)
	persistentPeers = strings.Join(addressesIPs, ",")

	return appGenTxs, persistentPeers, nil
}

func parseTxsAndPeers(
	ctx context.Context, txJSONDecoder sdk.TxDecoder,
	moniker, jsonTxPath string,
	valAddrCodec sdkruntime.ValidatorAddressCodec,
	validator types.MessageValidator,
	balancesMap map[string]bankexported.GenesisBalance,
) (genTx sdk.Tx, peerAddr string, err error) {
	if err := ctx.Err(); err != nil {
		return nil, "", err
	}

	// get the genTx
	jsonRawTx, err := os.ReadFile(jsonTxPath)
	if err != nil {
		return nil, "", err
	}

	genTx, err = types.ValidateAndGetGenTx(jsonRawTx, txJSONDecoder, validator)
	if err != nil {
		return nil, "", err
	}

	// the memo flag is used to store
	// the ip and node-id, for example this may be:
	// "528fd3df22b31f4969b05652bfe8f0fe921321d5@192.168.2.37:26656"

	memoTx, ok := genTx.(sdk.TxWithMemo)
	if !ok {
		err = fmt.Errorf("expected TxWithMemo, got %T", genTx)
		return nil, "", err
	}
	nodeAddrIP := memoTx.GetMemo()

	// genesis transactions must be single-message
	msgs := genTx.GetMsgs()

	// TODO abstract out staking message validation back to staking
	msg := msgs[0].(*stakingtypes.MsgCreateValidator)

	// validate validator addresses and funds against the accounts in the state
	valAddr, err := valAddrCodec.StringToBytes(msg.ValidatorAddress)
	if err != nil {
		return genTx, "", err
	}

	valAccAddr := sdk.AccAddress(valAddr).String()

	delBal, delOk := balancesMap[valAccAddr]
	if !delOk {
		_, file, no, ok := runtime.Caller(1)
		if ok {
			fmt.Printf("CollectTxs-1, called from %s#%d\n", file, no)
		}

		return genTx, "", fmt.Errorf("account %s balance not in genesis state: %+v", valAccAddr, balancesMap)
	}

	_, valOk := balancesMap[valAccAddr]
	if !valOk {
		_, file, no, ok := runtime.Caller(1)
		if ok {
			fmt.Printf("CollectTxs-2, called from %s#%d - %s\n", file, no, sdk.AccAddress(msg.ValidatorAddress).String())
		}
		return genTx, "", fmt.Errorf("account %s balance not in genesis state: %+v", valAddr, balancesMap)
	}

	if delBal.GetCoins().AmountOf(msg.Value.Denom).LT(msg.Value.Amount) {
		return genTx, "", fmt.Errorf(
			"insufficient fund for delegation %v: %v < %v",
			delBal.GetAddress(), delBal.GetCoins().AmountOf(msg.Value.Denom), msg.Value.Amount,
		)
	}

	// exclude itself from persistent peers
	if msg.Description.Moniker != moniker {
		peerAddr = nodeAddrIP
	}

	return genTx, peerAddr, nil
}
