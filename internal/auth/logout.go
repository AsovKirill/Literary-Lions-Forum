package auth

import (
	"database/sql"
	"net/http"
	"time"
)

// NewLogoutHandler returns a handler for logging out users
func NewLogoutHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Try to get the session cookie from the user’s request
		cookie, err := r.Cookie("session_id")
		if err == nil && cookie.Value != "" {
			// If found, remove (delete) the session from the database
			dbConn.Exec(`DELETE FROM sessions WHERE id = ?`, cookie.Value)

			// Clear the session cookie in the user's browser by setting it expired
			http.SetCookie(w, &http.Cookie{
				Name:     "session_id",
				Value:    "",
				Path:     "/",
				Expires:  time.Unix(0, 0), // Expire immediately
				MaxAge:   -1,              // Remove from browser
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			})
		}

		// Redirect the user to the home page after logout
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
