package cli

import (
	"fmt"

	"github.com/iancoleman/strcase"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
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
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			clientConn := getClientConn(ctx)
			input := inputType.New()
			output := outputType.New()
			err := clientConn.Invoke(ctx, methodName, input.Interface(), output.Interface())
			if err != nil {
				return err
			}

			bz, err := protojson.Marshal(output.Interface())
			if err != nil {
				return err
			}

			_, err = cmd.OutOrStdout().Write(bz)
			return err
		},
	}

	numFields := inputDesc.Fields().Len()
	for i := 0; i < numFields; i++ {
		b.addFieldFlag(cmd.Flags(), inputDesc.Fields().Get(i))
	}

	return cmd
}

func protoNameToCliName(name protoreflect.Name) string {
	return strcase.ToKebab(string(name))
}
