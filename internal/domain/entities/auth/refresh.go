package entities

type RefreshRequest struct {
	RefreshToken string
}

type RefreshResponse struct {
	AccessToken  string
	RefreshToken string
	ClientID     string
	*TokenUserInfo
}
