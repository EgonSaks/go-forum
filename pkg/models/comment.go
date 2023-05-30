package models

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"forum/logger"
)

type Comment struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	PostID        string    `json:"post_id"`
	Content       string    `json:"content"`
	CreatedAt     time.Time `json:"created_at"`
	User          User      `json:"user"`
	Post          Post      `json:"post"`
	PostTitle     string    `json:"post_title"`
	CommentsCount int       `json:"comments_count"`
	IsLoggedIn    bool
	LoggedInUser  User
	Likes         int
	Dislikes      int
}

func CreateComment(db *sql.DB, comment Comment) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "INSERT INTO comments (id, user_id, post_id, content, created_at) VALUES (?, ?, ?, ?, ?)"
	statement, err := db.PrepareContext(ctx, query)
	if err != nil {
		logger.ErrorLogger.Printf("failed to prepare create comment statement: %v", err)
		return comment.ID, fmt.Errorf("failed to prepare create comment statement: %v", err)
	}

	_, err = statement.ExecContext(ctx, &comment.ID, &comment.UserID, &comment.PostID, &comment.Content, &comment.CreatedAt)
	if err != nil {
		logger.ErrorLogger.Printf("failed to create comment: %v", err)
		return comment.ID, fmt.Errorf("failed to create comment: %v", err)
	}

	return comment.ID, nil
}

func GetAllCommentsByPostID(db *sql.DB, postID string) ([]Comment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var comments []Comment

	query := `
		SELECT comments.id, comments.user_id, comments.post_id, comments.content , comments.created_at, users.id, users.name, users.email,  users.created_at, posts.id, posts.user_id, posts.title, posts.content, posts.created_at
		FROM comments
		JOIN users ON comments.user_id = users.id
		JOIN posts ON comments.post_id = posts.id 
		WHERE comments.post_id = ? 
		ORDER BY comments.created_at DESC
	`
	rows, err := db.QueryContext(ctx, query, postID)
	if err != nil {
		logger.ErrorLogger.Printf("Failed to get comments: %v", err)
		return nil, fmt.Errorf("failed to get comments: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var comment Comment
		var user User
		var post Post

		err := rows.Scan(&comment.ID, &comment.UserID, &comment.PostID, &comment.Content, &comment.CreatedAt, &user.ID, &user.Name, &user.Email, &user.CreatedAt, &post.ID, &post.UserID, &post.Title, &post.Content, &post.CreatedAt)
		if err != nil {
			logger.ErrorLogger.Printf("Failed to scan comment: %v", err)
			return nil, fmt.Errorf("failed to scan comment: %v", err)
		}

		comment.User = user
		comment.Post = post
		comments = append(comments, comment)
	}

	if len(comments) == 0 {
		logger.InfoLogger.Printf("No comments found for post ID %s", postID)
	}

	return comments, nil
}

func CommentCountByPostID(db *sql.DB, postID string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var count int

	query := `SELECT COUNT(*) FROM comments WHERE post_id = ?`
	err := db.QueryRowContext(ctx, query, postID).Scan(&count)
	if err != nil {
		logger.ErrorLogger.Printf("Failed to get comment count: %v", err)
		return 0, fmt.Errorf("failed to get comment count: %v", err)
	}

	return count, nil
}

func GetAllCommentsByUserID(db *sql.DB, userID string) ([]Comment, error) {
	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	logger.InfoLogger.Println("Getting all comments for user:", userID)

	var comments []Comment

	query := `
        SELECT comments.id, comments.user_id, comments.post_id, comments.content, comments.created_at, users.id, users.name, users.email, users.created_at, posts.id, posts.user_id, posts.title, posts.content, posts.created_at
        FROM comments
        JOIN users ON comments.user_id = users.id
        JOIN posts ON comments.post_id = posts.id
        WHERE comments.user_id = ?
        ORDER BY comments.created_at DESC
    `
	rows, err := db.QueryContext(context, query, userID)
	if err != nil {
		logger.ErrorLogger.Println("Failed to get comments:", err)
		return nil, fmt.Errorf("failed to get comments: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var comment Comment
		var user User
		var post Post

		err := rows.Scan(&comment.ID, &comment.UserID, &comment.PostID, &comment.Content, &comment.CreatedAt, &user.ID, &user.Name, &user.Email, &user.CreatedAt, &post.ID, &post.UserID, &post.Title, &post.Content, &post.CreatedAt)
		if err != nil {
			logger.ErrorLogger.Println("Failed to scan comment:", err)
			return nil, fmt.Errorf("failed to scan comment: %v", err)
		}

		comment.User = user
		comment.Post = post
		comments = append(comments, comment)
	}

	return comments, nil
}

func GetAllComments(db *sql.DB) ([]Comment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "SELECT * FROM comments"
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		logger.ErrorLogger.Printf("Failed to get comments: %v", err)
		return nil, fmt.Errorf("failed to get comments: %v", err)
	}
	defer rows.Close()

	comments := []Comment{}
	for rows.Next() {
		comment := Comment{}
		err = rows.Scan(&comment.ID, &comment.UserID, &comment.PostID, &comment.Content, &comment.CreatedAt)
		if err != nil {
			logger.ErrorLogger.Printf("Failed to scan comment: %v", err)
			return nil, fmt.Errorf("failed to scan comment: %v", err)
		}
		comments = append(comments, comment)
	}

	if err = rows.Err(); err != nil {
		logger.ErrorLogger.Printf("Failed to get comments: %v", err)
		return nil, fmt.Errorf("failed to get comments: %v", err)
	}

	return comments, nil
}

func GetUserLikedCommentByCommentIDAndUserID(db *sql.DB, commentID string, userID string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var count int

	query := `SELECT COUNT(*) FROM comment_reactions WHERE comment_id = ? AND reaction_type = 'like' AND user_id = ?`
	err := db.QueryRowContext(ctx, query, commentID, userID).Scan(&count)
	if err != nil {
		logger.ErrorLogger.Printf("Failed to get liked comments count: %v", err)
		return 0, fmt.Errorf("failed to get liked comments count: %v", err)
	}

	return count, nil
}

func GetUserDislikedCommentByCommentIDAndUserID(db *sql.DB, commentID string, userID string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var count int

	query := `SELECT COUNT(*) FROM comment_reactions WHERE comment_id = ? AND reaction_type = 'dislike' AND user_id = ?`
	err := db.QueryRowContext(ctx, query, commentID, userID).Scan(&count)
	if err != nil {
		logger.ErrorLogger.Printf("Failed to get disliked comments count: %v", err)
		return 0, fmt.Errorf("failed to get disliked comments count: %v", err)
	}
	return count, nil
}
