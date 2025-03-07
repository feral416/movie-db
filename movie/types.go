package movie

import (
	"net/http"
	"sync"
	"time"
)

// type for a context key to avoid collision
type sessionContextKey string

const S sessionContextKey = "session"

var Sessions *SessionsStore

type MovContext struct {
	*Movie
	Last string
}

type User struct {
	Id           int
	Password     string
	Username     string
	RegisterDate string
	Admin        bool
	Banned       bool
	BanUntil     string
}

type Comment struct {
	CommentId   int
	UserId      int
	CommentText string
	PostedDT    string
	Username    string
	MovieId     string
}

type CommentsContext struct {
	Comment
	Last  bool
	Owner bool //Comment owned by user
}

type Session struct {
	UserId   int
	Username string
	Admin    bool
	Expires  time.Time
}

type SessionsStore struct {
	Sessions map[string]Session
	mu       sync.RWMutex
}

func (ss *SessionsStore) Create(s Session, token string) {
	defer ss.mu.Unlock()
	ss.mu.Lock()
	ss.Sessions[token] = s
}

func (ss *SessionsStore) Get(token string) (Session, bool) {
	defer ss.mu.RUnlock()
	ss.mu.RLock()
	s, ok := ss.Sessions[token]
	return s, ok
}

// Get session info from a request
func (ss *SessionsStore) GetSessionInfo(r *http.Request) *Session {
	if r == nil {
		return nil
	}
	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		return nil
	}
	session, ok := ss.Get(cookie.Value)
	if !ok {
		return nil
	}
	return &session
}

func (ss *SessionsStore) Delete(token string) {
	defer ss.mu.Unlock()
	ss.mu.Lock()
	delete(ss.Sessions, token)
}

func NewSessionsStore() *SessionsStore {
	return &SessionsStore{
		Sessions: make(map[string]Session),
	}
}
