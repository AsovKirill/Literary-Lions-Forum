package pages

import (
	"database/sql"
	"fmt"
	"literary-lions/internal/models"
	"literary-lions/internal/views"
	"net/http"
)

// AboutPageHandler renders the "About" page.
// It prepares template data, including login status and categories for navigation/sidebar.
func AboutPageHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			LoggedIn   bool
			Username   string
			Categories []models.Category
		}{}
		err := views.Templates.ExecuteTemplate(w, "about.html", data)
		fmt.Println("About handler called!")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
