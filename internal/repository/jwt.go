package repository

import (
	"api/pkg/errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTRepository struct {
	secret string
}

func NewJWTRepository(secret string) *JWTRepository {
	return &JWTRepository{secret: secret}
}

func (j *JWTRepository) GenerateToken(userID uint, isAdmin bool) (string, error) {
	if j.secret == "" {
		return "", errors.NewInternalError("JWT secret not configured", nil)
	}

	claims := jwt.MapClaims{
		"user_id":  userID,
		"is_admin": isAdmin,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(j.secret))
	if err != nil {
		return "", errors.NewInternalError("Failed to sign token", err)
	}

	return signedToken, nil
}

func (j *JWTRepository) ValidateToken(tokenStr string) (*jwt.Token, error) {
	if j.secret == "" {
		return nil, errors.NewInternalError("JWT secret not configured", nil)
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.NewUnauthorizedError("Invalid signing method", errors.ErrInvalidToken)
		}
		return []byte(j.secret), nil
	})

	if err != nil {
		return nil, errors.NewUnauthorizedError("Invalid token", err)
	}

	if !token.Valid {
		return nil, errors.NewUnauthorizedError("Token is not valid", errors.ErrInvalidToken)
	}

	return token, nil
}

func (j *JWTRepository) GetClaimsFromToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := j.ValidateToken(tokenStr)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.NewUnauthorizedError("Invalid token claims", errors.ErrInvalidToken)
	}

	return claims, nil
}
