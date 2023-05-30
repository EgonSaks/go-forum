package main

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"
	"time"

	"forum/configs"
	"forum/logger"
	"forum/pkg/models"
	"forum/pkg/models/sqlite"
	"forum/utils"
)

const (
	host = "https://localhost"
	port = ":10443"
)

type application struct {
	templateCache map[string]*template.Template
	posts         *models.Post
	comments      *models.Comment
	users         *models.User
	session       *models.Session
	db            *sql.DB
}

func init() {
	// Initialize logger
	logger.InitLogger()

	// Initialize env
	env := utils.GetEnvironment()
	fmt.Println("Running in environment:", strings.ToUpper(env))

	configFile := "configs/config/" + env + ".env"
	file, err := utils.LoadConfigFile(configFile)
	if err != nil {
		logger.ErrorLogger.Println("Error opening config file:", err)
		os.Exit(1)
	}
	defer file.Close()

	err = utils.SetEnvironmentVariables(file)
	if err != nil {
		logger.ErrorLogger.Println("Error setting environment variables:", err)
		os.Exit(1)
	}

	configs.GoogleOauthConfig.ClientID = os.Getenv("GOOGLE_KEY")
	configs.GoogleOauthConfig.ClientSecret = os.Getenv("GOOGLE_SECRET")
}

func main() {
	// Load templates
	templateCache, err := newTemplateCache("./ui/html/")
	if err != nil {
		logger.ErrorLogger.Fatalf("Error loading template cache: %v", err)
	}

	// Connect to database
	app := &application{
		templateCache: templateCache,
		posts:         &models.Post{},
		comments:      &models.Comment{},
		users:         &models.User{},
		session:       &models.Session{},
	}

	app.db, err = sqlite.ConnectDB()
	if err != nil {
		logger.ErrorLogger.Fatalf("Error connecting to the database: %v", err)
	}
	defer app.db.Close()

	// configs.InsertDummyData(app.db)

	// Configure TLS
	tlsConfig := &tls.Config{
		PreferServerCipherSuites: true,
		CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	// Start server
	srv := &http.Server{
		Addr:           port,
		Handler:        wwwRedirect(app.routes()),
		IdleTimeout:    time.Minute,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		TLSConfig:      tlsConfig,
	}

	go func() {
		logger.ErrorLogger.Fatal(http.ListenAndServe(":8080", http.HandlerFunc(redirectHTTPS)))
	}()

	logger.InfoLogger.Printf("Starting application on port %v, to shut it down press CTRL + c\n", port)
	logger.InfoLogger.Printf("Open application: %v%v\n", host, port)

	fmt.Printf("Starting application on port %v, to shut it down press "+"\033[10;31m"+"CTRL + c\n"+"\033[0m", port)
	fmt.Println("Open application:", "\033[10;32m"+host+port+"\033[0m")

	if err := srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem"); err != nil {
		logger.ErrorLogger.Fatalf("Error starting server: %v", err)
	}
}
