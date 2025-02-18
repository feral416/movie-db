package movie

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"movie_db/db"
	"movie_db/utils"
	"net/http"
	"os"
	"strconv"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
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
	id := r.PathValue("id")
	query := `DELETE FROM movies WHERE movieId = ?`
	result, err := db.DB.Exec(query, id)
	if err != nil {
		fmt.Println(err)
		context.Error = "Error while deleting a movie!"
	} else if n, _ := result.RowsAffected(); n == 0 {
		context.Error = "Movie doesn't exist!"
	} else {
		context.Msg = "Movie deleted in database!"
		os.Remove("assets/posters/" + id + ".png")
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
		if err := db.DB.QueryRow(query, id).Scan(&exists); err != nil {
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

func (h *Handler) PostRegister(w http.ResponseWriter, r *http.Request) {
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")
	const passwordMinLength, passwordMaxLength = 8, 128
	if username == "" || password == "" {
		fmt.Println("Username or password can't be empty!")
		return
	}
	if !utils.PasswordAnalysis(password, passwordMinLength, passwordMaxLength) {
		fmt.Println("Password doesn't meet the requirements!")
		return
	}

	var exists bool
	query := `SELECT EXISTS(SELECT * FROM users WHERE username = ?)`
	if err := db.DB.QueryRow(query, username).Scan(&exists); err != nil {
		fmt.Println("No connection to db!")
		return
	}
	if exists {
		fmt.Println("Username already exists!")
		return
	}
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		fmt.Println("Error hashing password!")
		return
	}
	query = `INSERT INTO users (username, password) VALUES (?, ?)`
	result, err := db.DB.Exec(query, username, hashedPassword)
	if err != nil {
		fmt.Println(err)
		return
	}
	id, err := result.LastInsertId()
	if err != nil {
		fmt.Println("Error getting last insert ID:", err)
		return
	}
	fmt.Println("User added successfully. ID:", id)

}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")
	if username == "" || password == "" {
		http.Error(w, "Username or password can't be empty!", http.StatusBadRequest)
		fmt.Println("Username or password can't be empty!")
		return
	}
	query := `SELECT userId, password FROM users WHERE username = ?`
	user := &User{}
	if err := db.DB.QueryRow(query, username).Scan(&user.Id, &user.Password); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			fmt.Println("Invalid username or password", err)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		fmt.Println("Error while login", err)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		fmt.Println("Wrong password!")
		return
	}
	const tokenLenght = 32
	token, err := utils.GenerateToken(tokenLenght)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		fmt.Println("Error generating token!")
		return
	}
	session := Session{user.Id, username, time.Now().Add(time.Hour * 24)}
	Sessions.Create(session, token)
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Expires:  session.Expires,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})
	//http.Redirect(w, r, "/", http.StatusSeeOther)
	fmt.Println("User found:", w)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "No session token", http.StatusBadRequest)
		fmt.Println("No session token")
		return
	}
	Sessions.Delete(cookie.Value)
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})
	//http.Redirect(w, r, "/", http.StatusSeeOther)
	fmt.Println("User logged out")
}

func (h *Handler) PostComment(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	comment := r.PostFormValue("comment")
	if comment == "" {
		http.Error(w, "Comment can't be empty", http.StatusBadRequest)
		return
	}
	movieId := r.PostFormValue("movieId")
	if movieId == "" {
		http.Error(w, "No movie id", http.StatusBadRequest)
		return
	}
	session, ok := Sessions.Get(cookie.Value)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		fmt.Println("Unauthorized")
		return
	}
	query := `SELECT EXISTS(SELECT * FROM movies WHERE movieId = ?)`
	var exists bool
	if err := db.DB.QueryRow(query, movieId).Scan(&exists); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "Movie doesn't exist", http.StatusBadRequest)
		return
	}

	query = `INSERT INTO comments (userId, movieId, comment) VALUES (?, ?, ?)`
	res, err := db.DB.Exec(query, session.UserId, movieId, comment)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	commentId, err := res.LastInsertId()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	fmt.Println("Comment added successfully. ID:", commentId)
	// TODO: add response logic
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
