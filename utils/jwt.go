package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	AccessTokenTTL  = time.Hour * 7 * 24 // 7 days
	RefreshTokenTTL = time.Hour * 24 * 1 // 1 day
)
type Claims struct {
	UserID string `json:"user_id"`
	Type   string `json:"type"`
	jwt.RegisteredClaims
}

func GenerateTokens(userID string) (string, string, error) {
	jwtKey := os.Getenv("JWT_SECRET")
	if jwtKey == "" {
		fmt.Println("JWT_SECRET not set in .env")
		return "", "", fmt.Errorf("JWT_SECRET not set")
	}

	if userID == "" {
		fmt.Println("Error: userID is empty")
		return "", "", fmt.Errorf("userID cannot be empty")
	}
	fmt.Printf("Generating tokens for userID: %s\n", userID)

	accessClaims := &Claims{
		UserID: userID,
		Type:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(jwtKey))
	if err != nil {
		fmt.Printf("Error signing access token: %v\n", err)
		return "", "", fmt.Errorf("failed to sign access token: %v", err)
	}
	fmt.Println("Generated Access Token:", accessTokenString)

	refreshClaims := &Claims{
		UserID: userID,
		Type:   "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(RefreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(jwtKey))
	if err != nil {
		fmt.Printf("Error signing refresh token: %v\n", err)
		return "", "", fmt.Errorf("failed to sign refresh token: %v", err)
	}
	fmt.Println("Generated Refresh Token:", refreshTokenString)

	return accessTokenString, refreshTokenString, nil
}

func RefreshAccessToken(refreshTokenString string) (string, error) {
	jwtKey := os.Getenv("JWT_SECRET")
	if jwtKey == "" {
		fmt.Println("JWT_SECRET not set in .env")
		return "", fmt.Errorf("JWT_SECRET not set")
	}

	claims := &Claims{}

	// Parse and validate the token
	token, err := jwt.ParseWithClaims(refreshTokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtKey), nil
	})
	if err != nil {
		fmt.Printf("Error parsing refresh token: %v\n", err)
		if err == jwt.ErrTokenMalformed {
			return "", fmt.Errorf("malformed token")
		} else if err == jwt.ErrTokenExpired {
			return "", fmt.Errorf("token has expired")
		} else if err == jwt.ErrTokenSignatureInvalid {
			return "", fmt.Errorf("invalid token signature")
		}
		return "", fmt.Errorf("failed to parse token: %v", err)
	}

	if !token.Valid {
		fmt.Println("Token is not valid")
		return "", fmt.Errorf("token is not valid")
	}

	if claims.Type != "refresh" {
		fmt.Println("Provided token is not a refresh token")
		return "", fmt.Errorf("provided token is not a refresh token")
	}

	if claims.UserID == "" {
		fmt.Println("Error: userID is empty in refresh token claims")
		return "", fmt.Errorf("userID cannot be empty")
	}

	accessClaims := &Claims{
		UserID: claims.UserID,
		Type:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)

	tokenString, err := accessToken.SignedString([]byte(jwtKey))
	if err != nil {
		fmt.Printf("Error signing new access token: %v\n", err)
		return "", fmt.Errorf("failed to sign access token: %v", err)
	}
	fmt.Println("Generated New Access Token:", tokenString)

	return tokenString, nil
}