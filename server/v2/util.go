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

// SetServerContext sets the logger and viper in the context.
// The server manager expects the logger and viper to be set in the context.
func SetServerContext(ctx context.Context, viper *viper.Viper, logger log.Logger) (context.Context, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	ctx = context.WithValue(ctx, corectx.LoggerContextKey, logger)
	ctx = context.WithValue(ctx, corectx.ViperContextKey, viper)
	return ctx, nil
}

// SetCmdServerContext sets a command's Context value to the provided argument.
// The server manager expects the logger and viper to be set in the context.
// If the context has not been set, set the given context as the default.
func SetCmdServerContext(cmd *cobra.Command, viper *viper.Viper, logger log.Logger) error {
	ctx, err := SetServerContext(cmd.Context(), viper, logger)
	if err != nil {
		return err
	}
	cmd.SetContext(ctx)
	return nil
}

// GetViperFromContext returns the viper instance from the context.
func GetViperFromContext(ctx context.Context) *viper.Viper {
	value := ctx.Value(corectx.ViperContextKey)
	v, ok := value.(*viper.Viper)
	if !ok {
		panic(fmt.Sprintf("incorrect viper type %T: expected *viper.Viper. Have you forgot to set the viper in the context?", value))
	}
	return v
}

// GetViperFromCmd returns the viper instance from the command context.
func GetViperFromCmd(cmd *cobra.Command) *viper.Viper {
	return GetViperFromContext(cmd.Context())
}

// GetLoggerFromContext returns the logger instance from the context.
func GetLoggerFromContext(ctx context.Context) log.Logger {
	v := ctx.Value(corectx.LoggerContextKey)
	logger, ok := v.(log.Logger)
	if !ok {
		panic(fmt.Sprintf("incorrect logger type %T: expected log.Logger. Have you forgot to set the logger in the context?", v))
	}

	return logger
}

// GetLoggerFromCmd returns the logger instance from the command context.
func GetLoggerFromCmd(cmd *cobra.Command) log.Logger {
	return GetLoggerFromContext(cmd.Context())
}

// ExternalIP https://stackoverflow.com/questions/23558425/how-do-i-get-the-local-ip-address-in-go
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
