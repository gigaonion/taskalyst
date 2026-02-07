package password

import (
	"fmt"

	"github.com/alexedwards/argon2id"
)

func Hash(plain string) (string, error) {
	hash, err := argon2id.CreateHash(plain, argon2id.DefaultParams)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return hash, nil
}

func Check(plain, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(plain, hash)
	if err != nil {
		return false, fmt.Errorf("failed to compare password and hash: %w", err)
	}
	return match, nil
}
