package genesis

import (
	"encoding/json"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/pkg/errors"

	cmn "github.com/tendermint/tmlibs/common"
)

// KeyDelimiter is used to separate module and key in
// the options
const KeyDelimiter = "/"

// Option just holds module/key/value triples from
// parsing the genesis file
type Option struct {
	Module string
	Key    string
	Value  string
}

// InitStater is anything that can handle app options
// from genesis file. Setting the merkle store, config options,
// or anything else
type InitStater interface {
	InitState(module, key, value string) error
}

// Load parses the genesis file and sets the initial
// state based on that
func Load(app InitStater, filePath string) error {
	opts, err := GetOptions(filePath)
	if err != nil {
		return err
	}

	// execute all the genesis init options
	// abort on any error
	for _, opt := range opts {
		err = app.InitState(opt.Module, opt.Key, opt.Value)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetOptions parses the genesis file in a format
// that can easily be handed into InitStaters
func GetOptions(path string) ([]Option, error) {
	genDoc, err := load(path)
	if err != nil {
		return nil, err
	}

	opts := genDoc.AppOptions
	cnt := 1 + len(opts.Accounts) + len(opts.pluginOptions)
	res := make([]Option, 0, cnt)
	res = append(res, Option{sdk.ModuleNameBase, sdk.ChainKey, genDoc.ChainID})

	// set accounts
	for _, acct := range opts.Accounts {
		res = append(res, Option{"coin", "account", string(acct)})
	}

	// set plugin options
	for _, kv := range opts.pluginOptions {
		module, key := splitKey(kv.Key)
		res = append(res, Option{module, key, kv.Value})
	}

	return res, nil
}

type keyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// FullDoc - includes tendermint (in the json, we ignore here)
type FullDoc struct {
	ChainID    string `json:"chain_id"`
	AppOptions *Doc   `json:"app_options"`
}

// Doc - All genesis values
type Doc struct {
	Accounts      []json.RawMessage `json:"accounts"`
	PluginOptions []json.RawMessage `json:"plugin_options"`

	pluginOptions []keyValue // unmarshaled rawmessages
}

func load(filePath string) (*FullDoc, error) {
	bytes, err := cmn.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrap(err, "loading genesis file")
	}

	// the basecoin genesis go-wire/data :)
	genDoc := new(FullDoc)
	err = json.Unmarshal(bytes, genDoc)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshaling genesis file")
	}

	if genDoc.AppOptions == nil {
		genDoc.AppOptions = new(Doc)
	}

	pluginOpts, err := parseList(genDoc.AppOptions.PluginOptions)
	if err != nil {
		return nil, err
	}
	genDoc.AppOptions.pluginOptions = pluginOpts
	return genDoc, nil
}

func parseList(kvzIn []json.RawMessage) (kvz []keyValue, err error) {
	if len(kvzIn)%2 != 0 {
		return nil, errors.New("genesis cannot have an odd number of items.  Format = [key1, value1, key2, value2, ...]")
	}

	for i := 0; i < len(kvzIn); i += 2 {
		kv := keyValue{}
		rawK := []byte(kvzIn[i])
		err := json.Unmarshal(rawK, &(kv.Key))
		if err != nil {
			return nil, errors.Errorf("Non-string key: %s", string(rawK))
		}
		// convert value to string if possible (otherwise raw json)
		rawV := kvzIn[i+1]
		err = json.Unmarshal(rawV, &(kv.Value))
		if err != nil {
			kv.Value = string(rawV)
		}
		kvz = append(kvz, kv)
	}
	return kvz, nil
}

// Splits the string at the first '/'.
// if there are none, assign default module ("base").
func splitKey(key string) (string, string) {
	if strings.Contains(key, KeyDelimiter) {
		keyParts := strings.SplitN(key, KeyDelimiter, 2)
		return keyParts[0], keyParts[1]
	}
	return sdk.ModuleNameBase, key
}
