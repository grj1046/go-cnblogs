package db

import (
	"database/sql"
	"io/ioutil"

	//github.com/mattn/go-sqlite3
	_ "github.com/mattn/go-sqlite3"
)

var dbPath = "./cnblogs.db"

//InitialDB initial conblogs.ing database structures
func InitialDB() error {
	//load sql
	sqlbuf, err := ioutil.ReadFile("./cnblogs.sql")
	if err != nil {
		return err
	}
	sqlStmt := string(sqlbuf)

	db, err := sql.Open("sqlite3", dbPath)
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

//Execute insert update  delete
func Execute(strSQL string, args ...interface{}) (*sql.Result, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	//"insert into foo(id, name) values(?, ?)"
	stmt, err := tx.Prepare(strSQL)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(args)
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
	db, err := sql.Open("sqlite3", dbPath)
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
