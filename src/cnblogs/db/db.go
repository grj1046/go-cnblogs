package db

import (
	"database/sql"
	"io/ioutil"

	//github.com/mattn/go-sqlite3
	_ "github.com/mattn/go-sqlite3"
)

//cache=shared&mode=rwc&_busy_timeout=3000
var dbPath = "./data/cnblogs.db?_busy_timeout=5000"
var dbPathOrigin = "./data/cnblogs_origin.db?_busy_timeout=5000"

//GetDB Get Sqlite database instance
func GetDB() (*sql.DB, error) {
	return sql.Open("sqlite3", dbPath)
}

//GetDBOrigin Get Sqlite database instance
func GetDBOrigin() (*sql.DB, error) {
	return sql.Open("sqlite3", dbPathOrigin)
}

//InitialDB initial conblogs.ing database structures
func InitialDB() error {
	err := initialCnblogs()
	if err != nil {
		return err
	}
	err = initialCnblogsOrigin()
	if err != nil {
		return err
	}
	return nil
}

func initialCnblogs() error {
	//load sql
	sqlbuf, err := ioutil.ReadFile("./sql/cnblogs.sql")
	if err != nil {
		return err
	}
	sqlStmt := string(sqlbuf)

	db, err := GetDB()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(sqlStmt)
	if err != nil {
		return err
	}
	return nil
}

func initialCnblogsOrigin() error {
	//load sql
	sqlbuf, err := ioutil.ReadFile("./sql/cnblogs_origin.sql")
	if err != nil {
		return err
	}
	sqlStmt := string(sqlbuf)

	db, err := GetDBOrigin()
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(sqlStmt)
	if err != nil {
		return err
	}
	return nil
}

//Execute Execute Sql
func Execute(strSQL string, args ...interface{}) (*sql.Result, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	result, err := db.Exec(strSQL, args...)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

//ExecuteTrans insert update delete
func ExecuteTrans(strSQL string, args ...interface{}) (*sql.Result, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	stmt, err := tx.Prepare(strSQL)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(args...)
	if err != nil {
		return nil, err
	}
	tx.Commit()
	return &result, nil
}

/*
Query executes a query that returns rows, typically a SELECT.
The args are for any placeholder parameters in the query.
Example:
	for rows.Next() {
		var id int
		var name string
		err = rows.Scan(&id, &name)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(id, name)
	}
*/
func Query(query string, args ...interface{}) (*sql.Rows, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return rows, nil
}

//QueryRow var name string; row.Scan(&name)
func QueryRow(query string, args ...interface{}) (*sql.Row, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	stmt, err := db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	return stmt.QueryRow(args...), nil
}
