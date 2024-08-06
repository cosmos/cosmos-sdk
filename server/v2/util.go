package serverv2

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	corectx "cosmossdk.io/core/context"
	"cosmossdk.io/log"
)

// SetCmdServerContext sets a command's Context value to the provided argument.
// If the context has not been set, set the given context as the default.
func SetCmdServerContext(cmd *cobra.Command, viper *viper.Viper, logger log.Logger) error {
	var cmdCtx context.Context
	if cmd.Context() == nil {
		cmdCtx = context.Background()
	} else {
		cmdCtx = cmd.Context()
	}

	cmdCtx = context.WithValue(cmdCtx, corectx.LoggerContextKey, logger)
	cmdCtx = context.WithValue(cmdCtx, corectx.ViperContextKey, viper)
	cmd.SetContext(cmdCtx)

	return nil
}

func GetViperFromCmd(cmd *cobra.Command) *viper.Viper {
	value := cmd.Context().Value(corectx.ViperContextKey)
	v, ok := value.(*viper.Viper)
	if !ok {
		panic(fmt.Sprintf("incorrect viper type %T: expected *viper.Viper. Have you forgot to set the viper in the command context?", value))
	}
	return v
}

func GetLoggerFromCmd(cmd *cobra.Command) log.Logger {
	v := cmd.Context().Value(corectx.LoggerContextKey)
	logger, ok := v.(log.Logger)
	if !ok {
		panic(fmt.Sprintf("incorrect logger type %T: expected log.Logger. Have you forgot to set the logger in the command context?", v))
	}

	return logger
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
