package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/mattn/go-sqlite3"
)

// NonceService handles nonce operations
type NonceService struct {
	db          *sql.DB
	nonceExpiry time.Duration
}

// NewNonceService creates a new NonceService
func NewNonceService(db *sql.DB, nonceExpiry time.Duration) *NonceService {
	return &NonceService{
		db:          db,
		nonceExpiry: nonceExpiry,
	}
}

// GenerateNonce generates a random nonce and stores it in the database
func (s *NonceService) GenerateNonce() (string, error) {
	// Generate random bytes
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Convert to hex string
	nonce := hex.EncodeToString(bytes)

	// Calculate expiry time
	expiresAt := time.Now().Add(s.nonceExpiry).Unix()

	// Store in database
	_, err := s.db.Exec("INSERT INTO nonces (nonce, expires_at) VALUES (?, ?)", nonce, expiresAt)
	if err != nil {
		// Check for unique constraint violation
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.Code == sqlite3.ErrConstraint {
			return "", errors.New("failed to generate unique nonce, please try again")
		}
		return "", fmt.Errorf("failed to store nonce: %w", err)
	}

	return nonce, nil
}

// ValidateNonce checks if a nonce is valid and not expired
func (s *NonceService) ValidateNonce(nonce string) error {
	// Check if nonce exists and not expired
	var expiresAt int64
	err := s.db.QueryRow("SELECT expires_at FROM nonces WHERE nonce = ?", nonce).Scan(&expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("invalid nonce")
		}
		return fmt.Errorf("error querying nonce: %w", err)
	}

	// Check if expired
	if time.Now().Unix() > expiresAt {
		return errors.New("expired nonce")
	}

	// Delete nonce to prevent reuse
	_, err = s.db.Exec("DELETE FROM nonces WHERE nonce = ?", nonce)
	if err != nil {
		return fmt.Errorf("error deleting used nonce: %w", err)
	}

	return nil
}

// CleanupExpiredNonces removes expired nonces from the database
func (s *NonceService) CleanupExpiredNonces() error {
	_, err := s.db.Exec("DELETE FROM nonces WHERE expires_at < ?", time.Now().Unix())
	if err != nil {
		return fmt.Errorf("error cleaning up expired nonces: %w", err)
	}
	return nil
}
