package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"forum/configs"
	"forum/logger"
	"forum/pkg/models"

	"github.com/google/uuid"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	loggedInUser, isLoggedIn := app.GetUserFromSession(r)

	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	posts, err := models.GetAllPosts(app.db)
	if err != nil {
		logger.ErrorLogger.Printf("Error getting posts: %v\n", err)
		http.Error(w, "Post(s) not found", http.StatusNotFound)
		return
	}

	for i := range posts {
		count, err := models.CommentCountByPostID(app.db, posts[i].ID)
		if err != nil {
			logger.ErrorLogger.Printf("Error getting comment count: %v\n", err)
			http.Error(w, "Failed to get comment count", http.StatusInternalServerError)
			return
		}
		posts[i].CommentsCount = count
	}

	for i := range posts {
		likes, err := models.PostLikeCountByPostID(app.db, posts[i].ID)
		if err != nil {
			logger.ErrorLogger.Printf("Error getting post likes: %v\n", err)
			http.Error(w, "Like(s) not found", http.StatusNotFound)
		}
		posts[i].Likes = likes

		dislikes, err := models.PostDislikeCountByPostID(app.db, posts[i].ID)
		if err != nil {
			logger.ErrorLogger.Printf("Error getting post dislikes: %v\n", err)
			http.Error(w, "Dislike(s) not found", http.StatusNotFound)
		}
		posts[i].Dislikes = dislikes
	}

	data := &templateData{
		Posts:        posts,
		IsLoggedIn:   isLoggedIn,
		LoggedInUser: loggedInUser,
	}

	if err := app.renderTemplate(w, r, "home.page.html", data); err != nil {
		logger.ErrorLogger.Printf("Error rendering template: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// post handlers, showPost, createPost
func (app *application) showPost(w http.ResponseWriter, r *http.Request) {
	loggedInUser, isLoggedIn := app.GetUserFromSession(r)

	if r.URL.Path != "/post/" {
		http.NotFound(w, r)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.NotFound(w, r)
		return
	}

	post, err := models.GetPostByID(app.db, id)
	if err != nil {
		logger.ErrorLogger.Println("Error getting post:", err)
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	comments, err := models.GetAllCommentsByPostID(app.db, post.ID)
	if err != nil {
		logger.ErrorLogger.Println("Error getting comments:", err)
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	for i := range comments {
		comments[i].IsLoggedIn = isLoggedIn
	}

	for i := range comments {
		commentLikes, err := models.CommentLikeCountByCommentID(app.db, comments[i].ID)
		if err != nil {
			logger.ErrorLogger.Println("Error getting comment likes:", err)
			http.Error(w, "Like(s) not found", http.StatusNotFound)
			return
		}
		comments[i].Likes = commentLikes

		commentDislikes, err := models.CommentDislikeCountByCommentID(app.db, comments[i].ID)
		if err != nil {
			logger.ErrorLogger.Println("Error getting comment dislikes:", err)
			http.Error(w, "Dislike(s) not found", http.StatusNotFound)
			return
		}
		comments[i].Dislikes = commentDislikes
	}

	commentsCount, err := models.CommentCountByPostID(app.db, id)
	if err != nil {
		logger.ErrorLogger.Println("Error getting comments count:", err)
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	postLikes, err := models.PostLikeCountByPostID(app.db, post.ID)
	if err != nil {
		logger.ErrorLogger.Println("Error getting post likes:", err)
		http.Error(w, "Like(s) not found", http.StatusNotFound)
		return
	}

	postDislikes, err := models.PostDislikeCountByPostID(app.db, post.ID)
	if err != nil {
		logger.ErrorLogger.Println("Error getting post dislikes:", err)
		http.Error(w, "Dislike(s) not found", http.StatusNotFound)
		return
	}

	data := &templateData{
		Post:          post,
		IsLoggedIn:    isLoggedIn,
		LoggedInUser:  loggedInUser,
		Comments:      comments,
		CommentsCount: commentsCount,
		PostLikes:     postLikes,
		PostDislikes:  postDislikes,
	}

	if err := app.renderTemplate(w, r, "show.page.html", data); err != nil {
		logger.ErrorLogger.Printf("Error rendering template: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// user profile
func (app *application) userProfile(w http.ResponseWriter, r *http.Request) {
	loggedInUser, isLoggedIn := app.GetUserFromSession(r)

	if r.URL.Path != "/user/profile" {
		http.NotFound(w, r)
		return
	}

	data := &templateData{
		IsLoggedIn:   isLoggedIn,
		LoggedInUser: loggedInUser,
	}

	if err := app.renderTemplate(w, r, "userprofile.page.html", data); err != nil {
		logger.ErrorLogger.Printf("Error rendering template: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// user profile sub-pages
func (app *application) userProfilePostsPage(w http.ResponseWriter, r *http.Request) {
	loggedInUser, isLoggedIn := app.GetUserFromSession(r)

	if r.URL.Path != "/user/profile/posts" {
		http.NotFound(w, r)
		return
	}

	posts, err := models.GetAllPostsByUserID(app.db, loggedInUser.ID)
	if err != nil {
		logger.ErrorLogger.Println("Error getting post:", err)
		http.Error(w, "Post(s) not found", http.StatusNotFound)
		return
	}

	data := &templateData{
		Posts:        posts,
		IsLoggedIn:   isLoggedIn,
		LoggedInUser: loggedInUser,
	}

	if err := app.renderTemplate(w, r, "userprofile.posts.page.html", data); err != nil {
		logger.ErrorLogger.Printf("Error rendering template: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (app *application) userProfileCommentsPage(w http.ResponseWriter, r *http.Request) {
	loggedInUser, isLoggedIn := app.GetUserFromSession(r)

	if r.URL.Path != "/user/profile/comments" {
		http.NotFound(w, r)
		return
	}

	comments, err := models.GetAllCommentsByUserID(app.db, loggedInUser.ID)
	if err != nil {
		logger.ErrorLogger.Println("Error getting comment:", err)
		http.Error(w, "Comment(s) not found", http.StatusNotFound)
		return
	}

	data := &templateData{
		Comments:     comments,
		IsLoggedIn:   isLoggedIn,
		LoggedInUser: loggedInUser,
	}

	if err := app.renderTemplate(w, r, "userprofile.comments.page.html", data); err != nil {
		logger.ErrorLogger.Printf("Error rendering template: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (app *application) userProfilePostReaction(w http.ResponseWriter, r *http.Request) {
	loggedInUser, isLoggedIn := app.GetUserFromSession(r)

	if r.URL.Path != "/user/profile/post/reactions" {
		http.NotFound(w, r)
		return
	}

	allPosts, err := models.GetAllPosts(app.db)
	if err != nil {
		logger.ErrorLogger.Println("Error getting post:", err)
		http.Error(w, "Post(s) not found", http.StatusNotFound)
		return
	}

	// user liked disliked posts
	var userLikedDislikedPosts []models.Post

	for i := range allPosts {
		likes, err := models.GetUserLikedPostsByPostIDAndUserID(app.db, allPosts[i].ID, loggedInUser.ID)
		if err != nil {
			logger.ErrorLogger.Println("Error getting like:", err)
			http.Error(w, "Like(s) not found", http.StatusNotFound)
		}
		allPosts[i].Likes = likes

		dislikes, err := models.GetUserDislikedPostsByPostIDAndUserID(app.db, allPosts[i].ID, loggedInUser.ID)
		if err != nil {
			logger.ErrorLogger.Println("Error getting dislike:", err)
			http.Error(w, "Dislike(s) not found", http.StatusNotFound)
		}
		allPosts[i].Dislikes = dislikes

		if likes > 0 || dislikes > 0 {
			userLikedDislikedPosts = append(userLikedDislikedPosts, allPosts[i])
		}
	}

	data := &templateData{
		IsLoggedIn:             isLoggedIn,
		LoggedInUser:           loggedInUser,
		UserLikedDislikedPosts: userLikedDislikedPosts,
	}

	if err := app.renderTemplate(w, r, "userprofile.post.reactions.page.html", data); err != nil {
		logger.ErrorLogger.Printf("Error rendering template: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (app *application) userProfileCommentReaction(w http.ResponseWriter, r *http.Request) {
	loggedInUser, isLoggedIn := app.GetUserFromSession(r)

	if r.URL.Path != "/user/profile/comment/reactions" {
		http.NotFound(w, r)
		return
	}

	allComments, err := models.GetAllComments(app.db)
	if err != nil {
		logger.ErrorLogger.Println("Error getting comment:", err)
		http.Error(w, "Comment(s) not found", http.StatusNotFound)
		return
	}

	// user liked disliked comments
	var userLikedDislikedComments []models.Comment

	for i := range allComments {
		likes, err := models.GetUserLikedCommentByCommentIDAndUserID(app.db, allComments[i].ID, loggedInUser.ID)
		if err != nil {
			logger.ErrorLogger.Println("Error getting like:", err)
			http.Error(w, "Like(s) not found", http.StatusNotFound)
		}
		allComments[i].Likes = likes

		dislikes, err := models.GetUserDislikedCommentByCommentIDAndUserID(app.db, allComments[i].ID, loggedInUser.ID)
		if err != nil {
			logger.ErrorLogger.Println("Error getting dislike:", err)
			http.Error(w, "Dislike(s) not found", http.StatusNotFound)
		}
		allComments[i].Dislikes = dislikes

		if likes > 0 || dislikes > 0 {
			postTitle, _ := models.GetPostTitleByCommentID(app.db, allComments[i].ID)
			if err != nil {
				logger.ErrorLogger.Println("Error getting comment:", err)
				http.Error(w, "Post not found", http.StatusNotFound)
				return
			}
			allComments[i].PostTitle = postTitle
			userLikedDislikedComments = append(userLikedDislikedComments, allComments[i])
		}
	}

	data := &templateData{
		IsLoggedIn:                isLoggedIn,
		LoggedInUser:              loggedInUser,
		UserLikedDislikedComments: userLikedDislikedComments,
	}

	if err := app.renderTemplate(w, r, "userprofile.comment.reaction.page.html", data); err != nil {
		logger.ErrorLogger.Printf("Error rendering template: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (app *application) userActivity(w http.ResponseWriter, r *http.Request) {
	loggedInUser, isLoggedIn := app.GetUserFromSession(r)

	if r.URL.Path != "/user/profile/activity" {
		http.NotFound(w, r)
		return
	}

	allPosts, err := models.GetAllPosts(app.db)
	if err != nil {
		logger.ErrorLogger.Println("Error getting post:", err)
		http.Error(w, "Post(s) not found", http.StatusNotFound)
		return
	}

	allComments, err := models.GetAllComments(app.db)
	if err != nil {
		logger.ErrorLogger.Println("Error getting comment:", err)
		http.Error(w, "Comment(s) not found", http.StatusNotFound)
		return
	}

	posts, err := models.GetAllPostsByUserID(app.db, loggedInUser.ID)
	if err != nil {
		logger.ErrorLogger.Println("Error getting post:", err)
		http.Error(w, "Post(s) not found", http.StatusNotFound)
		return
	}

	comments, err := models.GetAllCommentsByUserID(app.db, loggedInUser.ID)
	if err != nil {
		logger.ErrorLogger.Println("Error getting comment:", err)
		http.Error(w, "Comment(s) not found", http.StatusNotFound)
		return
	}

	// user liked disliked posts
	var userLikedDislikedPosts []models.Post

	for i := range allPosts {
		likes, err := models.GetUserLikedPostsByPostIDAndUserID(app.db, allPosts[i].ID, loggedInUser.ID)
		if err != nil {
			logger.ErrorLogger.Println("Error getting like:", err)
			http.Error(w, "Like(s) not found", http.StatusNotFound)
		}
		allPosts[i].Likes = likes

		dislikes, err := models.GetUserDislikedPostsByPostIDAndUserID(app.db, allPosts[i].ID, loggedInUser.ID)
		if err != nil {
			logger.ErrorLogger.Println("Error getting dislike:", err)
			http.Error(w, "Dislike(s) not found", http.StatusNotFound)
		}
		allPosts[i].Dislikes = dislikes

		if likes > 0 || dislikes > 0 {
			userLikedDislikedPosts = append(userLikedDislikedPosts, allPosts[i])
		}
	}

	// user liked disliked comments
	var userLikedDislikedComments []models.Comment

	for i := range allComments {
		likes, err := models.GetUserLikedCommentByCommentIDAndUserID(app.db, allComments[i].ID, loggedInUser.ID)
		if err != nil {
			logger.ErrorLogger.Println("Error getting like:", err)
			http.Error(w, "Like(s) not found", http.StatusNotFound)
		}
		allComments[i].Likes = likes

		dislikes, err := models.GetUserDislikedCommentByCommentIDAndUserID(app.db, allComments[i].ID, loggedInUser.ID)
		if err != nil {
			logger.ErrorLogger.Println("Error getting dislike:", err)
			http.Error(w, "Dislike(s) not found", http.StatusNotFound)
		}
		allComments[i].Dislikes = dislikes

		if likes > 0 || dislikes > 0 {
			postTitle, _ := models.GetPostTitleByCommentID(app.db, allComments[i].ID)
			if err != nil {
				logger.ErrorLogger.Println("Error getting post:", err)
				http.Error(w, "Post not found", http.StatusNotFound)
				return
			}
			allComments[i].PostTitle = postTitle
			userLikedDislikedComments = append(userLikedDislikedComments, allComments[i])
		}
	}

	data := &templateData{
		Posts:                     posts,
		Comments:                  comments,
		IsLoggedIn:                isLoggedIn,
		LoggedInUser:              loggedInUser,
		UserLikedDislikedPosts:    userLikedDislikedPosts,
		UserLikedDislikedComments: userLikedDislikedComments,
	}

	if err := app.renderTemplate(w, r, "userprofile.activity.page.html", data); err != nil {
		logger.ErrorLogger.Printf("Error rendering template: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (app *application) createPost(w http.ResponseWriter, r *http.Request) {
	loggedInUser, loggedIn := app.GetUserFromSession(r)

	if r.URL.Path != "/post/create" {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		data := &templateData{
			IsLoggedIn:   loggedIn,
			LoggedInUser: loggedInUser,
			CurrentPage:  r.URL.Path,
		}

		if err := app.renderTemplate(w, r, "create.page.html", data); err != nil {
			logger.ErrorLogger.Printf("Error rendering template: %v\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}

	case http.MethodPost:
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			logger.ErrorLogger.Printf("Error parsing multipart form: %s\n", err)
			http.Error(w, "Unable to parse form", http.StatusBadRequest)
			return
		}
		title := r.PostForm.Get("title")
		content := r.PostForm.Get("content")
		categories := r.PostForm["categories"]
		category := strings.Join(categories, "; ")

		// Get the file from the form data
		image, handler, err := r.FormFile("image")
		if err != nil {
			logger.ErrorLogger.Printf("Error getting image from form data: %s\n", err)
			image = nil
			handler = nil
		} else {
			defer image.Close()
		}

		var formErrors map[string]string
		var post models.Post

		if image != nil {
			// an image was uploaded
			extension := filepath.Ext(handler.Filename)

			formErrors = validateCreatePostForm(title, content, extension, categories, handler)
			if len(formErrors) == 0 {

				filePath, err := app.UploadImage(image, extension)
				if err != nil {
					logger.ErrorLogger.Printf("Error creating post with image: %s\n", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				post = models.Post{
					ID:            uuid.New().String(),
					UserID:        loggedInUser.ID,
					Title:         title,
					Content:       content,
					ImageFullPath: filePath,
					Category:      category,
					CreatedAt:     time.Now(),
				}
			}
		} else {
			// no image was uploaded
			formErrors = validateCreatePostFormWithoutImage(title, content, categories)

			if len(formErrors) == 0 {
				post = models.Post{
					ID:        uuid.New().String(),
					UserID:    loggedInUser.ID,
					Title:     title,
					Content:   content,
					Category:  category,
					CreatedAt: time.Now(),
				}
			}
		}

		if len(formErrors) > 0 {
			data := &templateData{
				FormErrors: formErrors,
				FormData:   r.PostForm,
			}

			if err := app.renderTemplate(w, r, "create.page.html", data); err != nil {
				logger.ErrorLogger.Printf("Error rendering template: %v\n", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		if _, err := models.CreatePost(app.db, post); err != nil {
			logger.ErrorLogger.Printf("Error creating post: %v\n", err)
			http.Error(w, "Unable to create post", http.StatusInternalServerError)
			return
		}

		logger.InfoLogger.Printf("Post created: ID=%s, Title=%s, Author=%s\n", post.ID, post.Title, loggedInUser.Name)
		http.Redirect(w, r, "/post?id="+post.ID, http.StatusSeeOther)

	default:
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
}

// comment handlers createComment
func (app *application) createComment(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/post/comment" {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			logger.ErrorLogger.Printf("Error parsing a form: %s\n", err)
			http.Error(w, "Unable to parse form", http.StatusBadRequest)
			return
		}

		post_id := r.FormValue("post_id")

		user, isLoggedIn := app.GetUserFromSession(r)
		if user.ID == "" {
			http.Error(w, "Missing user ID", http.StatusBadRequest)
			return
		}

		comment := r.PostForm.Get("comment")
		formErrors := validateCreateCommentForm(comment)

		if len(formErrors) > 0 {

			post, err := models.GetPostByID(app.db, post_id)
			if err != nil {
				logger.ErrorLogger.Println("Error getting post:", err)
				http.Error(w, "Post not found", http.StatusNotFound)
				return
			}

			comments, err := models.GetAllCommentsByPostID(app.db, post.ID)
			if err != nil {
				logger.ErrorLogger.Println("Error getting comment:", err)
				http.Error(w, "Comment not found", http.StatusNotFound)
				return
			}

			for i := range comments {
				comments[i].IsLoggedIn = isLoggedIn
				commentLikes, err := models.CommentLikeCountByCommentID(app.db, comments[i].ID)
				if err != nil {
					logger.ErrorLogger.Println("Error getting like:", err)
					http.Error(w, "Like(s) not found", http.StatusNotFound)
					return
				}
				comments[i].Likes = commentLikes

				commentDislikes, err := models.CommentDislikeCountByCommentID(app.db, comments[i].ID)
				if err != nil {
					logger.ErrorLogger.Println("Error getting dislike:", err)
					http.Error(w, "Dislike(s) not found", http.StatusNotFound)
					return
				}
				comments[i].Dislikes = commentDislikes
			}

			commentsCount, err := models.CommentCountByPostID(app.db, post.ID)
			if err != nil {
				logger.ErrorLogger.Println("Error getting post:", err)
				http.Error(w, "Post not found", http.StatusNotFound)
				return
			}

			postLikes, err := models.PostLikeCountByPostID(app.db, post.ID)
			if err != nil {
				logger.ErrorLogger.Println("Error getting like:", err)
				http.Error(w, "Like(s) not found", http.StatusNotFound)
				return
			}

			postDislikes, err := models.PostDislikeCountByPostID(app.db, post.ID)
			if err != nil {
				logger.ErrorLogger.Println("Error getting dislike:", err)
				http.Error(w, "Dislike(s) not found", http.StatusNotFound)
				return
			}

			data := &templateData{
				Post:          post,
				IsLoggedIn:    isLoggedIn,
				LoggedInUser:  user,
				Comments:      comments,
				CommentsCount: commentsCount,
				PostLikes:     postLikes,
				PostDislikes:  postDislikes,
				FormErrors:    formErrors,
				FormData:      r.PostForm,
			}

			if err := app.renderTemplate(w, r, "show.page.html", data); err != nil {
				logger.ErrorLogger.Printf("Error rendering template: %v\n", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		comment_content := models.Comment{
			ID:        uuid.New().String(),
			UserID:    user.ID,
			PostID:    post_id,
			Content:   comment,
			CreatedAt: time.Now(),
		}

		if _, err := models.CreateComment(app.db, comment_content); err != nil {
			logger.ErrorLogger.Println("Error with creating comment:", err)
			http.Error(w, "Unable to create comment", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/post?id="+post_id, http.StatusSeeOther)

	default:
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// likes and dislikes handling for post
func (app *application) createPostReaction(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/post/reaction" {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			logger.ErrorLogger.Printf("Error parsing a form: %s\n", err)
			http.Error(w, "Unable to parse form", http.StatusBadRequest)
			return
		}

		user, _ := app.GetUserFromSession(r)
		if user.ID == "" {
			http.Error(w, "Missing user ID", http.StatusBadRequest)
			return
		}

		post_id := r.FormValue("post_id")
		reactionType := r.FormValue("reaction_type")

		reaction := models.PostReaction{
			ID:           uuid.New().String(),
			UserID:       user.ID,
			PostID:       post_id,
			ReactionType: reactionType,
			CreatedAt:    time.Now(),
		}

		if _, err := models.CreatePostReaction(app.db, reaction); err != nil {
			logger.ErrorLogger.Printf("Error with creating a reaction: %s\n", err)
			http.Error(w, "Unable to create reaction", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/post?id="+post_id, http.StatusSeeOther)

	default:
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// likes and dislikes handling for comments
func (app *application) createCommentReaction(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/post/comment/reaction" {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			logger.ErrorLogger.Printf("Error parsing a form: %s\n", err)
			http.Error(w, "Unable to parse form", http.StatusBadRequest)
			return
		}

		user, _ := app.GetUserFromSession(r)
		if user.ID == "" {
			http.Error(w, "Missing user ID", http.StatusBadRequest)
			return
		}

		post_id := r.FormValue("post_id")
		comment_id := r.FormValue("comment_id")
		reactionType := r.FormValue("reaction_type")

		reaction := models.CommentReaction{
			ID:           uuid.New().String(),
			UserID:       user.ID,
			PostID:       post_id,
			CommentID:    comment_id,
			ReactionType: reactionType,
			CreatedAt:    time.Now(),
		}

		if _, err := models.CreateCommentReaction(app.db, reaction); err != nil {
			logger.ErrorLogger.Printf("Error with creating reaction %s\n", err)
			http.Error(w, "Unable to create reaction", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/post?id="+post_id, http.StatusSeeOther)

	default:
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

// search handler
func (app *application) search(w http.ResponseWriter, r *http.Request) {
	loggedInUser, isLoggedIn := app.GetUserFromSession(r)

	if r.URL.Path != "/search" {
		http.NotFound(w, r)
		return
	}

	searchKey := r.FormValue("search")
	errors := checkSearch(searchKey)
	if len(errors) > 0 {
		data := &templateData{
			FormErrors: errors,
			FormData:   r.PostForm,
		}
		if err := app.renderTemplate(w, r, "home.page.html", data); err != nil {
			logger.ErrorLogger.Printf("Error rendering template: %v\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	posts, err := models.GetAllBySearchKey(app.db, searchKey)
	if err != nil {
		logger.ErrorLogger.Println("Error getting post:", err)
		http.Error(w, "Post(s) not found in search", http.StatusNotFound)
	}

	for i := range posts {
		count, err := models.CommentCountByPostID(app.db, posts[i].ID)
		if err != nil {
			logger.ErrorLogger.Println("Error getting comment:", err)
			http.Error(w, "Failed to get comment count", http.StatusInternalServerError)
			return
		}
		posts[i].CommentsCount = count
	}

	for i := range posts {
		likes, err := models.PostLikeCountByPostID(app.db, posts[i].ID)
		if err != nil {
			logger.ErrorLogger.Println("Error getting like:", err)
			http.Error(w, "Like(s) not found", http.StatusNotFound)
		}
		posts[i].Likes = likes

		dislikes, err := models.PostDislikeCountByPostID(app.db, posts[i].ID)
		if err != nil {
			logger.ErrorLogger.Println("Error getting dislike:", err)
			http.Error(w, "Dislike(s) not found", http.StatusNotFound)
		}
		posts[i].Dislikes = dislikes
	}

	data := &templateData{
		Posts:        posts,
		IsLoggedIn:   isLoggedIn,
		LoggedInUser: loggedInUser,
	}

	if err := app.renderTemplate(w, r, "home.page.html", data); err != nil {
		logger.ErrorLogger.Printf("Error rendering template: %v\n", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// filter handler
func (app *application) filter(w http.ResponseWriter, r *http.Request) {
	loggedInUser, isLoggedIn := app.GetUserFromSession(r)

	if r.URL.Path != "/filter" {
		http.NotFound(w, r)
		return
	}

	// Get filter values
	switch r.Method {
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			logger.ErrorLogger.Printf("Error parsing a form: %s\n", err)
			http.Error(w, "Unable to parse form", http.StatusBadRequest)
			return
		}

		categories := r.PostForm["category-filter"]
		fromDate := r.FormValue("date-filter")
		likesStr := r.PostForm["likes-filter"]

		// Convert likes filter to []int
		var likes []int
		for _, likeStr := range likesStr {
			like, err := strconv.Atoi(likeStr)
			if err != nil {
				fmt.Println(err)
				continue
			}
			likes = append(likes, like)
		}

		// Get filtered posts
		posts, err := models.GetPostsWithFilters(app.db, categories, fromDate, likes)
		if err != nil {
			logger.ErrorLogger.Println("Error getting post:", err)
			http.Error(w, "Post(s) not found", http.StatusNotFound)
			return
		}

		// Update post counts based on filters
		for i := range posts {
			count, err := models.CommentCountByPostID(app.db, posts[i].ID)
			if err != nil {
				logger.ErrorLogger.Println("Error getting comment:", err)
				http.Error(w, "Failed to get comment count", http.StatusInternalServerError)
				return
			}
			posts[i].CommentsCount = count

			likes, err := models.PostLikeCountByPostID(app.db, posts[i].ID)
			if err != nil {
				logger.ErrorLogger.Println("Error getting like:", err)
				http.Error(w, "Like(s) not found", http.StatusNotFound)
			}
			posts[i].Likes = likes

			dislikes, err := models.PostDislikeCountByPostID(app.db, posts[i].ID)
			if err != nil {
				logger.ErrorLogger.Println("Error getting dislike:", err)
				http.Error(w, "Dislike(s) not found", http.StatusNotFound)
			}
			posts[i].Dislikes = dislikes
		}

		data := &templateData{
			Posts:        posts,
			IsLoggedIn:   isLoggedIn,
			LoggedInUser: loggedInUser,
		}

		if err := app.renderTemplate(w, r, "home.page.html", data); err != nil {
			logger.ErrorLogger.Printf("Error rendering template: %v\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

func (app *application) signup(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/user/signup" {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		if err := app.renderTemplate(w, r, "signup.page.html", nil); err != nil {
			logger.ErrorLogger.Printf("Error rendering template: %v\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}

	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			logger.ErrorLogger.Printf("Error parsing a form: %s\n", err)
			http.Error(w, "Unable to parse form", http.StatusBadRequest)
			return
		}

		name := r.PostForm.Get("name")
		email := r.PostForm.Get("email")
		password := r.PostForm.Get("password")

		formErrors := app.validateSignUpForm(name, email, password)

		if len(formErrors) > 0 {
			data := &templateData{
				FormErrors: formErrors,
				FormData:   r.PostForm,
			}
			if err := app.renderTemplate(w, r, "signup.page.html", data); err != nil {
				logger.ErrorLogger.Printf("Error rendering template: %v\n", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		hashedPassword, err := HashPassword(password)
		if err != nil {
			log.Fatal(err)
			logger.ErrorLogger.Println("Error creating user:", err)
			http.Error(w, "Unable to create user", http.StatusInternalServerError)
			return
		}

		user := models.User{
			ID:             uuid.New().String(),
			Name:           name,
			Email:          email,
			HashedPassword: hashedPassword,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		if _, err := models.CreateUser(app.db, user); err != nil {
			logger.ErrorLogger.Println("Error creating user:", err)
			http.Error(w, "Unable to create user", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/user/login", http.StatusSeeOther)

	default:
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
}

func (app *application) login(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/user/login" {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		if err := app.renderTemplate(w, r, "login.page.html", nil); err != nil {
			logger.ErrorLogger.Printf("Error rendering template: %v\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}

	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			logger.ErrorLogger.Printf("Error parsing a form: %s\n", err)
			http.Error(w, "Unable to parse form", http.StatusBadRequest)
			return
		}

		email := r.PostForm.Get("email")
		password := r.PostForm.Get("password")

		errors := validateSingInForm(email, password)

		id, err := models.AuthenticateUser(app.db, email, password)
		if err != nil {
			errors["generic"] = "Email or Password is incorrect"
			data := &templateData{
				FormErrors: errors,
				FormData:   r.PostForm,
			}
			app.renderTemplate(w, r, "login.page.html", data)
			return
		}

		if len(errors) > 0 {
			data := &templateData{
				FormErrors: errors,
				FormData:   r.PostForm,
			}
			if err := app.renderTemplate(w, r, "login.page.html", data); err != nil {
				logger.ErrorLogger.Printf("Error rendering template: %v\n", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		app.SetSession(w, r, id)
		http.Redirect(w, r, "/", http.StatusSeeOther)

	default:
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
}

func (app *application) logout(w http.ResponseWriter, r *http.Request) {
	_, isLoggedIn := app.GetUserFromSession(r)

	if r.URL.Path != "/user/logout" {
		http.NotFound(w, r)
		return
	}
	if !isLoggedIn {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	app.DeleteSession(w, r)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// google login / logout
func (app *application) handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/login/google" {
		http.NotFound(w, r)
		return
	}
	url := configs.GoogleOauthConfig.AuthCodeURL(configs.OauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (app *application) handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/GoogleCallback" {
		http.NotFound(w, r)
		return
	}
	state := r.FormValue("state")
	if state != configs.OauthStateString {
		logger.ErrorLogger.Printf("Error, invalid oauth state, expected '%s', got '%s'\n", configs.OauthStateString, state)
		fmt.Printf("invalid oauth state, expected '%s', got '%s'\n", configs.OauthStateString, state)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	token, err := configs.GetGoogleAccessToken(r, logger.ErrorLogger)
	logger.InfoLogger.Println("Got Google access token")
	if err != nil {
		logger.ErrorLogger.Printf("Error, code exchange failed with '%v'\n", err)
		fmt.Printf("Code exchange failed with '%v'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	GoogleUser, err := configs.GetGoogleData(token, logger.ErrorLogger)
	logger.InfoLogger.Println("Successfully retrieved data from Google API.")
	if err != nil {
		logger.ErrorLogger.Printf("Error with getting user info: %v\n", err)
		fmt.Printf("Failed to get user info: %v\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	dbUser, _ := models.GetUserByEmail(app.db, GoogleUser.Email)
	if dbUser.Email == GoogleUser.Email {
		// Set the session cookie
		cookie := http.Cookie{
			Name:    "session",
			Value:   token.AccessToken,
			Expires: time.Now().Add(2 * time.Hour),
			Path:    "/",
			Secure:  true,
		}
		http.SetCookie(w, &cookie)

		session := models.Session{
			ID:        cookie.Value,
			UserID:    dbUser.ID,
			ExpiresAt: cookie.Expires,
		}

		_, err = models.CreateSession(app.db, session)
		if err != nil {
			log.Fatal(err)
		}
	} else if dbUser.Email != GoogleUser.Email {
		google_user := models.User{
			ID:             GoogleUser.ID,
			Name:           GoogleUser.GivenName,
			Email:          GoogleUser.Email,
			HashedPassword: []byte{0o00},
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		if _, err := models.CreateUser(app.db, google_user); err != nil {
			fmt.Println(err)
			logger.ErrorLogger.Printf("Error with creating user: %v\n", err)
			http.Error(w, "Unable to create user", http.StatusInternalServerError)
			return
		}

		// Set the session cookie
		cookie := http.Cookie{
			Name:    "session",
			Value:   token.AccessToken,
			Expires: time.Now().Add(2 * time.Hour),
			Path:    "/",
		}
		http.SetCookie(w, &cookie)

		session := models.Session{
			ID:        cookie.Value,
			UserID:    google_user.ID,
			ExpiresAt: cookie.Expires,
		}

		_, err = models.CreateSession(app.db, session)
		if err != nil {
			log.Fatal(err)
		}
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// github login / logout
func (app *application) handleGithubLogin(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/login/github/" {
		http.NotFound(w, r)
		return
	}
	githubClientID := os.Getenv("GITHUB_KEY")

	redirectURL := fmt.Sprintf("https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s", githubClientID, "https://localhost:10443/login/github/callback")

	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

func (app *application) handleGithubCallback(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/login/github/callback" {
		http.NotFound(w, r)
		return
	}
	code := r.URL.Query().Get("code")

	githubAccessToken := configs.GetGithubAccessToken(code)
	githubData, dataErr := configs.GetGithubData(githubAccessToken)
	if dataErr != nil {
		logger.ErrorLogger.Printf("Error with unmarshaling JSON data: %v\n", dataErr)
		fmt.Printf("Failed to unmarshal JSON data: %v\n", dataErr)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	var GithubUser configs.GithubData
	err := json.Unmarshal(githubData, &GithubUser)
	if err != nil {
		logger.ErrorLogger.Printf("Error with unmarshaling JSON data: %v\n", err)
		fmt.Printf("Failed to unmarshal JSON data: %v\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	if GithubUser.UserInfo.Email == "" && len(GithubUser.EmailInfo) > 0 {
		GithubUser.UserInfo.Email = GithubUser.EmailInfo[0].Email
	}

	dbUser, _ := models.GetUserByEmail(app.db, GithubUser.UserInfo.Email)
	if dbUser.Email == GithubUser.UserInfo.Email {

		cookie := http.Cookie{
			Name:    "session",
			Value:   githubAccessToken,
			Expires: time.Now().Add(2 * time.Hour),
			Path:    "/",
			Secure:  true,
		}
		http.SetCookie(w, &cookie)

		session := models.Session{
			ID:        cookie.Value,
			UserID:    dbUser.ID,
			ExpiresAt: cookie.Expires,
		}

		_, err = models.CreateSession(app.db, session)
		if err != nil {
			log.Fatal(err)
		}
	} else if dbUser.Email != GithubUser.UserInfo.Email {
		github_user := models.User{
			ID:             uuid.New().String(),
			Name:           GithubUser.UserInfo.Name,
			Email:          GithubUser.UserInfo.Email,
			HashedPassword: []byte{0o00},
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		if _, err := models.CreateUser(app.db, github_user); err != nil {
			fmt.Println(err)
			logger.ErrorLogger.Printf("Error with creating user: %v\n", err)
			http.Error(w, "Unable to create user", http.StatusInternalServerError)
			return
		}

		cookie := http.Cookie{
			Name:    "session",
			Value:   githubAccessToken,
			Expires: time.Now().Add(2 * time.Hour),
			Path:    "/",
			Secure:  true,
		}
		http.SetCookie(w, &cookie)

		session := models.Session{
			ID:        cookie.Value,
			UserID:    github_user.ID,
			ExpiresAt: cookie.Expires,
		}

		_, err = models.CreateSession(app.db, session)
		if err != nil {
			log.Fatal(err)
		}

	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
