package container_test

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/container"
)

type KVStoreKey struct {
	name string
}

type ModuleKey string

type MsgClientA struct {
	key ModuleKey
}

type KeeperA struct {
	key KVStoreKey
}

type KeeperB struct {
	key        KVStoreKey
	msgClientA MsgClientA
}

type Handler struct {
	Handle func()
}

func (Handler) IsOnePerModuleType() {}

type Command struct {
	Run func()
}

func (Command) IsAutoGroupType() {}

func ProvideKVStoreKey(moduleKey container.ModuleKey) KVStoreKey {
	return KVStoreKey{name: moduleKey.Name()}
}

func ProvideModuleKey(moduleKey container.ModuleKey) (ModuleKey, error) {
	return ModuleKey(moduleKey.Name()), nil
}

func ProvideMsgClientA(_ container.ModuleKey, key ModuleKey) MsgClientA {
	return MsgClientA{key}
}

type ModuleA struct{}

func (ModuleA) Provide(key KVStoreKey) (KeeperA, Handler, Command) {
	return KeeperA{key}, Handler{}, Command{}
}

type ModuleB struct{}

type BDependencies struct {
	container.In

	Key KVStoreKey
	A   MsgClientA
}

type BProvides struct {
	container.Out

	KeeperB  KeeperB
	Commands []Command
}

func (ModuleB) Provide(dependencies BDependencies, _ container.ModuleKey) (BProvides, Handler, error) {
	return BProvides{
		KeeperB: KeeperB{
			key:        dependencies.Key,
			msgClientA: dependencies.A,
		},
		Commands: []Command{{}, {}},
	}, Handler{}, nil
}

func TestScenario(t *testing.T) {
	require.NoError(t,
		container.Run(
			func(handlers map[string]Handler, commands []Command, a KeeperA, b KeeperB) {
				require.Len(t, handlers, 2)
				require.Equal(t, Handler{}, handlers["a"])
				require.Equal(t, Handler{}, handlers["b"])
				require.Len(t, commands, 3)
				require.Equal(t, KeeperA{
					key: KVStoreKey{name: "a"},
				}, a)
				require.Equal(t, KeeperB{
					key: KVStoreKey{name: "b"},
					msgClientA: MsgClientA{
						key: "b",
					},
				}, b)
			},
			container.Provide(
				ProvideKVStoreKey,
				ProvideModuleKey,
				ProvideMsgClientA,
			),
			container.ProvideInModule("a", wrapMethod0(ModuleA{})),
			container.ProvideInModule("b", wrapMethod0(ModuleB{})),
		))
}

func wrapMethod0(module interface{}) interface{} {
	methodFn := reflect.TypeOf(module).Method(0).Func.Interface()
	ctrInfo, err := container.ExtractProviderDescriptor(methodFn)
	if err != nil {
		panic(err)
	}

	ctrInfo.Inputs = ctrInfo.Inputs[1:]
	fn := ctrInfo.Fn
	ctrInfo.Fn = func(values []reflect.Value) ([]reflect.Value, error) {
		return fn(append([]reflect.Value{reflect.ValueOf(module)}, values...))
	}
	return ctrInfo
}

func TestResolveError(t *testing.T) {
	require.Error(t, container.Run(
		func(x string) {},
		container.Provide(
			func(x float64) string { return fmt.Sprintf("%f", x) },
			func(x int) float64 { return float64(x) },
			func(x float32) int { return int(x) },
		),
	))
}

func TestCyclic(t *testing.T) {
	require.Error(t, container.Run(
		func(x string) {},
		container.Provide(
			func(x int) float64 { return float64(x) },
			func(x float64) (int, string) { return int(x), "hi" },
		),
	))
}

func TestErrorOption(t *testing.T) {
	err := container.Run(func() {}, container.Error(fmt.Errorf("an error")))
	require.Error(t, err)
}

func TestBadCtr(t *testing.T) {
	_, err := container.ExtractProviderDescriptor(KeeperA{})
	require.Error(t, err)
}

func TestInvoker(t *testing.T) {
	require.NoError(t, container.Run(func() {}))
	require.NoError(t, container.Run(func() error { return nil }))
	require.Error(t, container.Run(func() error { return fmt.Errorf("error") }))
	require.Error(t, container.Run(func() int { return 0 }))
}

func TestErrorFunc(t *testing.T) {
	_, err := container.ExtractProviderDescriptor(
		func() (error, int) { return nil, 0 },
	)
	require.Error(t, err)

	_, err = container.ExtractProviderDescriptor(
		func() (int, error) { return 0, nil },
	)
	require.NoError(t, err)

	require.Error(t,
		container.Run(
			func(x int) {
			},
			container.Provide(func() (int, error) {
				return 0, fmt.Errorf("the error")
			}),
		))

	require.Error(t,
		container.Run(func() error {
			return fmt.Errorf("the error")
		}), "the error")
}

func TestSimple(t *testing.T) {
	require.NoError(t,
		container.Run(
			func(x int) {
				require.Equal(t, 1, x)
			},
			container.Provide(
				func() int { return 1 },
			),
		),
	)

	require.Error(t,
		container.Run(func(int) {},
			container.Provide(
				func() int { return 0 },
				func() int { return 1 },
			),
		),
	)
}

func TestModuleScoped(t *testing.T) {
	require.Error(t,
		container.Run(func(int) {},
			container.Provide(
				func(container.ModuleKey) int { return 0 },
			),
		),
	)

	require.Error(t,
		container.Run(func(float64) {},
			container.Provide(
				func(container.ModuleKey) int { return 0 },
				func() int { return 1 },
			),
			container.ProvideInModule("a",
				func(x int) float64 { return float64(x) },
			),
		),
	)

	require.Error(t,
		container.Run(func(float64) {},
			container.Provide(
				func() int { return 0 },
				func(container.ModuleKey) int { return 1 },
			),
			container.ProvideInModule("a",
				func(x int) float64 { return float64(x) },
			),
		),
	)

	require.Error(t,
		container.Run(func(float64) {},
			container.Provide(
				func(container.ModuleKey) int { return 0 },
				func(container.ModuleKey) int { return 1 },
			),
			container.ProvideInModule("a",
				func(x int) float64 { return float64(x) },
			),
		),
	)

	require.NoError(t,
		container.Run(func(float64) {},
			container.Provide(
				func(container.ModuleKey) int { return 0 },
			),
			container.ProvideInModule("a",
				func(x int) float64 { return float64(x) },
			),
		),
	)

	require.Error(t,
		container.Run(func(float64) {},
			container.Provide(
				func(container.ModuleKey) int { return 0 },
			),
			container.ProvideInModule("",
				func(x int) float64 { return float64(x) },
			),
		),
	)

	require.NoError(t,
		container.Run(func(float64, float32) {},
			container.Provide(
				func(container.ModuleKey) int { return 0 },
			),
			container.ProvideInModule("a",
				func(x int) float64 { return float64(x) },
				func(x int) float32 { return float32(x) },
			),
		),
		"use module dep twice",
	)
}

type OnePerModuleInt int

func (OnePerModuleInt) IsOnePerModuleType() {}

func TestOnePerModule(t *testing.T) {
	require.Error(t,
		container.Run(
			func(OnePerModuleInt) {},
		),
		"bad input type",
	)

	require.NoError(t,
		container.Run(
			func(x map[string]OnePerModuleInt, y string) {
				require.Equal(t, map[string]OnePerModuleInt{
					"a": 3,
					"b": 4,
				}, x)
				require.Equal(t, "7", y)
			},
			container.ProvideInModule("a",
				func() OnePerModuleInt { return 3 },
			),
			container.ProvideInModule("b",
				func() OnePerModuleInt { return 4 },
			),
			container.Provide(func(x map[string]OnePerModuleInt) string {
				sum := 0
				for _, v := range x {
					sum += int(v)
				}
				return fmt.Sprintf("%d", sum)
			}),
		),
	)

	require.Error(t,
		container.Run(
			func(map[string]OnePerModuleInt) {},
			container.ProvideInModule("a",
				func() OnePerModuleInt { return 0 },
				func() OnePerModuleInt { return 0 },
			),
		),
		"duplicate",
	)

	require.Error(t,
		container.Run(
			func(map[string]OnePerModuleInt) {},
			container.Provide(
				func() OnePerModuleInt { return 0 },
			),
		),
		"out of scope",
	)

	require.Error(t,
		container.Run(
			func(map[string]OnePerModuleInt) {},
			container.Provide(
				func() map[string]OnePerModuleInt { return nil },
			),
		),
		"bad return type",
	)

	require.NoError(t,
		container.Run(
			func(map[string]OnePerModuleInt) {},
		),
		"no providers",
	)
}

type AutoGroupInt int

func (AutoGroupInt) IsAutoGroupType() {}

func TestAutoGroup(t *testing.T) {
	require.NoError(t,
		container.Run(
			func(xs []AutoGroupInt, sum string) {
				require.Len(t, xs, 2)
				require.Contains(t, xs, AutoGroupInt(4))
				require.Contains(t, xs, AutoGroupInt(9))
				require.Equal(t, "13", sum)
			},
			container.Provide(
				func() AutoGroupInt { return 4 },
				func() AutoGroupInt { return 9 },
				func(xs []AutoGroupInt) string {
					sum := 0
					for _, x := range xs {
						sum += int(x)
					}
					return fmt.Sprintf("%d", sum)
				},
			),
		),
	)

	require.Error(t,
		container.Run(
			func(AutoGroupInt) {},
			container.Provide(
				func() AutoGroupInt { return 0 },
			),
		),
		"bad input type",
	)

	require.NoError(t,
		container.Run(
			func([]AutoGroupInt) {},
		),
		"no providers",
	)
}

func TestSupply(t *testing.T) {
	require.NoError(t,
		container.Run(func(x int) {
			require.Equal(t, 3, x)
		},
			container.Supply(3),
		),
	)

	require.Error(t,
		container.Run(func(x int) {},
			container.Supply(3),
			container.Provide(func() int { return 4 }),
		),
		"can't supply then provide",
	)

	require.Error(t,
		container.Run(func(x int) {},
			container.Supply(3),
			container.Provide(func() int { return 4 }),
		),
		"can't provide then supply",
	)

	require.Error(t,
		container.Run(func(x int) {},
			container.Supply(3, 4),
		),
		"can't supply twice",
	)
}

type TestInput struct {
	container.In

	X int `optional:"true"`
	Y float64
}

type TestOutput struct {
	container.Out

	X string
}

func TestStructArgs(t *testing.T) {
	require.Error(t, container.Run(
		func(input TestInput) {},
	))

	require.NoError(t, container.Run(
		func(input TestInput) {
			require.Equal(t, 0, input.X)
			require.Equal(t, 1.3, input.Y)
		},
		container.Supply(1.3),
	))

	require.NoError(t, container.Run(
		func(input TestInput) {
			require.Equal(t, 1, input.X)
			require.Equal(t, 1.3, input.Y)
		},
		container.Supply(1.3, 1),
	))

	require.NoError(t, container.Run(
		func(x string) {
			require.Equal(t, "A", x)
		},
		container.Provide(func() (TestOutput, error) {
			return TestOutput{X: "A"}, nil
		}),
	))

	require.Error(t, container.Run(
		func(x string) {},
		container.Provide(func() (TestOutput, error) {
			return TestOutput{}, fmt.Errorf("error")
		}),
	))
}

func TestLogging(t *testing.T) {
	var logOut string
	var dotGraph string

	outfile, err := os.CreateTemp("", "out")
	require.NoError(t, err)
	stdout := os.Stdout
	os.Stdout = outfile
	defer func() { os.Stdout = stdout }()
	defer os.Remove(outfile.Name())

	graphfile, err := os.CreateTemp("", "graph")
	require.NoError(t, err)
	defer os.Remove(graphfile.Name())

	require.NoError(t, container.RunDebug(
		func() {},
		container.DebugOptions(
			container.Logger(func(s string) {
				logOut += s
			}),
			container.Visualizer(func(g string) {
				dotGraph = g
			}),
			container.LogVisualizer(),
			container.FileVisualizer(graphfile.Name(), "svg"),
			container.StdoutLogger(),
		),
	))

	require.Contains(t, logOut, "digraph")
	require.Contains(t, dotGraph, "digraph")

	outfileContents, err := os.ReadFile(outfile.Name())
	require.NoError(t, err)
	require.Contains(t, string(outfileContents), "digraph")

	graphfileContents, err := os.ReadFile(graphfile.Name())
	require.NoError(t, err)
	require.Contains(t, string(graphfileContents), "<svg")
}
