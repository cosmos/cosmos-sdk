package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// Select prompts the user to select an option from a list of choices.
// It takes a label string to display above the selection prompt and a slice of string options to choose from.
func Select(label string, options []string) (string, error) {
	fmt.Printf("%s:\n", label)
	for i, option := range options {
		fmt.Printf("[%d] %s\n", i+1, option)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter number: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("failed to read input: %w", err)
		}

		input = strings.TrimSpace(input)
		choice, err := parseSelection(input, options)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		return choice, nil
	}
}

// parseSelection converts user input to a selection from options.
// Accepts either the number (1-based) or exact text of the option.
func parseSelection(input string, options []string) (string, error) {
	// Try to parse as a number
	var index int
	if _, err := fmt.Sscanf(input, "%d", &index); err == nil {
		if index < 1 || index > len(options) {
			return "", fmt.Errorf("invalid selection: must be between 1 and %d", len(options))
		}
		return options[index-1], nil
	}

	// Check if input matches any option exactly
	for _, option := range options {
		if strings.EqualFold(input, option) {
			return option, nil
		}
	}

	return "", fmt.Errorf("invalid selection: must be a number between 1 and %d or match an option", len(options))
}

// PromptString prompts the user for a string input with the given label.
// It validates the input using the provided validate function.
func PromptString(label string, validate func(string) error) (string, error) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s: ", label)
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("failed to read input: %w", err)
		}

		input = strings.TrimSpace(input)

		if validate != nil {
			if err := validate(input); err != nil {
				fmt.Println(err.Error())
				continue
			}
		}

		return input, nil
	}
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
