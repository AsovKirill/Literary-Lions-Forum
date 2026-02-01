package middleware

import (
	"html/template"
	"literary-lions/internal/models"
	"log"
	"net/http"
	"path/filepath"
)

// errorTemplate is compiled once and used for rendering all error pages.
var errorTemplate = template.Must(
	template.ParseFiles(filepath.Join("web", "templates", "error.html")),
)

// ErrorHandler renders a custom error page (error.html) with the provided HTTP status code and message.
// The user's login state and username can be passed to personalize the error template.
func ErrorHandler(w http.ResponseWriter, status int, message string, loggedIn bool, username string) {
	// Set the HTTP response status code.
	w.WriteHeader(status)
	// Fill the ErrorData struct for the template.
	data := models.ErrorData{
		Status:   status,
		Message:  message,
		LoggedIn: loggedIn,
		Username: username,
	}
	// Try to execute the template. If it fails, fall back to the default http.Error.
	if err := errorTemplate.Execute(w, data); err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "An error occurred.", http.StatusInternalServerError)
	}
}

// RecoverMiddleware is a middleware that recovers from panics in handlers.
// If a panic occurs, it logs the panic and renders a custom 500 error page instead of crashing the server.
func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Defer a function to catch panics in the handler.
		defer func() {
			if rec := recover(); rec != nil {
				// Log the panic for debugging.
				log.Printf("Panic recovered: %v", rec)
				// Display a friendly 500 error page using your template.
				ErrorHandler(w, http.StatusInternalServerError, "Something went wrong (500)", false, "")
			}
		}()
		// Call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}
