package postgres

//func (tu *tableInfo) exec(ctx context.Context, conn *pgx.Conn, update indexerbase.EntityUpdate) error {
//	var sql string
//	var err error
//	if update.Delete {
//		sql, err = tu.deleteSql()
//	} else {
//		sql, err = tu.insertOrUpdateSql()
//	}
//	if err != nil {
//		return err
//	}
//	fmt.Println(sql)
//
//	params, err := tu.bindParams(update)
//
//	_, err = conn.Exec(ctx, sql, params)
//	return err
//}
