package bip39

import (
	"strings"

	"github.com/bartekn/go-bip39"
)

// ValidSentenceLen defines the mnemonic sentence lengths supported by this BIP 39 library.
type ValidSentenceLen uint8

const (
	// FundRaiser is the sentence length used during the cosmos fundraiser (12 words).
	FundRaiser ValidSentenceLen = 12
	// Size of the checksum employed for the fundraiser
	FundRaiserChecksumSize = 4
	// FreshKey is the sentence length used for newly created keys (24 words).
	FreshKey ValidSentenceLen = 24
	// Size of the checksum employed for new keys
	FreshKeyChecksumSize = 8
)

// NewMnemonic will return a string consisting of the mnemonic words for
// the given sentence length.
func NewMnemonic(len ValidSentenceLen) (words []string, err error) {
	// len = (entropySize + checksum) / 11
	var entropySize int
	switch len {
	case FundRaiser:
		// entropySize = 128
		entropySize = int(len)*11 - FundRaiserChecksumSize
	case FreshKey:
		// entropySize = 256
		entropySize = int(len)*11 - FreshKeyChecksumSize
	}
	var entropy []byte
	entropy, err = bip39.NewEntropy(entropySize)
	if err != nil {
		return
	}
	var mnemonic string
	mnemonic, err = bip39.NewMnemonic(entropy)
	if err != nil {
		return
	}
	words = strings.Split(mnemonic, " ")
	return
}

// MnemonicToSeed creates a BIP 39 seed from the passed mnemonic (with an empty BIP 39 password).
// This method does not validate the mnemonics checksum.
func MnemonicToSeed(mne string) (seed []byte) {
	// we do not checksum here...
	seed = bip39.NewSeed(mne, "")
	return
}

// MnemonicToSeedWithErrChecking returns the same seed as MnemonicToSeed.
// It creates a BIP 39 seed from the passed mnemonic (with an empty BIP 39 password).
//
// Different from MnemonicToSeed it validates the checksum.
// For details on the checksum see the BIP 39 spec.
func MnemonicToSeedWithErrChecking(mne string) (seed []byte, err error) {
	seed, err = bip39.NewSeedWithErrorChecking(mne, "")
	return
}
