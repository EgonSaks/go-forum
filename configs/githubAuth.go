package configs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"forum/logger"
)

type GithubData struct {
	UserInfo  UserInfo    `json:"user_info"`
	EmailInfo []EmailInfo `json:"email_info,omitempty"`
}

type UserInfo struct {
	AvatarURL string `json:"avatar_url"`
	Name      string `json:"name"`
	Email     string `json:"email"`
}

type EmailInfo struct {
	Email      string `json:"email"`
	Primary    bool   `json:"primary"`
	Verified   bool   `json:"verified"`
	Visibility string `json:"visibility"`
}

const (
	githubAccessTokenURL = "https://github.com/login/oauth/access_token"
	githubAPIUserURL     = "https://api.github.com/user"
	githubAPIEmailsURL   = "https://api.github.com/user/emails"
	contentTypeJSON      = "application/json"
	acceptTypeJSON       = "application/json"
)

func GetGithubAccessToken(code string) string {
	clientID := os.Getenv("GITHUB_KEY")
	clientSecret := os.Getenv("GITHUB_SECRET")

	// Set us the request body as JSON
	requestBodyMap := map[string]string{
		"client_id":     clientID,
		"client_secret": clientSecret,
		"code":          code,
	}
	requestJSON, _ := json.Marshal(requestBodyMap)

	// POST request to set URL
	req, reqerr := http.NewRequest("POST", githubAccessTokenURL, bytes.NewBuffer(requestJSON))
	if reqerr != nil {
		logger.ErrorLogger.Fatalf("Request creation failed: %s", reqerr)
	}
	req.Header.Set("Content-Type", contentTypeJSON)
	req.Header.Set("Accept", acceptTypeJSON)

	// Get the response
	resp, resperr := http.DefaultClient.Do(req)
	if resperr != nil {
		logger.ErrorLogger.Fatalf("Request failed: %s", resperr)
	}

	// Response body converted to stringified JSON
	respbody, _ := ioutil.ReadAll(resp.Body)

	// Represents the response received from Github
	type githubAccessTokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
	}

	// Convert stringified JSON to a struct object of type githubAccessTokenResponse
	var ghresp githubAccessTokenResponse
	json.Unmarshal(respbody, &ghresp)

	logger.InfoLogger.Println("Got Github access token")

	// Return the access token (as the rest of the
	// details are relatively unnecessary for us)
	return ghresp.AccessToken
}

func GetGithubData(accessToken string) ([]byte, error) {
	// Get request to a set URL
	req, reqerr := http.NewRequest("GET", githubAPIUserURL, nil)
	if reqerr != nil {
		logger.ErrorLogger.Fatal("API Request creation failed:", reqerr)
	}

	// Set the Authorization header before sending the request
	// Authorization: token XXXXXXXXXXXXXXXXXXXXXXXXXXX
	authorizationHeaderValue := fmt.Sprintf("token %s", accessToken)
	req.Header.Set("Authorization", authorizationHeaderValue)

	// Make the request
	resp, resperr := http.DefaultClient.Do(req)
	if resperr != nil {
		logger.ErrorLogger.Fatal("Request failed:", resperr)
	}

	// Read the response as a byte slice
	respbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.ErrorLogger.Fatal("Error reading response body:", err)
	}

	// Make a new request to retrieve the user's emails
	req, reqerr = http.NewRequest("GET", githubAPIEmailsURL, nil)
	if reqerr != nil {
		logger.ErrorLogger.Fatal("API Request creation failed:", reqerr)
	}

	// Set the Authorization header before sending the request
	req.Header.Set("Authorization", authorizationHeaderValue)

	// Make the request
	resp, resperr = http.DefaultClient.Do(req)
	if resperr != nil {
		logger.ErrorLogger.Fatal("Request failed:", resperr)
	}

	// Read the response as a byte slice
	emailResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.ErrorLogger.Fatal("Error reading response body:", err)
	}

	// Combine the user's basic info and email info into a single JSON object
	// and return as a byte slice
	combinedData := fmt.Sprintf(`{"user_info": %s, "email_info": %s}`, respbody, emailResp)
	logger.InfoLogger.Println("Successfully retrieved data from GitHub API.")
	return []byte(combinedData), nil
}
