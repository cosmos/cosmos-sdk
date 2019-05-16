package lcd

import (
	"bytes"
	"fmt"
	crkeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/viper"
	"github.com/tendermint/go-amino"
	tmcfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/cli"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var cdc = amino.NewCodec()

func init() {
	ctypes.RegisterAmino(cdc)
}

// CreateAddr adds an address to the key store and returns an address and seed.
// It also requires that the key could be created.
func CreateAddr(name, password string, kb crkeys.Keybase) (sdk.AccAddress, string, error) {
	var (
		err  error
		info crkeys.Info
		seed string
	)
	info, seed, err = kb.CreateMnemonic(name, crkeys.English, password, crkeys.Secp256k1)
	return sdk.AccAddress(info.GetPubKey().Address()), seed, err
}

// CreateAddr adds multiple address to the key store and returns the addresses and associated seeds in lexographical order by address.
// It also requires that the keys could be created.
func CreateAddrs(kb crkeys.Keybase, numAddrs int) (addrs []sdk.AccAddress, seeds, names, passwords []string) {
	var (
		err  error
		info crkeys.Info
		seed string
	)

	addrSeeds := AddrSeedSlice{}

	for i := 0; i < numAddrs; i++ {
		name := fmt.Sprintf("test%d", i)
		password := "1234567890"
		info, seed, err = kb.CreateMnemonic(name, crkeys.English, password, crkeys.Secp256k1)
		if err != nil {
			panic(err)
		}
		addrSeeds = append(addrSeeds, AddrSeed{Address: sdk.AccAddress(info.GetPubKey().Address()), Seed: seed, Name: name, Password: password})
	}

	sort.Sort(addrSeeds)

	for i := range addrSeeds {
		addrs = append(addrs, addrSeeds[i].Address)
		seeds = append(seeds, addrSeeds[i].Seed)
		names = append(names, addrSeeds[i].Name)
		passwords = append(passwords, addrSeeds[i].Password)
	}

	return addrs, seeds, names, passwords
}

// AddrSeed combines an Address with the mnemonic of the private key to that address
type AddrSeed struct {
	Address  sdk.AccAddress
	Seed     string
	Name     string
	Password string
}

// AddrSeedSlice implements `Interface` in sort package.
type AddrSeedSlice []AddrSeed

func (b AddrSeedSlice) Len() int {
	return len(b)
}

// Less sorts lexicographically by Address
func (b AddrSeedSlice) Less(i, j int) bool {
	// bytes package already implements Comparable for []byte.
	switch bytes.Compare(b[i].Address.Bytes(), b[j].Address.Bytes()) {
	case -1:
		return true
	case 0, 1:
		return false
	default:
		panic("not fail-able with `bytes.Comparable` bounded [-1, 1].")
	}
}

func (b AddrSeedSlice) Swap(i, j int) {
	b[j], b[i] = b[i], b[j]
}

// InitClientHome initialises client home dir.
func InitClientHome(dir string) string {
	var err error
	if dir == "" {
		dir, err = ioutil.TempDir("", "lcd_test")
		if err != nil {
			panic(err)
		}
	}
	// TODO: this should be set in NewRestServer
	// and pass down the CLIContext to achieve
	// parallelism.
	viper.Set(cli.HomeFlag, dir)
	return dir
}

// makePathname creates a unique pathname for each test. It will panic if it
// cannot get the current working directory.
func makePathname() string {
	p, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	sep := string(filepath.Separator)
	return strings.Replace(p, sep, "_", -1)
}

// GetConfig returns a Tendermint config for the test cases.
func GetConfig() *tmcfg.Config {
	pathname := makePathname()
	config := tmcfg.ResetTestRoot(pathname)

	tmAddr, _, err := server.FreeTCPAddr()
	if err != nil {
		panic(err)
	}

	rcpAddr, _, err := server.FreeTCPAddr()
	if err != nil {
		panic(err)
	}

	grpcAddr, _, err := server.FreeTCPAddr()
	if err != nil {
		panic(err)
	}

	config.P2P.ListenAddress = tmAddr
	config.RPC.ListenAddress = rcpAddr
	config.RPC.GRPCListenAddress = grpcAddr

	return config
}
