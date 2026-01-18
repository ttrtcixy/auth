package entities

type EmailVerifyRequest struct {
	EmailToken string
}

type EmailVerifyResponse struct {
	AccessToken  string
	RefreshToken string
	ClientUUID   string
}
