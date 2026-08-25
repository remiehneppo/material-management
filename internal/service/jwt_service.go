package service

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/remiehneppo/material-management/types"
)

type jwtAccessClaims struct {
	UserId          string `json:"user_id"`
	Username        string `json:"username"`
	ManagementLevel int    `json:"management_level"`
	WorkspaceRole   string `json:"workspace_role"`
	Workspace       string `json:"workspace"`
	SessionID       string `json:"session_id"`
	jwt.RegisteredClaims
}

type JWTService struct {
	secretKey string
	issuer    string
	exp       int64
}

func NewJWTService(secretKey, issuer string, exp int64) *JWTService {
	return &JWTService{
		secretKey: secretKey,
		issuer:    issuer,
		exp:       exp,
	}
}

func (j *JWTService) GenerateAccessToken(user *types.User) (string, error) {
	return j.GenerateAccessTokenForSession(user, user.SessionID)
}

func (j *JWTService) GenerateAccessTokenForSession(user *types.User, sessionID string) (string, error) {
	expiresIn := 6 * time.Hour
	if j.exp > 0 {
		expiresIn = time.Duration(j.exp) * time.Second
	}
	claims := &jwtAccessClaims{
		UserId:        user.ID,
		Username:      user.Username,
		WorkspaceRole: user.WorkspaceRole,
		Workspace:     user.Workspace,
		SessionID:     sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func (j *JWTService) ValidateAccessToken(token string) (*types.User, error) {

	parsedToken, err := jwt.ParseWithClaims(token, &jwtAccessClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.secretKey), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithIssuer(j.issuer))
	if err != nil {
		return nil, err
	}
	if claims, ok := parsedToken.Claims.(*jwtAccessClaims); ok && parsedToken.Valid {
		return &types.User{
			Username:      claims.Username,
			ID:            claims.Subject,
			WorkspaceRole: claims.WorkspaceRole,
			Workspace:     claims.Workspace,
			SessionID:     claims.SessionID,
		}, nil
	}
	return nil, jwt.ErrInvalidKey
}
