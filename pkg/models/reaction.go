package models

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"forum/logger"
)

type PostReaction struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	PostID       string    `json:"post_id"`
	ReactionType string    `json:"reaction_type"`
	CreatedAt    time.Time `json:"created_at"`
}

type CommentReaction struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	PostID       string    `json:"post_id"`
	CommentID    string    `json:"comment_id"`
	ReactionType string    `json:"reaction_type"`
	CreatedAt    time.Time `json:"created_at"`
}

func CreatePostReaction(db *sql.DB, reaction PostReaction) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "INSERT INTO post_reactions (id, user_id, post_id, reaction_type, created_at) VALUES (?, ?, ?, ?, ?)"
	statement, err := db.PrepareContext(ctx, query)
	if err != nil {
		logger.ErrorLogger.Printf("Failed to prepare create post reaction statement: %v\n", err)
		return reaction.ID, fmt.Errorf("failed to prepare create post reaction statement: %v", err)
	}

	_, err = statement.ExecContext(ctx, &reaction.ID, &reaction.UserID, &reaction.PostID, &reaction.ReactionType, &reaction.CreatedAt)
	if err != nil {
		logger.ErrorLogger.Printf("Failed to create post reaction: %v\n", err)
		return reaction.ID, fmt.Errorf("failed to create post reaction: %v", err)
	}

	return reaction.ID, nil
}

func PostLikeCountByPostID(db *sql.DB, postID string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var count int

	query := `SELECT COUNT(*) FROM post_reactions WHERE post_id = ? AND reaction_type = 'like'`
	err := db.QueryRowContext(ctx, query, postID).Scan(&count)
	if err != nil {
		logger.ErrorLogger.Printf("Failed to get like count: %v\n", err)
		return 0, fmt.Errorf("failed to get like count: %v", err)
	}

	return count, nil
}

func PostDislikeCountByPostID(db *sql.DB, postID string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var count int

	query := `SELECT COUNT(*) FROM post_reactions WHERE post_id = ? AND reaction_type = 'dislike'`
	err := db.QueryRowContext(ctx, query, postID).Scan(&count)
	if err != nil {
		logger.ErrorLogger.Printf("Failed to get like count: %v", err)
		return 0, fmt.Errorf("failed to get like count: %v", err)
	}

	return count, nil
}

func CreateCommentReaction(db *sql.DB, reaction CommentReaction) (string, error) {
	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "INSERT INTO comment_reactions (id, user_id, post_id, comment_id, reaction_type, created_at) VALUES (?, ?, ?, ?, ?, ?)"
	statement, err := db.PrepareContext(context, query)
	if err != nil {
		logger.ErrorLogger.Printf("Failed to prepare create comment reaction statement: %v", err)
		return reaction.ID, fmt.Errorf("failed to prepare create comment reaction statement: %v", err)
	}

	_, err = statement.ExecContext(context, &reaction.ID, &reaction.UserID, &reaction.PostID, &reaction.CommentID, &reaction.ReactionType, &reaction.CreatedAt)
	if err != nil {
		logger.ErrorLogger.Printf("Failed to create comment reaction: %v", err)
		return reaction.ID, fmt.Errorf("failed to create comment reaction: %v", err)
	}

	return reaction.ID, nil
}

func CommentLikeCountByCommentID(db *sql.DB, commentID string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var count int

	query := `SELECT COUNT(*) FROM comment_reactions WHERE comment_id = ? AND reaction_type = 'like'`
	err := db.QueryRowContext(ctx, query, commentID).Scan(&count)
	if err != nil {
		logger.ErrorLogger.Printf("failed to get like count: %v", err)
		return 0, fmt.Errorf("failed to get like count: %v", err)
	}

	return count, nil
}

func CommentDislikeCountByCommentID(db *sql.DB, commentID string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var count int

	query := `SELECT COUNT(*) FROM comment_reactions WHERE comment_id = ? AND reaction_type = 'dislike'`
	err := db.QueryRowContext(ctx, query, commentID).Scan(&count)
	if err != nil {
		logger.ErrorLogger.Printf("failed to get dislike count: %v", err)
		return 0, fmt.Errorf("failed to get dislike count: %v", err)
	}

	return count, nil
}
