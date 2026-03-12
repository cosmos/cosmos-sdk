package migration

import (
	"go/ast"

	"github.com/rs/zerolog/log"
)

// StructFieldRemoval defines a struct field to remove from a type declaration.
type StructFieldRemoval struct {
	// StructName is the name of the struct (e.g., "SimApp").
	StructName string
	// FieldName is the name of the field to remove (e.g., "CircuitKeeper").
	FieldName string
}

// StructFieldModification defines a struct field to modify.
type StructFieldModification struct {
	// StructName is the name of the struct (e.g., "SimApp").
	StructName string
	// FieldName is the name of the field to modify (e.g., "EpochsKeeper").
	FieldName string
	// MakePointer wraps the field type in a pointer (*T).
	MakePointer bool
}

// updateStructFieldRemovals removes fields from struct type declarations.
func updateStructFieldRemovals(node *ast.File, removals []StructFieldRemoval) (bool, error) {
	if len(removals) == 0 {
		return false, nil
	}

	modified := false

	ast.Inspect(node, func(n ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return true
		}

		for _, removal := range removals {
			if typeSpec.Name.Name != removal.StructName {
				continue
			}

			var newFields []*ast.Field
			for _, field := range structType.Fields.List {
				shouldRemove := false
				for _, name := range field.Names {
					if name.Name == removal.FieldName {
						shouldRemove = true
						break
					}
				}
				if shouldRemove {
					log.Debug().Msgf("Removed field %s from struct %s", removal.FieldName, removal.StructName)
					modified = true
				} else {
					newFields = append(newFields, field)
				}
			}
			structType.Fields.List = newFields
		}

		return true
	})

	return modified, nil
}

// updateStructFieldModifications modifies fields in struct type declarations.
func updateStructFieldModifications(node *ast.File, mods []StructFieldModification) (bool, error) {
	if len(mods) == 0 {
		return false, nil
	}

	modified := false

	ast.Inspect(node, func(n ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return true
		}

		for _, mod := range mods {
			if typeSpec.Name.Name != mod.StructName {
				continue
			}

			for _, field := range structType.Fields.List {
				for _, name := range field.Names {
					if name.Name != mod.FieldName {
						continue
					}

					if mod.MakePointer {
						// Check if already a pointer
						if _, isPtr := field.Type.(*ast.StarExpr); !isPtr {
							field.Type = &ast.StarExpr{X: field.Type}
							log.Debug().Msgf("Made field %s.%s a pointer type", mod.StructName, mod.FieldName)
							modified = true
						}
					}
				}
			}
		}

		return true
	})

	return modified, nil
}
