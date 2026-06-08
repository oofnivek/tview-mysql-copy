package db

import (
	"database/sql"
	"fmt"

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
