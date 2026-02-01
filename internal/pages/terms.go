package pages

import (
	"database/sql"
	"fmt"
	"literary-lions/internal/middleware"
	"literary-lions/internal/models"
	"literary-lions/internal/views"
	"net/http"
)

// TermsPageHandler returns an HTTP handler that renders the Terms of Service page.
// It passes login status and username (if available) to the template for a personalized UI.
func TermsPageHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the current user (if logged in) from session
		_, username, logged := middleware.CurrentUser(r)
		// Prepare template data: Login state, username, categories (optional)
		data := struct {
			LoggedIn   bool
			Username   string
			Categories []models.Category
		}{
			LoggedIn: logged,
			Username: username,
		}
		// Render the terms.html template with the prepared data
		err := views.Templates.ExecuteTemplate(w, "terms.html", data)
		fmt.Println("Terms handler called!")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
