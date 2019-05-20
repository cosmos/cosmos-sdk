package types

import (
	"fmt"
	"testing"
)

var (
	id          uint64 = 1
	owner              = userAddr1
	name               = "cool token"
	description        = "a very cool token"
	image              = "https://google.com/token-1.png"
	tokenURI           = "https://google.com/token-1.json"
)

func TestBaseNFTGetMethods(t *testing.T) {

	testNFT := NewBaseNFT(id, owner, tokenURI, description, image, name)

	_id := testNFT.GetID()
	if id != _id {
		t.Errorf("ID was incorrect, want: %d, got %d", id, _id)
	}

	_owner := testNFT.GetOwner()
	if !owner.Equals(_owner) {
		t.Errorf("GetOwner was incorrect, want %s, got %s", owner, _owner)
	}

	_name := testNFT.GetName()
	if name != _name {
		t.Errorf("Name was incorrect, want %s, got %s", name, _name)
	}

	_description := testNFT.GetDescription()
	if description != _description {
		t.Errorf("Description was incorrect, want %s, got %s", description, _description)
	}

	_image := testNFT.GetImage()
	if image != _image {
		t.Errorf("Image was incorrect, want %s, got %s", image, _image)
	}

	_tokenURI := testNFT.GetTokenURI()
	if tokenURI != _tokenURI {
		t.Errorf("TokenURI was incorrect, want %s, got %s", tokenURI, _tokenURI)
	}

}

func TestBaseNFTSetMethods(t *testing.T) {
	owner2 := userAddr2
	name2 := "cooler token"
	description2 := "a super cool token"
	image2 := "https://google.com/token-2.png"
	tokenURI2 := "https://google.com/token-2.json"

	testNFT := NewBaseNFT(id, owner, tokenURI, description, image, name)

	testNFT.SetOwner(owner2)
	_owner := testNFT.GetOwner()
	if !owner2.Equals(_owner) {
		t.Errorf("SetOwner was incorrect, want %s, got %s", owner2, _owner)
	}

	testNFT.EditMetadata(name2, description2, image2, tokenURI2)

	_name := testNFT.GetName()
	if name2 != _name {
		t.Errorf("EditMetadata - Name was incorrect, want %s, got %s", name2, _name)
	}

	_description := testNFT.GetDescription()
	if description2 != _description {
		t.Errorf("EditMetadata - Description was incorrect, want %s, got %s", description2, _description)
	}

	_image := testNFT.GetImage()
	if image2 != _image {
		t.Errorf("EditMetadata - Image was incorrect, want %s, got %s", image2, _image)
	}

	_tokenURI := testNFT.GetTokenURI()
	if tokenURI2 != _tokenURI {
		t.Errorf("EditMetadata - TokenURI was incorrect, want %s, got %s", tokenURI2, _tokenURI)
	}
}

func TestBaseNFTStringFormat(t *testing.T) {
	testNFT := NewBaseNFT(id, owner, tokenURI, description, image, name)
	expected := fmt.Sprintf(`ID:				%d
Owner:			%s
Name:			%s
Description: 	%s
Image:			%s
TokenURI:		%s`,
		id, owner, name, description, image, tokenURI)
	received := testNFT.String()
	if expected != received {
		t.Errorf("BaseNFT String was inforrect, want %s, got %s", received, expected)
	}
}
