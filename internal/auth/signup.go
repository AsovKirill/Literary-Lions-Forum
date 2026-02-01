package auth

import (
	"database/sql"
	"log"
	"net/http"
	"strings"

	"literary-lions/internal/middleware"
	"literary-lions/internal/views"

	"golang.org/x/crypto/bcrypt"
)

// NewSignupHandler returns an HTTP handler for user registration/signup
func NewSignupHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("🟡 SignupHandler received request:", r.Method)

		// Struct to hold form data and errors for template rendering
		type FormData struct {
			Error    string
			Email    string
			Username string
		}

		switch r.Method {
		case http.MethodGet:
			// Render signup form for GET requests (show empty form)
			_ = views.Templates.ExecuteTemplate(w, "signup.html", nil)
			return

		case http.MethodPost:
			// --- Collect and validate form input ---
			email := strings.TrimSpace(r.FormValue("email"))
			username := strings.TrimSpace(r.FormValue("username"))
			password := r.FormValue("password")

			// Check for empty fields (required)
			if email == "" || username == "" || password == "" {
				// 400 Bad Request for missing data
				w.WriteHeader(http.StatusBadRequest)
				_ = views.Templates.ExecuteTemplate(w, "signup.html", FormData{
					Error:    "All fields are required",
					Email:    email,
					Username: username,
				})
				return
			}

			// Validate email format (very basic check)
			if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
				w.WriteHeader(http.StatusBadRequest)
				_ = views.Templates.ExecuteTemplate(w, "signup.html", FormData{
					Error:    "Invalid email address",
					Email:    email,
					Username: username,
				})
				return
			}

			// --- Check for existing user by email or username ---
			var exists int
			err := dbConn.QueryRow(
				"SELECT COUNT(*) FROM users WHERE email = ? OR username = ?",
				email, username,
			).Scan(&exists)
			if err != nil {
				// 500 Internal Server Error for DB issues
				log.Println("DB error:", err)
				middleware.ErrorHandler(w, http.StatusInternalServerError, "Database error", false, "")
				return
			}
			if exists > 0 {
				// 400 Bad Request for duplicate email/username
				w.WriteHeader(http.StatusBadRequest)
				_ = views.Templates.ExecuteTemplate(w, "signup.html", FormData{
					Error:    "Email or username already in use",
					Email:    email,
					Username: username,
				})
				return
			}

			// --- Hash password securely using bcrypt ---
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				log.Println("Hash error:", err)
				middleware.ErrorHandler(w, http.StatusInternalServerError, "Server error", false, "")
				return
			}

			// --- Store the new user in the database ---
			_, err = dbConn.Exec(
				"INSERT INTO users (username, email, password) VALUES (?, ?, ?)",
				username, email, string(hashedPassword),
			)
			if err != nil {
				// 500 Internal Server Error if insert fails
				log.Println("Insert error:", err)
				middleware.ErrorHandler(w, http.StatusInternalServerError, "Could not create user", false, "")
				return
			}

			// Log success and redirect to login page
			log.Printf("✅ User %s registered successfully\n", username)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return

		default:
			// Only allow GET or POST (others: 405 Method Not Allowed)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
