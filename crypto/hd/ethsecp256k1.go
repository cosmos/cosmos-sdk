package hd

import (
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil/hdkeychain"
	bip39 "github.com/cosmos/go-bip39"

	ethaccounts "github.com/ethereum/go-ethereum/accounts"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ethsecp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/types"
)

var (
	// EthSecp256k1 uses the Ethereum secp256k1 ECDSA parameters.
	EthSecp256k1 = ethSecp256k1Algo{}
)

type ethSecp256k1Algo struct {
}

// Name returns ethsecp256k1
func (s ethSecp256k1Algo) Name() PubKeyType {
	return EthSecp256k1Type
}

// Derive derives and returns the ethsecp256k1 private key for the given mnemonic and HD path.
func (s ethSecp256k1Algo) Derive() DeriveFn {
	return func(mnemonic string, bip39Passphrase, path string) ([]byte, error) {
		hdpath, err := ethaccounts.ParseDerivationPath(path)
		if err != nil {
			return nil, err
		}

		seed, err := bip39.NewSeedWithErrorChecking(mnemonic, bip39Passphrase)
		if err != nil {
			return nil, err
		}

		masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
		if err != nil {
			return nil, err
		}

		key := masterKey
		for _, n := range hdpath {
			key, err = key.Derive(n)
			if err != nil {
				return nil, err
			}
		}

		privateKey, err := key.ECPrivKey()
		if err != nil {
			return nil, err
		}

		privateKeyECDSA := privateKey.ToECDSA()
		derivedKey := ethcrypto.FromECDSA(privateKeyECDSA)

		return derivedKey, nil
	}
}

// Generate generates a ethsecp256k1 private key from the given bytes.
func (s ethSecp256k1Algo) Generate() GenerateFn {
	return func(bz []byte) types.PrivKey {
		var bzArr = make([]byte, ethsecp256k1.PrivKeySize)
		copy(bzArr, bz)

		return &ethsecp256k1.PrivKey{Key: bzArr}
	}
}
