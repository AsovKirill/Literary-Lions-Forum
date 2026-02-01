package pages

import (
	"database/sql"
	"literary-lions/internal/db"
	"literary-lions/internal/middleware"
	"literary-lions/internal/models"
	"literary-lions/internal/views"
	"log"
	"net/http"
)

// NewHomeHandler returns a handler for the main forum page ("/").
// It loads all recent posts, categories, and popular posts for display.
// It handles error scenarios and passes user/session info to the template.
func NewHomeHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user info from session (if logged in)
		userID, username, loggedIn := middleware.CurrentUser(r)

		// Allow simulating a 500 error for testing by visiting /?force500=true
		if r.URL.Query().Get("force500") == "true" {
			middleware.ErrorHandler(w, http.StatusInternalServerError, "Simulated server error.", loggedIn, username)
			return
		}

		// Fetch all posts for the home feed, with respect to current user (for likes)
		posts, err := db.FetchPosts(dbConn, userID)
		// Fetch all categories for the sidebar or menu
		categories, _ := db.FetchCategories(dbConn)
		// Fetch top 6 popular posts (by likes)
		popularPosts, _ := db.FetchPopularPosts(dbConn, 6)

		// If there was a DB problem, show a friendly error page
		if err != nil {
			log.Println("DB problem:", err)
			middleware.ErrorHandler(w, http.StatusInternalServerError, "Something went wrong while loading the forum. Please try again later.", loggedIn, username)
			return
		}

		// For each post, fetch the number of likes and comments (show as badges)
		for i := range posts {
			likes, _ := db.CountLikesByPostID(dbConn, posts[i].ID)
			comments, _ := db.CountCommentsByPostID(dbConn, posts[i].ID)
			posts[i].Likes = likes
			posts[i].Comments = comments
		}

		// Prepare data to send to the template
		data := models.HomePageData{
			BasePageData: models.BasePageData{
				Username: username,
				LoggedIn: loggedIn,
			},
			Posts:        posts,
			Categories:   categories,
			PopularPosts: popularPosts,
		}
		// Render the home page template
		err = views.Templates.ExecuteTemplate(w, "index.html", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
