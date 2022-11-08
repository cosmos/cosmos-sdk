package ormdb

import (
	"context"
	"fmt"
	"sort"

	"github.com/cosmos/cosmos-sdk/orm/types/ormerrors"

	"github.com/cosmos/cosmos-sdk/orm/types/ormjson"

	"google.golang.org/protobuf/reflect/protoreflect"

	"cosmossdk.io/errors"
)

func (m moduleDB) DefaultJSON(target ormjson.WriteTarget) error {
	for name, table := range m.tablesByName {
		w, err := target.OpenWriter(name)
		if err != nil {
			return err
		}

		_, err = w.Write(table.DefaultJSON())
		if err != nil {
			return err
		}

		err = w.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (m moduleDB) ValidateJSON(source ormjson.ReadSource) error {
	errMap := map[protoreflect.FullName]error{}
	for name, table := range m.tablesByName {
		r, err := source.OpenReader(name)
		if err != nil {
			return err
		}

		if r == nil {
			continue
		}

		err = table.ValidateJSON(r)
		if err != nil {
			errMap[name] = err
		}

		err = r.Close()
		if err != nil {
			return err
		}
	}

	if len(errMap) != 0 {
		var allErrors string
		for name, err := range errMap {
			allErrors += fmt.Sprintf("Error in JSON for table %s: %v\n", name, err)
		}
		return ormerrors.JSONValidationError.Wrap(allErrors)
	}

	return nil
}

func (m moduleDB) ImportJSON(ctx context.Context, source ormjson.ReadSource) error {
	var names []string
	for name := range m.tablesByName {
		names = append(names, string(name))
	}
	sort.Strings(names)

	for _, name := range names {
		fullName := protoreflect.FullName(name)
		table := m.tablesByName[fullName]

		r, err := source.OpenReader(fullName)
		if err != nil {
			return errors.Wrapf(err, "table %s", fullName)
		}

		if r == nil {
			continue
		}

		err = table.ImportJSON(ctx, r)
		if err != nil {
			return errors.Wrapf(err, "table %s", fullName)
		}

		err = r.Close()
		if err != nil {
			return errors.Wrapf(err, "table %s", fullName)
		}
	}

	return nil
}

func (m moduleDB) ExportJSON(ctx context.Context, sink ormjson.WriteTarget) error {
	for name, table := range m.tablesByName {
		w, err := sink.OpenWriter(name)
		if err != nil {
			return err
		}

		err = table.ExportJSON(ctx, w)
		if err != nil {
			return err
		}

		err = w.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
