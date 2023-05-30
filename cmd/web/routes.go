package main

import (
	"net/http"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", app.home)

	// sign up, sing in, sign out
	mux.HandleFunc("/user/signup", app.signup)
	mux.HandleFunc("/user/login", app.login)
	mux.HandleFunc("/user/logout", app.logout)

	// google auth
	mux.HandleFunc("/login/google", app.handleGoogleLogin)
	mux.HandleFunc("/GoogleCallback", app.handleGoogleCallback)

	// github auth
	mux.HandleFunc("/login/github/", app.handleGithubLogin)
	mux.HandleFunc("/login/github/callback", app.handleGithubCallback)

	// search
	mux.HandleFunc("/search", app.search)

	// filter
	mux.HandleFunc("/filter", app.filter)

	// post handlers
	mux.HandleFunc("/post/", app.showPost)
	mux.HandleFunc("/post/create", app.requireLogin(app.createPost))

	// post like/dislike handler
	mux.HandleFunc("/post/reaction", app.requireLogin(app.createPostReaction))

	// comment handler
	mux.HandleFunc("/post/comment", app.requireLogin(app.createComment))

	// comment like/dislike handler
	mux.HandleFunc("/post/comment/reaction", app.requireLogin(app.createCommentReaction))

	// user profile
	mux.HandleFunc("/user/profile", app.requireLogin(app.userProfile))
	mux.HandleFunc("/user/profile/posts", app.requireLogin(app.userProfilePostsPage))
	mux.HandleFunc("/user/profile/comments", app.requireLogin(app.userProfileCommentsPage))
	mux.HandleFunc("/user/profile/post/reactions", app.requireLogin(app.userProfilePostReaction))
	mux.HandleFunc("/user/profile/comment/reactions", app.requireLogin(app.userProfileCommentReaction))
	mux.HandleFunc("/user/profile/activity", app.requireLogin(app.userActivity))

	fileServer := http.FileServer(http.Dir("./ui/static"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))
	
	return rateLimiter(secureHeaders(mux))
}
