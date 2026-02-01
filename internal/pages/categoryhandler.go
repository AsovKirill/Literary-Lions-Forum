package pages

import (
	"database/sql"
	"literary-lions/internal/db"
	"literary-lions/internal/middleware"
	"literary-lions/internal/models"
	"literary-lions/internal/views"
	"net/http"
	"strconv"
	"strings"
)

// CategoryPageData holds all the information needed to render a category page.

type CategoryPageData struct {
	models.BasePageData
	Category   models.Category
	Posts      []models.Post
	Categories []models.Category
}

// NewCategoryHandler returns an http.HandlerFunc that serves category pages.
// It expects URLs like /category/3 where 3 is the category ID.
func NewCategoryHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Get current user (for personalization, login controls).
		_, username, loggedIn := middleware.CurrentUser(r)
		// 2. Parse the category ID from the URL.
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		// Check if the path is well-formed: /category/{id}
		if len(parts) != 2 || parts[0] != "category" {
			// If the category ID is not a valid number, show 404.
			http.NotFound(w, r)
			return
		}
		catID, err := strconv.Atoi(parts[1])
		if err != nil {
			http.NotFound(w, r)
			return
		}

		// 3. Fetch the specific category from the database.
		var category models.Category
		err = dbConn.QueryRow("SELECT id, name FROM categories WHERE id = ?", catID).Scan(&category.ID, &category.Name)
		if err != nil {
			// If category doesn't exist, show 404.
			http.NotFound(w, r)
			return
		}

		// 4. Fetch all posts in this category.
		posts, _ := db.FetchPostsByCategory(dbConn, catID)

		// 5. Fetch all categories (for displaying in the sidebar, navigation, etc).
		cats, _ := db.FetchCategories(dbConn)
		// 6. Prepare the template data.
		data := CategoryPageData{
			BasePageData: models.BasePageData{Username: username, LoggedIn: loggedIn},
			Category:     category,
			Posts:        posts,
			Categories:   cats,
		}
		// 7. Render the "category.html" template with the data.
		views.Templates.ExecuteTemplate(w, "category.html", data)
	}
}
