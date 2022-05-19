package container_test

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"gotest.tools/v3/golden"

	"github.com/cosmos/cosmos-sdk/container"
)

type KVStoreKey struct {
	name string
}

type MsgClientA struct {
	key string
}

type KeeperA struct {
	key  KVStoreKey
	name string
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

func (Command) IsManyPerContainerType() {}

func ProvideKVStoreKey(moduleKey container.ModuleKey) KVStoreKey {
	return KVStoreKey{name: moduleKey.Name()}
}

func ProvideMsgClientA(key container.ModuleKey) MsgClientA {
	return MsgClientA{key.Name()}
}

type ModuleA struct{}

func (ModuleA) Provide(key KVStoreKey, moduleKey container.OwnModuleKey) (KeeperA, Handler, Command) {
	return KeeperA{key: key, name: container.ModuleKey(moduleKey).Name()}, Handler{}, Command{}
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

func (ModuleB) Provide(dependencies BDependencies) (BProvides, Handler, error) {
	return BProvides{
		KeeperB: KeeperB{
			key:        dependencies.Key,
			msgClientA: dependencies.A,
		},
		Commands: []Command{{}, {}},
	}, Handler{}, nil
}

var scenarioConfig = container.Options(
	container.Provide(ProvideMsgClientA),
	container.ProvideInModule("runtime", ProvideKVStoreKey),
	container.ProvideInModule("a", wrapMethod0(ModuleA{})),
	container.ProvideInModule("b", wrapMethod0(ModuleB{})),
)

func TestScenario(t *testing.T) {
	var (
		handlers map[string]Handler
		commands []Command
		a        KeeperA
		b        KeeperB
	)
	require.NoError(t,
		container.Build(
			scenarioConfig,
			&handlers,
			&commands,
			&a,
			&b,
		))

	require.Len(t, handlers, 2)
	require.Equal(t, Handler{}, handlers["a"])
	require.Equal(t, Handler{}, handlers["b"])
	require.Len(t, commands, 3)
	require.Equal(t, KeeperA{
		key:  KVStoreKey{name: "a"},
		name: "a",
	}, a)
	require.Equal(t, KeeperB{
		key: KVStoreKey{name: "b"},
		msgClientA: MsgClientA{
			key: "b",
		},
	}, b)
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
	var x string
	require.Error(t, container.Build(
		container.Provide(
			func(x float64) string { return fmt.Sprintf("%f", x) },
			func(x int) float64 { return float64(x) },
			func(x float32) int { return int(x) },
		),
		&x,
	))
}

func TestCyclic(t *testing.T) {
	var x string
	require.Error(t, container.Build(
		container.Provide(
			func(x int) float64 { return float64(x) },
			func(x float64) (int, string) { return int(x), "hi" },
		),
		&x,
	))
}

func TestErrorOption(t *testing.T) {
	err := container.Build(container.Error(fmt.Errorf("an error")))
	require.Error(t, err)
}

func TestBadCtr(t *testing.T) {
	_, err := container.ExtractProviderDescriptor(KeeperA{})
	require.Error(t, err)
}

func TestTrivial(t *testing.T) {
	require.NoError(t, container.Build(container.Options()))
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

	var x int
	require.Error(t,
		container.Build(
			container.Provide(func() (int, error) {
				return 0, fmt.Errorf("the error")
			}),
			&x,
		))
}

func TestSimple(t *testing.T) {
	var x int
	require.NoError(t,
		container.Build(
			container.Provide(
				func() int { return 1 },
			),
			&x,
		),
	)

	require.Error(t,
		container.Build(
			container.Provide(
				func() int { return 0 },
				func() int { return 1 },
			),
			&x,
		),
	)
}

func TestModuleScoped(t *testing.T) {
	var x int
	require.Error(t,
		container.Build(
			container.Provide(
				func(container.ModuleKey) int { return 0 },
			),
			&x,
		),
	)

	var y float64
	require.Error(t,
		container.Build(
			container.Options(
				container.Provide(
					func(container.ModuleKey) int { return 0 },
					func() int { return 1 },
				),
				container.ProvideInModule("a",
					func(x int) float64 { return float64(x) },
				),
			),
			&y,
		),
	)

	require.Error(t,
		container.Build(
			container.Options(
				container.Provide(
					func() int { return 0 },
					func(container.ModuleKey) int { return 1 },
				),
				container.ProvideInModule("a",
					func(x int) float64 { return float64(x) },
				),
			),
			&y,
		),
	)

	require.Error(t,
		container.Build(
			container.Options(
				container.Provide(
					func(container.ModuleKey) int { return 0 },
					func(container.ModuleKey) int { return 1 },
				),
				container.ProvideInModule("a",
					func(x int) float64 { return float64(x) },
				),
			),
			&y,
		),
	)

	require.NoError(t,
		container.Build(
			container.Options(
				container.Provide(
					func(container.ModuleKey) int { return 0 },
				),
				container.ProvideInModule("a",
					func(x int) float64 { return float64(x) },
				),
			),
			&y,
		),
	)

	require.Error(t,
		container.Build(
			container.Options(
				container.Provide(
					func(container.ModuleKey) int { return 0 },
				),
				container.ProvideInModule("",
					func(x int) float64 { return float64(x) },
				),
			),
			&y,
		),
	)

	var z float32
	require.NoError(t,
		container.Build(
			container.Options(
				container.Provide(
					func(container.ModuleKey) int { return 0 },
				),
				container.ProvideInModule("a",
					func(x int) float64 { return float64(x) },
					func(x int) float32 { return float32(x) },
				),
			),
			&y, &z,
		),
		"use module dep twice",
	)
}

type OnePerModuleInt int

func (OnePerModuleInt) IsOnePerModuleType() {}

func TestOnePerModule(t *testing.T) {
	var x OnePerModuleInt
	require.Error(t,
		container.Build(container.Options(), &x),
		"bad input type",
	)

	var y map[string]OnePerModuleInt
	var z string
	require.NoError(t,
		container.Build(
			container.Options(
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
			&y,
			&z,
		),
	)

	require.Equal(t, map[string]OnePerModuleInt{
		"a": 3,
		"b": 4,
	}, y)
	require.Equal(t, "7", z)

	var m map[string]OnePerModuleInt
	require.Error(t,
		container.Build(
			container.ProvideInModule("a",
				func() OnePerModuleInt { return 0 },
				func() OnePerModuleInt { return 0 },
			),
			&m,
		),
		"duplicate",
	)

	require.Error(t,
		container.Build(
			container.Provide(
				func() OnePerModuleInt { return 0 },
			),
			&m,
		),
		"out of scope",
	)

	require.Error(t,
		container.Build(
			container.Provide(
				func() map[string]OnePerModuleInt { return nil },
			),
			&m,
		),
		"bad return type",
	)

	require.NoError(t,
		container.Build(
			container.Options(),
			&m,
		),
		"no providers",
	)
}

type ManyPerContainerInt int

func (ManyPerContainerInt) IsManyPerContainerType() {}

func TestManyPerContainer(t *testing.T) {
	var xs []ManyPerContainerInt
	var sum string
	require.NoError(t,
		container.Build(
			container.Provide(
				func() ManyPerContainerInt { return 4 },
				func() ManyPerContainerInt { return 9 },
				func(xs []ManyPerContainerInt) string {
					sum := 0
					for _, x := range xs {
						sum += int(x)
					}
					return fmt.Sprintf("%d", sum)
				},
			),
			&xs,
			&sum,
		),
	)
	require.Len(t, xs, 2)
	require.Contains(t, xs, ManyPerContainerInt(4))
	require.Contains(t, xs, ManyPerContainerInt(9))
	require.Equal(t, "13", sum)

	var z ManyPerContainerInt
	require.Error(t,
		container.Build(
			container.Provide(
				func() ManyPerContainerInt { return 0 },
			),
			&z,
		),
		"bad input type",
	)

	require.NoError(t,
		container.Build(
			container.Options(),
			&xs,
		),
		"no providers",
	)
}

func TestSupply(t *testing.T) {
	var x int
	require.NoError(t,
		container.Build(
			container.Supply(3),
			&x,
		),
	)
	require.Equal(t, 3, x)

	require.Error(t,
		container.Build(
			container.Options(
				container.Supply(3),
				container.Provide(func() int { return 4 }),
			),
			&x,
		),
		"can't supply then provide",
	)

	require.Error(t,
		container.Build(
			container.Options(
				container.Supply(3),
				container.Provide(func() int { return 4 }),
			),
			&x,
		),
		"can't provide then supply",
	)

	require.Error(t,
		container.Build(
			container.Supply(3, 4),
			&x,
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
	Y int64
}

func TestStructArgs(t *testing.T) {
	var input TestInput
	require.Error(t, container.Build(container.Options(), &input))

	require.NoError(t, container.Build(
		container.Supply(1.3),
		&input,
	))
	require.Equal(t, 0, input.X)
	require.Equal(t, 1.3, input.Y)

	require.NoError(t, container.Build(
		container.Supply(1.3, 1),
		&input,
	))
	require.Equal(t, 1, input.X)
	require.Equal(t, 1.3, input.Y)

	var x string
	var y int64
	require.NoError(t, container.Build(
		container.Provide(func() (TestOutput, error) {
			return TestOutput{X: "A", Y: -10}, nil
		}),
		&x, &y,
	))
	require.Equal(t, "A", x)
	require.Equal(t, int64(-10), y)

	require.Error(t, container.Build(
		container.Provide(func() (TestOutput, error) {
			return TestOutput{}, fmt.Errorf("error")
		}),
		&x,
	))
}

func TestDebugOptions(t *testing.T) {
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

	require.NoError(t, container.BuildDebug(
		container.DebugOptions(
			container.Logger(func(s string) {
				logOut += s
			}),
			container.Visualizer(func(g string) {
				dotGraph = g
			}),
			container.LogVisualizer(),
			container.FileVisualizer(graphfile.Name()),
			container.StdoutLogger(),
		),
		container.Options(),
	))

	require.Contains(t, logOut, "digraph")
	require.Contains(t, dotGraph, "digraph")

	outfileContents, err := os.ReadFile(outfile.Name())
	require.NoError(t, err)
	require.Contains(t, string(outfileContents), "digraph")

	graphfileContents, err := os.ReadFile(graphfile.Name())
	require.NoError(t, err)
	require.Contains(t, string(graphfileContents), "digraph")
}

func TestGraphAndLogOutput(t *testing.T) {
	var graphOut string
	var b KeeperB
	debugOpts := container.DebugOptions(
		container.Visualizer(func(dotGraph string) {
			graphOut = dotGraph
		}))
	require.NoError(t, container.BuildDebug(debugOpts, scenarioConfig, &b))
	golden.Assert(t, graphOut, "example.dot")

	badConfig := container.Options(
		container.ProvideInModule("runtime", ProvideKVStoreKey),
		container.ProvideInModule("a", wrapMethod0(ModuleA{})),
		container.ProvideInModule("b", wrapMethod0(ModuleB{})),
	)
	require.Error(t, container.BuildDebug(debugOpts, badConfig, &b))
	golden.Assert(t, graphOut, "example_error.dot")
}

func TestConditionalDebugging(t *testing.T) {
	logs := ""
	success := false
	conditionalDebugOpt := container.DebugOptions(
		container.OnError(container.Logger(func(s string) {
			logs += s + "\n"
		})),
		container.OnSuccess(container.DebugCleanup(func() {
			success = true
		})))

	var input TestInput
	require.Error(t, container.BuildDebug(
		conditionalDebugOpt,
		container.Options(),
		&input,
	))
	require.Contains(t, logs, `Initializing logger`)
	require.Contains(t, logs, `Registering providers`)
	require.Contains(t, logs, `Registering outputs`)
	require.False(t, success)

	logs = ""
	success = false
	require.NoError(t, container.BuildDebug(
		conditionalDebugOpt,
		container.Options(),
	))
	require.Empty(t, logs)
	require.True(t, success)
}
