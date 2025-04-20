package crypto

import (
	"context"
	"strings"

	"github.com/trranminhquang/go-boilerplate/pkg/utils"
	"golang.org/x/crypto/bcrypt"
)

type HashCost = int

const (
	// DefaultHashCost represents the default
	// hashing cost for any hashing algorithm.
	DefaultHashCost HashCost = iota

	// QuickHashCosts represents the quickest
	// hashing cost for any hashing algorithm,
	// useful for tests only.
	QuickHashCost HashCost = iota

	Argon2Prefix         = "$argon2"
	FirebaseScryptPrefix = "$fbscrypt"
	FirebaseScryptKeyLen = 32 // Firebase uses AES-256 which requires 32 byte keys: https://pkg.go.dev/golang.org/x/crypto/scrypt#Key
)

// PasswordHashCost is the current pasword hashing cost
// for all new hashes generated with
// GenerateHashFromPassword.
var PasswordHashCost = DefaultHashCost

// GenerateFromPassword generates a password hash from a
// password, using PasswordHashCost. Context can be used to cancel the hashing
// if the algorithm supports it.
func GenerateFromPassword(ctx context.Context, password string) (string, error) {
	hashCost := bcrypt.DefaultCost

	switch PasswordHashCost {
	case QuickHashCost:
		hashCost = bcrypt.MinCost
	}

	// attributes := []attribute.KeyValue{
	// 	attribute.String("alg", "bcrypt"),
	// 	attribute.Int("bcrypt_cost", hashCost),
	// }

	// generateFromPasswordSubmittedCounter.Add(ctx, 1, metric.WithAttributes(attributes...))
	// defer generateFromPasswordCompletedCounter.Add(ctx, 1, metric.WithAttributes(attributes...))

	hash := utils.Must(bcrypt.GenerateFromPassword([]byte(password), hashCost))

	return string(hash), nil
}

// GeneratePassword generates a random password of the specified length
// that contains at least one character from each of the required character sets.
func GeneratePassword(requiredChars []string, length int) string {
	passwordBuilder := strings.Builder{}
	passwordBuilder.Grow(length)

	// Add required characters
	for _, group := range requiredChars {
		if len(group) > 0 {
			randomIndex := secureRandomInt(len(group))

			passwordBuilder.WriteByte(group[randomIndex])
		}
	}

	// Define a default character set for random generation (if needed)
	const allChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// Fill the rest of the password
	for passwordBuilder.Len() < length {
		randomIndex := secureRandomInt(len(allChars))
		passwordBuilder.WriteByte(allChars[randomIndex])
	}

	// Convert to byte slice for shuffling
	passwordBytes := []byte(passwordBuilder.String())

	// Secure shuffling
	for i := len(passwordBytes) - 1; i > 0; i-- {
		j := secureRandomInt(i + 1)

		passwordBytes[i], passwordBytes[j] = passwordBytes[j], passwordBytes[i]
	}

	return string(passwordBytes)
}
