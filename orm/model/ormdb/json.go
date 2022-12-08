package ormdb

import (
	"context"
	"fmt"
	"sort"

	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/reflect/protoreflect"

	"cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/orm/types/ormerrors"
	"github.com/cosmos/cosmos-sdk/orm/types/ormjson"
)

func (m moduleDB) DefaultJSON(target ormjson.WriteTarget) error {
	tableNames := maps.Keys(m.tablesByName)
	sort.Slice(tableNames, func(i, j int) bool {
		ti, tj := tableNames[i], tableNames[j]
		return ti.Name() < tj.Name()
	})

	for _, name := range tableNames {
		table := m.tablesByName[name]
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
	names := maps.Keys(m.tablesByName)
	sort.Slice(names, func(i, j int) bool {
		ti, tj := names[i], names[j]
		return ti.Name() < tj.Name()
	})
	for _, name := range names {
		r, err := source.OpenReader(name)
		if err != nil {
			return err
		}

		if r == nil {
			continue
		}

		table := m.tablesByName[name]
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
	// Ensure that we export the tables in a deterministic order.
	tableNames := maps.Keys(m.tablesByName)
	sort.Slice(tableNames, func(i, j int) bool {
		ti, tj := tableNames[i], tableNames[j]
		return ti.Name() < tj.Name()
	})

	for _, name := range tableNames {
		w, err := sink.OpenWriter(name)
		if err != nil {
			return err
		}

		table := m.tablesByName[name]
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
