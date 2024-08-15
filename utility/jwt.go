package utility

import (
	"errors"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

type JWTAuthClaims struct {
	UserID      int32  `json:"user_id"`
	RoleLevel   int32  `json:"role_level"`
	ResidenceID *int32 `json:"residence_id"`
	jwt.RegisteredClaims
}

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func CreateJWTToken(jwtAuthClaims JWTAuthClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtAuthClaims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ParseJWTToken(tokenString string) (*JWTAuthClaims, error) {
	jwtAuthClaims := &JWTAuthClaims{}
	token, err := jwt.ParseWithClaims(tokenString, jwtAuthClaims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return jwtAuthClaims, nil
}
