package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	kr "github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	types2 "github.com/cosmos/cosmos-sdk/docs/migrations/types"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/go-bip39"
	"io"
	"net/http"
	"os"
)

// CreateMnemonic creates a new mnemonic
func CreateMnemonic() (string, error) {
	entropySeed, err := bip39.NewEntropy(256)
	if err != nil {
		return "", err
	}

	mnemonic, err := bip39.NewMnemonic(entropySeed)
	if err != nil {
		return "", err
	}

	return mnemonic, nil
}

func main() {
	url := "http://localhost:8080/"
	fmt.Println("URL:>", url)

	//mnemonic, err := CreateMnemonic()
	//if err != nil {
	//	return
	//}

	fmt.Println("mnemonic:>", types2.Mnemonic)

	keyring := kr.NewInMemory()

	info, err := keyring.NewAccount("s", types2.Mnemonic, "", "", hd.Secp256k1)
	fmt.Println("info:>", info)
	senderAddr := info.GetAddress()
	fmt.Println("sender:>", senderAddr)
	receiverAddr := secp256k1.GenPrivKey().PubKey().Address()

	receiverA := types.AccAddress(receiverAddr)

	body := &types2.BankSendBody{
		AccountNumber: 0,
		Sequence:      0,
		Sender:        senderAddr,
		Receiver:      receiverA,
		Denom:         "uatom",
		Amount:        1000,
		ChainID:       "chain1",
		Memo:          "Hello",
		Fee:           5000,
		GasAdjustment: 1.5,
		Gas:           200000,
	}

	payloadBuf := new(bytes.Buffer)
	json.NewEncoder(payloadBuf).Encode(body)
	req, err := http.NewRequest("POST", url, payloadBuf)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	fmt.Printf("response meta: %+v\n", resp)

	fmt.Println("response Status:", resp.Status)
	io.Copy(os.Stdout, resp.Body)

	defer resp.Body.Close()
}
