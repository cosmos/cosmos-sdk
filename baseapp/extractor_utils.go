package baseapp

import (
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
	"io"
	"os"
	"reflect"
	"strings"
)

func recordTxData(app *BaseApp, txBytes []byte, tx sdk.Tx, result abci.ResponseDeliverTx) {
	ctx := app.getContextForTx(runTxModeDeliver, txBytes)
	txHash := fmt.Sprintf("%X", tmhash.Sum(txBytes))

	sdktx, _ := tx.(auth.StdTx)
	jsonTags, _ := codec.Cdc.MarshalJSON(sdk.TagsToStringTags(result.Tags))
	jsonMsgs := MsgsToString(sdktx.GetMsgs())
	jsonFee, _ := codec.Cdc.MarshalJSON(sdktx.Fee)
	jsonMemo, _ := codec.Cdc.MarshalJSON(sdktx.GetMemo())

	if err := recordMsgData(ctx, tx, txHash); err != nil {
		panic(fmt.Sprintf("error: (%v) while trying to record message data\n", err))
	}

	f, err := os.OpenFile(fmt.Sprintf("./extract/progress/txs.%d.%s", ctx.BlockHeight(), ctx.ChainID()), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		panic(fmt.Sprintf("error: (%v) while opening progress/txs file\n", err))
	}
	defer f.Close()

	f.WriteString(fmt.Sprintf("%s,%d,%d,%d,%d,$%s$,$%s$,$%s$,$%s$,$%s$,%s,%s\n",
		txHash,
		ctx.BlockHeight(),
		uint32(result.Code),
		int64(result.GasWanted),
		int64(result.GasUsed),
		strings.ReplaceAll(result.Log, "$", "\\$"),
		strings.ReplaceAll(string(jsonMemo), "$", "\\$"),
		strings.ReplaceAll(string(jsonFee), "$", "\\$"),
		strings.ReplaceAll(string(jsonTags), "$", "\\$"),
		strings.ReplaceAll(string(jsonMsgs), "$", "\\$"),
		ctx.BlockHeader().Time.Format("2006-01-02 15:04:05"),
		ctx.ChainID()))
}

func recordMsgData(ctx sdk.Context, tx sdk.Tx, txHash string) error {
	f, err := os.OpenFile(fmt.Sprintf("./extract/progress/messages.%d.%s", ctx.BlockHeight(), ctx.ChainID()), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("error: (%w) while trying to open progress/messages file", err)
	}
	defer f.Close()

	for idx, msg := range tx.GetMsgs() {
		msgString, _ := codec.Cdc.MarshalJSON(msg)
		f.WriteString(fmt.Sprintf("%s,%d,%s,$%s$,%s,%s\n",
			txHash,
			idx,
			msg.Type(),
			strings.ReplaceAll(string(msgString), "$", "\\$"),
			ctx.BlockHeader().Time.Format("2006-01-02 15:04:05"),
			ctx.ChainID()))
		if err := extractAddresses(msg, string(txHash), ctx.BlockHeight(), idx, ctx.ChainID()); err != nil {
			return fmt.Errorf("error: (%w) while trying to extract the address", err)
		}
	}

	return nil
}

func getAddresses(msg sdk.Msg) []sdk.Address {
	var addrs []sdk.Address
	m := BasicMsgStruct{}
	_ = json.Unmarshal(msg.GetSignBytes(), &m)
	ref := reflect.ValueOf(&m.Value).Elem()
	for i := 0; i < ref.NumField(); i++ {
		addr := ref.Field(i).Interface()
		sdkAddr, ok := addr.(sdk.Address) // cast to address interface so we have access to the String() method, which bech32ifies the address
		if ok && !sdkAddr.Empty() {
			addrs = append(addrs, sdkAddr)
		}
	}
	return addrs
}

func extractAddresses(msg sdk.Msg, hash string, height int64, idx int, chainid string) error {
	addrs := getAddresses(msg)
	if len(addrs) != 0 {
		f, err := os.OpenFile(fmt.Sprintf("./extract/progress/addresses.%d.%s", height, chainid), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return fmt.Errorf("error: (%w) while trying to open progress/addresses file", err)
		}
		defer f.Close()

		for _, addr := range addrs {
			f.WriteString(fmt.Sprintf("%s,%d,%s,%s\n", hash, idx, addr.String(), chainid))
		}
	}

	return nil
}

func copyFile(destination string, source string) error {
	sourceFile, err := os.OpenFile(source, os.O_RDONLY, 0644)
	if err != nil {
		// It is not an error if source file does not exists
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("error: (%v) while trying to open source file while copying", err)
	}
	defer sourceFile.Close()

	destinationFile, err := os.OpenFile(destination, os.O_WRONLY | os.O_CREATE | os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("error: (%v) while trying to open destination file while copying", err)
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return fmt.Errorf("error: (%v) while trying to copy source file into destination", err)
	}

	return nil
}

func commitUncheckedFiles(ctx sdk.Context) {
	for _, key := range []string{"delegations", "unbond", "balance", "rewards"} {
		err := copyFile(fmt.Sprintf("./extract/progress/%s.%d.%s", key, ctx.BlockHeight(), ctx.ChainID()), fmt.Sprintf("./extract/unchecked/%s.%d.%s", key, ctx.BlockHeight(), ctx.ChainID()))
		if err != nil {
			panic(fmt.Sprintf("error: (%v) while commiting unchecked file\n", err))
		}
		// No need for the file now
		if err := os.Remove(fmt.Sprintf("./extract/unchecked/%s.%d.%s", key, ctx.BlockHeight(), ctx.ChainID())); err != nil && !os.IsNotExist(err) {
			fmt.Printf("error: (%v) while removing unchecked file after commiting\n", err)
		}
	}
}

func deleteUncheckedFiles(ctx sdk.Context) {
	for _, key := range []string{"delegations", "unbond", "balance", "rewards"} {
		if err := os.Remove(fmt.Sprintf("./extract/unchecked/%s.%d.%s", key, ctx.BlockHeight(), ctx.ChainID())); err != nil && !os.IsNotExist(err) {
			fmt.Printf("error: (%v) while removing unchecked file\n", err)
		}
	}
}
