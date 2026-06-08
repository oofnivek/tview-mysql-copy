package db

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"tview-mysql-copy/config"
)

var systemDBs = map[string]bool{
	"information_schema": true,
	"performance_schema": true,
	"mysql":              true,
	"sys":                true,
}

func dsn(c config.Connection, database string) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?timeout=5s&readTimeout=10s",
		c.User, c.Password, c.Host, c.Port, database)
}

func ListDatabases(c config.Connection) ([]string, error) {
	db, err := sql.Open("mysql", dsn(c, ""))
	if err != nil {
		return nil, err
	}
	defer db.Close()
	rows, err := db.Query("SHOW DATABASES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var dbs []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		if !systemDBs[name] {
			dbs = append(dbs, name)
		}
	}
	return dbs, rows.Err()
}

func ListTables(c config.Connection, database string) ([]string, error) {
	db, err := sql.Open("mysql", dsn(c, database))
	if err != nil {
		return nil, err
	}
	defer db.Close()
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, name)
	}
	return tables, rows.Err()
}

var (
	reAutoIncrementOpt = regexp.MustCompile(`(?i)\s+AUTO_INCREMENT=\d+`)
	reAutoIncrementCol = regexp.MustCompile(`(?i)\s+AUTO_INCREMENT\b`)
)

func stripAutoIncrement(ddl string) string {
	ddl = reAutoIncrementOpt.ReplaceAllString(ddl, "")
	ddl = reAutoIncrementCol.ReplaceAllString(ddl, "")
	return ddl
}

func CopyTable(src, dst config.Connection, srcDB, srcTable, dstDB, dstTable string, onProgress func(copied int)) error {
	srcConn, err := sql.Open("mysql", dsn(src, srcDB))
	if err != nil {
		return fmt.Errorf("connect source: %w", err)
	}
	defer srcConn.Close()

	dstConn, err := sql.Open("mysql", dsn(dst, dstDB))
	if err != nil {
		return fmt.Errorf("connect destination: %w", err)
	}
	defer dstConn.Close()

	// Get DDL from source
	var tblName, createStmt string
	if err := srcConn.QueryRow("SHOW CREATE TABLE `"+srcTable+"`").Scan(&tblName, &createStmt); err != nil {
		return fmt.Errorf("get DDL: %w", err)
	}

	createStmt = stripAutoIncrement(createStmt)
	// Rename table in DDL if dst table name differs
	createStmt = strings.Replace(createStmt, "`"+srcTable+"`", "`"+dstTable+"`", 1)

	// Disable FK checks on destination for the duration of the copy
	if _, err := dstConn.Exec("SET FOREIGN_KEY_CHECKS=0"); err != nil {
		return fmt.Errorf("disable FK checks: %w", err)
	}
	defer dstConn.Exec("SET FOREIGN_KEY_CHECKS=1") //nolint:errcheck

	// Drop and recreate destination table
	if _, err := dstConn.Exec("DROP TABLE IF EXISTS `" + dstTable + "`"); err != nil {
		return fmt.Errorf("drop table: %w", err)
	}
	if _, err := dstConn.Exec(createStmt); err != nil {
		return fmt.Errorf("create table: %w", err)
	}

	// Stream rows from source and batch insert into destination
	rows, err := srcConn.Query("SELECT * FROM `" + srcTable + "`")
	if err != nil {
		return fmt.Errorf("select source: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("columns: %w", err)
	}
	if len(cols) == 0 {
		return nil
	}

	colList := "`" + strings.Join(cols, "`, `") + "`"
	placeholder := "(" + strings.Repeat("?,", len(cols)-1) + "?)"

	const batchSize = 500
	var batch []interface{}
	placeholders := make([]string, 0, batchSize)

	copied := 0
	flush := func() error {
		if len(batch) == 0 {
			return nil
		}
		query := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES %s", dstTable, colList, strings.Join(placeholders, ","))
		if _, err := dstConn.Exec(query, batch...); err != nil {
			return fmt.Errorf("insert batch: %w", err)
		}
		copied += len(placeholders)
		if onProgress != nil {
			onProgress(copied)
		}
		batch = batch[:0]
		placeholders = placeholders[:0]
		return nil
	}

	vals := make([]interface{}, len(cols))
	ptrs := make([]interface{}, len(cols))
	for i := range vals {
		ptrs[i] = &vals[i]
	}

	for rows.Next() {
		if err := rows.Scan(ptrs...); err != nil {
			return fmt.Errorf("scan row: %w", err)
		}
		row := make([]interface{}, len(cols))
		copy(row, vals)
		batch = append(batch, row...)
		placeholders = append(placeholders, placeholder)
		if len(placeholders) >= batchSize {
			if err := flush(); err != nil {
				return err
			}
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read rows: %w", err)
	}
	return flush()
}
