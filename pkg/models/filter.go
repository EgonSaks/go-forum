package models

import (
	"database/sql"
	"fmt"
	"strings"

	"forum/logger"

	_ "github.com/mattn/go-sqlite3"
)

func GetPostsWithFilters(db *sql.DB, categories []string, fromDate string, likes []int) ([]Post, error) {
	var posts []Post
	var query string

	// Date filer
	if fromDate != "" {
		query = query + fmt.Sprintf(" AND created_at >= '%s'", fromDate)
	}

	// Category filter
	if len(categories) > 0 {
		categoryConditions := []string{}
		for _, category := range categories {
			if category != "all_categories" {
				categoryConditions = append(categoryConditions, fmt.Sprintf("category='%s'", category))
			}
		}
		if len(categoryConditions) > 0 {
			query = query + fmt.Sprintf(" AND (%s)", strings.Join(categoryConditions, " OR "))
		}
	}

	// Likes filter
	if len(likes) > 0 {
		likesConditions := []string{}
		for _, like := range likes {
			if like == -1 {
				continue
			}
			if like == 0 {
				likesConditions = append(likesConditions, "(SELECT COUNT(*) FROM post_reactions WHERE reaction_type='like' AND post_id=posts.id) = 0")
			} else {
				likesConditions = append(likesConditions, fmt.Sprintf("(SELECT COUNT(*) FROM post_reactions WHERE reaction_type='like' AND post_id=posts.id) >= %d", like))
			}
		}
		if len(likesConditions) > 0 {
			query = query + fmt.Sprintf(" AND (%s)", strings.Join(likesConditions, " OR "))
		}
	}

	query = fmt.Sprintf("SELECT * FROM posts WHERE 1=1 %s", query)

	// Log the query being executed
	logger.InfoLogger.Printf("Executing query: %s", query)

	rows, err := db.Query(query)
	if err != nil {
		// Log the error and return it
		logger.ErrorLogger.Printf("Error executing query: %s", err)
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var post Post
		err := rows.Scan(&post.ID, &post.UserID, &post.Title, &post.Content, &post.ImageFullPath, &post.Category, &post.CreatedAt)
		if err != nil {
			// Log the error and return it
			logger.ErrorLogger.Printf("Error scanning rows: %s", err)
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}
