package cli

import (
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/x/nfts"
)

const (
	flagName        = "name"
	flagDescription = "description"
	flagImage       = "image"
	flagTokenURI    = "tokenURI"
)

func parseEditMetadataFlags() (nfts.MsgEditNFTMetadata, error) {
	msg := nfts.MsgEditNFTMetadata{}

	name := viper.GetString(flagName)
	if name != "" {
		msg.EditName = true
		msg.Name = name
	}

	description := viper.GetString(flagDescription)
	if description != "" {
		msg.EditDescription = true
		msg.Description = description
	}

	image := viper.GetString(flagImage)
	if image != "" {
		msg.EditImage = true
		msg.Image = image
	}

	tokenURI := viper.GetString(flagTokenURI)
	if tokenURI != "" {
		msg.EditTokenURI = true
		msg.TokenURI = tokenURI
	}

	return msg, nil
}
