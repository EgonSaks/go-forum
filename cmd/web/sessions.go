package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"forum/configs"
	"forum/logger"
	"forum/pkg/models"

	"github.com/google/uuid"
)

func (app *application) SetSession(w http.ResponseWriter, r *http.Request, id string) (*http.Cookie, error) {
	user := models.User{
		ID: id,
	}
	cookie, err := r.Cookie("session")
	fmt.Println("Looking for an active session...")

	if err != nil {
		fmt.Println("Didn't find active session. Creating a new session")
		cookie = &http.Cookie{
			Name:     "session",
			Value:    uuid.New().String(),
			Path:     "/",
			HttpOnly: true,
			Expires:  time.Now().Add(2 * time.Hour),
			Secure:   true,
			// SameSite: http.SameSiteStrictMode,
		}
		http.SetCookie(w, cookie)

		session := models.Session{
			ID:        cookie.Value,
			UserID:    user.ID,
			ExpiresAt: cookie.Expires,
		}

		_, err = models.CreateSession(app.db, session)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Println("Getting active session", cookie)
		session, err := models.GetSessionByID(app.db, cookie.Value)
		if err != nil {
			fmt.Println("No matching session found in the database, delete the cookie")
			app.DeleteSession(w, r)

			// Create a new session cookie for the current user
			cookie = &http.Cookie{
				Name:     "session",
				Value:    uuid.New().String(),
				Path:     "/",
				HttpOnly: true,
				Expires:  time.Now().Add(2 * time.Hour),
				Secure:   true,
			}
			http.SetCookie(w, cookie)

			session := models.Session{
				ID:        cookie.Value,
				UserID:    user.ID,
				ExpiresAt: cookie.Expires,
			}

			_, err = models.CreateSession(app.db, session)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Created new session for user", user.ID)
		} else if session.UserID != user.ID {
			fmt.Println("Session belongs to a different user, delete the cookie")
			app.DeleteSession(w, r)

			// Create a new session cookie for the current user
			cookie = &http.Cookie{
				Name:     "session",
				Value:    uuid.New().String(),
				Path:     "/",
				HttpOnly: true,
				Expires:  time.Now().Add(2 * time.Hour),
				Secure:   true,
			}
			http.SetCookie(w, cookie)

			session := models.Session{
				ID:        cookie.Value,
				UserID:    user.ID,
				ExpiresAt: cookie.Expires,
			}

			_, err = models.CreateSession(app.db, session)
			if err != nil {
				logger.ErrorLogger.Println("Error with creating session:", err)
				log.Fatal(err)
			}
			fmt.Println("Created new session for user", user.ID)
		}
	}

	return cookie, nil
}

func (app *application) GetUserFromSession(r *http.Request) (models.User, bool) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return models.User{}, false
	}

	var userID string
	err = app.db.QueryRow(`SELECT user_id FROM sessions WHERE id = ? AND expires_at > ?`, cookie.Value, time.Now()).Scan(&userID)
	if err != nil {
		return models.User{}, false
	}

	var user models.User
	err = app.db.QueryRow(`SELECT id, name, email, hashed_password, created_at, updated_at FROM users WHERE id = ?`, userID).Scan(&user.ID, &user.Name, &user.Email, &user.HashedPassword, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return models.User{}, false
	}
	// fmt.Println("Logged in", user.ID)
	return user, true
}

func (app *application) DeleteSession(w http.ResponseWriter, r *http.Request) (*http.Cookie, error) {
	cookie, err := r.Cookie("session")
	if err != nil {
		fmt.Println("No cookie found")
		return nil, nil

	}

	err = configs.RevokeGoogleToken(cookie.Value, logger.ErrorLogger)
	if err != nil {
		logger.ErrorLogger.Printf("Failed to revoke token: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	err = models.DeleteSession(app.db, cookie.Value)
	if err != nil {
		return nil, err
	}

	cookie = &http.Cookie{
		Name:   "session",
		Value:  "",
		MaxAge: -1,
		Path:   "/",
	}
	http.SetCookie(w, cookie)
	return cookie, nil
}
