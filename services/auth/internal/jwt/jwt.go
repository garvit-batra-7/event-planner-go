package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

type JWTClaims struct {
	UserID int `json:"userId"`
	jwt.StandardClaims
}

// GenerateToken creates a JWT token with proper claims
func GenerateToken(userID int, secret string, expiry time.Duration) (string, error) {
	// Create claims with STANDARD fields
	claims := JWTClaims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(expiry).Unix(),
			IssuedAt:  time.Now().Unix(),
			NotBefore: time.Now().Unix(),         
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken verifies a token and returns the user ID
func ValidateToken(tokenString, secret string) (int, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return 0, fmt.Errorf("failed to parse token: %w", err)
	}

	// Extract claims
	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		// Check if token is expired (automatic with StandardClaims)
		if claims.VerifyExpiresAt(time.Now().Unix(), true) {
			return claims.UserID, nil
		}
		return 0, fmt.Errorf("token expired")
	}

	return 0, fmt.Errorf("invalid token")
}