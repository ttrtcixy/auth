package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	entities "github.com/ttrtcixy/users/internal/domain/entities/auth"
	apperrors "github.com/ttrtcixy/users/internal/domain/errors"
)

type Config struct {
	JwtSecret        string
	EmailJwtExpiry   time.Duration `env:"JWT_EMAIL_TOKEN_EXPIRE"`
	AccessJwtExpiry  time.Duration `env:"JWT_ACCESS_TOKEN_EXPIRE"`
	RefreshJwtExpiry time.Duration `env:"JWT_REFRESH_TOKEN_EXPIRE"`
}

type JwtTokenService struct {
	cfg *Config
}

func New(cfg *Config) *JwtTokenService {
	return &JwtTokenService{
		cfg,
	}
}

type EmailVerificationClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

type UserClaims struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	RoleId   string `json:"role_id"`
	jwt.RegisteredClaims
}

// ParseVerificationToken - UserError: ErrEmailTokenExpired, ErrInvalidEmailToken, ErrInvalidVerificationToken
func (t *JwtTokenService) ParseVerificationToken(jwtToken string) (email string, err error) {
	const op = "JwtTokenService.ParseVerificationToken"
	// parse and validate token
	token, err := jwt.ParseWithClaims(jwtToken, &EmailVerificationClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(t.cfg.JwtSecret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", apperrors.ErrEmailTokenExpired
		}
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return "", apperrors.ErrInvalidEmailVerifyToken
		}
		return "", apperrors.Wrap(op, err)
	}

	// get user email to activate account
	claims, ok := token.Claims.(*EmailVerificationClaims)
	if !ok {
		return "", apperrors.ErrInvalidEmailVerifyToken
	}

	if claims.Email == "" {
		return "", apperrors.ErrInvalidEmailVerifyToken
	}

	return claims.Email, nil
}

// AccessToken - generate access token with user username, email, roleId.
func (t *JwtTokenService) AccessToken(user *entities.TokenUserInfo) (token string, err error) {
	const op = "JwtTokenService.AccessToken"

	exp := time.Now().Add(t.cfg.AccessJwtExpiry)

	claims := &UserClaims{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		RoleId:   user.RoleID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			Issuer:    "auth_grpc_app",
		},
	}

	if token, err = t.jwt(t.cfg.JwtSecret, claims); err != nil {
		return "", apperrors.Wrap(op, err)
	}

	return token, nil
}

func (t *JwtTokenService) ParseAccessToken(jwtToken string) (result *entities.TokenUserInfo, err error) {
	const op = "JwtTokenService.ParseAccessToken"
	token, err := jwt.ParseWithClaims(jwtToken, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			// todo usererr
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(t.cfg.JwtSecret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, apperrors.ErrAccessTokenExpired
		}
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return nil, apperrors.ErrInvalidAccessToken
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, apperrors.ErrInvalidAccessToken
		}
		return nil, apperrors.Wrap(op, err)
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return nil, apperrors.ErrInvalidAccessToken
	}

	userId := claims.ID
	if userId == "" {
		return nil, apperrors.ErrInvalidAccessToken
	}

	username := claims.Username
	if username == "" {
		return nil, apperrors.ErrInvalidAccessToken
	}

	email := claims.Email
	if email == "" {
		return nil, apperrors.ErrInvalidAccessToken
	}

	roleID := claims.RoleId

	result = &entities.TokenUserInfo{
		ID:       userId,
		Username: username,
		Email:    email,
		RoleID:   roleID,
	}

	return result, nil
}

type RefreshTokenClaims struct {
	ClientID string `json:"client_id"`
	jwt.RegisteredClaims
}

// RefreshToken - generate jwt Token.
func (t *JwtTokenService) RefreshToken(clientID, tokenID string, exp time.Time) (token string, err error) {
	const op = "JwtTokenService.RefreshToken"

	claims := &RefreshTokenClaims{
		ClientID: clientID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "auth_grpc_app",
			ID:        tokenID,
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}

	if token, err = t.jwt(t.cfg.JwtSecret, claims); err != nil {
		return "", apperrors.Wrap(op, err)
	}

	return token, nil
}

// ParseRefreshToken - parse refresh token generated with RefreshToken()
func (t *JwtTokenService) ParseRefreshToken(jwtToken string) (clientID, jtl string, err error) {
	const op = "JwtTokenService.ParseRefreshToken"

	token, err := jwt.ParseWithClaims(jwtToken, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			// todo
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(t.cfg.JwtSecret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", "", apperrors.ErrRefreshTokenExpired
		}
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			return "", "", apperrors.ErrInvalidRefreshToken
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return "", "", apperrors.ErrInvalidRefreshToken
		}
		return "", "", apperrors.Wrap(op, err)
	}

	claims, ok := token.Claims.(*RefreshTokenClaims)
	if !ok {
		return "", "", apperrors.ErrInvalidRefreshToken
	}

	if claims.ClientID == "" {
		return "", "", apperrors.ErrInvalidRefreshToken
	}

	if claims.ID == "" {
		return "", "", apperrors.ErrInvalidRefreshToken
	}

	return claims.ClientID, claims.ID, nil
}

// EmailVerificationToken - generate jwtToken with user email to verify.
func (t *JwtTokenService) EmailVerificationToken(email string) (token string, err error) {
	const op = "JwtTokenService.EmailVerificationToken"

	expAt := time.Now().Add(t.cfg.EmailJwtExpiry)

	claims := &EmailVerificationClaims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "auth_grpc_app",
			ExpiresAt: jwt.NewNumericDate(expAt),
		},
	}

	if token, err = t.jwt(t.cfg.JwtSecret, claims); err != nil {
		return "", apperrors.Wrap(op, err)
	}

	return token, nil
}

func (t *JwtTokenService) jwt(secret string, claims jwt.Claims) (token string, err error) {
	const op = "JwtTokenService.jwt"
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	if token, err = jwtToken.SignedString([]byte(secret)); err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}
