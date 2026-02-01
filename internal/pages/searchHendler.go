package pages

import (
	"database/sql"
	"fmt"
	"literary-lions/internal/db"
	"literary-lions/internal/middleware"
	"literary-lions/internal/models"
	"literary-lions/internal/views"
	"net/http"
	"strings"
	"time"
)

// NewSearchHandler returns an HTTP handler for processing search requests.
// Supports:
//   - Redirect to category page if query matches a category name.
//   - Redirect to user profile if query matches a username.
//   - Searches post titles and lists results if found.
//   - Redirects to the post if only one result, otherwise shows results list.
func NewSearchHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user session for personalized features.
		userID, username, loggedIn := middleware.CurrentUser(r)
		// Get and trim the search query from the URL.
		query := strings.TrimSpace(r.URL.Query().Get("q"))
		if query == "" {
			// If no query provided, redirect back to the home page.
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		// --- 1. Check for category name match (case-insensitive) ---
		var categoryID int
		err := dbConn.QueryRow("SELECT id FROM categories WHERE LOWER(name) = LOWER(?)", query).Scan(&categoryID)
		if err == nil {
			// Redirect to the matched category page.
			http.Redirect(w, r, fmt.Sprintf("/category/%d", categoryID), http.StatusSeeOther)
			return
		}

		// --- 2. Check for exact username match (case-insensitive) ---
		var usernameFound string
		err = dbConn.QueryRow("SELECT username FROM users WHERE LOWER(username) = LOWER(?)", query).Scan(&usernameFound)
		if err == nil {
			// Redirect to the user profile page.
			http.Redirect(w, r, "/u/"+usernameFound, http.StatusSeeOther)
			return
		}
		// --- 3. Search for posts by title (case-insensitive, partial match) ---
		rows, err := dbConn.Query(`
			SELECT p.id, p.user_id, p.category_id, p.title, p.content, COALESCE(p.image, ''), 
			       p.created_at, u.username, c.name,
			       COALESCE(l.value, 0)
			FROM posts p
			JOIN users u ON p.user_id = u.id
			JOIN categories c ON p.category_id = c.id
			LEFT JOIN post_likes l ON l.post_id = p.id AND l.user_id = ?
			WHERE LOWER(p.title) LIKE LOWER(?)
			ORDER BY p.created_at DESC
			LIMIT 50
		`, userID, "%"+query+"%")
		if err != nil {
			http.Error(w, "DB error", 500)
			return
		}
		defer rows.Close()

		var results []models.Post
		for rows.Next() {
			var post models.Post
			var createdAt time.Time
			err := rows.Scan(
				&post.ID, &post.UserID, &post.CategoryID, &post.Title, &post.Content,
				&post.Image, &createdAt, &post.Author, &post.Category,
				&post.UserLikeValue,
			)
			if err == nil {
				post.CreatedAt = createdAt
				post.Likes, _ = db.CountLikesByPostID(dbConn, post.ID)
				post.Comments, _ = db.CountCommentsByPostID(dbConn, post.ID)
				results = append(results, post)
			}
		}

		categories, _ := db.FetchCategories(dbConn)
		// --- 4. If exactly one post found, redirect directly to it. ---
		if len(results) == 1 {
			http.Redirect(w, r, fmt.Sprintf("/post/%d", results[0].ID), http.StatusSeeOther)
			return
		}

		// --- 5. Otherwise, render the search results page ---
		data := struct {
			models.BasePageData
			Query      string
			NotFound   bool
			Results    []models.Post
			Categories []models.Category
		}{
			BasePageData: models.BasePageData{
				Username: username,
				LoggedIn: loggedIn,
			},
			Query:      query,
			NotFound:   len(results) == 0,
			Results:    results,
			Categories: categories,
		}
		// Render the template with search results (or not found message)
		views.Templates.ExecuteTemplate(w, "search_results.html", data)
	}
}
