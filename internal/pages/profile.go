package pages

import (
	"database/sql"
	"literary-lions/internal/db"
	"literary-lions/internal/middleware"
	"literary-lions/internal/models"
	"literary-lions/internal/views"
	"net/http"
	"strings"
	"time"
)

// ProfilePageData structures the data needed for rendering a user profile page.
type ProfilePageData struct {
	models.BasePageData
	MyPosts    []models.Post // Posts authored by this user
	MyLikes    []models.Post // Posts this user has liked
	Categories []models.Category // All available categories
}

// NewProfileHandler serves a user's public profile at /u/{username}.
// It shows their posts, their liked posts, and categories.
// If the username does not exist, a 404 is returned.
func NewProfileHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get session info for highlighting or display (who's viewing)
		_, sessionUsername, loggedIn := middleware.CurrentUser(r)

		// Parse the URL to extract the profile username (/u/{username})
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) != 2 || parts[0] != "u" {
			http.NotFound(w, r)
			return
		}
		profileUsername := parts[1]

		// Look up the user ID for the given username
		var userID int
		if err := dbConn.QueryRow(
			"SELECT id FROM users WHERE username = ?", profileUsername).
			Scan(&userID); err != nil {
			http.NotFound(w, r)
			return
		}

		// --- Fetch posts authored by this user ---
		postRows, err := dbConn.Query(`
			SELECT
				p.id, p.user_id, p.category_id,
				p.title, p.content, COALESCE(p.image,''),
				p.created_at,
				u.username, c.name,
				(SELECT COUNT(*) FROM post_likes WHERE post_id=p.id AND value=1) AS likes,
				(SELECT COUNT(*) FROM comments   WHERE post_id=p.id)            AS comments
			FROM posts p
			JOIN users      u ON u.id = p.user_id
			JOIN categories c ON c.id = p.category_id
			WHERE p.user_id = ?
			ORDER BY p.created_at DESC;
		`, userID)
		if err != nil {
			http.Error(w, "Error loading posts", 500)
			return
		}
		defer postRows.Close()

		var myPosts []models.Post
		for postRows.Next() {
			var (
				p           models.Post
				createdText string 
			)
			if err := postRows.Scan(
				&p.ID, &p.UserID, &p.CategoryID,
				&p.Title, &p.Content, &p.Image,
				&p.CreatedAt,
				&p.Author, &p.Category,
				&p.Likes, &p.Comments,
			); err != nil {
				continue
			}
			if t, err := time.Parse("2006-01-02 15:04:05", createdText); err == nil {
				p.CreatedAt = t
			}
			myPosts = append(myPosts, p)
		}

		// --- Fetch posts liked by this user ---
		likeRows, err := dbConn.Query(`
			SELECT
				p.id, p.user_id, p.category_id,
				p.title, p.content, COALESCE(p.image,''),
				p.created_at,
				u.username, c.name,
				(SELECT COUNT(*) FROM post_likes WHERE post_id=p.id AND value=1) AS likes,
				(SELECT COUNT(*) FROM comments   WHERE post_id=p.id)            AS comments
			FROM post_likes l
			JOIN posts      p ON p.id = l.post_id
			JOIN users      u ON u.id = p.user_id
			JOIN categories c ON c.id = p.category_id
			WHERE l.user_id = ? AND l.value = 1
			ORDER BY p.created_at DESC;
		`, userID)
		if err != nil {
			http.Error(w, "Error loading liked posts", 500)
			return
		}
		defer likeRows.Close()

		var myLikes []models.Post
		for likeRows.Next() {
			var (
				p           models.Post
				createdText string
			)
			if err := likeRows.Scan(
				&p.ID, &p.UserID, &p.CategoryID,
				&p.Title, &p.Content, &p.Image,
				&p.CreatedAt,
				&p.Author, &p.Category,
				&p.Likes, &p.Comments,
			); err != nil {
				continue
			}
			if t, err := time.Parse("2006-01-02 15:04:05", createdText); err == nil {
				p.CreatedAt = t
			}
			myLikes = append(myLikes, p)
		}

		// --- Fetch all categories for sidebar ---
		categories, _ := db.FetchCategories(dbConn)
		// --- Prepare and render the profile page ---
		data := ProfilePageData{
			BasePageData: models.BasePageData{
				Username: sessionUsername,
				LoggedIn: loggedIn,
			},
			MyPosts:    myPosts,
			MyLikes:    myLikes,
			Categories: categories,
		}
		views.Templates.ExecuteTemplate(w, "profile.html", data)
	}
}

// NewMyProfileHandler redirects the logged-in user from /profile to their /u/{username} page.
// This lets users easily see their own profile without needing to type the URL.
func NewMyProfileHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, username, loggedIn := middleware.CurrentUser(r)
		if !loggedIn || username == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, "/u/"+username, http.StatusSeeOther)
	}
}
