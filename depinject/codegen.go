package depinject

import (
	"fmt"
	"go/ast"
	"go/token"
	"reflect"
	"strconv"
	"strings"
)

func (c *container) createIdent(namePrefix string) *ast.Ident {
	return c.doCreateIdent(namePrefix, nil)
}

func (c *container) doCreateIdent(namePrefix string, handle interface{}) *ast.Ident {
	// TODO reserved names: keywords, builtin types, imports, err
	v := namePrefix
	i := 2
	for {
		_, ok := c.idents[v]
		if !ok {
			c.idents[v] = handle
			if handle != nil {
				c.reverseIdents[handle] = v
			}
			return ast.NewIdent(v)
		}

		v = fmt.Sprintf("%s%d", namePrefix, i)
		i++
	}
}

func (c *container) getOrCreateIdent(namePrefix string, handle interface{}) (v *ast.Ident, created bool) {
	if v, ok := c.reverseIdents[handle]; ok {
		return ast.NewIdent(v), false
	}

	return c.doCreateIdent(namePrefix, handle), true
}

func (c *container) codegenStmt(stmt ast.Stmt) {
	c.codegenFunc.Body.List = append(c.codegenFunc.Body.List, stmt)
}

func (c *container) valueExpr(value reflect.Value) (ast.Expr, error) {
	typ := value.Type()
	switch typ.Kind() {

	case reflect.Bool:
		return &ast.BasicLit{Kind: token.IDENT, Value: fmt.Sprintf("%t", value.Bool())}, nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("%d", value.Uint())}, nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("%d", value.Int())}, nil

	case reflect.Float32, reflect.Float64:
		return &ast.BasicLit{Kind: token.FLOAT, Value: strconv.FormatFloat(value.Float(), 'e', -1, 64)}, nil

	case reflect.Complex64, reflect.Complex128:
		return &ast.BasicLit{Kind: token.FLOAT, Value: strconv.FormatComplex(value.Complex(), 'e', -1, 128)}, nil

	case reflect.Array:
		return c.arraySliceExpr(value)

	case reflect.Map:
		if value.IsNil() {
			return ast.NewIdent("nil"), nil
		}

		t, err := c.typeExpr(typ)
		if err != nil {
			return nil, err
		}

		n := value.Len()
		lit := &ast.CompositeLit{
			Type: t,
			Elts: make([]ast.Expr, n),
		}

		for i, key := range value.MapKeys() {
			k, err := c.valueExpr(key)
			if err != nil {
				return nil, err
			}

			v, err := c.valueExpr(value.MapIndex(key))
			if err != nil {
				return nil, err
			}

			lit.Elts[i] = &ast.KeyValueExpr{Key: k, Value: v}
		}

		return lit, nil

	case reflect.Slice:
		if value.IsNil() {
			return ast.NewIdent("nil"), nil
		}

		return c.arraySliceExpr(value)

	case reflect.String:
		return &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", value.String())}, nil

	case reflect.Struct:
		t, err := c.typeExpr(typ)
		if err != nil {
			return nil, err
		}

		n := typ.NumField()
		lit := &ast.CompositeLit{
			Type: t,
		}

		for i := 0; i < n; i++ {
			f := typ.Field(i)
			v := value.FieldByName(f.Name)
			if v.IsZero() {
				continue
			}

			vExpr, err := c.valueExpr(v)
			if err != nil {
				return nil, err
			}

			lit.Elts = append(lit.Elts, &ast.KeyValueExpr{
				Key:   ast.NewIdent(f.Name),
				Value: vExpr,
			})
		}

		return lit, nil
	case reflect.Pointer:
		if value.IsNil() {
			return ast.NewIdent("nil"), nil
		}

		if typ.Elem().Kind() == reflect.Struct {
			v, err := c.valueExpr(value.Elem())
			if err != nil {
				return nil, err
			}

			return &ast.UnaryExpr{Op: token.AND, X: v}, nil
		} else {
			return nil, fmt.Errorf("invalid type %s", typ)
		}
	case reflect.Invalid, reflect.Uintptr, reflect.Chan, reflect.Func, reflect.Interface, reflect.UnsafePointer:
		return nil, fmt.Errorf("invalid type %s", typ)

	default:
		return nil, fmt.Errorf("invalid type %s", typ)
	}
}

func (c *container) arraySliceExpr(value reflect.Value) (ast.Expr, error) {
	astTyp, err := c.typeExpr(value.Type())
	if err != nil {
		return nil, err
	}

	n := value.Len()
	lit := &ast.CompositeLit{Type: astTyp, Elts: make([]ast.Expr, n)}

	for i := 0; i < n; i++ {
		lit.Elts[i], err = c.valueExpr(value.Index(i))
		if err != nil {
			return nil, err
		}
	}

	return lit, nil
}

func (c *container) typeExpr(typ reflect.Type) (ast.Expr, error) {
	if name := typ.Name(); name != "" {
		importPrefix := c.getOrAddImport(typ.PkgPath())
		if importPrefix == "" {
			return ast.NewIdent(name), nil
		}

		return ast.NewIdent(fmt.Sprintf("%s.%s", importPrefix, name)), nil
	}

	switch typ.Kind() {

	case reflect.Array:
		elt, err := c.typeExpr(typ.Elem())
		if err != nil {
			return nil, err
		}
		return &ast.ArrayType{
			Len: &ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("%d", typ.Len())},
			Elt: elt,
		}, nil

	case reflect.Slice:
		elt, err := c.typeExpr(typ.Elem())
		if err != nil {
			return nil, err
		}
		return &ast.ArrayType{Elt: elt}, nil

	case reflect.Chan:
		elt, err := c.typeExpr(typ.Elem())
		if err != nil {
			return nil, err
		}
		e := &ast.ChanType{Value: elt}
		switch typ.ChanDir() {
		case reflect.SendDir:
			e.Dir = ast.SEND
		case reflect.RecvDir:
			e.Dir = ast.RECV
		default:
			e.Dir = ast.SEND | ast.RECV
		}
		return e, nil

	case reflect.Func:
		e := &ast.FuncType{
			Params:  &ast.FieldList{},
			Results: &ast.FieldList{},
		}

		numIn := typ.NumIn()
		for i := 0; i < numIn; i++ {
			in, err := c.typeExpr(typ.In(i))
			if err != nil {
				return nil, err
			}
			e.Params.List = append(e.Params.List, &ast.Field{Type: in})
		}

		if typ.IsVariadic() {
			in, err := c.typeExpr(typ.In(numIn - 1).Elem())
			if err != nil {
				return nil, err
			}

			e.Params.List[numIn-1] = &ast.Field{Type: &ast.Ellipsis{Elt: in}}
		}

		for i := 0; i < typ.NumOut(); i++ {
			out, err := c.typeExpr(typ.Out(i))
			if err != nil {
				return nil, err
			}
			e.Results.List = append(e.Results.List, &ast.Field{Type: out})
		}

		return e, nil

	case reflect.Map:
		k, err := c.typeExpr(typ.Key())
		if err != nil {
			return nil, err
		}

		v, err := c.typeExpr(typ.Elem())
		if err != nil {
			return nil, err
		}

		return &ast.MapType{Key: k, Value: v}, nil

	case reflect.Pointer:
		elem, err := c.typeExpr(typ.Elem())
		if err != nil {
			return nil, err
		}

		return &ast.StarExpr{X: elem}, nil

	default:
		return nil, fmt.Errorf("unexpected type %v", typ)
	}
}

func (c *container) getOrAddImport(pkgPath string) (shortName string) {
	if pkgPath == "" || pkgPath == c.codegenPkgPath {
		return ""
	}

	if c.pkgImportMap == nil {
		c.pkgImportMap = map[string]*ast.ImportSpec{}
	}

	if c.shortNameImportMap == nil {
		c.shortNameImportMap = map[string]*ast.ImportSpec{}
	}

	if i, ok := c.pkgImportMap[pkgPath]; ok {
		return i.Name.Name
	}

	imp := &ast.ImportSpec{
		Path: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", pkgPath)},
	}

	pkgParts := strings.Split(pkgPath, "/")
	prefix := pkgParts[len(pkgParts)-1]
	shortName = prefix
	i := 2
	for {
		if _, ok := c.shortNameImportMap[shortName]; !ok {
			break
		}

		shortName = fmt.Sprintf("%s%d", prefix, i)
		i++
	}

	imp.Name = ast.NewIdent(shortName)
	c.imports = append(c.imports, imp)
	c.pkgImportMap[pkgPath] = imp
	c.shortNameImportMap[shortName] = imp
	return shortName
}
