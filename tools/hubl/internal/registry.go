package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/manifoldco/promptui"
)

type ChainRegistryEntry struct {
	APIs ChainRegistryAPIs `json:"apis"`
}

type ChainRegistryAPIs struct {
	GRPC []*APIEntry `json:"grpc"`
}

type APIEntry struct {
	Address  string
	Provider string
}

func GetChainRegistryEntry(chain string) (*ChainRegistryEntry, error) {
	res, err := http.Get(fmt.Sprintf("https://raw.githubusercontent.com/cosmos/chain-registry/master/%v/chain.json", chain))
	if err != nil {
		return nil, err
	}

	bz, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	data := &ChainRegistryEntry{}
	err = json.Unmarshal(bz, data)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Found data for %s in the chain registry\n", chain)

	return data, nil
}

func SelectGRPCEndpoints(chain string) (string, error) {
	entry, err := GetChainRegistryEntry(chain)
	if err != nil {
		fmt.Printf("Unable to load data for %s in the chain registry. You'll have to specify a gRPC endpoint manually.\n", chain)
		prompt := &promptui.Prompt{
			Label: "Enter a gRPC endpoint that you trust",
		}
		return prompt.Run()
	}

	var items []string
	if entry != nil {
		for _, apiEntry := range entry.APIs.GRPC {
			items = append(items, fmt.Sprintf("%s: %s", apiEntry.Provider, apiEntry.Address))
		}
	}
	prompt := promptui.SelectWithAdd{
		Label:    fmt.Sprintf("Select a gRPC endpoint that you trust for the %s network", chain),
		Items:    items,
		AddLabel: "Custom endpoint:",
	}

	i, ep, err := prompt.Run()
	if err != nil {
		return "", err
	}

	// user selected a custom endpoint
	if i == -1 {
		return ep, nil
	}

	return entry.APIs.GRPC[i].Address, nil
}
