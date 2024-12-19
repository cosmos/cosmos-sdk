package server

import (
	"errors"
	"net"
	"os"

	cmtcfg "github.com/cometbft/cometbft/config"
	"github.com/spf13/viper"

	"cosmossdk.io/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ServerContextKey defines the context key used to retrieve a server.Context from
// a command's Context.
const ServerContextKey = sdk.ContextKey("server.context")

// Context is the server context.
// Prefer using we use viper a it tracks track all config.
// See core/context/server_context.go.
type Context struct {
	Viper  *viper.Viper
	Config *cmtcfg.Config
	Logger log.Logger
}

func NewDefaultContext() *Context {
	return NewContext(
		viper.New(),
		cmtcfg.DefaultConfig(),
		log.NewLogger(os.Stdout),
	)
}

func NewContext(v *viper.Viper, config *cmtcfg.Config, logger log.Logger) *Context {
	return &Context{v, config, logger}
}

// ExternalIP https://stackoverflow.com/questions/23558425/how-do-i-get-the-local-ip-address-in-go
// TODO there must be a better way to get external IP
func ExternalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range ifaces {
		if skipInterface(iface) {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}

		for _, addr := range addrs {
			ip := addrToIP(addr)
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}

func skipInterface(iface net.Interface) bool {
	if iface.Flags&net.FlagUp == 0 {
		return true // interface down
	}

	if iface.Flags&net.FlagLoopback != 0 {
		return true // loopback interface
	}

	return false
}

func addrToIP(addr net.Addr) net.IP {
	var ip net.IP

	switch v := addr.(type) {
	case *net.IPNet:
		ip = v.IP
	case *net.IPAddr:
		ip = v.IP
	}
	return ip
}
