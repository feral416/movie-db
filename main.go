package main

import (
	"log"
	"movie_db/db"
	"movie_db/movie"
	"net/http"
	"time"
)

func server() {
	const port string = "8080"
	router := http.NewServeMux()
	loadRoutes(router)
	server := http.Server{
		Addr:    ":" + port,
		Handler: router,
	}
	log.Printf("Server listening on port: %s", port)
	server.ListenAndServe()

}

func hourly() {
	go func() {
		for {
			if err := movie.SM.Shrink(); err != nil {
				log.Printf("Error shrinking sessions: %s", err)
			}
			time.Sleep(time.Hour)
		}
	}()
}

func main() {
	movie.Sessions = movie.NewSessionsStore()
	db.Connect()
	defer db.DB.Close()
	movie.SM = &movie.SessionManager{DB: db.DB, Cache: movie.Sessions}
	movie.SM.InitSync()
	hourly()
	server()
}
