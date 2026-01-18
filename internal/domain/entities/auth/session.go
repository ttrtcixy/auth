package entities

import "time"

type Session struct {
	UserID           int
	ClientUUID       string
	RefreshTokenUUID string
	ExpiresAt        time.Time
}

type UpdateSessionRequest struct {
	ClientUUID          string
	OldRefreshTokenUUID string
	NewRefreshTokenUUID string
	ExpiresAt           time.Time
}

type CreateSessionResponse struct {
	RefreshToken string
	ClientUUID   string
}

type TokenUserInfo struct {
	ID       string
	Username string
	Email    string
	RoleID   string
}
