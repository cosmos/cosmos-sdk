package genesis

import (
	"encoding/json"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cmn "github.com/tendermint/tmlibs/common"
)

const genesisFilepath = "./testdata/genesis.json"

func TestParseList(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	bytes, err := cmn.ReadFile(genesisFilepath)
	require.Nil(err, "loading genesis file %+v", err)

	// the basecoin genesis go-wire/data :)
	genDoc := new(FullDoc)
	err = json.Unmarshal(bytes, genDoc)
	require.Nil(err, "unmarshaling genesis file %+v", err)

	pluginOpts, err := parseList(genDoc.AppOptions.PluginOptions)
	require.Nil(err, "%+v", err)
	genDoc.AppOptions.pluginOptions = pluginOpts

	assert.Equal(genDoc.AppOptions.pluginOptions[0].Key, "plugin1/key1")
	assert.Equal(genDoc.AppOptions.pluginOptions[1].Key, "plugin1/key2")
	assert.Equal(genDoc.AppOptions.pluginOptions[0].Value, "value1")
	assert.Equal(genDoc.AppOptions.pluginOptions[1].Value, "value2")
}

func TestGetOptions(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

	opts, err := GetOptions(genesisFilepath)
	require.Nil(err, "loading genesis file %+v", err)

	require.Equal(4, len(opts))
	chain := opts[0]
	assert.Equal(sdk.ModuleNameBase, chain.Module)
	assert.Equal(sdk.ChainKey, chain.Key)
	assert.Equal("foo_bar_chain", chain.Value)

	acct := opts[1]
	assert.Equal("coin", acct.Module)
	assert.Equal("account", acct.Key)

	p1 := opts[2]
	assert.Equal("plugin1", p1.Module)
	assert.Equal("key1", p1.Key)
	assert.Equal("value1", p1.Value)

	p2 := opts[3]
	assert.Equal("plugin1", p2.Module)
	assert.Equal("key2", p2.Key)
	assert.Equal("value2", p2.Value)
}

func TestSplitKey(t *testing.T) {
	assert := assert.New(t)
	prefix, suffix := splitKey("foo/bar")
	assert.EqualValues("foo", prefix)
	assert.EqualValues("bar", suffix)

	prefix, suffix = splitKey("foobar")
	assert.EqualValues("base", prefix)
	assert.EqualValues("foobar", suffix)

	prefix, suffix = splitKey("some/complex/issue")
	assert.EqualValues("some", prefix)
	assert.EqualValues("complex/issue", suffix)

}
