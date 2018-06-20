package crypto

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"golang.org/x/crypto/openpgp/armor"
)

func EncodeArmor(blockType string, headers map[string]string, data []byte) string {
	buf := new(bytes.Buffer)
	w, err := armor.Encode(buf, blockType, headers)
	if err != nil {
		panic(fmt.Errorf("could not encode ascii armor: %s", err))
	}
	_, err = w.Write(data)
	if err != nil {
		panic(fmt.Errorf("could not encode ascii armor: %s", err))
	}
	err = w.Close()
	if err != nil {
		panic(fmt.Errorf("could not encode ascii armor: %s", err))
	}
	return buf.String()
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
