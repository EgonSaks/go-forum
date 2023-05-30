package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"forum/logger"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	HashedPassword []byte    `json:"hashed_password"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func CreateUser(db *sql.DB, user User) (string, error) {
	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := "INSERT INTO users (id, name, email, hashed_password, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)"
	statement, err := db.PrepareContext(context, query)
	if err != nil {
		logger.ErrorLogger.Printf("Failed to prepare create user statement: %v\n", err)
		return user.ID, fmt.Errorf("failed to prepare create user statement: %v", err)
	}

	_, err = statement.ExecContext(context, &user.ID, &user.Name, &user.Email, &user.HashedPassword, &user.CreatedAt, time.Now())
	if err != nil {
		logger.ErrorLogger.Printf("Failed to create user: %v\n", err)
		return user.ID, fmt.Errorf("failed to create user: %v", err)
	}

	return user.ID, nil
}

func GetUserByEmail(db *sql.DB, email string) (User, error) {
	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user User
	query := "SELECT id, name, email, hashed_password FROM users WHERE email = ? LIMIT 1"
	err := db.QueryRowContext(context, query, email).Scan(&user.ID, &user.Name, &user.Email, &user.HashedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.InfoLogger.Printf("No user found with email: %s", email)
			return User{}, nil
		}
		logger.ErrorLogger.Printf("Failed to get user by email: %v", err)
		return User{}, fmt.Errorf("failed to get user by email: %v", err)
	}

	return user, nil
}

func GetUserByName(db *sql.DB, name string) (User, error) {
	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user User
	query := "SELECT id, name, email, hashed_password FROM users WHERE name = ? LIMIT 1"
	err := db.QueryRowContext(context, query, name).Scan(&user.ID, &user.Name, &user.Email, &user.HashedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.InfoLogger.Printf("No user found with name: %s", name)
			return User{}, nil
		}
		logger.ErrorLogger.Printf("Failed to get user by name: %v", err)
		return User{}, fmt.Errorf("failed to get user by name: %v", err)
	}

	return user, nil
}

func AuthenticateUser(db *sql.DB, email, password string) (string, error) {
	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user User
	query := "SELECT id, hashed_password FROM users WHERE email = ?"
	err := db.QueryRowContext(context, query, email).Scan(&user.ID, &user.HashedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.ErrorLogger.Printf("email not found: %v\n", err)
			return "", errors.New("email not found")
		}
		logger.ErrorLogger.Printf("error retrieving user from database: %v\n", err)
		return "", fmt.Errorf("error retrieving user from database: %v", err)
	}

	err = bcrypt.CompareHashAndPassword(user.HashedPassword, []byte(password))
	if err != nil {
		logger.ErrorLogger.Println("incorrect password")
		return "", errors.New("incorrect password")
	}
	return user.ID, nil
}
