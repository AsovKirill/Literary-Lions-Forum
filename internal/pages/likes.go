package pages

import (
	"database/sql"
	"literary-lions/internal/middleware"
	"net/http"
	"strconv"
)

// NewLikeHandler returns a handler to process likes/dislikes for posts.
// Accepts POST requests with user session, post_id, and value (+1, -1, or 0).
// - 1: Like, -1: Dislike, 0: Remove existing like/dislike.
// Redirects back to the originating page.
func NewLikeHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST (disallow GET, etc.)
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// User must be logged in to like/dislike
		userID, _, loggedIn := middleware.CurrentUser(r)
		if !loggedIn || userID == 0 {
			http.Error(w, "You must be logged in to like/dislike", http.StatusUnauthorized)
			return
		}

		// Parse post_id and value from the form
		postID, err := strconv.Atoi(r.FormValue("post_id"))
		if err != nil {
			http.Error(w, "Invalid post_id", http.StatusBadRequest)
			return
		}
		value, err := strconv.Atoi(r.FormValue("value"))
		if err != nil || (value != 1 && value != -1 && value != 0) {
			http.Error(w, "Invalid like value", http.StatusBadRequest)
			return
		}

		if value == 0 {
			// Remove like/dislike (unlike/undislike)
			_, err = dbConn.Exec(`DELETE FROM post_likes WHERE post_id = ? AND user_id = ?`, postID, userID)
		} else {
			// Insert or update like/dislike (upsert)
			_, err = dbConn.Exec(`
        INSERT INTO post_likes (post_id, user_id, value)
        VALUES (?, ?, ?)
        ON CONFLICT(user_id, post_id) DO UPDATE SET value=excluded.value
    `, postID, userID, value)
		}
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Redirect back to the page user came from or home if not specified
		returnTo := r.FormValue("return_to")
		if returnTo == "" {
			returnTo = "/"
		}
		http.Redirect(w, r, returnTo, http.StatusSeeOther)
	}
}

// NewCommentLikeHandler handles likes/dislikes for comments.
// Only POST requests from logged-in users are allowed.
// - value: 1 (like), -1 (dislike), 0 (remove reaction).
// Redirects user back to the referring page (usually the post).
func NewCommentLikeHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// POST only
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// User must be authenticated
		userID, _, ok := middleware.CurrentUser(r)
		if !ok {
			http.Error(w, "unauthorised", http.StatusUnauthorized)
			return
		}

		// Parse comment_id and like/dislike value
		cID, err := strconv.Atoi(r.FormValue("comment_id"))
		if err != nil {
			http.Error(w, "bad id", 400)
			return
		}

		value, err := strconv.Atoi(r.FormValue("value")) // 1 or -1 or 0
		if err != nil || (value != 1 && value != -1 && value != 0) {
			http.Error(w, "invalid value", 400)
			return
		}

		if value == 0 {
			// Remove like/dislike
			_, err = dbConn.Exec(
				`DELETE FROM comment_likes WHERE comment_id=? AND user_id=?`,
				cID, userID)
		} else {
			// Insert or update like/dislike
			_, err = dbConn.Exec(`
			  INSERT INTO comment_likes(comment_id,user_id,value)
			  VALUES(?,?,?)
			  ON CONFLICT(user_id,comment_id)
			  DO UPDATE SET value=excluded.value`,
				cID, userID, value)
		}
		if err != nil {
			http.Error(w, "DB error", 500)
			return
		}

		// Redirect back to the referring page (where like was clicked)
		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
	}
}
