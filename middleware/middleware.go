package middleware

import (
	"bytes"
	"context"
	"log"
	"movie_db/movie"
	"net/http"
)

type Middleware func(http.Handler) http.Handler

func CreateStack(xs ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := range xs {
			next = xs[len(xs)-1-i](next)
		}
		return next
	}
}

type ResponseRecorder struct {
	http.ResponseWriter
	StatusCode int
	Body       *bytes.Buffer
}

func (rr *ResponseRecorder) WriteHeader(code int) {
	rr.StatusCode = code
	rr.ResponseWriter.WriteHeader(code)
}

func (rr *ResponseRecorder) Write(b []byte) (int, error) {
	rr.Body.Write(b)
	return rr.ResponseWriter.Write(b)
}

func NewResponseRecorder(w http.ResponseWriter) *ResponseRecorder {
	return &ResponseRecorder{ResponseWriter: w, StatusCode: http.StatusOK, Body: &bytes.Buffer{}}
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rr := NewResponseRecorder(w)
		next.ServeHTTP(rr, r)
		if rr.StatusCode >= 400 {
			log.Printf("Request method: %s, URL: %s, Response status code: %d, Response body: %s",
				r.Method, r.URL.Path, rr.StatusCode, rr.Body.String())
		}
	})
}

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil || cookie.Value == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		session, ok := movie.Sessions.Get(cookie.Value)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		r = r.WithContext(context.WithValue(r.Context(), movie.S, session))

		next.ServeHTTP(w, r)
	})
}
