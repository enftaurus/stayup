package api

import (
	"net/http"
	"strings"

	"github.com/enftaurus/stayup/config"
	"github.com/enftaurus/stayup/utils"
	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	frontendURL := config.GetEnv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == frontendURL || frontendURL == "*" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else if origin != "" && strings.HasPrefix(origin, "http://localhost:") {
			// Allow local dev frontend ports
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Missing Authorization header", "")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid Authorization header format", "")
			c.Abort()
			return
		}

		tokenString := parts[1]
		jwtSecret := config.GetEnv("JWT_SECRET")
		claims, err := utils.ValidateAccessToken(tokenString, jwtSecret)
		if err != nil {
			RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired access token", "")
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Next()
	}
}

func GetUserID(c *gin.Context) string {
	val, exists := c.Get("userID")
	if !exists {
		return ""
	}
	return val.(string)
}
