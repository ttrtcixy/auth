package entities

type SigninRequest struct {
	Username string
	Email    string
	Password string
}

type SigninResponse struct {
	AccessToken  string
	RefreshToken string
	ClientUUID   string
}
