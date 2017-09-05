package client

import (
	"errors"

	"github.com/tendermint/light-client/certifiers"
	certclient "github.com/tendermint/light-client/certifiers/client"
	"github.com/tendermint/light-client/certifiers/files"

	rpcclient "github.com/tendermint/tendermint/rpc/client"
)

// GetNode prepares a simple rpc.Client for the given endpoint
func GetNode(url string) rpcclient.Client {
	return rpcclient.NewHTTP(url, "/websocket")
}

// GetRPCProvider retuns a certifier compatible data source using
// tendermint RPC
func GetRPCProvider(url string) certifiers.Provider {
	return certclient.NewHTTP(url)
}

// GetLocalProvider returns a reference to a file store of headers
// wrapped with an in-memory cache
func GetLocalProvider(dir string) certifiers.Provider {
	return certifiers.NewCacheProvider(
		certifiers.NewMemStoreProvider(),
		files.NewProvider(dir),
	)
}

// GetCertifier initializes an inquiring certifier given a fixed chainID
// and a local source of trusted data with at least one seed
func GetCertifier(chainID string, trust certifiers.Provider,
	source certifiers.Provider) (*certifiers.InquiringCertifier, error) {

	// this gets the most recent verified seed
	seed, err := certifiers.LatestSeed(trust)
	if certifiers.IsSeedNotFoundErr(err) {
		return nil, errors.New("Please run init first to establish a root of trust")
	}
	if err != nil {
		return nil, err
	}
	cert := certifiers.NewInquiring(chainID, seed, trust, source)
	return cert, nil
}
