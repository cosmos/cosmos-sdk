package container_test

import (
	"fmt"
	"io/ioutil"
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

type Command struct {
	Run func()
}

func ProvideKVStoreKey(scope container.Scope) KVStoreKey {
	return KVStoreKey{name: scope.Name()}
}

func ProvideModuleKey(scope container.Scope) (ModuleKey, error) {
	return ModuleKey(scope.Name()), nil
}

func ProvideMsgClientA(_ container.Scope, key ModuleKey) MsgClientA {
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

func (ModuleB) Provide(dependencies BDependencies, _ container.Scope) (BProvides, Handler, error) {
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
			container.AutoGroupTypes(reflect.TypeOf(Command{})),
			container.OnePerScopeTypes(reflect.TypeOf(Handler{})),
			container.Provide(
				ProvideKVStoreKey,
				ProvideModuleKey,
				ProvideMsgClientA,
			),
			container.ProvideWithScope("a", wrapMethod0(ModuleA{})),
			container.ProvideWithScope("b", wrapMethod0(ModuleB{})),
		))
}

func wrapMethod0(module interface{}) interface{} {
	methodFn := reflect.TypeOf(module).Method(0).Func.Interface()
	ctrInfo, err := container.ExtractConstructorInfo(methodFn)
	if err != nil {
		panic(err)
	}

	ctrInfo.In = ctrInfo.In[1:]
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
	_, err := container.ExtractConstructorInfo(KeeperA{})
	require.Error(t, err)
}

func TestInvoker(t *testing.T) {
	require.NoError(t, container.Run(func() {}))
	require.NoError(t, container.Run(func() error { return nil }))
	require.Error(t, container.Run(func() error { return fmt.Errorf("error") }))
	require.Error(t, container.Run(func() int { return 0 }))
}

func TestErrorFunc(t *testing.T) {
	_, err := container.ExtractConstructorInfo(
		func() (error, int) { return nil, 0 },
	)
	require.Error(t, err)

	_, err = container.ExtractConstructorInfo(
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

func TestScoped(t *testing.T) {
	require.Error(t,
		container.Run(func(int) {},
			container.Provide(
				func(container.Scope) int { return 0 },
			),
		),
	)

	require.Error(t,
		container.Run(func(float64) {},
			container.Provide(
				func(container.Scope) int { return 0 },
				func() int { return 1 },
			),
			container.ProvideWithScope("a",
				func(x int) float64 { return float64(x) },
			),
		),
	)

	require.Error(t,
		container.Run(func(float64) {},
			container.Provide(
				func() int { return 0 },
				func(container.Scope) int { return 1 },
			),
			container.ProvideWithScope("a",
				func(x int) float64 { return float64(x) },
			),
		),
	)

	require.Error(t,
		container.Run(func(float64) {},
			container.Provide(
				func(container.Scope) int { return 0 },
				func(container.Scope) int { return 1 },
			),
			container.ProvideWithScope("a",
				func(x int) float64 { return float64(x) },
			),
		),
	)

	require.NoError(t,
		container.Run(func(float64) {},
			container.Provide(
				func(container.Scope) int { return 0 },
			),
			container.ProvideWithScope("a",
				func(x int) float64 { return float64(x) },
			),
		),
	)

	require.Error(t,
		container.Run(func(float64) {},
			container.Provide(
				func(container.Scope) int { return 0 },
			),
			container.ProvideWithScope("",
				func(x int) float64 { return float64(x) },
			),
		),
	)

	require.NoError(t,
		container.Run(func(float64, float32) {},
			container.Provide(
				func(container.Scope) int { return 0 },
			),
			container.ProvideWithScope("a",
				func(x int) float64 { return float64(x) },
				func(x int) float32 { return float32(x) },
			),
		),
		"use scope dep twice",
	)
}

var intType = reflect.TypeOf(0)

func TestOnePerScope(t *testing.T) {
	require.Error(t,
		container.Run(
			func() {},
			container.OnePerScopeTypes(intType),
			container.AutoGroupTypes(intType),
		),
	)

	require.Error(t,
		container.Run(
			func(int) {},
			container.OnePerScopeTypes(intType),
		),
		"bad input type",
	)

	require.NoError(t,
		container.Run(
			func(x map[string]int, y string) {
				require.Equal(t, map[string]int{
					"a": 3,
					"b": 4,
				}, x)
				require.Equal(t, "7", y)
			},
			container.OnePerScopeTypes(intType),
			container.ProvideWithScope("a",
				func() int { return 3 },
			),
			container.ProvideWithScope("b",
				func() int { return 4 },
			),
			container.Provide(func(x map[string]int) string {
				sum := 0
				for _, v := range x {
					sum += v
				}
				return fmt.Sprintf("%d", sum)
			}),
		),
	)

	require.Error(t,
		container.Run(
			func(map[string]int) {},
			container.OnePerScopeTypes(intType),
			container.ProvideWithScope("a",
				func() int { return 0 },
				func() int { return 0 },
			),
		),
		"duplicate",
	)

	require.Error(t,
		container.Run(
			func(map[string]int) {},
			container.OnePerScopeTypes(intType),
			container.Provide(
				func() int { return 0 },
			),
		),
		"out of scope",
	)

	require.Error(t,
		container.Run(
			func(map[string]int) {},
			container.OnePerScopeTypes(intType),
			container.Provide(
				func() map[string]int { return nil },
			),
		),
		"bad return type",
	)
}

func TestAutoGroup(t *testing.T) {
	require.NoError(t,
		container.Run(
			func() {},
			container.AutoGroupTypes(intType),
		),
	)

	require.Error(t,
		container.Run(
			func() {},
			container.AutoGroupTypes(reflect.SliceOf(intType)),
		),
	)

	require.Error(t,
		container.Run(
			func() {},
			container.AutoGroupTypes(intType),
			container.OnePerScopeTypes(intType),
		),
	)

	require.NoError(t,
		container.Run(
			func(xs []int, sum string) {
				require.Len(t, xs, 2)
				require.Contains(t, xs, 4)
				require.Contains(t, xs, 9)
				require.Equal(t, "13", sum)
			},
			container.AutoGroupTypes(intType),
			container.Provide(
				func() int { return 4 },
				func() int { return 9 },
				func(xs []int) string {
					sum := 0
					for _, x := range xs {
						sum += x
					}
					return fmt.Sprintf("%d", sum)
				},
			),
		),
	)

	require.Error(t,
		container.Run(
			func(int) {},
			container.AutoGroupTypes(intType),
			container.Provide(
				func() int { return 0 },
			),
		),
		"bad input type",
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

	outfile, err := ioutil.TempFile("", "out")
	require.NoError(t, err)
	stdout := os.Stdout
	os.Stdout = outfile
	defer func() { os.Stdout = stdout }()
	defer os.Remove(outfile.Name())

	graphfile, err := ioutil.TempFile("", "graph")
	require.NoError(t, err)
	defer os.Remove(graphfile.Name())

	require.NoError(t, container.Run(
		func() {},
		container.Logger(func(s string) {
			logOut += s
		}),
		container.Visualizer(func(g string) {
			dotGraph = g
		}),
		container.LogVisualizer(),
		container.FileVisualizer(graphfile.Name(), "svg"),
		container.StdoutLogger(),
	))

	require.Contains(t, logOut, "digraph")
	require.Contains(t, dotGraph, "digraph")

	outfileContents, err := ioutil.ReadFile(outfile.Name())
	require.NoError(t, err)
	require.Contains(t, string(outfileContents), "digraph")

	graphfileContents, err := ioutil.ReadFile(graphfile.Name())
	require.NoError(t, err)
	require.Contains(t, string(graphfileContents), "<svg")
}
