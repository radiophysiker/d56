package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/radiophysiker/d56/internal/domain/user"
)

const (
	TokenExpiration = 24 * time.Hour
)

type Claims struct {
	UserID user.UserID `json:"user_id"`
	jwt.RegisteredClaims
}

type Service struct {
	signingKey []byte
}

func NewService(jwtSecretKey string) (*Service, error) {
	if jwtSecretKey == "" {
		return nil, fmt.Errorf("JWT secret key cannot be empty")
	}

	return &Service{
		signingKey: []byte(jwtSecretKey),
	}, nil
}

func (s *Service) GenerateToken(userID user.UserID) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.signingKey)
}

func (s *Service) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.signingKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}
