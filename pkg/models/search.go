package models

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"forum/logger"
)

func GetAllBySearchKey(db *sql.DB, searchKey string) ([]Post, error) {
	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "SELECT id, user_id, title, content, category, created_at FROM posts WHERE title LIKE '%' || ? || '%' OR content LIKE '%' || ? || '%' OR category LIKE '%' || ? || '%' ORDER BY created_at DESC LIMIT 15"

	rows, err := db.QueryContext(context, query, searchKey, searchKey, searchKey)
	if err != nil {
		logger.ErrorLogger.Printf("failed to execute get all posts query: %v", err)
		return nil, fmt.Errorf("failed to execute get all posts query: %v", err)
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		err := rows.Scan(&post.ID, &post.UserID, &post.Title, &post.Content, &post.Category, &post.CreatedAt)
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
