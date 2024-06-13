package main

import (
	"fmt"
	"movie_db/db"
	"net/http"
)

func server() {
	router := http.NewServeMux()
	loadRoutes(router)
	server := http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	fmt.Println("Server listening on port: 8080")
	server.ListenAndServe()

}

func main() {
	db.Connect()
	defer db.DB.Close()
	server()
}
