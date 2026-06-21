// Package token issues and validates the JWT access tokens used for
// authentication. It is deliberately small: a Manager holds the signing secret
// and token lifetime, and exposes Generate/Parse.
package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// ErrInvalidToken is returned when a token is malformed, expired, or signed
// with the wrong key.
var ErrInvalidToken = errors.New("invalid or expired token")

// Claims is the payload carried by an access token. UserID is the subject of
// the token; Email is included for convenience so handlers don't have to hit
// the database just to read it.
type Claims struct {
	UserID uuid.UUID `json:"uid"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}

// Manager signs and verifies tokens with a shared HMAC secret.
type Manager struct {
	secret []byte
	expiry time.Duration
}

// NewManager builds a Manager. secret must be kept private; expiry is how long
// an issued token stays valid.
func NewManager(secret string, expiry time.Duration) *Manager {
	return &Manager{secret: []byte(secret), expiry: expiry}
}

// Generate signs a new access token for the given user.
func (m *Manager) Generate(userID uuid.UUID, email string) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.expiry)),
		},
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString(m.secret)
}

// Parse verifies a token string and returns its claims. It rejects tokens
// signed with any method other than HMAC.
func (m *Manager) Parse(tokenString string) (*Claims, error) {
	claims := &Claims{}
	tok, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	})
	if err != nil || !tok.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
