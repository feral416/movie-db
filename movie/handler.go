package movie

import (
	"fmt"
	"html/template"
	"io"
	"movie_db/db"
	"movie_db/utils"
	"net/http"
	"os"
	"strconv"
	"unicode/utf8"
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

	context := &Context{Movie: movie, Genres: movie.SplitGenresString()}

	err = utils.TemplateWrap(tmpl, w, "movie", context, "index", "")
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (h *Handler) GetPoster(w http.ResponseWriter, r *http.Request) {
	fileBytes, err := os.ReadFile("assets/posters/" + r.PathValue("id") + ".png")
	if err != nil {
		bytes, err := os.ReadFile("assets/posters/no-poster.png")
		if err != nil {
			return
		}
		w.Write(bytes)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(fileBytes)
}

func (h *Handler) AddMoviePage(w http.ResponseWriter, r *http.Request) {
	err := utils.TemplateWrap(tmpl, w, "add-movie", newEmptyContextAddMovie(), "index", "")
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (h *Handler) PostMovie(w http.ResponseWriter, r *http.Request) {
	//time.Sleep(5 * time.Second)
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
	context := NewMovieActionsContext()
	query := `DELETE FROM movies WHERE movieId = ?`
	result, err := db.DB.Exec(query, r.PathValue("id"))
	if err != nil {
		fmt.Println(err)
		context.Error = "Error while deleting a movie!"
	} else if n, _ := result.RowsAffected(); n == 0 {
		context.Error = "Movie doesn't exist!"
	} else {
		context.Msg = "Movie deleted!"
	}
	tmpl.ExecuteTemplate(w, "movie-actions-result", context)
}

func (h *Handler) GetEditMovieForm(w http.ResponseWriter, r *http.Request) {
	context := NewMovieActionsContext()
	query := `SELECT * FROM movies WHERE movieId = ?`
	err := db.DB.QueryRow(query, r.PathValue("id")).Scan(&context.ID, &context.Title, &context.Genres)
	if err != nil {
		context.Error = "Error while getting movie information from the DB!"
	}
	tmpl.ExecuteTemplate(w, "movie-actions-result", context)
}

func (h *Handler) EmptyResponse(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "")
}

func (h *Handler) UpdateMovie(w http.ResponseWriter, r *http.Request) {
	context := NewMovieActionsContext()
	title := r.PostFormValue("title")
	if utf8.RuneCount([]byte(title)) < 1 {
		context.Error = "Movie not updated: title too short!"
	} else {
		query := `UPDATE movies SET title = ?, genres = ? WHERE movieId = ?`
		result, err := db.DB.Exec(query, title, r.PostFormValue("genres"), r.PathValue("id"))
		if err != nil {
			context.Error = "Error while updating a movie in database!"
		} else if n, _ := result.RowsAffected(); n == 1 {
			context.Msg = "Movie updated successfully."
		} else {
			context.Error = "Something wrong: movie might have been deleted!"
		}
	}
	tmpl.ExecuteTemplate(w, "movie-actions-result", context)
}

func (h *Handler) UpdatePoster(w http.ResponseWriter, r *http.Request) {
	context := func() *MovieActionsContext {
		context := NewMovieActionsContext()
		const MB int64 = 1 << 20 //megabyte size
		id := r.PathValue("id")

		if id == "" {
			context.Error = "Wrong movie id!"
			tmpl.ExecuteTemplate(w, "movie-actions-result", context)
			return context
		}
		var exists bool
		query := `SELECT EXISTS(SELECT * FROM movies WHERE movieId = ?)`
		if err := db.DB.QueryRow(query, r.PathValue("id")).Scan(&exists); err != nil {
			context.Error = "No connection to db or movie doesn't exist in db!"
			return context
		}
		if err := r.ParseMultipartForm(MB); err != nil {
			context.Error = "Error while parsing the file!"
			return context
		}
		formFile, header, err := r.FormFile("file")
		if err != nil {
			context.Error = "Error retrieving the file!"
			return context
		}
		defer formFile.Close()
		if header.Size > 1*MB {
			context.Error = "File larger than 1MB!"
			return context
		}
		formFileBytes, err := io.ReadAll(formFile)
		if err != nil {
			context.Error = "Error reading the file!"
			return context
		}
		if ct := http.DetectContentType(formFileBytes); ct != "image/png" {
			context.Error = "File is not a png image!"
			return context
		}
		newFile, err := os.Create("assets/posters/" + id + ".png")
		if err != nil {
			context.Error = "Error creating the file!"
			return context
		}
		defer newFile.Close()
		newFile.Write(formFileBytes)
		context.Msg = "File uploaded. Size: " + strconv.FormatInt(header.Size, 10)
		return context
	}()

	tmpl.ExecuteTemplate(w, "movie-actions-result", context)

}

func (h *Handler) GetAllMovies(w http.ResponseWriter, r *http.Request) {

	err := utils.TemplateWrap(tmpl, w, "all-movies", "", "index", "")
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (h *Handler) RealodSearchCatalog(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "search-catalog", "")
}

func (h *Handler) GetAllMoviesHTMX(w http.ResponseWriter, r *http.Request) {
	const recordsPerPage int = 20
	const idColumnName = " movieId "
	const titleColumnName = " title "
	prompt := r.PostFormValue("prompt")
	lastElement := r.PostFormValue("last-el")
	sortBy := r.PostFormValue("sort-by")
	order := r.PostFormValue("order")
	Movies := []MovContext{}
	whereColumnName := idColumnName
	whereCondition := " > ? "
	orderQueryPart := " ASC "
	sortPrmVal := lastElement
	prompt += "%"
	//id for date(cuz id rises every insertion)
	if order == "desc" {
		orderQueryPart = " DESC "
		if lastElement != "" {
			whereCondition = ` < ? `
		}
	}
	if sortBy == "title" {
		whereColumnName = titleColumnName
		sortPrmVal = lastElement
	}

	query := `SELECT * FROM movies WHERE title LIKE ? AND ` + whereColumnName + whereCondition + `ORDER BY` + whereColumnName + orderQueryPart + `LIMIT ?`

	rows, err := db.DB.Query(query, prompt, sortPrmVal, recordsPerPage)
	if err != nil {
		fmt.Println("Error in GetAllMovies :", err)
		return
	}
	for rows.Next() {
		movie := &MovContext{Movie: &Movie{}}
		err := rows.Scan(&movie.ID, &movie.Title, &movie.Genres)
		if err != nil {
			fmt.Println("Error in GetAllMovies while scanning a row :", err)
		}
		Movies = append(Movies, *movie)
	}

	if len(Movies) > 0 {
		lastIndex := len(Movies) - 1
		Movies[lastIndex].Last = strconv.Itoa(Movies[lastIndex].ID)
		if sortBy == "title" {
			Movies[lastIndex].Last = Movies[lastIndex].Title
		}
	}
	//lag for testing loading indicator
	//time.Sleep(5 * time.Second)
	tmpl.ExecuteTemplate(w, "movie-rows", Movies)
}

func (h *Handler) SearchByTitle(w http.ResponseWriter, r *http.Request) {
	//time.Sleep(5 * time.Second)
	const searchLimit int = 10
	userInput := r.PostFormValue("search")
	Movies := make([]Movie, 0)
	if utf8.RuneCountInString(userInput) < 1 {
		return
	}
	userInput += "%"
	query := `SELECT movieID, title FROM movies WHERE title LIKE ? LIMIT ?`
	rows, err := db.DB.Query(query, userInput, searchLimit)
	if err != nil {
		//TODO: add error logging
		return
	}
	for rows.Next() {
		movie := &Movie{}

		if err := rows.Scan(&movie.ID, &movie.Title); err != nil {
			fmt.Println("Error while scanning a row!")
			return
		}
		Movies = append(Movies, *movie)
	}
	tmpl.ExecuteTemplate(w, "search-results", Movies)
}
