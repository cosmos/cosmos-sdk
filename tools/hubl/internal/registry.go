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
		if idx := strings.Index(apiEntry.Address, "://"); idx != -1 {
			data.APIs.GRPC[i].Address = apiEntry.Address[idx+3:]
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
	return data, nil
}

func SelectGRPCEndpoints(chain string) (string, error) {
	entry, err := GetChainRegistryEntry(chain)
	if err != nil || len(entry.APIs.GRPC) == 0 {
		if err != nil {
			// print error here so that user can know what happened and decide what to do next
			fmt.Printf("Failed to load data for %s in the chain registry: %v\n", chain, err)
		} else {
			fmt.Printf("Found empty gRPC endpoint of %s in the chain registry.\n", chain)
		}
		fmt.Println("Specify a custom gRPC endpoint manually.")
		prompt := &promptui.Prompt{
			Label: "Enter a gRPC endpoint that you trust",
		}
		return prompt.Run()
	}
	fmt.Printf("Found data for %s in the chain registry\n", chain)

	var items []string
	for _, apiEntry := range entry.APIs.GRPC {
		items = append(items, fmt.Sprintf("%s: %s", apiEntry.Provider, apiEntry.Address))
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
