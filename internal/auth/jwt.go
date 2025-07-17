package auth

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject:   userID.String(),
	})

	signed, err := token.SignedString([]byte(tokenSecret))

	if err != nil {
		log.Println("auth: failed to sign token")
		return "", err
	}

	return signed, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})
	date, _ := token.Claims.GetExpirationTime()
	log.Printf("Expiration time: %v", date)
	if err != nil {
		log.Println("validateJwt: failed to parse claims. invalid or expired", err)
		return uuid.Nil, err
	}

	sub, err := token.Claims.GetSubject()

	if err != nil {
		log.Println("auth: failed to sign token")
		return uuid.Nil, err
	}

	return uuid.MustParse(sub), nil
}

func MakeRefreshToken() (string, error) {
	randomBytes := make([]byte, 32)
	// Fill random bytes with random bytes. never returns an error
	if _, err := rand.Read(randomBytes); err != nil {
		panic(err)
	}

	return hex.EncodeToString(randomBytes), nil
}
