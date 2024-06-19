package postgres

//func (tu *tableInfo) bindParams(update indexerbase.EntityUpdate) (pgx.NamedArgs, error) {
//	params := map[string]any{}
//
//	if len(tu.table.KeyColumns) == 0 {
//		params["_id"] = 1
//	} else if len(tu.table.KeyColumns) == 1 {
//		params[tu.table.KeyColumns[0].Name] = update.Key
//	} else {
//		ks, ok := update.Key.([]any)
//		if !ok {
//			return nil, fmt.Errorf("expected array key, got %T", update.Key)
//		}
//		if len(ks) != len(tu.table.KeyColumns) {
//			return nil, fmt.Errorf("expected %d key columns, got %d", len(tu.table.KeyColumns), len(ks))
//		}
//		for i, col := range tu.table.KeyColumns {
//			params[col.Name] = ks[i]
//		}
//	}
//
//	if !update.Delete {
//		if len(tu.table.ValueColumns) == 1 {
//			params[tu.table.ValueColumns[0].Name] = update.Value
//		} else {
//			vs, ok := update.Value.([]any)
//			if !ok {
//				return nil, fmt.Errorf("expected array key, got %T", update.Key)
//			}
//			if len(vs) != len(tu.table.ValueColumns) {
//				return nil, fmt.Errorf("expected %d value columns, got %d", len(tu.table.ValueColumns), len(vs))
//			}
//			for i, col := range tu.table.ValueColumns {
//				params[col.Name] = vs[i]
//			}
//		}
//	}
//
//	return params, nil
//}
