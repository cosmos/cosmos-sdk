package remote

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"path"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	reflectionv1 "cosmossdk.io/api/cosmos/reflection/v1"
	"github.com/hashicorp/go-multierror"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

const DefaultDirName = ".cosmcli"

type ChainInfo struct {
	ModuleOptions     map[string]*autocliv1.ModuleOptions
	OpenClient        func() (*grpc.ClientConn, error)
	FileDescriptorSet *protoregistry.Files
	Context           context.Context
}

func LoadChainInfo(configDir, chain string, config *ChainConfig, reload bool) (*ChainInfo, error) {
	var client *grpc.ClientConn
	ctx := context.Background()

	cacheDir := path.Join(configDir, "cache")
	err := os.MkdirAll(cacheDir, 0755)
	if err != nil {
		return nil, err
	}

	fdSet := &descriptorpb.FileDescriptorSet{}
	//var listServicesRes *grpc_reflection_v1alpha.ListServiceResponse
	fdsFilename := path.Join(cacheDir, fmt.Sprintf("%s.fds", chain))
	if _, err := os.Stat(fdsFilename); os.IsNotExist(err) || reload {
		client, err = openClient(config)
		if err != nil {
			return nil, err
		}

		reflectionClient := reflectionv1.NewReflectionServiceClient(client)
		fdRes, err := reflectionClient.FileDescriptors(ctx, &reflectionv1.FileDescriptorsRequest{})
		if err != nil {
			fdSet, err = loadFileDescriptorsCompat(ctx, client)
			if err != nil {
				return nil, err
			}
		} else {
			fdSet = &descriptorpb.FileDescriptorSet{File: fdRes.Files}
		}

		bz, err := proto.Marshal(fdSet)
		if err != nil {
			return nil, err
		}

		err = os.WriteFile(fdsFilename, bz, 0644)
		if err != nil {
			return nil, err
		}
	} else {
		bz, err := os.ReadFile(fdsFilename)
		if err != nil {
			return nil, err
		}

		err = proto.Unmarshal(bz, fdSet)
		if err != nil {
			return nil, err
		}
	}

	files, err := protodesc.FileOptions{AllowUnresolvable: true}.NewFiles(fdSet)
	if err != nil {
		return nil, err
	}

	var appOpts map[string]*autocliv1.ModuleOptions
	appOptsFilename := path.Join(cacheDir, fmt.Sprintf("%s.autocli", chain))
	if _, err := os.Stat(appOptsFilename); os.IsNotExist(err) || reload {
		if client == nil {
			client, err = openClient(config)
			if err != nil {
				return nil, err
			}
		}

		autocliQueryClient := autocliv1.NewQueryClient(client)
		appOptionsRes, err := autocliQueryClient.AppOptions(ctx, &autocliv1.AppOptionsRequest{})
		if err != nil {
			appOptionsRes = guessAutocli(files)
		}

		bz, err := proto.Marshal(appOptionsRes)
		if err != nil {
			return nil, err
		}

		err = os.WriteFile(appOptsFilename, bz, 0644)
		if err != nil {
			return nil, err
		}

		appOpts = appOptionsRes.ModuleOptions
	} else {
		bz, err := os.ReadFile(appOptsFilename)
		if err != nil {
			return nil, err
		}

		var appOptsRes autocliv1.AppOptionsResponse
		err = proto.Unmarshal(bz, &appOptsRes)
		if err != nil {
			return nil, err
		}

		appOpts = appOptsRes.ModuleOptions
	}

	return &ChainInfo{
		ModuleOptions: appOpts,
		OpenClient: func() (*grpc.ClientConn, error) {
			if client != nil {
				return client, nil
			}

			return openClient(config)
		},
		Context:           ctx,
		FileDescriptorSet: files,
	}, nil
}

func openClient(config *ChainConfig) (*grpc.ClientConn, error) {
	var errors error
	for _, endpoint := range config.GRPCEndpoints {
		var err error
		var creds credentials.TransportCredentials
		if endpoint.Insecure {
			creds = insecure.NewCredentials()
		} else {
			creds = credentials.NewTLS(&tls.Config{
				MinVersion: tls.VersionTLS12,
			})
		}
		client, err := grpc.Dial(endpoint.Endpoint, grpc.WithTransportCredentials(creds))
		if err != nil {
			errors = multierror.Append(errors, err)
			continue
		}

		return client, nil
	}

	return nil, errors
}
