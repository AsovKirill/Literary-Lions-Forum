package main

import (
	"log"
	"net/http"

	"literary-lions/internal/auth"
	"literary-lions/internal/db"
	"literary-lions/internal/middleware"
	"literary-lions/internal/pages"
	"literary-lions/internal/views"
)

func main() {
	// Open (or create) the SQLite database file
	dbConn, err := db.InitDB("forum.db")
	if err != nil {
		log.Fatal(err)
	}

	// Parse and cache all HTML templates for rendering pages
	views.InitTemplates()

	// Create the main HTTP request router (ServeMux)
	mux := http.NewServeMux()

	// Serve static files (CSS, JS, images) from /static/ URL
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// Page routes:
	mux.HandleFunc("/about", pages.AboutPageHandler(dbConn))
	mux.HandleFunc("/terms", pages.TermsPageHandler(dbConn))
	mux.HandleFunc("/signup", auth.NewSignupHandler(dbConn))
	mux.HandleFunc("/login", auth.NewLoginHandler(dbConn))
	mux.HandleFunc("/logout", auth.NewLogoutHandler(dbConn))
	mux.HandleFunc("/u/", pages.NewProfileHandler(dbConn))
	mux.HandleFunc("/profile", pages.NewMyProfileHandler(dbConn))
	mux.HandleFunc("/createpost", pages.NewCreatePostHandler(dbConn))
	mux.HandleFunc("/like", pages.NewLikeHandler(dbConn))
	mux.HandleFunc("/deletecomment", pages.NewDeleteCommentHandler(dbConn))
	mux.HandleFunc("/deletepost", pages.NewDeletePostHandler(dbConn))
	mux.Handle("/category/", pages.NewCategoryHandler(dbConn))
	mux.HandleFunc("/post/", pages.NewShowPostHandler(dbConn))
	mux.HandleFunc("/comment-like", pages.NewCommentLikeHandler(dbConn))
	mux.HandleFunc("/search", pages.NewSearchHandler(dbConn))

	// fallback for "/" and 404s
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Home page for "/"
		if r.URL.Path == "/" {
			pages.NewHomeHandler(dbConn)(w, r)
			return
		}
		// Everything else = 404
		username, loggedIn := "", false
		if _, n, ok := middleware.CurrentUser(r); ok {
			username = n
			loggedIn = true
		}
		middleware.ErrorHandler(w, http.StatusNotFound, "Page not found", loggedIn, username)
	})

	// Compose middleware: session management + panic recovery
	handler := middleware.RecoverMiddleware(middleware.NewWithSession(dbConn)(mux))

	log.Println("Server starting at http://localhost:8080...")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
