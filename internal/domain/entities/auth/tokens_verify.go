package entities

type TokensVerifyRequest struct {
	AccessToken  string
	RefreshToken string
}

type TokensVerifyResponse struct {
	*Tokens
	*TokenUserInfo
}

type Tokens struct {
	AccessToken  *string
	RefreshToken *string
}
