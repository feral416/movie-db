package db

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func Connect() (err error) {
	//TODO: move confidential to env!
	DB, err = sql.Open("mysql", "user:user@(localhost:3306)/movies?parseTime=true")
	if err != nil {
		fmt.Println("Error during connection: ", err)
		return err
	}
	DB.Ping()
	return nil
}
