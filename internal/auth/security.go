package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// SecurityService handles request security validation
type SecurityService struct {
	secretKey        string
	timestampMaxDiff int64 // Maximum allowed difference in seconds
}

// NewSecurityService creates a new SecurityService
func NewSecurityService(secretKey string, timestampMaxDiff int64) *SecurityService {
	return &SecurityService{
		secretKey:        secretKey,
		timestampMaxDiff: timestampMaxDiff,
	}
}

// ValidateTimestamp validates if the timestamp is within the allowed time window
func (s *SecurityService) ValidateTimestamp(timestamp string) error {
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return errors.New("invalid timestamp format")
	}

	// Convert to seconds if in milliseconds
	if ts > 1000000000000 {
		ts = ts / 1000
	}

	currentTime := time.Now().Unix()
	diff := currentTime - ts

	if diff < 0 {
		diff = -diff // Handle case where client time is ahead of server
	}

	if diff > s.timestampMaxDiff {
		return fmt.Errorf("timestamp expired, difference of %d seconds exceeds maximum allowed %d seconds", diff, s.timestampMaxDiff)
	}

	return nil
}

// GenerateSignature generates a HMAC-SHA256 signature for the given parameters
func (s *SecurityService) GenerateSignature(params map[string]string) string {
	// Sort parameters by key
	keys := make([]string, 0, len(params))
	for k := range params {
		if k != "sign" { // Exclude sign parameter
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// Build string to sign
	var sb strings.Builder
	for i, k := range keys {
		if i > 0 {
			sb.WriteString("&")
		}
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(params[k])
	}
	stringToSign := sb.String()

	// Generate HMAC-SHA256 signature
	h := hmac.New(sha256.New, []byte(s.secretKey))
	h.Write([]byte(stringToSign))
	signature := hex.EncodeToString(h.Sum(nil))

	return signature
}

// ValidateSignature validates the signature of a request
func (s *SecurityService) ValidateSignature(params map[string]string, providedSignature string) error {
	// Generate signature for comparison
	expectedSignature := s.GenerateSignature(params)

	// Compare signatures
	if !hmac.Equal([]byte(providedSignature), []byte(expectedSignature)) {
		return errors.New("invalid signature")
	}

	return nil
}

// ExtractParams extracts parameters from query string and form values
func ExtractParams(queryString string, formValues url.Values) map[string]string {
	params := make(map[string]string)

	// Extract from query string
	if queryString != "" {
		query, err := url.ParseQuery(queryString)
		if err == nil {
			for k, v := range query {
				if len(v) > 0 {
					params[k] = v[0]
				}
			}
		}
	}

	// Extract from form values
	for k, v := range formValues {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}

	return params
}
