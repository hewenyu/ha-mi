package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/boringsoft/ha-mi/internal/auth"
)

// AuthMiddleware authenticates requests using JWT
func AuthMiddleware(jwtService *auth.JWTService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Get the Authorization header
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		// Check if it's a Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			return
		}

		// Validate the token
		token := parts[1]
		claims, err := jwtService.ValidateToken(token)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token: " + err.Error()})
			return
		}

		// Check if it's an access token
		if claims.Type != auth.AccessToken {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token type"})
			return
		}

		// Store user information in the context
		ctx.Set("userId", claims.UserID)
		ctx.Set("email", claims.Email)
		ctx.Set("role", claims.Role)

		ctx.Next()
	}
}

// SecurityMiddleware validates request security parameters (timestamp, nonce, sign)
func SecurityMiddleware(nonceService *auth.NonceService, securityService *auth.SecurityService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Skip for some endpoints that handle their own security
		path := ctx.Request.URL.Path
		if strings.HasSuffix(path, "/auth/login") ||
			strings.HasSuffix(path, "/auth/refresh") ||
			strings.HasSuffix(path, "/auth/nonce") {
			ctx.Next()
			return
		}

		// Get timestamp from query or header
		timestamp := ctx.Query("timestamp")
		if timestamp == "" {
			timestamp = ctx.GetHeader("X-Timestamp")
		}

		if timestamp == "" {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Missing timestamp"})
			return
		}

		// Validate timestamp
		if err := securityService.ValidateTimestamp(timestamp); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get nonce from query or header
		nonce := ctx.Query("nonce")
		if nonce == "" {
			nonce = ctx.GetHeader("X-Nonce")
		}

		if nonce == "" {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Missing nonce"})
			return
		}

		// Validate nonce
		if err := nonceService.ValidateNonce(nonce); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get sign from query or header
		sign := ctx.Query("sign")
		if sign == "" {
			sign = ctx.GetHeader("X-Sign")
		}

		if sign == "" {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Missing signature"})
			return
		}

		// Extract parameters for signature validation
		params := make(map[string]string)

		// Extract parameters from query string
		for k, v := range ctx.Request.URL.Query() {
			if k != "sign" && len(v) > 0 {
				params[k] = v[0]
			}
		}

		// Add parameters from headers
		for _, k := range []string{"X-Timestamp", "X-Nonce"} {
			if v := ctx.GetHeader(k); v != "" {
				params[strings.ToLower(strings.TrimPrefix(k, "X-"))] = v
			}
		}

		// Parse form if it's a form request
		if ctx.Request.Method == "POST" || ctx.Request.Method == "PUT" {
			contentType := ctx.GetHeader("Content-Type")
			if strings.Contains(contentType, "application/x-www-form-urlencoded") ||
				strings.Contains(contentType, "multipart/form-data") {
				_ = ctx.Request.ParseForm()
				for k, v := range ctx.Request.Form {
					if k != "sign" && len(v) > 0 {
						params[k] = v[0]
					}
				}
			}
		}

		// Validate signature
		if err := securityService.ValidateSignature(params, sign); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx.Next()
	}
}
