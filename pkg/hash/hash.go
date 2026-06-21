// Package hash wraps bcrypt so the rest of the codebase never touches the
// crypto library directly. Passwords are hashed before they reach the database
// and verified on login.
package hash

import "golang.org/x/crypto/bcrypt"

// Cost is the bcrypt work factor. Higher is slower (and safer); the default of
// 10 is a sensible balance for an API.
const Cost = bcrypt.DefaultCost

// Password hashes a plaintext password. The returned string embeds the salt and
// cost, so it is safe to store as-is.
func Password(plaintext string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plaintext), Cost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Compare reports whether plaintext matches a previously hashed password.
func Compare(hashed, plaintext string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plaintext)) == nil
}
