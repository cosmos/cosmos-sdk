package prompt

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Select prompts the user to select an option from a list of choices.
// It takes a label string to display above the selection prompt and a slice of string options to choose from.
func Select(label string, options []string) (string, error) {
	selectUi := promptui.Select{
		Label: label,
		Items: options,
	}

	_, selectedProposalType, err := selectUi.Run()
	if err != nil {
		return "", fmt.Errorf("failed to prompt proposal types: %w", err)
	}

	return selectedProposalType, nil
}

// PromptString prompts the user for a string input with the given label.
// It validates the input using the provided validate function.
func PromptString(label string, validate func(string) error) (string, error) {
	promptUi := promptui.Prompt{
		Label:    label,
		Validate: validate,
	}

	return promptUi.Run()
}

// SetDefaults sets default values on a protobuf message based on a map of field names to values.
// It iterates through the message fields and sets values from the defaults map if the field name
// and type match.
func SetDefaults(msg protoreflect.Message, defaults map[string]interface{}) {
	fields := msg.Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		fieldName := string(field.Name())

		if v, ok := defaults[fieldName]; ok {
			// Get the field's kind
			fieldKind := field.Kind()

			switch v.(type) {
			case string:
				if fieldKind == protoreflect.StringKind {
					msg.Set(field, protoreflect.ValueOf(v))
				}
			case int64:
				if fieldKind == protoreflect.Int64Kind {
					msg.Set(field, protoreflect.ValueOf(v))
				}
			case int32:
				if fieldKind == protoreflect.Int32Kind {
					msg.Set(field, protoreflect.ValueOf(v))
				}
			case bool:
				if fieldKind == protoreflect.BoolKind {
					msg.Set(field, protoreflect.ValueOf(v))
				}
			}
		}
	}
}
