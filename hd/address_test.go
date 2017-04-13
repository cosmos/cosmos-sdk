package hd

/*

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestManual(t *testing.T) {
	bytes, _ := hex.DecodeString("dfac699f1618c9be4df2befe94dc5f313946ebafa386756bd4926a1ecfd7cf2438426ede521d1ee6512391bc200b7910bcbea593e68d52b874c29bdc5a308ed1")
	fmt.Println(bytes)
	puk, prk, ch, se := ComputeMastersFromSeed(string(bytes))
	fmt.Println(puk, ch, se)

	pubBytes2 := DerivePublicKeyForPath(
		HexDecode(puk),
		HexDecode(ch),
		//"44'/118'/0'/0/0",
		"0/0",
	)
	fmt.Printf("PUB2 %X\n", pubBytes2)

	privBytes := DerivePrivateKeyForPath(
		HexDecode(prk),
		HexDecode(ch),
		//"44'/118'/0'/0/0",
		//"0/0",
		"44'/118'/0'/0/0",
	)
	fmt.Printf("PRIV %X\n", privBytes)
	pubBytes := PubKeyBytesFromPrivKeyBytes(privBytes, true)
	fmt.Printf("PUB  %X\n", pubBytes)
}

*/
