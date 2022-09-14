package runtime

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"

	"github.com/cosmos/gogoproto/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
)

type appConfigService struct {
	appv1alpha1.UnimplementedQueryServer
	appConfig *appv1alpha1.Config
	files     *descriptorpb.FileDescriptorSet
}

func newAppConfigService(appConfig *appv1alpha1.Config) (*appConfigService, error) {
	allFds := proto.AllFileDescriptors()
	fds := &descriptorpb.FileDescriptorSet{}

	for _, compressedBz := range allFds {
		rdr, err := gzip.NewReader(bytes.NewReader(compressedBz))
		if err != nil {
			return nil, err
		}

		bz, err := io.ReadAll(rdr)
		if err != nil {
			return nil, err
		}

		fd := &descriptorpb.FileDescriptorProto{}
		err = proto.Unmarshal(bz, fd)
		if err != nil {
			return nil, err
		}

		fds.File = append(fds.File, fd)
	}

	return &appConfigService{appConfig: appConfig, files: fds}, nil
}

func (a *appConfigService) Config(context.Context, *appv1alpha1.QueryConfigRequest) (*appv1alpha1.QueryConfigResponse, error) {
	return &appv1alpha1.QueryConfigResponse{Config: a.appConfig}, nil
}

func (a *appConfigService) FileDescriptorSet(context.Context, *appv1alpha1.QueryFileDescriptorSetRequest) (*appv1alpha1.QueryFileDescriptorSetResponse, error) {
	return &appv1alpha1.QueryFileDescriptorSetResponse{Files: a.files}, nil
}

var _ appv1alpha1.QueryServer = &appConfigService{}

type autocliService struct {
	autocliv1.UnimplementedRemoteInfoServiceServer

	moduleOptions map[string]*autocliv1.ModuleOptions
}

func newAutocliService(cliConfigs map[string]CLIConfig) *autocliService {
	moduleOptions := map[string]*autocliv1.ModuleOptions{}
	for module, cfg := range cliConfigs {
		moduleOptions[module] = cfg.AutoCLIOptions
	}
	return &autocliService{
		moduleOptions: moduleOptions,
	}
}

func (a autocliService) AppOptions(context.Context, *autocliv1.AppOptionsRequest) (*autocliv1.AppOptionsResponse, error) {
	return &autocliv1.AppOptionsResponse{
		ModuleOptions: a.moduleOptions,
	}, nil
}

var _ autocliv1.RemoteInfoServiceServer = &autocliService{}
