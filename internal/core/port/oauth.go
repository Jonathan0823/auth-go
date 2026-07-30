package port

import "net/http"

type OAuthProfile struct {
	UserID    string
	Email     string
	Name      string
	Provider  string
	AvatarURL string
}

type OAuthClient interface {
	BeginAuth(w http.ResponseWriter, r *http.Request, provider string)
	CompleteAuth(w http.ResponseWriter, r *http.Request, provider string) (OAuthProfile, error)
}
