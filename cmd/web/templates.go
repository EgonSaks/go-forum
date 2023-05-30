package main

import (
	"html/template"
	"net/url"
	"path/filepath"
	"time"

	"forum/logger"
	"forum/pkg/models"
)

type templateData struct {
	Post                      models.Post
	Posts                     []models.Post
	Comments                  []models.Comment
	CommentsCount             int
	User                      models.User
	Sessions                  models.Session
	FormData                  url.Values
	FormErrors                map[string]string
	CurrentYear               int
	IsLoggedIn                bool
	LoggedInUser              models.User
	CurrentPage               string
	PostLikes                 int
	PostDislikes              int
	UserLikedDislikedPosts    []models.Post
	UserLikedDislikedComments []models.Comment
}

func humanDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Local().Format("15:04 on 02 Jan 2006")
}

var functions = template.FuncMap{
	"humanDate": humanDate,
}

func newTemplateCache(dir string) (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := filepath.Glob(filepath.Join(dir, "*.page.html"))
	if err != nil {
		logger.ErrorLogger.Println("Error finding page templates:", err)
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			logger.ErrorLogger.Printf("Error parsing page template %s: %v", name, err)
			return nil, err
		}

		ts, err = ts.ParseGlob(filepath.Join(dir, "*.layout.html"))
		if err != nil {
			logger.ErrorLogger.Printf("Error parsing layout template in %s: %v", name, err)
			return nil, err
		}

		ts, err = ts.ParseGlob(filepath.Join(dir, "*.partial.html"))
		if err != nil {
			logger.ErrorLogger.Printf("Error parsing partial template in %s: %v", name, err)
			return nil, err
		}

		cache[name] = ts
		logger.InfoLogger.Printf("Template %s has been parsed successfully", name)
	}

	return cache, nil
}
