package pages

import (
	"database/sql"
	"literary-lions/internal/db"
	"literary-lions/internal/middleware"
	"net/http"
	"strconv"
)

// NewDeleteCommentHandler returns an HTTP handler for deleting comments.
// Only allows POST, only allows the owner of the comment to delete.
// Responds with appropriate HTTP error codes and messages.
func NewDeleteCommentHandler(dbConn *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
         // Only allow POST requests for this handler (security and REST best practice)
        if r.Method != http.MethodPost {
            http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
            return
        }
        // Get the current user's ID, username, and login status from the session/context.
        userID, username, loggedIn := middleware.CurrentUser(r)
        if !loggedIn {
            // If not logged in, redirect to the login page.
            http.Redirect(w, r, "/login", http.StatusSeeOther)
            return
        }
        // Parse the comment ID from the form value.
        commentIDStr := r.FormValue("comment_id")
        commentID, err := strconv.Atoi(commentIDStr)
        if err != nil {
            // Invalid or missing comment ID, show a 400 error page.
            middleware.ErrorHandler(w, 400, "Invalid comment ID", loggedIn, username)
            return
        }

        // Attempt to delete the comment, but only if it belongs to the current user.
        err = db.DeleteCommentByID(dbConn, commentID, userID)
        if err != nil {
            if err == sql.ErrNoRows {
                middleware.ErrorHandler(w, 403, "You are not allowed to delete this comment.", loggedIn, username)
            } else {
                middleware.ErrorHandler(w, 500, "Could not delete comment.", loggedIn, username)
            }
            return
        }
        // Redirect to home 
        http.Redirect(w, r, "/", http.StatusSeeOther)
    }
}
