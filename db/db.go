package db

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func Connect() (err error) {
	//TODO: move confidential to env!
	DB, err = sql.Open("mysql", "user:user@(localhost:3306)/movies?parseTime=true&loc=Local")
	if err != nil {
		log.Printf("Error openig connection to db: %s", err)
		return err
	}
	err = DB.Ping()
	if err != nil {
		log.Printf("Error during connection to db: %s", err)
		return err
	}
	return nil
}
