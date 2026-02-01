package pages

import (
	"database/sql"
	"io"
	"literary-lions/internal/middleware"
	"literary-lions/internal/models"
	"literary-lions/internal/views"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// CreatePostPageData holds data required to render the "Create Post" page.
type CreatePostPageData struct {
	Username   string
	LoggedIn   bool
	Categories []models.Category
}

// NewCreatePostHandler returns an HTTP handler for the /createpost route.
// It handles both GET (render the form) and POST (process post creation).
func NewCreatePostHandler(dbConn *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Require login for both GET and POST!
        userID, username, loggedIn := middleware.CurrentUser(r)
        if !loggedIn {
			// If not logged in, redirect to login page.
            http.Redirect(w, r, "/login", http.StatusSeeOther)
            return
        }

		// Handle GET request: render the post creation form.
        if r.Method == http.MethodGet {
			// Fetch all categories for the category dropdown.
            rows, err := dbConn.Query("SELECT id, name FROM categories")
            var categories []models.Category
            if err == nil {
                defer rows.Close()
                for rows.Next() {
                    var c models.Category
                    rows.Scan(&c.ID, &c.Name)
                    categories = append(categories, c)
                }
            }

			// Render the form with username, login status, and categories.
            data := CreatePostPageData{
                Username:   username,
                LoggedIn:   loggedIn,
                Categories: categories,
            }
            views.Templates.ExecuteTemplate(w, "createpost.html", data)
            return
        }

		// Handle POST request: process submitted form and create a new post.
        if r.Method == http.MethodPost {
			// Read form values.
            title := r.FormValue("title")
            content := r.FormValue("content")
            categoryID, _ := strconv.Atoi(r.FormValue("category_id"))

			// Handle optional image upload.
            var imagePath string
            file, handler, err := r.FormFile("image")
            if err == nil && handler != nil {
                defer file.Close()
                os.MkdirAll("web/static/uploads", os.ModePerm)
                ext := filepath.Ext(handler.Filename)
                imageFileName := "post_" + strconv.FormatInt(time.Now().UnixNano(), 10) + ext
                imagePath = "/static/uploads/" + imageFileName
                dst, err := os.Create("web/static/uploads/" + imageFileName)
                if err == nil {
                    defer dst.Close()
                    io.Copy(dst, file)
                }
            }

			// Insert the new post into the database.
            _, err = dbConn.Exec(`
                INSERT INTO posts (user_id, category_id, title, content, image)
                VALUES (?, ?, ?, ?, ?)`,
                userID, categoryID, title, content, imagePath,
            )
            if err != nil {
				// If there is a DB error, show a friendly error page.
				_, username, loggedIn := middleware.CurrentUser(r)
				middleware.ErrorHandler(w, 500, "Error when creating post: "+err.Error(), loggedIn, username)
				return
            }

			// On success, redirect to the home page (or you can redirect to the new post).
            http.Redirect(w, r, "/", http.StatusSeeOther)
            return
        }

        // If the method is not GET or POST, respond with 405 Method Not Allowed.
        http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
    }
}
