package proxy

import (
	"github.com/cosmos/cosmos-sdk/lcd/files"
	certclient "github.com/cosmos/cosmos-sdk/lcd/client"
	"github.com/cosmos/cosmos-sdk/lcd"
)

func GetCertifier(chainID, rootDir, nodeAddr string) (*lcd.InquiringCertifier, error) {
	trust := lcd.NewCacheProvider(
		lcd.NewMemStoreProvider(),
		files.NewProvider(rootDir),
	)

	source := certclient.NewHTTPProvider(nodeAddr)

	// XXX: total insecure hack to avoid `init`
	fc, err := source.LatestCommit()
	/* XXX
	// this gets the most recent verified commit
	fc, err := trust.LatestCommit()
	if certerr.IsCommitNotFoundErr(err) {
		return nil, errors.New("Please run init first to establish a root of trust")
	}*/
	if err != nil {
		return nil, err
	}

	cert, err := lcd.NewInquiringCertifier(chainID, fc, trust, source)
	if err != nil {
		return nil, err
	}

	return cert, nil
}
