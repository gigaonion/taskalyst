package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

func GeneratePAT() (token string, hash string, err error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}
	
	rawToken := hex.EncodeToString(bytes)
	userToken := "pat_" + rawToken

	hashBytes := sha256.Sum256([]byte(userToken))
	tokenHash := hex.EncodeToString(hashBytes[:])

	return userToken, tokenHash, nil
}

func HashPAT(rawToken string) string {
	hashBytes := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(hashBytes[:])
}
