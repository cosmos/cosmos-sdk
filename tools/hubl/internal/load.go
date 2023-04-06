package internal

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"path"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	reflectionv1 "cosmossdk.io/api/cosmos/reflection/v1"
	"github.com/cockroachdb/errors"
	"github.com/hashicorp/go-multierror"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

const DefaultConfigDirName = ".hubl"

type ChainInfo struct {
	client *grpc.ClientConn

	ConfigDir     string
	Chain         string
	ModuleOptions map[string]*autocliv1.ModuleOptions
	ProtoFiles    *protoregistry.Files
	Context       context.Context
	Config        *ChainConfig
}

func NewChainInfo(configDir, chain string, config *ChainConfig) *ChainInfo {
	return &ChainInfo{
		ConfigDir: configDir,
		Chain:     chain,
		Config:    config,
		Context:   context.Background(),
	}
}

func (c *ChainInfo) getCacheDir() (string, error) {
	cacheDir := path.Join(c.ConfigDir, "cache")
	return cacheDir, os.MkdirAll(cacheDir, 0o755)
}

func (c *ChainInfo) fdsCacheFilename() (string, error) {
	cacheDir, err := c.getCacheDir()
	if err != nil {
		return "", err
	}
	return path.Join(cacheDir, fmt.Sprintf("%s.fds", c.Chain)), nil
}

func (c *ChainInfo) appOptsCacheFilename() (string, error) {
	cacheDir, err := c.getCacheDir()
	if err != nil {
		return "", err
	}
	return path.Join(cacheDir, fmt.Sprintf("%s.autocli", c.Chain)), nil
}

func (c *ChainInfo) Load(reload bool) error {
	fdSet := &descriptorpb.FileDescriptorSet{}
	fdsFilename, err := c.fdsCacheFilename()
	if err != nil {
		return err
	}

	if _, err := os.Stat(fdsFilename); os.IsNotExist(err) || reload {
		client, err := c.OpenClient()
		if err != nil {
			return err
		}

		reflectionClient := reflectionv1.NewReflectionServiceClient(client)
		fdRes, err := reflectionClient.FileDescriptors(c.Context, &reflectionv1.FileDescriptorsRequest{})
		if err != nil {
			fdSet, err = loadFileDescriptorsGRPCReflection(c.Context, client)
			if err != nil {
				return err
			}
		} else {
			fdSet = &descriptorpb.FileDescriptorSet{File: fdRes.Files}
		}

		bz, err := proto.Marshal(fdSet)
		if err != nil {
			return err
		}

		if err = os.WriteFile(fdsFilename, bz, 0o600); err != nil {
			return err
		}
	} else {
		bz, err := os.ReadFile(fdsFilename)
		if err != nil {
			return err
		}

		if err = proto.Unmarshal(bz, fdSet); err != nil {
			return err
		}
	}

	c.ProtoFiles, err = protodesc.FileOptions{AllowUnresolvable: true}.NewFiles(fdSet)
	if err != nil {
		return fmt.Errorf("error building protoregistry.Files: %w", err)
	}

	appOptsFilename, err := c.appOptsCacheFilename()
	if err != nil {
		return err
	}

	if _, err := os.Stat(appOptsFilename); os.IsNotExist(err) || reload {
		client, err := c.OpenClient()
		if err != nil {
			return err
		}

		autocliQueryClient := autocliv1.NewQueryClient(client)
		appOptionsRes, err := autocliQueryClient.AppOptions(c.Context, &autocliv1.AppOptionsRequest{})
		if err != nil {
			appOptionsRes = guessAutocli(c.ProtoFiles)
		}

		bz, err := proto.Marshal(appOptionsRes)
		if err != nil {
			return err
		}

		err = os.WriteFile(appOptsFilename, bz, 0o600)
		if err != nil {
			return err
		}

		c.ModuleOptions = appOptionsRes.ModuleOptions
	} else {
		bz, err := os.ReadFile(appOptsFilename)
		if err != nil {
			return err
		}

		var appOptsRes autocliv1.AppOptionsResponse
		err = proto.Unmarshal(bz, &appOptsRes)
		if err != nil {
			return err
		}

		c.ModuleOptions = appOptsRes.ModuleOptions
	}

	return nil
}

func (c *ChainInfo) OpenClient() (*grpc.ClientConn, error) {
	if c.client != nil {
		return c.client, nil
	}

	var res error
	for _, endpoint := range c.Config.GRPCEndpoints {
		var creds credentials.TransportCredentials
		if endpoint.Insecure {
			creds = insecure.NewCredentials()
		} else {
			creds = credentials.NewTLS(&tls.Config{
				MinVersion: tls.VersionTLS12,
			})
		}

		var err error
		c.client, err = grpc.Dial(endpoint.Endpoint, grpc.WithTransportCredentials(creds))
		if err != nil {
			res = multierror.Append(res, err)
			continue
		}

		return c.client, nil
	}

	return nil, errors.Wrapf(res, "error loading gRPC client")
}
