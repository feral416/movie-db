package main

import (
	"crypto/tls"
	"log"
	"movie_db/db"
	"movie_db/movie"
	"net/http"
	"time"
)

func server() {
	const httpsPort string = "443"
	const httpPort string = "80"
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}
	router := http.NewServeMux()
	loadRoutes(router)
	httpsServer := http.Server{
		Addr:      ":" + httpsPort,
		Handler:   router,
		TLSConfig: tlsConfig,
	}
	//HTTP server for redirection to https
	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/", redirectToHTTPS)
	httpServer := http.Server{
		Addr:    ":" + httpPort,
		Handler: http.Handler(httpMux),
	}
	go func() {
		log.Printf("HTTPS server listening on port: %s", httpsPort)
		err := httpsServer.ListenAndServeTLS("cert.pem", "key.pem")
		if err != nil {
			log.Fatalf("HTTPS server failed: %v", err)
		}
	}()
	log.Printf("HTTP server listening on port: %s", httpPort)
	err := httpServer.ListenAndServe()
	if err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}

func redirectToHTTPS(w http.ResponseWriter, r *http.Request) {
	target := "https://" + r.Host + r.URL.String()
	w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	http.Redirect(w, r, target, http.StatusMovedPermanently)
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
