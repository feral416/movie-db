package movie

import (
	"database/sql"
	"html/template"
	"io"
	"log"
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
	const templateName string = "index"
	if err := tmpl.ExecuteTemplate(w, templateName, nil); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error executing template %s: %s", templateName, err)
		return
	}
}

func (h *Handler) GetMovieByID(w http.ResponseWriter, r *http.Request) {
	const wrapperName, contentName string = "index", "movie"
	query := `SELECT movieId, title, genres FROM movies WHERE movieId = ?`
	movie := &Movie{}
	err := db.DB.QueryRow(query, r.PathValue("id")).Scan(&movie.ID, &movie.Title, &movie.Genres)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Movie doesn't exist!", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("Error getting movie: %s", err)
		return
	}

	context := &Context{Movie: movie, Genres: movie.SplitGenresString()}

	if err = utils.TemplateWrap(tmpl, w, contentName, context, wrapperName, nil); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error wrapping template %s with template %s: %s", contentName, wrapperName, err)
		return
	}
}

func (h *Handler) GetPoster(w http.ResponseWriter, r *http.Request) {
	fileBytes, err := os.ReadFile("assets/posters/" + r.PathValue("id") + ".png")
	if err != nil {
		fileBytes, err = os.ReadFile("assets/posters/no-poster.png")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Printf("Error reading the default poster: %s", err)
			return
		}
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	if _, err = w.Write(fileBytes); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error writing writing image to response: %s", err)
		return
	}
}

func (h *Handler) AddMoviePage(w http.ResponseWriter, r *http.Request) {
	const wrapperName, contentName string = "index", "add-movie"
	if err := utils.TemplateWrap(tmpl, w, "add-movie", nil, "index", ""); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error wrapping template %s with template %s: %s", contentName, wrapperName, err)
		return
	}
}

func (h *Handler) PostMovie(w http.ResponseWriter, r *http.Request) {
	const templateName string = "add-movie"
	title := r.PostFormValue("title")
	if title == "" {
		http.Error(w, "Title can't be empty!", http.StatusBadRequest)
		return
	}
	query := `INSERT INTO movies (title, genres) VALUES (?, ?)`
	result, err := db.DB.Exec(query, title, r.PostFormValue("genres"))
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error writing movie to db: %s", err)
		return
	}
	Id, _ := result.LastInsertId()
	w.WriteHeader(http.StatusCreated)
	if err = tmpl.ExecuteTemplate(w, templateName, Id); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error executing template %s: %s", templateName, err)
		return
	}
}

func (h *Handler) DeleteMovie(w http.ResponseWriter, r *http.Request) {
	const templateName string = "deleted-movie"
	idStr := r.PathValue("id")
	if id, err := strconv.Atoi(idStr); err != nil || id < 0 {
		http.Error(w, "Wrong movie id!", http.StatusBadRequest)
		return
	}
	query := `DELETE FROM movies WHERE movieId = ?`
	result, err := db.DB.Exec(query, idStr)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error while deleting a movie: %s", err)
		return
	}
	if n, _ := result.RowsAffected(); n == 0 {
		http.Error(w, "Movie doesn't exist!", http.StatusBadRequest)
		return
	}
	//Potential error is not handled because poster may not exist
	os.Remove("assets/posters/" + idStr + ".png")
	if err = tmpl.ExecuteTemplate(w, templateName, nil); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error executing template %s: %s", templateName, err)
		return
	}
}

func (h *Handler) GetEditMovieForm(w http.ResponseWriter, r *http.Request) {
	movie := &Movie{}
	const formName string = "movie-edit-form"
	id := r.PathValue("id")
	if v, err := strconv.Atoi(id); err != nil || v < 0 {
		http.Error(w, "Wrong movie id!", http.StatusBadRequest)
		return
	}
	query := `SELECT movieId, title, genres FROM movies WHERE movieId = ?`
	err := db.DB.QueryRow(query, id).Scan(&movie.ID, &movie.Title, &movie.Genres)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error while getting movie information from a DB: %s", err)
		return
	}
	if err = tmpl.ExecuteTemplate(w, formName, movie); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error executing template %s: %s", formName, err)
		return
	}

}

func (h *Handler) EmptyResponse(w http.ResponseWriter, r *http.Request) {
	//fmt.Fprint(w, "")
}

func (h *Handler) UpdateMovie(w http.ResponseWriter, r *http.Request) {
	strId := r.PathValue("id")
	id, err := strconv.Atoi(strId)
	title := r.PostFormValue("title")
	if title == "" || id < 0 || err != nil {
		http.Error(w, "Invalid id or title!", http.StatusBadRequest)
		return
	}
	query := `UPDATE movies SET title = ?, genres = ? WHERE movieId = ?`
	result, err := db.DB.Exec(query, title, r.PostFormValue("genres"), id)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error while updating a movie in database: %s", err)
		return
	}
	if n, _ := result.RowsAffected(); n == 0 {
		http.Error(w, "Something wrong: movie not updated!", http.StatusBadRequest)
		return
	}
	w.Header().Add("HX-Redirect", "/movie/"+strId)
}

func (h *Handler) UpdatePoster(w http.ResponseWriter, r *http.Request) {
	const MB int64 = 1 << 20 //megabyte size
	strId := r.PathValue("id")
	id, err := strconv.Atoi(strId)
	if err != nil || id < 0 {
		http.Error(w, "Wrong id!", http.StatusBadRequest)
		return
	}

	var exists bool
	query := `SELECT EXISTS(SELECT * FROM movies WHERE movieId = ?)`
	if err := db.DB.QueryRow(query, id).Scan(&exists); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Movie doesn't exist!", http.StatusBadRequest)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Problem with db: %s", err)
		return
	}
	if err := r.ParseMultipartForm(MB); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error while parsing the file: %s", err)
		return
	}
	formFile, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error retrieving the file: %s", err)
		return
	}
	defer func() {
		if err = formFile.Close(); err != nil {
			log.Printf("Error closing the file: %s", err)
			return
		}
	}()
	if header.Size > 1*MB {
		http.Error(w, "File larger than 1MB!", http.StatusBadRequest)
		return
	}
	formFileBytes, err := io.ReadAll(formFile)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error reading the file: %s", err)
		return
	}
	if ct := http.DetectContentType(formFileBytes); ct != "image/png" {
		http.Error(w, "File is not a png image!", http.StatusBadRequest)
		return
	}
	newFile, err := os.Create("assets/posters/" + strId + ".png")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error creating the file: %s", err)
		return
	}
	defer func() {
		if err := newFile.Close(); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Printf("Error closing new file on server: %s", err)
			return
		}
	}()
	if _, err = newFile.Write(formFileBytes); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error writing the wile on disc: %s", err)
		return
	}
	//"File uploaded. Size: " + strconv.FormatInt(header.Size, 10)
	w.Header().Add("HX-Redirect", "/movie/"+strId)
}

func (h *Handler) GetAllMovies(w http.ResponseWriter, r *http.Request) {
	const wrapperName, contentName string = "index", "all-movies"
	if err := utils.TemplateWrap(tmpl, w, contentName, "", wrapperName, ""); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error wrapping template %s with template %s: %s", contentName, wrapperName, err)
		return
	}
}

func (h *Handler) RealodSearchCatalog(w http.ResponseWriter, r *http.Request) {
	const templateName string = "search-catalog"
	if err := tmpl.ExecuteTemplate(w, templateName, nil); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error executing template %s: %s", templateName, err)
		return
	}
}

func (h *Handler) PostRegister(w http.ResponseWriter, r *http.Request) {
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")
	confirmPassword := r.PostFormValue("confirmPassword")
	const passwordMinLength, passwordMaxLength = 8, 128
	if username == "" || password == "" || confirmPassword == "" {
		http.Error(w, "Username or password can't be empty!", http.StatusBadRequest)
		return
	}
	if password != confirmPassword {
		http.Error(w, "Passwords don't match!", http.StatusBadRequest)
		return
	}
	if !utils.PasswordAnalysis(password, passwordMinLength, passwordMaxLength) {
		http.Error(w, "Password doesn't meet the requirements!", http.StatusBadRequest)
		return
	}

	var exists bool
	query := `SELECT EXISTS(SELECT * FROM users WHERE username = ?)`
	if err := db.DB.QueryRow(query, username).Scan(&exists); err != nil {
		http.Error(w, "DB error!", http.StatusInternalServerError)
		log.Printf("Error checking username existanse in db: %s", err)
		return
	}
	if exists {
		http.Error(w, "Username already exists!", http.StatusBadRequest)
		return
	}
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		http.Error(w, "Error hashing password!", http.StatusInternalServerError)
		log.Printf("Error hashing password: %s", err)
		return
	}
	query = `INSERT INTO users (username, password) VALUES (?, ?)`
	result, err := db.DB.Exec(query, username, hashedPassword)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error inserting user to db: %s", err)
		return
	}
	_, err = result.LastInsertId()
	if err != nil {
		http.Error(w, "Error getting last insert ID:", http.StatusInternalServerError)
		log.Printf("Error getting last insert ID: %s", err)
		return
	}
	w.Header().Add("HX-Redirect", "/login")
}

func (h *Handler) GetRegistrationPage(w http.ResponseWriter, r *http.Request) {
	const wrapperName, contentName string = "index", "register-block"
	err := utils.TemplateWrap(tmpl, w, "contentName", nil, wrapperName, nil)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("Error wrapping template %s with template %s: %s", contentName, wrapperName, err)
		return
	}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")
	if username == "" || password == "" {
		http.Error(w, "Username or password can't be empty!", http.StatusBadRequest)
		return
	}
	query := `SELECT userId, password FROM users WHERE username = ?`
	user := &User{}
	if err := db.DB.QueryRow(query, username).Scan(&user.Id, &user.Password); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error getting user from db: %s", err)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}
	const tokenLenght = 32
	token, err := utils.GenerateToken(tokenLenght)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error generating token: %s", err)
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
	w.Header().Add("HX-Redirect", "/")
}

func (h *Handler) GetLoginPage(w http.ResponseWriter, r *http.Request) {
	const wrapperName, contentName string = "index", "login-block"
	if err := utils.TemplateWrap(tmpl, w, contentName, nil, wrapperName, nil); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("Error wrapping template %s with template %s: %s", contentName, wrapperName, err)
		return
	}

}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error getting cookie: %s", err)
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
	w.Header().Add("HX-Redirect", "/")
}

func (h *Handler) PostComment(w http.ResponseWriter, r *http.Request) {
	const templateName string = "comments-section"
	comment := r.PostFormValue("comment")
	if comment == "" {
		http.Error(w, "Comment can't be empty!", http.StatusBadRequest)
		return
	}
	movieId := r.PostFormValue("movieId")
	if v, err := strconv.Atoi(movieId); err != nil || v < 0 {
		http.Error(w, "Wrong movie id!", http.StatusBadRequest)
		return
	}
	session, ok := r.Context().Value(S).(Session)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error getting session from context in PostComment")
		return
	}
	query := `SELECT EXISTS(SELECT * FROM movies WHERE movieId = ?)`
	var exists bool
	if err := db.DB.QueryRow(query, movieId).Scan(&exists); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("Error checking movie existance in db: %s", err)
		return
	}
	if !exists {
		http.Error(w, "Movie doesn't exist!", http.StatusBadRequest)
		return
	}

	query = `INSERT INTO comments (userId, movieId, comment) VALUES (?, ?, ?)`
	res, err := db.DB.Exec(query, session.UserId, movieId, comment)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error inserting comment into db: %s", err)
		return
	}
	_, err = res.LastInsertId()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error getting last insert ID: %s", err)
		return
	}
	if err := tmpl.ExecuteTemplate(w, templateName, movieId); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error executing template %s: %s", templateName, err)
		return
	}
}

func (h *Handler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	const templateName string = "deleted-comment"
	commentId := r.PathValue("commentId")
	if v, err := strconv.Atoi(commentId); err != nil || v < 0 {
		http.Error(w, "Wrong comment id!", http.StatusBadRequest)
		return
	}
	session, ok := r.Context().Value(S).(Session)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error getting session from context in DeleteComment")
		return
	}
	query := `DELETE FROM comments WHERE commentId = ? AND userId = ?`
	sqlRes, err := db.DB.Exec(query, commentId, session.UserId)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error deleting comment from db: %s", err)
		return
	}
	if n, _ := sqlRes.RowsAffected(); n == 0 {
		http.Error(w, "Comment doesn't exist or you are not the author", http.StatusBadRequest)
		return
	}
	if err := tmpl.ExecuteTemplate(w, templateName, nil); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error executing template %s: %s", templateName, err)
		return
	}
}

func (h *Handler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	const templateName string = "comment"
	commentId := r.PostFormValue("commentId")
	if v, err := strconv.Atoi(commentId); err != nil || v < 0 {
		http.Error(w, "Wrong comment id!", http.StatusBadRequest)
		return
	}
	comment := r.PostFormValue("comment")
	if comment == "" {
		http.Error(w, "Comment can't be empty!", http.StatusBadRequest)
		return
	}
	session, ok := r.Context().Value(S).(Session)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error getting session from context in UpdateComment")
		return
	}
	query := `UPDATE comments SET comment = ? WHERE commentId = ? AND userId = ?`
	sqlRes, err := db.DB.Exec(query, comment, commentId, session.UserId)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("Error updating comment in db: %s", err)
		return
	}
	if n, _ := sqlRes.RowsAffected(); n == 0 {
		http.Error(w, "Comment doesn't exist or you are not the author!", http.StatusBadRequest)
		return
	}
	err = tmpl.ExecuteTemplate(w, templateName, struct {
		CommentId   string
		CommentText string
	}{commentId, comment})
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error executing template %s: %s", templateName, err)
		return
	}
}

func (h *Handler) GetComments(w http.ResponseWriter, r *http.Request) {
	const templateName string = "comments"
	const commentsPerQuery = 20
	session := Sessions.GetSessionInfo(r)
	ctxSlice := []CommentsContext{}
	movieId := r.PathValue("id")
	if v, err := strconv.Atoi(movieId); err != nil || v < 0 {
		http.Error(w, "Wrong movie id!", http.StatusBadRequest)
		return
	}
	lastCommentId := r.PathValue("last_comment_id")
	if v, err := strconv.Atoi(lastCommentId); err != nil || v < 0 {
		http.Error(w, "Wrong last comment id!", http.StatusBadRequest)
		return
	}
	query := "CALL GetComments(?, ?, ?)"
	rows, err := db.DB.Query(query, movieId, lastCommentId, commentsPerQuery)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error getting comments from db: %s", err)
		return
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Error closing rows: %s", err)
			return
		}
	}()
	for rows.Next() {
		comment := &CommentsContext{}
		err := rows.Scan(&comment.CommentText, &comment.PostedDT, &comment.UserId, &comment.CommentId, &comment.Username)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Printf("Error scanning a row: %s", err)
			return
		}
		if session != nil {
			comment.Owner = session.UserId == comment.UserId
		}
		ctxSlice = append(ctxSlice, *comment)
	}
	if len(ctxSlice) == commentsPerQuery {
		ctxSlice[commentsPerQuery-1].Last = true
		ctxSlice[commentsPerQuery-1].MovieId = movieId
	}
	if err := tmpl.ExecuteTemplate(w, templateName, ctxSlice); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error executing template %s: %s", templateName, err)
		return
	}
}

func (h *Handler) GetComment(w http.ResponseWriter, r *http.Request) {
	const templateName string = "comment"
	commentId, err := strconv.Atoi(r.PathValue("commentId"))
	if err != nil || commentId < 0 {
		http.Error(w, "Wrong comment id!", http.StatusBadRequest)
		return
	}
	query := `SELECT comment FROM comments WHERE commentId = ?`
	var commentText string
	if err := db.DB.QueryRow(query, commentId).Scan(&commentText); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Comment not found!", http.StatusBadRequest)
			return
		}
		http.Error(w, "Internal server error!", http.StatusInternalServerError)
		log.Printf("Error getting comment from db: %s", err)
		return
	}
	err = tmpl.ExecuteTemplate(w, templateName, struct {
		CommentId   int
		CommentText string
	}{commentId, commentText})
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error executing template %s: %s", templateName, err)
		return
	}
}

func (h *Handler) GetCommentEditForm(w http.ResponseWriter, r *http.Request) {
	const templateName string = "edit-comment"
	commentId, err := strconv.Atoi(r.PathValue("commentId"))
	if err != nil || commentId < 0 {
		http.Error(w, "Wrong comment id", http.StatusBadRequest)
		return
	}
	query := `SELECT comment FROM comments WHERE commentId = ?`
	var commentText string
	if err := db.DB.QueryRow(query, commentId).Scan(&commentText); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Comment not found!", http.StatusBadRequest)
			return
		}
		http.Error(w, "Internal server error!", http.StatusInternalServerError)
		log.Printf("Error getting comment from db: %s", err)
		return
	}
	err = tmpl.ExecuteTemplate(w, templateName, struct {
		CommentId   int
		CommentText string
	}{commentId, commentText})
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error executing template %s: %s", templateName, err)
		return
	}
}

func (h *Handler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	const templateName string = "auth-block"
	context := Sessions.GetSessionInfo(r)
	if err := tmpl.ExecuteTemplate(w, templateName, context); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error executing template %s: %s", templateName, err)
		return
	}
}

func (h *Handler) GetAllMoviesHTMX(w http.ResponseWriter, r *http.Request) {
	const templateName string = "movie-rows"
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
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error in GetAllMovies : %s", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		movie := &MovContext{Movie: &Movie{}}
		err := rows.Scan(&movie.ID, &movie.Title, &movie.Genres)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Printf("Error while scanning a row: %s", err)
			return
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
	if err := tmpl.ExecuteTemplate(w, templateName, Movies); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error executing template %s: %s", templateName, err)
		return
	}
}

func (h *Handler) SearchByTitle(w http.ResponseWriter, r *http.Request) {
	//time.Sleep(5 * time.Second)
	const templateName string = "search-results"
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
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error getting seraching movies in db: %s", err)
		return
	}
	for rows.Next() {
		movie := &Movie{}

		if err := rows.Scan(&movie.ID, &movie.Title); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Printf("Error while scanning a row: %s", err)
			return
		}
		Movies = append(Movies, *movie)
	}
	if err := tmpl.ExecuteTemplate(w, templateName, Movies); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error executing template %s: %s", templateName, err)
		return
	}
}

func (h *Handler) GetUserPage(w http.ResponseWriter, r *http.Request) {
	const wrapperName, contentName string = "index", "user-page"
	userId, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || userId < 0 {
		http.Error(w, "Wrong user id!", http.StatusBadRequest)
		return
	}
	user := &User{}
	query := `SELECT userId, username, registerDate, admin, banUntil FROM users WHERE userId = ?`
	if err := db.DB.QueryRow(query, userId).Scan(&user.Id, &user.Username, &user.RegisterDate, &user.Admin, &user.BanUntil); err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found!", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("Error getting user from db: %s", err)
		return
	}
	if user.BanUntil.After(time.Now()) {
		user.Banned = true
	}
	if err := utils.TemplateWrap(tmpl, w, contentName, user, wrapperName, nil); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error wrapping template %s with template %s: %s", contentName, wrapperName, err)
		return
	}
}
