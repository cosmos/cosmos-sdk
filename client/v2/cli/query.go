package cli

import (
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func (b *Builder) AddQueryService(command *cobra.Command, descriptor protoreflect.ServiceDescriptor) {
	methods := descriptor.Methods()
	n := methods.Len()
	for i := 0; i < n; i++ {
		cmd := b.QueryMethodToCommand(descriptor, methods.Get(i))
		command.AddCommand(cmd)
	}
}

func (b *Builder) QueryMethodToCommand(serviceDescriptor protoreflect.ServiceDescriptor, descriptor protoreflect.MethodDescriptor) *cobra.Command {
	docs := descriptor.ParentFile().SourceLocations().ByDescriptor(descriptor).LeadingComments
	getClientConn := b.GetClientConn
	methodName := fmt.Sprintf("/%s/%s", serviceDescriptor.FullName(), descriptor.Name())

	inputDesc := descriptor.Input()
	inputType := b.resolverMessageType(inputDesc)
	outputType := b.resolverMessageType(descriptor.Output())
	cmd := &cobra.Command{
		Use:  protoNameToCliName(descriptor.Name()),
		Long: docs,
	}

	flagHandler := b.registerMessageFlagSet(cmd.Flags(), inputType)

	jsonMarshalOptions := b.JSONMarshalOptions

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		clientConn := getClientConn(ctx)
		input := flagHandler.buildMessage()
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
