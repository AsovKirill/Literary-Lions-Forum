package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"time"
)

// ctxKey is a custom type to avoid context key collisions.
type ctxKey string

// userKey is the key under which user info is stored in the request context.
const userKey ctxKey = "user"

// NewWithSession returns a middleware that loads the logged-in user from the session cookie.
// If the session is valid, user information is attached to the request context for use in handlers.
func NewWithSession(dbConn *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Read the "session_id" cookie (if present)
			c, _ := r.Cookie("session_id")
			if c != nil {
				var (
					id       int
					username string
				)
				// Query the sessions and users tables to verify the session and fetch user info.
				err := dbConn.QueryRow(`
                    SELECT users.id, users.username
                    FROM sessions JOIN users ON users.id=sessions.user_id
                    WHERE sessions.id=? AND sessions.expires_at>?
                `, c.Value, time.Now()).Scan(&id, &username)
				if err == nil {
					// Add user info to the request context if session is valid.
					r = r.WithContext(context.WithValue(r.Context(), userKey, struct {
						ID       int
						Username string
					}{id, username}))
				}
				// If session is not valid, user is treated as logged out.
			}
			// Call the next handler, with context updated if user is logged in.
			next.ServeHTTP(w, r)
		})
	}
}

// CurrentUser extracts the user ID and username from the request context.
// Returns ok == true if a user is logged in, otherwise false.
func CurrentUser(r *http.Request) (id int, name string, ok bool) {
	u, ok := r.Context().Value(userKey).(struct {
		ID       int
		Username string
	})
	if ok {
		return u.ID, u.Username, true
	}
	return 0, "", false
}
