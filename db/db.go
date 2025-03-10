package db

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func Connect() (err error) {
	username := os.Getenv("MOVIE_DB_USER")
	pwd := os.Getenv("MOVIE_DB_PWD")
	dbAddr := os.Getenv("MYSQL_DB_ADDR")
	if username == "" || pwd == "" || dbAddr == "" {
		log.Fatalf("Environment variables for db are not set: username:%s pwd:%s dbAddr:%s", username, pwd, dbAddr)
	}
	DB, err = sql.Open("mysql", username+`:`+pwd+`@(`+dbAddr+`)/movies?parseTime=true&loc=Local`)
	if err != nil {
		log.Panicf("Error openig connection to db: %s", err)
	}
	err = DB.Ping()
	if err != nil {
		log.Printf("Error during connection to db: %s", err)
		return err
	}
	return nil
}
