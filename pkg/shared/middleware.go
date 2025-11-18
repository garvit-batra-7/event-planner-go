package shared

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

// AuthMiddleware validates JWT tokens and stores userID in context.
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	logger := GetLogger()

	return func(c *gin.Context) {
		logger.Debug("Auth middleware ENTER",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
		)

		// 1. Check if Authorization header exists
		authHeader := c.GetHeader("Authorization")
		logger.Debug("Auth header read", "raw_header", authHeader)

		if authHeader == "" {
			logger.Warn("Missing Authorization header",
				"path", c.Request.URL.Path,
				"ip", c.ClientIP(),
			)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		// 2. Extract Bearer token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			logger.Warn("Invalid Authorization header format (missing 'Bearer ')",
				"path", c.Request.URL.Path,
				"ip", c.ClientIP(),
			)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer token is required"})
			c.Abort()
			return
		}
		logger.Debug("Token extracted from header", "token_prefix", tokenString[:min(10, len(tokenString))])

		// 3. Parse and validate JWT token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Verify signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			logger.Warn("Token parsing failed",
				"error", err.Error(),
				"path", c.Request.URL.Path,
				"ip", c.ClientIP(),
			)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		if !token.Valid {
			logger.Warn("Invalid token (token.Valid=false)",
				"path", c.Request.URL.Path,
				"ip", c.ClientIP(),
			)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// 4. Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			logger.Error("Failed to cast token claims to MapClaims",
				"path", c.Request.URL.Path,
			)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}
		logger.Debug("Token claims extracted", "claims", claims)

		// 5. Extract user ID from claims
		userIDFloat, ok := claims["userId"].(float64)
		if !ok {
			logger.Error("Invalid or missing userId in token claims",
				"path", c.Request.URL.Path,
			)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid userId in token"})
			c.Abort()
			return
		}
		userID := int(userIDFloat)

		logger.LogUser(LevelDebug, fmt.Sprintf("%d", userID), "authenticated",
			"User authenticated successfully, storing in context",
			"path", c.Request.URL.Path,
		)

		// IMPORTANT: store userID (int), not user struct
		c.Set("userID", userID)

		logger.Debug("Auth middleware EXIT - userID set in context",
			"userID", userID,
		)

		c.Next()
	}
}

// small helper just to avoid slicing panic in logs
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetUserFromContext returns the userID from context (for backward compatibility)
func GetUserFromContext(c *gin.Context) (interface{}, bool) {
	return c.Get("userID")
}

// Helper with proper type + logs
func GetUserIDFromContext(c *gin.Context) (int, bool) {
	logger := GetLogger()

	v, ok := c.Get("userID")
	if !ok {
		logger.Warn("GetUserIDFromContext: userID not found in context",
			"path", c.Request.URL.Path,
		)
		return 0, false
	}

	id, ok := v.(int)
	if !ok {
		logger.Error("GetUserIDFromContext: userID in context is not int",
			"type", fmt.Sprintf("%T", v),
			"path", c.Request.URL.Path,
		)
		return 0, false
	}

	logger.Debug("GetUserIDFromContext: userID retrieved from context",
		"userID", id,
		"path", c.Request.URL.Path,
	)
	return id, true
}

func GetRawTokenFromContext(c *gin.Context) (string, bool) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", false
	}
	return authHeader, true
}
