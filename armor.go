package crypto

import (
	"bytes"
	"io/ioutil"

	. "github.com/tendermint/go-common"
	"golang.org/x/crypto/openpgp/armor"
)

func EncodeArmor(blockType string, headers map[string]string, data []byte) string {
	buf := new(bytes.Buffer)
	w, err := armor.Encode(buf, blockType, headers)
	if err != nil {
		PanicSanity("Error encoding ascii armor: " + err.Error())
	}
	_, err = w.Write(data)
	if err != nil {
		PanicSanity("Error encoding ascii armor: " + err.Error())
	}
	err = w.Close()
	if err != nil {
		PanicSanity("Error encoding ascii armor: " + err.Error())
	}
	return string(buf.Bytes())
}

func DecodeArmor(armorStr string) (blockType string, headers map[string]string, data []byte, err error) {
	buf := bytes.NewBufferString(armorStr)
	block, err := armor.Decode(buf)
	if err != nil {
		return "", nil, nil, err
	}
	data, err = ioutil.ReadAll(block.Body)
	if err != nil {
		return "", nil, nil, err
	}
	return block.Type, block.Header, data, nil
}
