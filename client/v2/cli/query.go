package cli

import (
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	"github.com/cosmos/cosmos-sdk/client/v2/cli/flag"
	"github.com/cosmos/cosmos-sdk/client/v2/internal/util"
)

// AddQueryServiceCommands adds a sub-command to the provided command for each
// method in the specified service and returns the command.
func (b *Builder) AddQueryServiceCommands(command *cobra.Command, serviceName protoreflect.FullName) *cobra.Command {
	resolver := b.FileResolver
	if resolver == nil {
		resolver = protoregistry.GlobalFiles
	}
	descriptor, err := resolver.FindDescriptorByName(serviceName)
	if err != nil {
		panic(err)
	}

	service := descriptor.(protoreflect.ServiceDescriptor)
	methods := service.Methods()
	n := methods.Len()
	for i := 0; i < n; i++ {
		cmd := b.CreateQueryMethodCommand(methods.Get(i))
		command.AddCommand(cmd)
	}
	return command
}

// CreateQueryMethodCommand creates a gRPC query command for the given service method.
func (b *Builder) CreateQueryMethodCommand(descriptor protoreflect.MethodDescriptor) *cobra.Command {
	serviceDescriptor := descriptor.Parent().(protoreflect.ServiceDescriptor)
	docs := util.DescriptorDocs(descriptor)
	getClientConn := b.GetClientConn
	methodName := fmt.Sprintf("/%s/%s", serviceDescriptor.FullName(), descriptor.Name())

	inputDesc := descriptor.Input()
	inputType := util.ResolveMessageType(b.TypeResolver, inputDesc)
	outputType := util.ResolveMessageType(b.TypeResolver, descriptor.Output())
	cmd := &cobra.Command{
		Use:  protoNameToCliName(descriptor.Name()),
		Long: docs,
	}

	binder := b.AddMessageFlags(cmd.Context(), cmd.Flags(), inputType, flag.Options{})

	jsonMarshalOptions := protojson.MarshalOptions{
		Indent:          "  ",
		UseProtoNames:   true,
		UseEnumNumbers:  false,
		EmitUnpopulated: true,
		Resolver:        b.TypeResolver,
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		clientConn := getClientConn(ctx)
		input := binder.BuildMessage()
		output := outputType.New()
		err := clientConn.Invoke(ctx, methodName, input.Interface(), output.Interface())
		if err != nil {
			return err
		}

		bz, err := jsonMarshalOptions.Marshal(output.Interface())
		if err != nil {
			return err
		}

		_, err = fmt.Fprintln(cmd.OutOrStdout(), string(bz))
		return err
	}

	return cmd
}

func protoNameToCliName(name protoreflect.Name) string {
	return strcase.ToKebab(string(name))
}
