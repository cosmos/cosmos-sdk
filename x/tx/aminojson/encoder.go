package aminojson

import (
	authapi "cosmossdk.io/api/cosmos/auth/v1beta1"
	"cosmossdk.io/math"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"google.golang.org/protobuf/reflect/protoreflect"
	"io"
)

func cosmosDecBytesEncoder(_ AminoJSON, v protoreflect.Value, w io.Writer) error {
	switch bz := v.Interface().(type) {
	case []byte:
		if len(bz) == 0 {
			return jsonMarshal(w, "0")
		}
		var dec math.LegacyDec
		err := dec.Unmarshal(bz)
		if err != nil {
			return err
		}
		return jsonMarshal(w, dec.String())
	default:
		return fmt.Errorf("unsupported type %T", bz)
	}
}

func cosmosDecEncoder(aj AminoJSON, v protoreflect.Value, w io.Writer) error {
	switch s := v.Interface().(type) {
	case string:
		if s == "" {
			return jsonMarshal(w, "0")
		}
		return jsonMarshal(w, s)
	default:
		return fmt.Errorf("unsupported type %T", s)
	}
}

// nullSliceAsEmptyEncoder replicates the behavior at:
// https://github.com/cosmos/cosmos-sdk/blob/be9bd7a8c1b41b115d58f4e76ee358e18a52c0af/types/coin.go#L199-L205
func nullSliceAsEmptyEncoder(aj AminoJSON, v protoreflect.Value, w io.Writer) error {
	switch list := v.Interface().(type) {
	case protoreflect.List:
		if list.Len() == 0 {
			_, err := w.Write([]byte("[]"))
			return err
		}
		return aj.marshalList(list, w)
	default:
		return fmt.Errorf("unsupported type %T", list)
	}
}

func emptyStringEncoder(_ AminoJSON, _ protoreflect.Value, w io.Writer) error {
	_, err := w.Write([]byte(`""`))
	return err
}

func jsonDefaultEncoder(_ AminoJSON, v protoreflect.Value, w io.Writer) error {
	switch val := v.Interface().(type) {
	case string, bool, int32, uint32, uint64, int64, protoreflect.EnumNumber:
		return jsonMarshal(w, val)
	default:
		return fmt.Errorf("unsupported type %T", val)
	}
}

func keyFieldEncoder(msg protoreflect.Message, w io.Writer) error {
	keyField := msg.Descriptor().Fields().ByName("key")
	if keyField == nil {
		return errors.New(`message encoder for key_field: no field named "key" found`)
	}

	bz := msg.Get(keyField).Bytes()

	if len(bz) == 0 {
		_, err := fmt.Fprint(w, "null")
		return err
	}

	_, err := fmt.Fprintf(w, `"%s"`, base64.StdEncoding.EncodeToString(bz))
	return err

}

type moduleAccountPretty struct {
	Address       string   `json:"address"`
	PubKey        string   `json:"public_key"`
	AccountNumber uint64   `json:"account_number"`
	Sequence      uint64   `json:"sequence"`
	Name          string   `json:"name"`
	Permissions   []string `json:"permissions"`
}

// moduleAccountEncoder replicates the behavior in
// https://github.com/cosmos/cosmos-sdk/blob/41a3dfeced2953beba3a7d11ec798d17ee19f506/x/auth/types/account.go#L230-L254
func moduleAccountEncoder(msg protoreflect.Message, w io.Writer) error {
	ma := msg.Interface().(*authapi.ModuleAccount)
	pretty := moduleAccountPretty{
		PubKey:      "",
		Name:        ma.Name,
		Permissions: ma.Permissions,
	}
	if ma.BaseAccount != nil {
		pretty.Address = ma.BaseAccount.Address
		pretty.AccountNumber = ma.BaseAccount.AccountNumber
		pretty.Sequence = ma.BaseAccount.Sequence
	} else {
		pretty.Address = ""
		pretty.AccountNumber = 0
		pretty.Sequence = 0
	}

	bz, err := json.Marshal(pretty)
	if err != nil {
		return err
	}
	_, err = w.Write(bz)
	return err
}
