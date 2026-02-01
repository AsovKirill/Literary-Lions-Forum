package models

import "time"

type BasePageData struct {
	Username string
	LoggedIn bool
}

type HomePageData struct {
	BasePageData
	Posts        []Post
	Categories   []Category
	PopularPosts []Post
}

type User struct {
	ID        int
	Username  string
	Email     string
	Password  string
	CreatedAt time.Time
}

type Post struct {
	ID            int
	UserID        int
	CategoryID    int
	Title         string
	Content       string
	Image         string
	CreatedAt     time.Time
	Likes         int
	Comments      int
	Author        string
	Category      string
	UserLikeValue int
}

type Comment struct {
	ID        int
	PostID    int
	UserID    int
	Content   string
	CreatedAt time.Time

	Author        string
	Likes         int // total # of likes
	UserLikeValue int
}

type Category struct {
	ID   int
	Name string
}

type ErrorData struct {
	Status   int
	Message  string
	LoggedIn bool
	Username string
}
type SearchResultsPageData struct {
	BasePageData
	Query      string
	NotFound   bool
	Results    []Post
	Categories []Category
}
