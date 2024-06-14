package main

import (
	"movie_db/movie"
	"net/http"
)

// Main router aggregator
func loadRoutes(router *http.ServeMux) {
	handler := &movie.Handler{}
	fileHandler := http.StripPrefix("/static/", http.FileServer(http.Dir("static")))

	router.HandleFunc("/{$}", handler.GetIndex)
	router.HandleFunc("GET /movie/{id}", handler.GetMovieByID)
	router.HandleFunc("GET /movies/", handler.GetAllMovies)
	router.HandleFunc("GET /movies/{id}", handler.GetAllMoviesHTMX)
	router.HandleFunc("GET /movie/add", handler.AddMoviePage)
	router.HandleFunc("POST /movie/", handler.PostMovie)
	router.HandleFunc("DELETE /movie/{id}", handler.DeleteMovie)
	router.HandleFunc("GET /movie/edit/{id}", handler.GetEditMovieForm)
	router.HandleFunc("PUT /movie/{id}", handler.UpdateMovie)
	router.HandleFunc("POST /search", handler.SearchByTitle)
	router.HandleFunc("GET /empty", handler.EmptyResponse)
	router.Handle("/static/", fileHandler)
}
