package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/boringsoft/ha-mi/internal/auth"
	"github.com/boringsoft/ha-mi/internal/config"
)

// AuthController handles authentication-related requests
type AuthController struct {
	jwtService      *auth.JWTService
	nonceService    *auth.NonceService
	securityService *auth.SecurityService
	config          *config.Config
}

// NewAuthController creates a new AuthController
func NewAuthController(jwtService *auth.JWTService, nonceService *auth.NonceService, securityService *auth.SecurityService, config *config.Config) *AuthController {
	return &AuthController{
		jwtService:      jwtService,
		nonceService:    nonceService,
		securityService: securityService,
		config:          config,
	}
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	Nonce     string `json:"nonce" binding:"required"`
	Timestamp string `json:"timestamp" binding:"required"`
	Sign      string `json:"sign" binding:"required"`
}

// TokenResponse represents the token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// RefreshRequest represents the refresh token request body
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
	Nonce        string `json:"nonce" binding:"required"`
	Timestamp    string `json:"timestamp" binding:"required"`
	Sign         string `json:"sign" binding:"required"`
}

// NonceResponse represents the nonce response
type NonceResponse struct {
	Nonce string `json:"nonce"`
}

// Login handles the login request
func (c *AuthController) Login(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Validate timestamp
	if err := c.securityService.ValidateTimestamp(req.Timestamp); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate nonce
	if err := c.nonceService.ValidateNonce(req.Nonce); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Extract parameters for signature validation
	params := map[string]string{
		"username":  req.Username,
		"password":  req.Password,
		"nonce":     req.Nonce,
		"timestamp": req.Timestamp,
	}

	// Validate signature
	if err := c.securityService.ValidateSignature(params, req.Sign); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate credentials
	if req.Username != c.config.Auth.User || req.Password != c.config.Auth.Password {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate tokens
	accessToken, refreshToken, err := c.jwtService.GenerateTokens(uuid.New().String(), req.Username, "admin")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	// Return tokens
	ctx.JSON(http.StatusOK, TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(c.config.Auth.AccessTokenExpiry.Seconds()),
		TokenType:    "Bearer",
	})
}

// Refresh handles the refresh token request
func (c *AuthController) Refresh(ctx *gin.Context) {
	var req RefreshRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Validate timestamp
	if err := c.securityService.ValidateTimestamp(req.Timestamp); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate nonce
	if err := c.nonceService.ValidateNonce(req.Nonce); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Extract parameters for signature validation
	params := map[string]string{
		"refresh_token": req.RefreshToken,
		"nonce":         req.Nonce,
		"timestamp":     req.Timestamp,
	}

	// Validate signature
	if err := c.securityService.ValidateSignature(params, req.Sign); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Refresh tokens
	accessToken, refreshToken, err := c.jwtService.RefreshTokens(req.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token: " + err.Error()})
		return
	}

	// Return tokens
	ctx.JSON(http.StatusOK, TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(c.config.Auth.AccessTokenExpiry.Seconds()),
		TokenType:    "Bearer",
	})
}

// GetNonce handles the nonce request
func (c *AuthController) GetNonce(ctx *gin.Context) {
	// Get timestamp from query or header
	timestamp := ctx.Query("timestamp")
	if timestamp == "" {
		timestamp = ctx.GetHeader("X-Timestamp")
	}

	if timestamp == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing timestamp"})
		return
	}

	// Validate timestamp
	if err := c.securityService.ValidateTimestamp(timestamp); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate nonce
	nonce, err := c.nonceService.GenerateNonce()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate nonce: " + err.Error()})
		return
	}

	// Return nonce
	ctx.JSON(http.StatusOK, NonceResponse{
		Nonce: nonce,
	})
}

// RegisterRoutes registers the auth routes
func (c *AuthController) RegisterRoutes(router *gin.RouterGroup) {
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/login", c.Login)
		authGroup.POST("/refresh", c.Refresh)
		authGroup.GET("/nonce", c.GetNonce)
	}
}
