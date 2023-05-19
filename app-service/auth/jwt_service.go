package auth

import (
	"fmt"
	"sonamusica-backend/app-service/identity"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

const (
	JWTToken_ExpiryTime_SetDefault = 0
)

type MainJWTClaims struct {
	UserID        identity.UserID            `json:"userId"`
	PrivilegeType identity.UserPrivilegeType `json:"privilegeType"`
	PurposeType   JWTTokenPurposeType        `json:"purposeType"`
	jwt.RegisteredClaims
}

type JWTTokenPurposeType int

const (
	JWTTokenPurposeType_None          JWTTokenPurposeType = iota
	JWTTokenPurposeType_Auth          JWTTokenPurposeType = 1
	JWTTokenPurposeType_ResetPassword JWTTokenPurposeType = 2
)

type JWTServiceConfig struct {
	SecretKey       []byte
	TokenExpiration time.Duration
}

type JWTService interface {
	CreateJWTToken(userID identity.UserID, privilegeType identity.UserPrivilegeType, tokenPurposeType JWTTokenPurposeType, expireAfter time.Duration) (string, error)
	VerifyTokenStringAndReturnClaims(tokenString string) (interface{}, error)
}

type jwtServiceImpl struct {
	config JWTServiceConfig
}

var _ JWTService = (*jwtServiceImpl)(nil)

func NewJWTServiceImpl(config JWTServiceConfig) *jwtServiceImpl {
	return &jwtServiceImpl{config: config}
}

func (s jwtServiceImpl) CreateJWTToken(userID identity.UserID, privilegeType identity.UserPrivilegeType, tokenPurposeType JWTTokenPurposeType, expireAfter time.Duration) (string, error) {
	exp := expireAfter
	if expireAfter == JWTToken_ExpiryTime_SetDefault {
		exp = s.config.TokenExpiration
	}

	purposeType := JWTTokenPurposeType_Auth
	if purposeType != JWTTokenPurposeType_None {
		purposeType = tokenPurposeType
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, MainJWTClaims{
		userID,
		privilegeType,
		tokenPurposeType,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(exp)),
		},
	})

	// Sign the token with the secret key
	tokenString, err := token.SignedString(s.config.SecretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s jwtServiceImpl) VerifyTokenStringAndReturnClaims(tokenString string) (interface{}, error) {
	token, err := jwt.ParseWithClaims(tokenString, &MainJWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return s.config.SecretKey, nil
	})
	if err != nil {
		return nil, err
	}

	// all invalid tokens will always returns error. so this code should be unreachable, and only used as a defensive code
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return token.Claims, nil
}
