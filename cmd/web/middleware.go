package main

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Frame-Options", "deny")

		next.ServeHTTP(w, r)
	})
}

func rateLimiter(next http.Handler) http.Handler {
	var (
		mutex   = &sync.Mutex{}
		counter = make(map[string]int)
	)

	go func() {
		for {
			time.Sleep(1 * time.Second)
			mutex.Lock()
			counter = make(map[string]int)
			mutex.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr

		mutex.Lock()
		counter[ip]++
		reqCount := counter[ip]
		mutex.Unlock()

		// Max 5 requests are allowed within a second,
		if reqCount > 5 {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) requireLogin(handler http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, loggedIn := app.GetUserFromSession(r)
		if !loggedIn {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}
		handler.ServeHTTP(w, r)
	})
}

func wwwRedirect(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if host := strings.TrimPrefix(r.Host, "www."); host != r.Host {
			// Request host has www. prefix. Redirect to host with www. trimmed.
			u := *r.URL
			u.Host = host
			u.Scheme = "https"
			http.Redirect(w, r, u.String(), http.StatusFound)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func redirectHTTPS(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://localhost"+port+r.RequestURI, http.StatusMovedPermanently)
}
