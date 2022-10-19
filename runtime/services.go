package runtime

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"

	"github.com/cosmos/gogoproto/proto"
	protov2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"

	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"cosmossdk.io/core/appmodule"
)

type appConfigService struct {
	appv1alpha1.UnimplementedQueryServer
	appConfig *appv1alpha1.Config
	files     *descriptorpb.FileDescriptorSet
}

func newAppConfigService(appConfig *appv1alpha1.Config) (*appConfigService, error) {
	fds := &descriptorpb.FileDescriptorSet{}

	// load gogo proto file descriptors
	allFds := proto.AllFileDescriptors()
	haveFileDescriptor := map[string]bool{}
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
		err = protov2.Unmarshal(bz, fd)
		if err != nil {
			return nil, err
		}

		fds.File = append(fds.File, fd)
		haveFileDescriptor[*fd.Name] = true
	}

	// load any protoregistry file descriptors not in gogo
	protoregistry.GlobalFiles.RangeFiles(func(fileDescriptor protoreflect.FileDescriptor) bool {
		if !haveFileDescriptor[fileDescriptor.Path()] {
			fds.File = append(fds.File, protodesc.ToFileDescriptorProto(fileDescriptor))
		}
		return true
	})

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

func newAutocliService(appModules map[string]appmodule.AppModule) *autocliService {
	moduleOptions := map[string]*autocliv1.ModuleOptions{}
	for modName, mod := range appModules {
		if autoCliMod, ok := mod.(interface {
			AutoCLIOptions() *autocliv1.ModuleOptions
		}); ok {
			moduleOptions[modName] = autoCliMod.AutoCLIOptions()
		}
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
