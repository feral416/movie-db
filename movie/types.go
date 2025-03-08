package movie

import (
	"database/sql"
	"log"
	"net/http"
	"sync"
	"time"
)

// type for a context key to avoid collision
type sessionContextKey string

const S sessionContextKey = "session"

var Sessions *SessionsStore

var SM *SessionManager

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

// Session store is session cache
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

func (ss *SessionsStore) KickUser(userId int) {
	ss.mu.RLock()
	keysToDelete := []string{}
	for k, v := range ss.Sessions {
		if v.UserId == userId {
			keysToDelete = append(keysToDelete, k)
		}
	}
	ss.mu.RUnlock()
	for _, k := range keysToDelete {
		ss.Delete(k)
	}
}

func (ss *SessionsStore) Delete(token string) {
	defer ss.mu.Unlock()
	ss.mu.Lock()
	delete(ss.Sessions, token)
}

func (ss *SessionsStore) Wipe() {
	defer ss.mu.Unlock()
	ss.mu.Lock()
	clear(ss.Sessions)
}

func (ss *SessionsStore) Reassign(nSS map[string]Session) {
	defer ss.mu.Unlock()
	ss.mu.Lock()
	ss.Sessions = nSS
}

func (ss *SessionsStore) Shrink() {
	ss.mu.RLock()
	keysToDelete := []string{}
	for k, v := range ss.Sessions {
		if time.Now().After(v.Expires) {
			keysToDelete = append(keysToDelete, k)
		}
	}
	ss.mu.RUnlock()
	for _, v := range keysToDelete {
		ss.Delete(v)
	}
}

func NewSessionsStore() *SessionsStore {
	return &SessionsStore{
		Sessions: make(map[string]Session),
	}
}

// Session manager manages sessions in remote DB and cache, where remote db data has priority
type SessionManager struct {
	DB    *sql.DB
	Cache *SessionsStore
}

func (sm *SessionManager) Create(s Session, token string) error {
	query := `INSERT INTO sessions(token, expirationDT, userId) VALUES(?, ?, ?)`
	if _, err := sm.DB.Exec(query, token, s.Expires, s.UserId); err != nil {
		return err
	}
	sm.Cache.Create(s, token)
	return nil
}

func (sm *SessionManager) Delete(token string) error {
	query := `DELETE FROM sessions WHERE token = ?`
	if _, err := sm.DB.Exec(query, token); err != nil {
		return err
	}
	sm.Cache.Delete(token)
	return nil
}

func (sm *SessionManager) KickUser(userId int) error {
	query := `DELETE FROM sessions WHERE userId = ?`
	if _, err := sm.DB.Exec(query, userId); err != nil {
		return err
	}
	sm.Cache.KickUser(userId)
	return nil
}

// Sync sessions on startup. Sync will block until completed to prevent drift
func (sm *SessionManager) InitSync() {
	const retryTime time.Duration = 10
	query := `SELECT s.token, s.expirationDT, s.userId, u.username, u.admin FROM sessions s JOIN users u ON s.userId = u.userId`
MainLoop:
	for {
		rows, err := sm.DB.Query(query)
		if err != nil {
			log.Printf("Error retrieving sessions from a DB: %s", err)
			time.Sleep(retryTime * time.Second)
			continue
		}
		newMap := make(map[string]Session)
		defer rows.Close()
		for rows.Next() {
			session := Session{}
			var token string
			err := rows.Scan(&token, &session.Expires, &session.UserId, &session.Username, &session.Admin)
			if err != nil {
				log.Printf("Error scanning a row from DB result (sessions): %s", err)
				time.Sleep(retryTime * time.Second)
				continue MainLoop
			}
			newMap[token] = session
		}
		sm.Cache.Reassign(newMap)
		break
	}
}

// Shrink checks and removes expired sessions
func (sm *SessionManager) Shrink() error {
	query := `DELETE FROM sessions WHERE expirationDT < NOW()`
	if _, err := sm.DB.Exec(query); err != nil {
		return err
	}
	sm.Cache.Shrink()
	return nil
}
