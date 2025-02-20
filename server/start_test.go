package server

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"reflect"
	"runtime"
	"testing"

	"github.com/spf13/cobra"
	tmcfg "github.com/tendermint/tendermint/config"
)

func TestAbciClientType(t *testing.T) {
	for _, tt := range []struct {
		clientType  string
		creatorName string
		wantErr     bool
	}{
		{
			clientType:  "committing",
			creatorName: "github.com/tendermint/tendermint/proxy.NewCommittingClientCreator",
		},
		{
			clientType:  "local",
			creatorName: "github.com/tendermint/tendermint/proxy.NewLocalClientCreator",
		},
		{
			clientType: "cool ranch",
			wantErr:    true,
		},
	} {
		t.Run(tt.clientType, func(t *testing.T) {
			creator, err := getAbciClientCreator(tt.clientType)
			if tt.wantErr {
				if err == nil {
					t.Error("wanted error, got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error %v", err)
				} else {
					creatorName := runtime.FuncForPC(reflect.ValueOf(creator).Pointer()).Name()
					if creatorName != tt.creatorName {
						t.Errorf(`want creator "%s", got "%s"`, tt.creatorName, creatorName)
					}
				}
			}
		})
	}
}

var errCancelledInPreRun = errors.New("cancelled in prerun")

func TestAbciClientPrecedence(t *testing.T) {
	for i, tt := range []struct {
		flag, toml, want string
	}{
		{
			want: "committing",
		},
		{
			flag: "foo",
			want: "foo",
		},
		{
			toml: "foo",
			want: "foo",
		},
		{
			flag: "foo",
			toml: "bar",
			want: "foo",
		},
	} {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			tempDir := t.TempDir()
			err := os.Mkdir(path.Join(tempDir, "config"), os.ModePerm)
			if err != nil {
				t.Fatalf("creating config dir failed: %v", err)
			}
			appTomlPath := path.Join(tempDir, "config", "app.toml")

			cmd := StartCmd(nil, tempDir)

			if tt.flag != "" {
				err = cmd.Flags().Set(FlagAbciClientType, tt.flag)
				if err != nil {
					t.Fatalf(`failed setting flag to "%s": %v`, tt.flag, err)
				}
			}

			if tt.toml != "" {
				writer, err := os.Create(appTomlPath)
				if err != nil {
					t.Fatalf(`failed creating "%s": %v`, appTomlPath, err)
				}
				_, err = writer.WriteString(fmt.Sprintf("%s = \"%s\"\n", FlagAbciClientType, tt.toml))
				if err != nil {
					t.Fatalf(`failed writing to app.toml: %v`, err)
				}
				err = writer.Close()
				if err != nil {
					t.Fatalf(`failed closing app.toml: %v`, err)
				}
			}

			// Compare to tests in util_test.go
			cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
				err := InterceptConfigsPreRunHandler(cmd, "", nil, tmcfg.DefaultConfig())
				if err != nil {
					return err
				}
				return errCancelledInPreRun
			}

			serverCtx := NewDefaultContext()
			ctx := context.WithValue(context.Background(), ServerContextKey, serverCtx)
			err = cmd.ExecuteContext(ctx)
			if err != errCancelledInPreRun {
				t.Fatal(err)
			}

			gotClientType := serverCtx.Viper.GetString(FlagAbciClientType)

			if gotClientType != tt.want {
				t.Errorf(`want client type "%s", got "%s"`, tt.want, gotClientType)
			}
		})
	}
}
