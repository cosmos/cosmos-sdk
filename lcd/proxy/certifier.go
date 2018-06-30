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
	//TODO this is just for proto type, here we should load the latest checkpoint. In the first run time, we should trust the validator set in genesis block
	fc, err := source.GetByHeight(1)

	if err != nil {
		return nil, err
	}

	cert, err := lcd.NewInquiringCertifier(chainID, fc, trust, source)
	if err != nil {
		return nil, err
	}

	return cert, nil
}
