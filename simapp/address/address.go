package address

import (
        sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
        Bech32MainPrefix = "insa"

        // PrefixAccount is the prefix for account keys
        PrefixAccount = "acc"
        // PrefixValidator is the prefix for validator keys
        PrefixValidator = "val"
        // PrefixConsensus is the prefix for consensus keys
        PrefixConsensus = "cons"
        // PrefixPublic is the prefix for public keys
        PrefixPublic = "pub"
        // PrefixOperator is the prefix for operator keys
        PrefixOperator = "oper"

        // PrefixAddress is the prefix for addresses
        PrefixAddress = "addr"

        // Bech32PrefixAccAddr defines the Bech32 prefix of an account's address
        Bech32PrefixAccAddr = Bech32MainPrefix
        // Bech32PrefixAccPub defines the Bech32 prefix of an account's public key
        Bech32PrefixAccPub = Bech32MainPrefix + PrefixPublic
        // Bech32PrefixValAddr defines the Bech32 prefix of a validator's operator address
        Bech32PrefixValAddr = Bech32MainPrefix + PrefixValidator + PrefixOperator
        // Bech32PrefixValPub defines the Bech32 prefix of a validator's operator public key
        Bech32PrefixValPub = Bech32MainPrefix + PrefixValidator + PrefixOperator + PrefixPublic
        // Bech32PrefixConsAddr defines the Bech32 prefix of a consensus node address
        Bech32PrefixConsAddr = Bech32MainPrefix + PrefixValidator + PrefixConsensus
        // Bech32PrefixConsPub defines the Bech32 prefix of a consensus node public key
        Bech32PrefixConsPub = Bech32MainPrefix + PrefixValidator + PrefixConsensus + PrefixPublic
)

func ConfigureBech32Prefix() {
        config := sdk.GetConfig()
        config.SetBech32PrefixForAccount(Bech32PrefixAccAddr, Bech32PrefixAccPub)
        config.SetBech32PrefixForValidator(Bech32PrefixValAddr, Bech32PrefixValPub)
        config.SetBech32PrefixForConsensusNode(Bech32PrefixConsAddr, Bech32PrefixConsPub)
        config.Seal()
}
