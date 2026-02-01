package db

import (
	"database/sql"
	"literary-lions/internal/models"
	"time"
)

// FetchPosts fetches the latest 20 posts, each with its author, category,
// and the requesting user's like/dislike value for each post (if any)
// Returns a slice of Post structs, or an error
func FetchPosts(db *sql.DB, userID int) ([]models.Post, error) {
	rows, err := db.Query(`
		SELECT 
			p.id, p.user_id, p.category_id, p.title, p.content, 
			COALESCE(p.image, ''),    
			p.created_at, 
			u.username, c.name,
			COALESCE(l.value, 0) -- like/dislike for this user, 0 if none
		FROM posts p
		JOIN users u ON p.user_id = u.id
		JOIN categories c ON p.category_id = c.id
		LEFT JOIN post_likes l ON l.post_id = p.id AND l.user_id = ?
		ORDER BY p.created_at DESC
		LIMIT 20
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		var createdAt string
		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.CategoryID,
			&post.Title,
			&post.Content,
			&post.Image,
			&createdAt,
			&post.Author,
			&post.Category,
			&post.UserLikeValue,
		)
		if err != nil {
			return nil, err
		}
		// Robust handling of date parsing (support for legacy/alternative formats)
		post.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAt)
		if err != nil {
			post.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
			if err != nil {
				post.CreatedAt = time.Now()
			}
		}
		posts = append(posts, post)
	}
	return posts, nil
}

// CountLikesByPostID returns the total number of likes (value = 1)
// for a given post ID. Returns 0 if none, or an error if the query fails
func CountLikesByPostID(db *sql.DB, postID int) (int, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM post_likes WHERE post_id = ? AND value = 1`, postID).Scan(&count)
	return count, err
}

// CountCommentsByPostID returns the total number of comments for a given post
func CountCommentsByPostID(db *sql.DB, postID int) (int, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM comments WHERE post_id = ?`, postID).Scan(&count)
	return count, err
}

// GetPost returns a Post struct by postID, along with the current user's like/dislike value.
// Uses both classic and RFC3339 date formats for robustness.
func GetPost(db *sql.DB, postID int, viewerID int) (models.Post, error) {
	var p models.Post
	var createdAt string
	err := db.QueryRow(`
        SELECT 
			p.id, 
			p.user_id, 
			p.category_id, 
			p.title, 
			p.content,
            COALESCE(p.image, ''), 
			p.created_at, 
			u.username, 
			c.name,
            COALESCE((SELECT value FROM post_likes WHERE post_id = p.id AND user_id = ?), 0)
        FROM posts p
        JOIN users u ON p.user_id = u.id
        JOIN categories c ON p.category_id = c.id
        WHERE p.id = ?`,
		viewerID, postID).
		Scan(
			&p.ID,
			&p.UserID,
			&p.CategoryID,
			&p.Title,
			&p.Content,
			&p.Image,
			&createdAt,
			&p.Author,
			&p.Category,
			&p.UserLikeValue,
		)
	if err != nil {
		return p, err
	}

	// Try to parse created_at using multiple formats; fallback to now if all fail
	p.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAt)
	if err != nil {
		p.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
		if err != nil {
			p.CreatedAt = time.Now() // fallback — чтоб не было мусора
		}
	}
	return p, nil
}

// GetComments returns all comments for a given postID, including the author's name,
// total likes for each comment, and the current user's like/dislike value for each comment
func GetComments(db *sql.DB, postID int, viewerID int) ([]models.Comment, error) {
	rows, err := db.Query(`
		SELECT
			c.id,
			c.post_id,
			c.user_id,
			u.username,                               -- author
			c.content,
			c.created_at,
			/* total likes for this comment */
			(SELECT COUNT(*) FROM comment_likes
			  WHERE comment_id = c.id AND value = 1)  AS likes,
			/* does the *viewer* like it?  1 / -1 / 0 */
			COALESCE((
				SELECT value FROM comment_likes
				WHERE comment_id = c.id AND user_id = ?
			), 0)                                       AS viewer_value
		FROM comments c
		JOIN users u ON u.id = c.user_id
		WHERE c.post_id = ?
		ORDER BY c.created_at DESC;
	`, viewerID, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.Comment
	for rows.Next() {
		var cm models.Comment
		if err := rows.Scan(
			&cm.ID,
			&cm.PostID,
			&cm.UserID,
			&cm.Author,
			&cm.Content,
			&cm.CreatedAt,
			&cm.Likes,         // ← add these fields in models.Comment
			&cm.UserLikeValue, // ← 1 / -1 / 0
		); err != nil {
			// Skip any broken row, but keep processing
			continue
		}
		list = append(list, cm)
	}
	return list, nil
}

// AddComment inserts a new comment for a given postID and userID with the given text.
// Returns an error if the insert fails
func AddComment(db *sql.DB, postID, userID int, text string) error {
	_, err := db.Exec(`INSERT INTO comments(post_id,user_id,content)
                       VALUES(?,?,?)`, postID, userID, text)
	return err
}

// FetchPostsByCategory fetches up to 50 posts for the given categoryID.
// Returns posts in newest-first order, or an error.
func FetchPostsByCategory(db *sql.DB, categoryID int) ([]models.Post, error) {
	rows, err := db.Query(`
        SELECT p.id, p.user_id, p.category_id, p.title, p.content, COALESCE(p.image,''), p.created_at, u.username, c.name
        FROM posts p
        JOIN users u ON p.user_id = u.id
        JOIN categories c ON p.category_id = c.id
        WHERE p.category_id = ?
        ORDER BY p.created_at DESC
        LIMIT 50
    `, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []models.Post
	for rows.Next() {
		var post models.Post
		var createdAt time.Time
		_ = rows.Scan(&post.ID, &post.UserID, &post.CategoryID, &post.Title, &post.Content, &post.Image, &createdAt, &post.Author, &post.Category)
		post.CreatedAt = createdAt
		posts = append(posts, post)
	}
	return posts, nil
}

// FetchCategories returns all categories sorted alphabetically.
func FetchCategories(db *sql.DB) ([]models.Category, error) {
	rows, err := db.Query("SELECT id, name FROM categories ORDER BY name ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cats []models.Category
	for rows.Next() {
		var c models.Category
		rows.Scan(&c.ID, &c.Name)
		cats = append(cats, c)
	}
	return cats, nil
}

// FetchPopularPosts returns the most-liked posts up to the specified limit.
// Each post includes its author and category names.
func FetchPopularPosts(db *sql.DB, limit int) ([]models.Post, error) {
	rows, err := db.Query(`
		SELECT p.id, p.user_id, p.category_id, p.title, p.content, p.image, p.created_at, u.username, c.name, 
		       IFNULL(SUM(pl.value), 0) as likes
		  FROM posts p
		  JOIN users u ON p.user_id = u.id
		  JOIN categories c ON p.category_id = c.id
		  LEFT JOIN post_likes pl ON p.id = pl.post_id
		 GROUP BY p.id
		 ORDER BY likes DESC, p.created_at DESC
		 LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var p models.Post
		var likes int
		err = rows.Scan(&p.ID, &p.UserID, &p.CategoryID, &p.Title, &p.Content, &p.Image, &p.CreatedAt, &p.Author, &p.Category, &likes)
		if err != nil {
			return nil, err
		}
		p.Likes = likes
		posts = append(posts, p)
	}
	return posts, nil
}

// DeleteCommentByID deletes a comment by its ID, only if it belongs to the given user.
// Returns sql.ErrNoRows if no such comment or not the owner.
func DeleteCommentByID(db *sql.DB, commentID, userID int) error {
	res, err := db.Exec(`DELETE FROM comments WHERE id = ? AND user_id = ?`, commentID, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows // not found or not owner
	}
	return nil
}

// DeletePostByID deletes a post by its ID, only if it belongs to the given user.
// Returns sql.ErrNoRows if no such post or not the owner.
func DeletePostByID(db *sql.DB, postID, userID int) error {
	res, err := db.Exec(`DELETE FROM posts WHERE id = ? AND user_id = ?`, postID, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows // Not found or not owner
	}
	return nil
}
