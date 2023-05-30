package configs

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleUser struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture,omitempty"`
	Locale        string `json:"locale,omitempty"`
}

const (
	googleRedirectURL    = "https://localhost:8080/GoogleCallback"
	googleAPIEndpoint    = "https://www.googleapis.com/oauth2/v2/userinfo"
	googleAccessTokenURL = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="
	googleRevokeTokenURL = "https://oauth2.googleapis.com/revoke"
	scopeProfile         = "https://www.googleapis.com/auth/userinfo.profile"
	scopeEmail           = "https://www.googleapis.com/auth/userinfo.email"
	oauthStateString     = "random"
)

var (
	GoogleOauthConfig = &oauth2.Config{
		RedirectURL: googleRedirectURL,
		Scopes:      []string{scopeProfile, scopeEmail},
		Endpoint:    google.Endpoint,
	}
	OauthStateString = oauthStateString
)

func GetGoogleAccessToken(r *http.Request, logger *log.Logger) (*oauth2.Token, error) {
	code := r.FormValue("code")
	token, err := GoogleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		logger.Printf("failed to exchange token: %v", err)
		return nil, err
	}
	return token, nil
}

func GetGoogleData(token *oauth2.Token, logger *log.Logger) (*GoogleUser, error) {
	response, err := http.Get(googleAccessTokenURL + token.AccessToken)
	if err != nil {
		logger.Printf("failed to get user info: %v", err)
		return nil, err
	}
	defer response.Body.Close()

	contents, err := io.ReadAll(response.Body)
	if err != nil {
		logger.Printf("failed to read response body: %v", err)
		return nil, err
	}

	var GoogleUser GoogleUser
	err = json.Unmarshal(contents, &GoogleUser)
	if err != nil {
		logger.Printf("failed to unmarshal JSON data: %v", err)
		return nil, err
	}
	return &GoogleUser, nil
}

func RevokeGoogleToken(token string, logger *log.Logger) error {
	req, err := http.NewRequest("POST", googleRevokeTokenURL, nil)
	if err != nil {
		logger.Printf("failed to create new request: %v", err)
		return err
	}

	q := req.URL.Query()
	q.Add("token", token)
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Printf("failed to revoke token: %v", err)
		return err
	}
	defer resp.Body.Close()

	return nil
}
