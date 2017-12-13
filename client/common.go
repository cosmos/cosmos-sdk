package client

import (
	"errors"

	"github.com/tendermint/tendermint/lite"
	certclient "github.com/tendermint/tendermint/lite/client"
	certerr "github.com/tendermint/tendermint/lite/errors"
	"github.com/tendermint/tendermint/lite/files"

	rpcclient "github.com/tendermint/tendermint/rpc/client"
)

// GetNode prepares a simple rpc.Client for the given endpoint
func GetNode(url string) rpcclient.Client {
	return rpcclient.NewHTTP(url, "/websocket")
}

// GetRPCProvider retuns a certifier compatible data source using
// tendermint RPC
func GetRPCProvider(url string) lite.Provider {
	return certclient.NewHTTPProvider(url)
}

// GetLocalProvider returns a reference to a file store of headers
// wrapped with an in-memory cache
func GetLocalProvider(dir string) lite.Provider {
	return lite.NewCacheProvider(
		lite.NewMemStoreProvider(),
		files.NewProvider(dir),
	)
}

// GetCertifier initializes an inquiring certifier given a fixed chainID
// and a local source of trusted data with at least one seed
func GetCertifier(chainID string, trust lite.Provider,
	source lite.Provider) (*lite.Inquiring, error) {

	// this gets the most recent verified commit
	fc, err := trust.LatestCommit()
	if certerr.IsCommitNotFoundErr(err) {
		return nil, errors.New("Please run init first to establish a root of trust")
	}
	if err != nil {
		return nil, err
	}
	cert := lite.NewInquiring(chainID, fc, trust, source)
	return cert, nil
}
