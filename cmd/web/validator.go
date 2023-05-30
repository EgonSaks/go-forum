package main

import (
	"io"
	"mime/multipart"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"forum/pkg/models"
)

func validateCreatePostForm(title, content, extension string, categories []string, handler *multipart.FileHeader) map[string]string {
	errors := make(map[string]string)

	title = strings.TrimSpace(title)
	if title == "" {
		errors["title"] = "Title is required"
	} else if utf8.RuneCountInString(title) > 50 {
		errors["title"] = "Title must not exceed 50 characters"
	}

	content = strings.TrimSpace(content)
	if content == "" {
		errors["content"] = "Content is required"
	} else if utf8.RuneCountInString(content) > 500 {
		errors["content"] = "Must not exceed 1000 characters"
	}

	if len(categories) < 1 || len(categories) > 3 {
		errors["categories"] = "Please select between 1 and 3 categories"
	}

	// Check if the file is actually an image by reading the first few bytes and comparing them to
	// the magic numbers of image file formats
	if handler != nil {
		image, err := handler.Open()
		if err != nil {
			errors["image"] = "Unable to read image file"
		} else {
			defer image.Close()

			// Read the first few bytes of the file
			buffer := make([]byte, 512)
			if _, err := io.ReadFull(image, buffer); err != nil {
				errors["image"] = "Unable to read image file"
			} else {
				// Check if the magic number matches one of the supported image formats
				// contentType := http.DetectContentType(buffer)
				contentType := handler.Header.Get("Content-Type")
				if !strings.HasPrefix(contentType, "image/") {
					errors["image"] = "Only image files are supported"
				} else if !strings.EqualFold(extension, ".jpeg") && !strings.EqualFold(extension, ".jpg") &&
					!strings.EqualFold(extension, ".png") && !strings.EqualFold(extension, ".svg") &&
					!strings.EqualFold(extension, ".gif") {
					errors["image"] = "Only JPEG, PNG, SVG, and GIF file formats are supported"
				}
			}
		}
	}

	if handler.Size > 20<<20 {
		errors["image"] = "file size should not exceed 20MB"
	}

	return errors
}

func validateCreatePostFormWithoutImage(title, content string, categories []string) map[string]string {
	errors := make(map[string]string)

	title = strings.TrimSpace(title)
	if title == "" {
		errors["title"] = "Title is required"
	} else if utf8.RuneCountInString(title) > 50 {
		errors["title"] = "Title must not exceed 50 characters"
	}

	content = strings.TrimSpace(content)
	if content == "" {
		errors["content"] = "Content is required"
	} else if utf8.RuneCountInString(content) > 500 {
		errors["content"] = "Must not exceed 1000 characters"
	}

	if len(categories) < 1 || len(categories) > 3 {
		errors["categories"] = "Please select between 1 and 3 categories"
	}

	return errors
}

func validateCreateCommentForm(comment string) map[string]string {
	errors := make(map[string]string)

	comment = strings.TrimSpace(comment)
	if comment == "" {
		errors["comment"] = "Comment can't be empty"
	} else if utf8.RuneCountInString(comment) > 500 {
		errors["comment"] = "Must not exceed 500 characters"
	}

	return errors
}

func (app *application) validateSignUpForm(name, email, password string) map[string]string {
	errors := make(map[string]string)

	name = strings.TrimSpace(name)
	if name == "" {
		errors["name"] = "Name is required"
	} else if utf8.RuneCountInString(name) < 2 {
		errors["name"] = "Name must be at least 2 characters long"
	} else if utf8.RuneCountInString(name) > 50 {
		errors["name"] = "Name must be max 50 characters"
	} else {
		checkNameErrors := checkName(name)
		for key, value := range checkNameErrors {
			errors[key] = value
		}
	}

	rxEmail := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

	email = strings.TrimSpace(email)
	if email == "" {
		errors["email"] = "Email is required"
	} else if len(email) > 254 || !rxEmail.MatchString(email) {
		errors["email"] = "Invalid Email Address"
	}

	password = strings.TrimSpace(string(password))
	if password == "" {
		errors["password"] = "Password is required"
	} else if !checkPassword(password) {
		errors["password"] = "Password must contain at least 6 characters, including at least one uppercase letter, one lowercase letter, one number, and one special character."
	}

	dbName, _ := models.GetUserByName(app.db, name)
	if name == "" {
		errors["name"] = "Name is required"
	} else if dbName.Name == name {
		errors["name"] = "Name already exists"
	}

	dbEmail, _ := models.GetUserByEmail(app.db, email)
	if email == "" {
		errors["email"] = "Email is required"
	} else if dbEmail.Email == email {
		errors["email"] = "Email already exists"
	}

	return errors
}

func validateSingInForm(email, password string) map[string]string {
	errors := make(map[string]string)

	rxEmail := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

	email = strings.TrimSpace(email)
	if email == "" {
		errors["email"] = "Email is required"
	} else if len(email) > 254 || !rxEmail.MatchString(email) {
		errors["email"] = "Invalid Email Address"
	}

	password = strings.TrimSpace(string(password))
	if password == "" {
		errors["password"] = "Password is required"
	}

	return errors
}

func checkPassword(password string) bool {
	var (
		minLen     = false
		maxLen     = false
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	if len(password) >= 6 {
		minLen = true
	}

	if len(password) <= 30 {
		maxLen = true
	}

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}
	return minLen && maxLen && hasLower && hasNumber && hasSpecial && hasUpper
}

func checkName(name string) map[string]string {
	errors := make(map[string]string)
	for _, char := range name {
		if !unicode.IsLetter(char) && !unicode.IsNumber(char) && char != '-' {
			errors["name"] = "Name can only contain letters, numbers, and hyphens (-)"
			break
		}
	}
	return errors
}

func checkSearch(searchKey string) map[string]string {
	errors := make(map[string]string)
	if len(searchKey) > 50 {
		errors["search"] = "Search string should be at most 50 characters"
		return errors
	}
	return errors
}
