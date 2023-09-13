package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/manifoldco/promptui"
)

type ChainRegistryEntry struct {
	APIs struct {
		GRPC []*APIEntry `json:"grpc"`
	} `json:"apis"`
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
	if err = json.Unmarshal(bz, data); err != nil {
		return nil, err
	}

	// clean-up the URL
	cleanEntries := make([]*APIEntry, 0)
	for i, apiEntry := range data.APIs.GRPC {
		// clean-up the http(s):// prefix
		if strings.Contains(apiEntry.Address, "https://") {
			data.APIs.GRPC[i].Address = strings.Replace(apiEntry.Address, "https://", "", 1)
		} else if strings.Contains(apiEntry.Address, "http://") {
			data.APIs.GRPC[i].Address = strings.Replace(apiEntry.Address, "http://", "", 1)
		}

		// remove trailing slashes
		data.APIs.GRPC[i].Address = strings.TrimSuffix(data.APIs.GRPC[i].Address, "/")

		// remove addresses without a port
		if !strings.Contains(data.APIs.GRPC[i].Address, ":") {
			continue
		}

		cleanEntries = append(cleanEntries, data.APIs.GRPC[i])
	}

	data.APIs.GRPC = cleanEntries
	fmt.Printf("Found data for %s in the chain registry\n", chain)
	return data, nil
}

func SelectGRPCEndpoints(chain string) (string, error) {
	entry, err := GetChainRegistryEntry(chain)
	if err != nil {
		fmt.Printf("Unable to load data for %s in the chain registry. Specify a custom gRPC endpoint manually.\n", chain)
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
