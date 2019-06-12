package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	id          = "1"
	owner       = userAddr1
	name        = "cool token"
	description = "a very cool token"
	image       = "https://google.com/token-1.png"
	tokenURI    = "https://google.com/token-1.json"
)

func TestBaseNFTGetMethods(t *testing.T) {

	testNFT := NewBaseNFT(id, owner, name, description, image, tokenURI)

	require.Equal(t, id, testNFT.GetID())
	require.Equal(t, owner, testNFT.GetOwner())
	require.Equal(t, name, testNFT.GetName())
	require.Equal(t, description, testNFT.GetDescription())
	require.Equal(t, image, testNFT.GetImage())
	require.Equal(t, tokenURI, testNFT.GetTokenURI())
}

func TestBaseNFTSetMethods(t *testing.T) {
	owner2 := userAddr2
	name2 := "cooler token"
	description2 := "a super cool token"
	image2 := "https://google.com/token-2.png"
	tokenURI2 := "https://google.com/token-2.json"

	testNFT := NewBaseNFT(id, owner, name, description, image, tokenURI)
	require.Equal(t, owner, testNFT.GetOwner())
	require.Equal(t, name, testNFT.GetName())
	require.Equal(t, description, testNFT.GetDescription())
	require.Equal(t, image, testNFT.GetImage())
	require.Equal(t, tokenURI, testNFT.GetTokenURI())

	testNFT.SetOwner(owner2)
	require.Equal(t, owner2, testNFT.GetOwner())

	testNFT.EditMetadata(name2, description2, image2, tokenURI2)
	require.Equal(t, name2, testNFT.GetName())
	require.Equal(t, description2, testNFT.GetDescription())
	require.Equal(t, image2, testNFT.GetImage())
	require.Equal(t, tokenURI2, testNFT.GetTokenURI())
}

func TestBaseNFTStringFormat(t *testing.T) {
	testNFT := NewBaseNFT(id, owner, name, description, image, tokenURI)
	expected := fmt.Sprintf(`ID:				%s
Owner:			%s
Name:			%s
Description: 	%s
Image:			%s
TokenURI:		%s`,
		id, owner, name, description, image, tokenURI)
	require.Equal(t, expected, testNFT.String())
}
