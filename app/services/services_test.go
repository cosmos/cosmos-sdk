package services

import (
	"context"
	"testing"

	"google.golang.org/grpc"

	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
)

type autoCLIModule struct {
	opts *autocliv1.ModuleOptions
}

func (m autoCLIModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return m.opts
}

func TestExtractAutoCLIOptions(t *testing.T) {
	expected := &autocliv1.ModuleOptions{
		Tx: &autocliv1.ServiceCommandDescriptor{Service: "custom.Msg"},
	}

	appModules := map[string]any{
		"explicit": autoCLIModule{opts: expected},
	}

	got := ExtractAutoCLIOptions(appModules)
	if got["explicit"] != expected {
		t.Fatal("expected explicit AutoCLIOptions to be preserved")
	}
}

func TestAutoCLIConfiguratorRegisterServiceWithUnknownService(t *testing.T) {
	cfg := &autocliConfigurator{}
	cfg.RegisterService(&grpc.ServiceDesc{ServiceName: "unknown.service.Name"}, nil)
	if cfg.err == nil {
		t.Fatal("expected registry lookup error for unknown service")
	}
}

func TestNewAutoCLIQueryServiceAppOptions(t *testing.T) {
	svc := NewAutoCLIQueryService(map[string]any{
		"m": autoCLIModule{
			opts: &autocliv1.ModuleOptions{
				Query: &autocliv1.ServiceCommandDescriptor{Service: "custom.Query"},
			},
		},
	})

	resp, err := svc.AppOptions(context.Background(), &autocliv1.AppOptionsRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ModuleOptions["m"] == nil {
		t.Fatal("expected module option for m")
	}
}

func TestAppQueryServiceConfig(t *testing.T) {
	cfg := &appv1alpha1.Config{}
	svc := NewAppQueryService(cfg)

	resp, err := svc.Config(context.Background(), &appv1alpha1.QueryConfigRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Config != cfg {
		t.Fatal("expected returned config pointer to match service config")
	}
}

func TestNewReflectionServiceFileDescriptors(t *testing.T) {
	svc, err := NewReflectionService()
	if err != nil {
		t.Fatalf("unexpected error constructing reflection service: %v", err)
	}

	resp, err := svc.FileDescriptors(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error fetching file descriptors: %v", err)
	}
	if len(resp.Files) == 0 {
		t.Fatal("expected at least one file descriptor")
	}
}
