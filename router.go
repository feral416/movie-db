package main

import (
	"movie_db/middleware"
	"movie_db/movie"
	"net/http"
)

// Main router aggregator
func loadRoutes(router *http.ServeMux) {
	protectedStack := middleware.CreateStack(
		middleware.Logging,
		middleware.Auth,
	)

	publicStack := middleware.CreateStack(
		middleware.Logging,
	)

	adminStack := middleware.CreateStack(
		middleware.Logging,
		middleware.Auth,
		middleware.Admin,
	)

	handler := &movie.Handler{}
	fileHandler := http.StripPrefix("/static/", http.FileServer(http.Dir("static")))
	//public routes
	public := http.NewServeMux()
	public.HandleFunc("/{$}", handler.GetIndex)
	public.HandleFunc("GET /movie/{id}", handler.GetMovieByID)
	public.HandleFunc("GET /movies/", handler.GetAllMovies)
	public.HandleFunc(`POST /movies/`, handler.GetAllMoviesHTMX)
	public.HandleFunc(`POST /movies/reload`, handler.RealodSearchCatalog)
	public.HandleFunc("GET /movie/poster/{id}", handler.GetPoster)
	public.HandleFunc("DELETE /movie/{id}", handler.DeleteMovie)
	public.HandleFunc("GET /movie/{id}/comments/{last_comment_id}", handler.GetComments)
	public.HandleFunc("GET /movie/comment/{commentId}", handler.GetComment)
	public.HandleFunc("POST /search", handler.SearchByTitle)
	public.HandleFunc("POST /user/register", handler.PostRegister)
	public.HandleFunc("GET /user/register", handler.GetRegistrationPage)
	public.HandleFunc("POST /user/login", handler.Login)
	public.HandleFunc("GET /user/{id}", handler.GetUserPage)
	public.HandleFunc("GET /user/userinfo", handler.GetUserInfo)
	public.HandleFunc("GET /login", handler.GetLoginPage)
	public.HandleFunc("GET /empty", handler.EmptyResponse)
	//routes that require auth
	protected := http.NewServeMux()
	protected.HandleFunc("POST /user/logout", handler.Logout)
	protected.HandleFunc("POST /movie/comment", handler.PostComment)
	protected.HandleFunc("GET /comment/edit/{commentId}", handler.GetCommentEditForm)
	protected.HandleFunc("PUT /comment/edit", handler.UpdateComment)
	protected.HandleFunc("DELETE /comment/delete/{commentId}", handler.DeleteComment)
	//admin routes
	admin := http.NewServeMux()
	admin.HandleFunc("GET /movie/add", handler.AddMoviePage)
	admin.HandleFunc("POST /movie/", handler.PostMovie)
	admin.HandleFunc("GET /movie/edit/{id}", handler.GetEditMovieForm)
	admin.HandleFunc("PUT /movie/{id}", handler.UpdateMovie)
	admin.HandleFunc("PUT /movie/poster/{id}", handler.UpdatePoster)
	//combining all routes
	router.Handle("/", publicStack(public))
	router.Handle("/auth/", http.StripPrefix("/auth", protectedStack(protected)))
	router.Handle("/admin/", http.StripPrefix("/admin", adminStack(admin)))

	router.Handle("/static/", fileHandler)
}
