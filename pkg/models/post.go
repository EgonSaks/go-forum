package models

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"forum/logger"
)

type Post struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	ImageFullPath string    `json:"image_url"`
	Category      string    `json:"category"`
	CreatedAt     time.Time `json:"created_at"`
	User          User      `json:"user"`
	Comments      []Comment `json:"comments"`
	CommentsCount int
	Likes         int
	Dislikes      int
}

func CreatePost(db *sql.DB, post Post) (string, error) {
	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "INSERT INTO posts (id, user_id, title, content, image_url, category, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)"
	statement, err := db.PrepareContext(context, query)
	if err != nil {
		logger.ErrorLogger.Printf("failed to prepare create post statement: %v", err)
		return post.ID, fmt.Errorf("failed to prepare create post statement: %v", err)
	}

	_, err = statement.ExecContext(context, &post.ID, &post.UserID, &post.Title, &post.Content, &post.ImageFullPath, &post.Category, &post.CreatedAt)
	if err != nil {
		logger.ErrorLogger.Printf("failed to create post: %v", err)
		return post.ID, fmt.Errorf("failed to create post: %v", err)
	}

	return post.ID, nil
}

func GetAllPosts(db *sql.DB) ([]Post, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "SELECT id, user_id, title, content, image_url, category, created_at FROM posts ORDER BY created_at DESC"
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		logger.ErrorLogger.Printf("failed to execute get all posts query: %v", err)
		return nil, fmt.Errorf("failed to execute get all posts query: %v", err)
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		err := rows.Scan(&post.ID, &post.UserID, &post.Title, &post.Content, &post.ImageFullPath, &post.Category, &post.CreatedAt)
		if err != nil {
			logger.ErrorLogger.Printf("failed to scan posts row: %v", err)
			return nil, fmt.Errorf("failed to scan posts row: %v", err)
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		logger.ErrorLogger.Printf("failed to iterate over rows to get posts: %v", err)
		return nil, fmt.Errorf("failed to iterate over rows to get posts: %v", err)
	}

	return posts, nil
}

func GetPostByID(db *sql.DB, id string) (Post, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var post Post

	query := `
        SELECT id, user_id, title, content, image_url, category, created_at
        FROM posts
        WHERE id = ?
        LIMIT 1
        `
	err := db.QueryRowContext(ctx, query, id).Scan(&post.ID, &post.UserID, &post.Title, &post.Content, &post.ImageFullPath, &post.Category, &post.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.ErrorLogger.Printf("no post found with ID %s", id)
			return Post{}, fmt.Errorf("no post found with ID %s", id)
		}
		logger.ErrorLogger.Printf("failed to get post: %v", err)
		return Post{}, fmt.Errorf("failed to get post: %v", err)
	}

	return post, nil
}

func GetAllPostsByUserID(db *sql.DB, userID string) ([]Post, error) {
	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "SELECT id, user_id, title, content, image_url, category, created_at FROM posts WHERE user_id=? ORDER BY created_at DESC"

	logger.InfoLogger.Printf("GetAllPostsByUserID query: %v", query)

	rows, err := db.QueryContext(context, query, userID)
	if err != nil {
		logger.ErrorLogger.Printf("Failed to execute get all posts query: %v", err)
		return nil, fmt.Errorf("failed to execute get all posts query: %v", err)
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		err := rows.Scan(&post.ID, &post.UserID, &post.Title, &post.Content, &post.ImageFullPath, &post.Category, &post.CreatedAt)
		if err != nil {
			logger.ErrorLogger.Printf("Failed to scan posts row: %v", err)
			return nil, fmt.Errorf("failed to scan posts row: %v", err)
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		logger.ErrorLogger.Printf("Failed to iterate over rows to get posts: %v", err)
		return nil, fmt.Errorf("failed to iterate over rows to get posts: %v", err)
	}

	return posts, nil
}

func GetUserLikedPostsByPostIDAndUserID(db *sql.DB, postID string, userID string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var count int

	query := `SELECT COUNT(*) FROM post_reactions WHERE post_id = ? AND reaction_type = 'like' AND user_id = ?`
	err := db.QueryRowContext(ctx, query, postID, userID).Scan(&count)
	if err != nil {
		logger.ErrorLogger.Printf("failed to get liked posts count: %v", err)
		return 0, fmt.Errorf("failed to get liked posts count: %v", err)
	}

	return count, nil
}

func GetUserDislikedPostsByPostIDAndUserID(db *sql.DB, postID string, userID string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var count int

	query := `SELECT COUNT(*) FROM post_reactions WHERE post_id = ? AND reaction_type = 'dislike' AND user_id = ?`
	err := db.QueryRowContext(ctx, query, postID, userID).Scan(&count)
	if err != nil {
		logger.ErrorLogger.Printf("failed to get disliked posts count: %v", err)
		return 0, fmt.Errorf("failed to get disliked posts count: %v", err)
	}

	return count, nil
}

func GetPostTitleByCommentID(db *sql.DB, commentID string) (string, error) {
	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `SELECT p.title
    FROM posts p 
    INNER JOIN comments c ON p.id = c.post_id
    WHERE c.id = ?`

	row := db.QueryRowContext(context, query, commentID)

	var postTitle string
	err := row.Scan(&postTitle)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.ErrorLogger.Printf("No post found for comment ID %s", commentID)
			return postTitle, fmt.Errorf("no post found for comment ID %s", commentID)
		}
		logger.ErrorLogger.Printf("Failed to get post title for comment ID %s: %v", commentID, err)
		return postTitle, fmt.Errorf("failed to get post title for comment ID %s: %v", commentID, err)
	}

	return postTitle, nil
}
