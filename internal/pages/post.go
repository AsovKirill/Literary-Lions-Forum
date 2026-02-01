package pages

import (
	"database/sql"
	"fmt"
	"literary-lions/internal/db"
	"literary-lions/internal/middleware"
	"literary-lions/internal/models"
	"literary-lions/internal/views"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// NewShowPostHandler serves the single post page and handles new comment submissions.
// - GET: Show the post, its comments, likes, etc.
// - POST: Add a new comment (only if logged in), then redirect back to the same post page.
func NewShowPostHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract the post ID from the URL path (e.g. /post/5)
		idStr := strings.TrimPrefix(r.URL.Path, "/post/")
		postID, _ := strconv.Atoi(idStr)

		// If the method is POST, process a new comment
		if r.Method == http.MethodPost {
			userID, _, ok := middleware.CurrentUser(r)
			if ok {
				if text := strings.TrimSpace(r.FormValue("comment")); text != "" {
					_ = db.AddComment(dbConn, postID, userID, text)
				}
			}
			// Always redirect back to the post after comment submission
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
			return
		}

		// Fetch the post by ID, using viewer info for like status, etc.
		viewerID, username, logged := middleware.CurrentUser(r)

		post, err := db.GetPost(dbConn, postID, viewerID)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		// Fetch the comments for the post, with per-user like info
		comments, _ := db.GetComments(dbConn, postID, viewerID)
		fmt.Printf("Found %d comments for post %d\n", len(comments), postID)
		// Fetch categories for sidebar
		categories, _ := db.FetchCategories(dbConn)

		// Recalculate likes/comments for accurate display
		likes, _ := db.CountLikesByPostID(dbConn, postID)
		post.Likes = likes
		post.Comments = len(comments)

		// Prepare the template data for rendering the post page
		data := struct {
			Post          models.Post
			Comments      []models.Comment
			LoggedIn      bool
			Username      string
			Categories    []models.Category
			CurrentUserID int
		}{
			Post:          post,
			Comments:      comments,
			LoggedIn:      logged,
			Username:      username,
			Categories:    categories,
			CurrentUserID: viewerID,
		}

		// Render the post.html template with the data
		err = views.Templates.ExecuteTemplate(w, "post.html", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// NewDeletePostHandler handles POST requests to delete a post.
// - Only the post owner can delete their post.
// - On success, redirects to home page. Otherwise, shows error or asks login.
func NewDeletePostHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Only POST allowed (for CSRF safety and RESTfulness)
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		// Must be logged in
		userID, _, ok := middleware.CurrentUser(r)
		if !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		// Parse post ID from the form
		postIDStr := r.FormValue("post_id")
		postID, err := strconv.Atoi(postIDStr)
		if err != nil {
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}
		// Try to delete the post (only if it belongs to the user)
		err = db.DeletePostByID(dbConn, postID, userID)
		if err != nil {
			// Not found or not the owner
			if err == sql.ErrNoRows {
				http.Error(w, "Post not found or not yours", http.StatusForbidden)
			} else {
				http.Error(w, "Could not delete post: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}
		log.Printf("User %d attempts to delete post %d", userID, postID)
		// On success, redirect to home page
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
