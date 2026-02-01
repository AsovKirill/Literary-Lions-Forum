package auth

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"literary-lions/internal/views"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// NewLoginHandler returns an HTTP handler for the login page and POST login submissions
func NewLoginHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// --- Show login form on GET ---
		if r.Method == http.MethodGet {
			views.Templates.ExecuteTemplate(w, "login.html", nil)
			return
		}

		// Grab submitted form fields
		email := r.FormValue("email")
		password := r.FormValue("password")

		// Prepare variables for user DB lookup
		var (
			id       int
			hash     string
			username string
		)

		// Try to find the user by email in the database
		err := dbConn.QueryRow(`SELECT id, password, username FROM users WHERE email=?`, email).
			Scan(&id, &hash, &username)

		// If no user found or DB error, show invalid credentials (do not reveal which failed)
		if err != nil {
			log.Printf("Login DB lookup failed for email %s: %v", email, err)
			views.Templates.ExecuteTemplate(w, "login.html", map[string]string{"LoginError": "Invalid credentials"})
			return
		}

		// Compare hashed password from DB to the submitted password
		if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
			log.Printf("Login password mismatch for user %s", email)
			views.Templates.ExecuteTemplate(w, "login.html", map[string]string{"LoginError": "Invalid credentials"})
			return
		}

		// Generate a new session ID (UUID)
		sID := uuid.NewString()

		// Insert session into DB, valid for 30 days
		_, err = dbConn.Exec(`INSERT INTO sessions(id,user_id,expires_at) VALUES(?,?,?)`,
			sID, id, time.Now().Add(30*24*time.Hour))
		if err != nil {
			log.Printf("Failed to create session for user %s: %v", email, err)
			http.Error(w, "server error", 500)
			return
		}

		// Set a secure session cookie in user's browser
		http.SetCookie(w, &http.Cookie{
			Name:     "session_id",
			Value:    sID,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		// Redirect user to the homepage after login
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
