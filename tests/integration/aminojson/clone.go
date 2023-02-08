package aminojson

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"pgregory.net/rapid"

	authapi "cosmossdk.io/api/cosmos/auth/v1beta1"
	authzapi "cosmossdk.io/api/cosmos/authz/v1beta1"
	"cosmossdk.io/x/tx/rapidproto"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
)

type typeUnderTest struct {
	gogo   gogoproto.Message
	pulsar proto.Message
}

type typeIndex struct {
	gogoFields   map[string]map[string]reflect.StructField
	pulsarFields map[string]map[string]reflect.StructField
	pulsarToGogo map[string]string
}

func fqTypeName(msg any) string {
	return fullyQualifiedTypeName(reflect.TypeOf(msg).Elem())
}

func fullyQualifiedTypeName(typ reflect.Type) string {
	pkgType := typ
	if typ.Kind() == reflect.Pointer || typ.Kind() == reflect.Slice || typ.Kind() == reflect.Map || typ.Kind() == reflect.Array {
		pkgType = typ.Elem()
	}
	pkgPath := pkgType.PkgPath()
	if pkgPath == "" {
		return fmt.Sprintf("%v", typ)
	}

	return fmt.Sprintf("%s/%v", pkgPath, typ)
}

func newTypeIndex(types []typeUnderTest) typeIndex {
	ti := typeIndex{
		gogoFields:   make(map[string]map[string]reflect.StructField),
		pulsarFields: make(map[string]map[string]reflect.StructField),
		pulsarToGogo: make(map[string]string),
	}
	for _, t := range types {
		gogoType := reflect.TypeOf(t.gogo).Elem()
		pulsarType := reflect.TypeOf(t.pulsar).Elem()

		ti.gogoFields[fqTypeName(t.gogo)] = make(map[string]reflect.StructField)
		ti.pulsarFields[fqTypeName(t.pulsar)] = make(map[string]reflect.StructField)
		ti.pulsarToGogo[fqTypeName(t.pulsar)] = fqTypeName(t.gogo)

		for i := 0; i < gogoType.NumField(); i++ {
			field := gogoType.Field(i)
			tag := field.Tag.Get("protobuf")
			if tag == "" {
				continue
			}
			n := strings.Split(tag, ",")[3]
			name := strings.Split(n, "=")[1]

			ti.gogoFields[fqTypeName(t.gogo)][name] = gogoType.Field(i)
		}
		for i := 0; i < pulsarType.NumField(); i++ {
			field := pulsarType.Field(i)
			tag := field.Tag.Get("protobuf")
			if tag == "" {
				continue
			}
			n := strings.Split(tag, ",")[3]
			name := strings.Split(n, "=")[1]
			ti.pulsarFields[fqTypeName(t.pulsar)][name] = pulsarType.Field(i)
		}
	}

	return ti
}

func (ti typeIndex) reflectedDeepClone(pulsar reflect.Value, gogo reflect.Value) {
	pulsarTypeName := fullyQualifiedTypeName(pulsar.Type())
	gogoTypeName := fullyQualifiedTypeName(gogo.Type())
	structFields, found := ti.pulsarFields[pulsarTypeName]
	if !found {
		panic(fmt.Sprintf("no pulsar fields for type %s", pulsarTypeName))
	}
	for n, pStructField := range structFields {
		gStructField := ti.gogoFields[gogoTypeName][n]
		pulsarField := pulsar.FieldByName(pStructField.Name)
		// todo init a new "gogo" since its nil
		gogoField := gogo.FieldByName(gStructField.Name)
		if !gogoField.IsValid() {
			gogoField = reflect.New(gStructField.Type)
			gogo.FieldByName(gStructField.Name).Set(gogoField)
		}
		//fmt.Printf("copying %s to %s\n", pStructField.Name, gStructField.Name)
		ti.setField(pulsarField, gogoField)
	}
}

func (ti typeIndex) deepClone(pulsar proto.Message, gogo gogoproto.Message) {
	pulsarTypeName := fqTypeName(pulsar)
	gogoTypeName := fqTypeName(gogo)
	structFields, found := ti.pulsarFields[pulsarTypeName]
	if !found {
		panic(fmt.Sprintf("no pulsar fields for type %s", pulsarTypeName))
	}
	for n, pStructField := range structFields {
		gStructField := ti.gogoFields[gogoTypeName][n]
		pulsarField := reflect.ValueOf(pulsar).Elem().FieldByName(pStructField.Name)
		gogoField := reflect.ValueOf(gogo).Elem().FieldByName(gStructField.Name)
		//fmt.Printf("copying %s to %s\n", pStructField.Name, gStructField.Name)
		ti.setField(pulsarField, gogoField)
	}
}

func (ti typeIndex) setField(pulsar reflect.Value, gogo reflect.Value) {
	switch pulsar.Type().Kind() {
	case reflect.Ptr:
		if !gogo.IsValid() {
			fmt.Printf("gogo field is invalid; gogo: %v\n", gogo)
		}
		if gogo.Type().Kind() != reflect.Ptr && gogo.Type().Kind() != reflect.Struct {
			panic(fmt.Sprintf("gogo field is not a pointer; pulsar: %s, gogo: %s", pulsar.Type(), gogo.Type()))
		}
		if pulsar.IsNil() {
			return
		}
		ti.setField(pulsar.Elem(), gogo)
		//panic(fmt.Sprintf("pointer not supported: %s", pulsar.Type()))
	case reflect.Struct:
		switch val := pulsar.Interface().(type) {
		case anypb.Any:
			//codectypes.NewAnyWithValue(val.Value)
			a := &codectypes.Any{
				TypeUrl: val.TypeUrl,
				Value:   val.Value,
			}
			gogo.Set(reflect.ValueOf(a))
		case timestamppb.Timestamp:
			// will panic if field is not of type *time.Time, hopefully it is.
			t := time.Unix(val.Seconds, int64(val.Nanos))
			gogo.Set(reflect.ValueOf(&t))
		default:
			if gogo.Type().Kind() == reflect.Ptr {
				gogoType := gogo.Type().Elem()
				newGogo := reflect.New(gogoType)
				gogo.Set(newGogo)
				ti.reflectedDeepClone(pulsar, gogo.Elem())
			} else {
				gogoType := gogo.Type()
				newGogo := reflect.New(gogoType).Elem()
				gogo.Set(newGogo)
				ti.reflectedDeepClone(pulsar, gogo)
			}
		}
	case reflect.Slice:
		// if slices are different types then we need to create a new slice
		if pulsar.Type().Elem() != gogo.Type().Elem() {
			gogoSlice := reflect.MakeSlice(gogo.Type(), 0, 0)
			gogoType := gogoSlice.Type().Elem()
			for i := 0; i < pulsar.Len(); i++ {
				p := pulsar.Index(i)
				g := reflect.New(gogoType).Elem()
				ti.setField(p, g)
				gogoSlice = reflect.Append(gogoSlice, g)
			}
			gogo.Set(gogoSlice)
			return
		}
		// otherwise we can just copy the slice
		fallthrough
	default:
		if pulsar.IsZero() {
			return
		}
		gogo.Set(pulsar)
	}
}

func newGogoMessage(t reflect.Type) gogoproto.Message {
	msg := reflect.New(t).Interface()
	switch msg.(type) {
	case *authtypes.ModuleAccount:
		return &authtypes.ModuleAccount{BaseAccount: &authtypes.BaseAccount{}}
	default:
		return msg.(gogoproto.Message)
	}
}

func TestDeepClone(t *testing.T) {
	ti := newTypeIndex(msgTypes)
	//tt := msgTypes[0]
	var anyTypeURLs []string
	for _, msgType := range msgTypes {
		anyTypeURLs = append(anyTypeURLs, string(msgType.pulsar.ProtoReflect().Descriptor().FullName()))
	}

	for _, tt := range msgTypes {
		fmt.Printf("testing %s\n", tt.pulsar.ProtoReflect().Descriptor().FullName())
		gen := rapidproto.MessageGenerator(tt.pulsar, rapidproto.GeneratorOptions{
			AnyTypeURLs: anyTypeURLs,
			Resolver:    protoregistry.GlobalTypes,
		})

		rapid.Check(t, func(rt *rapid.T) {
			msg := gen.Draw(rt, "msg").(proto.Message)
			//fmt.Printf("msg %v\n", msg)
			goMsg := reflect.New(reflect.TypeOf(tt.gogo).Elem()).Interface().(gogoproto.Message)
			//fmt.Println("clone")
			ti.deepClone(msg, goMsg)
			//fmt.Println("assert")
			//ti.assertEquals(t, msg, goMsg)
		})
	}
}

func (ti typeIndex) assertEquals(t *testing.T, pulsar proto.Message, gogo gogoproto.Message) {
	for n, pStructField := range ti.pulsarFields[fqTypeName(pulsar)] {
		gStructField := ti.gogoFields[fqTypeName(gogo)][n]
		pulsarField := reflect.ValueOf(pulsar).Elem().FieldByName(pStructField.Name)
		gogoField := reflect.ValueOf(gogo).Elem().FieldByName(gStructField.Name)
		ti.assertFieldEquals(t, pulsarField, gogoField)
	}
}

func (ti typeIndex) assertFieldEquals(t *testing.T, pulsarField reflect.Value, gogoField reflect.Value) {
	switch pulsarField.Type().Kind() {
	case reflect.Ptr:
		if gogoField.Type().Kind() != reflect.Ptr && gogoField.Type().Kind() != reflect.Struct {
			panic(fmt.Sprintf("gogo field is not a pointer; pulsar: %s", pulsarField.Type()))
		}
		if pulsarField.IsNil() {
			if gogoField.Type().Kind() == reflect.Struct {
				// TODO rewrite comparison as hash concatenation to avoid this hack and potential bug
				return
			} else if !gogoField.IsNil() {
				println("failing")
				require.Fail(t, "pulsar field is nil, but gogo field is not")
			} else {
				// both nil, return
				return
			}
		}
		// otherwise recurse
		ti.assertFieldEquals(t, pulsarField.Elem(), gogoField.Elem())
	//panic(fmt.Sprintf("pointer not supported: %s", pulsarField.Type()))
	case reflect.Struct:

	default:
		require.Equal(t, pulsarField.Interface(), gogoField.Interface())
	}
}

var msgTypes = []typeUnderTest{
	// auth
	{gogo: &authtypes.Params{}, pulsar: &authapi.Params{}},
	{gogo: &authtypes.BaseAccount{}, pulsar: &authapi.BaseAccount{}},

	// omitted from the test because of the custom MarshalJSON. It is tested separately. Pulsar types
	// may be generated with data which is invalid for the gogo type, but valid for the pulsar type.
	{gogo: &authtypes.ModuleAccount{}, pulsar: &authapi.ModuleAccount{}},

	// missing name extension, do we need it?
	// {gogo: &authtypes.ModuleCredential{}, pulsar: &authapi.ModuleCredential{}},

	{gogo: &authtypes.MsgUpdateParams{}, pulsar: &authapi.MsgUpdateParams{}},

	// authz
	{gogo: &authztypes.Grant{}, pulsar: &authzapi.Grant{}},
	{gogo: &authztypes.MsgGrant{}, pulsar: &authzapi.MsgGrant{}},
	{gogo: &authztypes.MsgRevoke{}, pulsar: &authzapi.MsgRevoke{}},
	{gogo: &authztypes.MsgExec{}, pulsar: &authzapi.MsgExec{}},
	{gogo: &authztypes.GenericAuthorization{}, pulsar: &authzapi.GenericAuthorization{}},
}
