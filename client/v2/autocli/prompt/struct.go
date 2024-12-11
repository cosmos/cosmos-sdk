package prompt

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"

	"github.com/manifoldco/promptui"
)

// PromptStruct prompts for values of a struct's fields interactively.
// It returns the populated struct and any error encountered.
func PromptStruct[T any](promptPrefix string, data T) (T, error) {
	return promptStruct(promptPrefix, data, nil)
}

// promptStruct prompts for values of a struct's fields interactively.
//
// For each field in the struct:
// - Pointer fields are initialized if nil and handled recursively if they contain structs
// - Struct fields are handled recursively
// - String and int slices are supported
// - String and int fields are prompted for and populated
// - Only String and int pointers are supported
// - Other types are skipped
func promptStruct[T any](promptPrefix string, data T, stdIn io.ReadCloser) (T, error) {
	v := reflect.ValueOf(&data).Elem()
	if v.Kind() == reflect.Interface {
		v = reflect.ValueOf(data)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldName := strings.ToLower(v.Type().Field(i).Name)

		// Handle pointer types
		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				field.Set(reflect.New(field.Type().Elem()))
			}
			if field.Elem().Kind() == reflect.Struct {
				result, err := promptStruct(promptPrefix+"."+fieldName, field.Interface(), stdIn)
				if err != nil {
					return data, err
				}
				field.Set(reflect.ValueOf(result))
				continue
			}
		}

		switch field.Kind() {
		case reflect.Struct:
			// For struct fields, create a new pointer to handle them
			structPtr := reflect.New(field.Type()).Interface()
			reflect.ValueOf(structPtr).Elem().Set(field)

			result, err := promptStruct(promptPrefix+"."+fieldName, structPtr, stdIn)
			if err != nil {
				return data, err
			}

			// Get the actual struct value from the result
			resultValue := reflect.ValueOf(result)
			if resultValue.Kind() == reflect.Ptr {
				resultValue = resultValue.Elem()
			}
			field.Set(resultValue)
			continue
		case reflect.Slice:
			if v.Field(i).Type().Elem().Kind() != reflect.String && v.Field(i).Type().Elem().Kind() != reflect.Int {
				continue
			}
		}

		// create prompts
		prompt := promptui.Prompt{
			Label:    fmt.Sprintf("Enter %s %s", promptPrefix, strings.Title(fieldName)), // nolint:staticcheck // strings.Title has a better API
			Validate: ValidatePromptNotEmpty,
			Stdin:    stdIn,
		}

		result, err := prompt.Run()
		if err != nil {
			return data, fmt.Errorf("failed to prompt for %s: %w", fieldName, err)
		}

		switch field.Kind() {
		case reflect.String:
			v.Field(i).SetString(result)
		case reflect.Int:
			resultInt, err := strconv.ParseInt(result, 10, 0)
			if err != nil {
				return data, fmt.Errorf("invalid value for int: %w", err)
			}
			v.Field(i).SetInt(resultInt)
		case reflect.Slice:
			switch v.Field(i).Type().Elem().Kind() {
			case reflect.String:
				v.Field(i).Set(reflect.ValueOf([]string{result}))
			case reflect.Int:
				resultInt, err := strconv.ParseInt(result, 10, 0)
				if err != nil {
					return data, fmt.Errorf("invalid value for int: %w", err)
				}

				v.Field(i).Set(reflect.ValueOf([]int{int(resultInt)}))
			}
		case reflect.Ptr:
			// Handle pointer fields by creating a new value and setting it
			ptrValue := reflect.New(field.Type().Elem())
			if ptrValue.Elem().Kind() == reflect.String {
				ptrValue.Elem().SetString(result)
				v.Field(i).Set(ptrValue)
			} else if ptrValue.Elem().Kind() == reflect.Int {
				resultInt, err := strconv.ParseInt(result, 10, 0)
				if err != nil {
					return data, fmt.Errorf("invalid value for int: %w", err)
				}
				ptrValue.Elem().SetInt(resultInt)
				v.Field(i).Set(ptrValue)
			}
		default:
			// skip any other types
			continue
		}
	}

	return data, nil
}
