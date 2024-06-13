package movie

import (
	"bytes"
	"fmt"
	"html/template"
	"movie_db/db"
	"net/http"
)

var tmpl = template.Must(template.ParseGlob("views/*.html"))

type Handler struct{}

func (h *Handler) GetIndex(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "index", "")
}

func (h *Handler) GetMovieByID(w http.ResponseWriter, r *http.Request) {
	query := `SELECT * FROM movies WHERE movieId = ?`
	movie := &Movie{}
	err := db.DB.QueryRow(query, r.PathValue("id")).Scan(&movie.ID, &movie.Title, &movie.Genres)
	if err != nil {
		fmt.Fprint(w, "Error in GetMovieByID :", err)
		return
	}

	type Context struct {
		Movie  *Movie
		Genres []string
	}

	context := &Context{Movie: movie, Genres: movie.SplitGenresString()}

	buff := &bytes.Buffer{}
	err = tmpl.ExecuteTemplate(buff, "movie", context)
	if err != nil {
		fmt.Println(err)
	}

	err = tmpl.ExecuteTemplate(w, "index", &struct {
		Htmlstr template.HTML
		Data    string
	}{Htmlstr: template.HTML(buff.String()), Data: ""})
	if err != nil {
		fmt.Println(err)
	}
	//w.Write(bytes.Join([][]byte{[]byte(movie.Title), []byte(movie.Genres)}, []byte(" ")))
}

type ContextAddMovie struct {
	Movie  *Movie
	Errors []string
	ID     int64
}

func newEmptyContextAddMovie() (c *ContextAddMovie) {
	return &ContextAddMovie{
		&Movie{},
		[]string{},
		0,
	}
}

func (h *Handler) AddMovie(w http.ResponseWriter, r *http.Request) {

	buff := &bytes.Buffer{}
	err := tmpl.ExecuteTemplate(buff, "add-movie", newEmptyContextAddMovie())
	if err != nil {
		fmt.Println(err)
	}

	err = tmpl.ExecuteTemplate(w, "index", &struct {
		Htmlstr template.HTML
		Data    string
	}{Htmlstr: template.HTML(buff.String()), Data: ""})
	if err != nil {
		fmt.Println(err)
	}
	//w.Write(bytes.Join([][]byte{[]byte(movie.Title), []byte(movie.Genres)}, []byte(" ")))
}

func (h *Handler) PostMovie(w http.ResponseWriter, r *http.Request) {
	movie := &Movie{Title: r.PostFormValue("title"), Genres: r.PostFormValue("genres")}
	context := newEmptyContextAddMovie()
	if movie.Title == "" {
		context.Errors = append(context.Errors, "Title can't be empty!")
		context.Movie = movie
	} else {
		query := `INSERT INTO movies (title, genres) VALUES (?, ?)`
		result, err := db.DB.Exec(query, movie.Title, movie.Genres)
		if err != nil {
			context.Errors = append(context.Errors, "Error while adding a movie into DB")
			context.Movie = movie
			fmt.Println(err)
		} else {
			context.ID, _ = result.LastInsertId()
		}
	}

	tmpl.ExecuteTemplate(w, "add-movie", context)
}

func (h *Handler) DeleteMovie(w http.ResponseWriter, r *http.Request) {
	query := `DELETE FROM movies WHERE movieId = ?`
	result, err := db.DB.Exec(query, r.PathValue("id"))
	var response string
	if err != nil {
		fmt.Println(err)
		response = "Error while deleting a movie!"
	} else if n, _ := result.RowsAffected(); n == 0 {
		response = "Movie doesn't exist!"
	} else {
		response = "Movie deleted!"
	}

	tmpl.ExecuteTemplate(w, "delete-result", response)
}

func (h *Handler) UpdateMovie(w http.ResponseWriter, r *http.Request) {
	query := `UPDATE movies SET title = ?, genres = ? WHERE movieId = ?`
	result, err := db.DB.Exec(query, r.PostFormValue("title"), r.PostFormValue("genres"), r.PathValue("id"))
	if err != nil {
		fmt.Fprint(w, "Error while updating a movie: ", err)
		return
	}
	fmt.Fprint(w, result)
}

func (h *Handler) GetAllMovies(w http.ResponseWriter, r *http.Request) {
	buff := &bytes.Buffer{}
	err := tmpl.ExecuteTemplate(buff, "all-movies", "")
	if err != nil {
		fmt.Println(err)
	}

	err = tmpl.ExecuteTemplate(w, "index", &struct {
		Htmlstr template.HTML
		Data    string
	}{Htmlstr: template.HTML(buff.String()), Data: ""})
	if err != nil {
		fmt.Println(err)
	}
}

func (h *Handler) GetAllMoviesHTMX(w http.ResponseWriter, r *http.Request) {
	const recordsPerPage int = 20
	type MovContext struct {
		Movie
		Last bool
	}
	Movies := []MovContext{}
	query := `SELECT * FROM movies WHERE movieId > ? ORDER BY movieId ASC LIMIT ?`
	rows, err := db.DB.Query(query, r.PathValue("id"), recordsPerPage)
	if err != nil {
		fmt.Fprint(w, "Error in GetAllMovies :", err)
		return
	}
	for rows.Next() {
		movie := &MovContext{Movie: Movie{}}
		err := rows.Scan(&movie.ID, &movie.Title, &movie.Genres)
		if err != nil {
			fmt.Println("Error in GetAllMovies while scanning a row :", err)
		}
		Movies = append(Movies, *movie)
	}

	if len(Movies) > 0 {
		Movies[len(Movies)-1].Last = true
	}
	tmpl.ExecuteTemplate(w, "movie-rows", Movies)
}
