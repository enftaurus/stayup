package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/enftaurus/stayup/config"
	"github.com/enftaurus/stayup/utils"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type GitHubTokenResp struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

type GitHubUserResp struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func isCookieSecure() bool {
	return strings.ToLower(config.GetEnv("COOKIE_SECURE")) == "true"
}

func setRefreshCookie(c *gin.Context, token string, maxAge int) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		"refresh_token",
		token,
		maxAge,
		"/",
		"",
		isCookieSecure(),
		true, // HttpOnly
	)
}

func GithubURL(c *gin.Context) {
	state, err := utils.GenerateRefreshToken()
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate oauth state", "")
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		"oauth_state",
		state,
		300,
		"/",
		"",
		isCookieSecure(),
		true,
	)

	clientID := config.GetEnv("GH_CLIENT")
	redirectURI := config.GetEnv("REDIRECT_URL")
	ghURL := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&state=%s&scope=read:user,user:email",
		url.QueryEscape(clientID),
		url.QueryEscape(redirectURI),
		url.QueryEscape(state),
	)

	c.Redirect(http.StatusFound, ghURL)
}

func GIthubCallback(c *gin.Context) {
	frontendURL := config.GetEnv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	code := c.Query("code")
	state := c.Query("state")
	if code == "" || state == "" {
		c.Redirect(http.StatusFound, frontendURL+"/?error=missing_code_or_state")
		return
	}

	storedState, err := c.Cookie("oauth_state")
	if err != nil || storedState != state {
		c.Redirect(http.StatusFound, frontendURL+"/?error=invalid_state")
		return
	}
	// Clear oauth state cookie
	c.SetCookie("oauth_state", "", -1, "/", "", isCookieSecure(), true)

	// Exchange code for GitHub token
	data := url.Values{}
	data.Set("client_id", config.GetEnv("GH_CLIENT"))
	data.Set("client_secret", config.GetEnv("GH_SECRET"))
	data.Set("code", code)
	data.Set("redirect_uri", config.GetEnv("REDIRECT_URL"))

	req, err := http.NewRequest(http.MethodPost, "https://github.com/login/oauth/access_token", strings.NewReader(data.Encode()))
	if err != nil {
		c.Redirect(http.StatusFound, frontendURL+"/?error=request_error")
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		c.Redirect(http.StatusFound, frontendURL+"/?error=oauth_exchange_failed")
		return
	}
	defer resp.Body.Close()

	var tokenResp GitHubTokenResp
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil || tokenResp.AccessToken == "" {
		c.Redirect(http.StatusFound, frontendURL+"/?error=token_decode_failed")
		return
	}

	// Fetch GitHub user
	userReq, err := http.NewRequest(http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		c.Redirect(http.StatusFound, frontendURL+"/?error=github_user_failed")
		return
	}
	userReq.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)
	userReq.Header.Set("Accept", "application/vnd.github+json")
	userReq.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	userResp, err := http.DefaultClient.Do(userReq)
	if err != nil || userResp.StatusCode != http.StatusOK {
		c.Redirect(http.StatusFound, frontendURL+"/?error=user_fetch_failed")
		return
	}
	defer userResp.Body.Close()

	var ghUser GitHubUserResp
	if err := json.NewDecoder(userResp.Body).Decode(&ghUser); err != nil {
		c.Redirect(http.StatusFound, frontendURL+"/?error=user_decode_failed")
		return
	}

	ctx := c.Request.Context()
	var userID string
	err = utils.DB.QueryRow(ctx, `
		INSERT INTO users (github_id, github_username, github_name, email, avatar_url, updated_at, last_login_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (github_id) DO UPDATE
		SET github_username = EXCLUDED.github_username,
		    github_name = COALESCE(EXCLUDED.github_name, users.github_name),
		    email = COALESCE(EXCLUDED.email, users.email),
		    avatar_url = COALESCE(EXCLUDED.avatar_url, users.avatar_url),
		    updated_at = NOW(),
		    last_login_at = NOW()
		RETURNING id::text
	`, ghUser.ID, ghUser.Login, ghUser.Name, ghUser.Email, ghUser.AvatarURL).Scan(&userID)

	if err != nil {
		log.Printf("Failed to insert/update user: %v", err)
		c.Redirect(http.StatusFound, frontendURL+"/?error=database_error")
		return
	}

	// Create session
	refreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		c.Redirect(http.StatusFound, frontendURL+"/?error=session_error")
		return
	}
	refreshTokenHash := HashToken(refreshToken)
	expiresAt := time.Now().UTC().Add(5 * 24 * time.Hour)
	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	_, err = utils.DB.Exec(ctx, `
		INSERT INTO user_sessions (user_id, refresh_token_hash, expires_at, ip_address, user_agent, created_at, last_used_at)
		VALUES ($1::uuid, $2, $3, $4, $5, NOW(), NOW())
	`, userID, refreshTokenHash, expiresAt, clientIP, userAgent)

	if err != nil {
		log.Printf("Failed to insert user session: %v", err)
		c.Redirect(http.StatusFound, frontendURL+"/?error=session_db_error")
		return
	}

	setRefreshCookie(c, refreshToken, 5*86400)
	c.Redirect(http.StatusFound, frontendURL+"/auth/done")
}

func Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Missing refresh token cookie", "")
		return
	}

	tokenHash := HashToken(refreshToken)
	ctx := c.Request.Context()

	var sessionID string
	var userID string
	err = utils.DB.QueryRow(ctx, `
		SELECT id::text, user_id::text
		FROM user_sessions
		WHERE refresh_token_hash = $1 AND expires_at > NOW()
	`, tokenHash).Scan(&sessionID, &userID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired session", "")
			return
		}
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Database error", "")
		return
	}

	// Rotate refresh token
	newRefreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate token", "")
		return
	}
	newHash := HashToken(newRefreshToken)
	newExpiresAt := time.Now().UTC().Add(5 * 24 * time.Hour)

	_, err = utils.DB.Exec(ctx, `
		UPDATE user_sessions
		SET refresh_token_hash = $1, expires_at = $2, last_used_at = NOW()
		WHERE id = $3::uuid
	`, newHash, newExpiresAt, sessionID)

	if err != nil {
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update session", "")
		return
	}

	setRefreshCookie(c, newRefreshToken, 5*86400)

	// Issue 15-minute access JWT
	jwtSecret := config.GetEnv("JWT_SECRET")
	accessToken, expiresIn, err := utils.GenerateAccessToken(userID, jwtSecret)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate access token", "")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
		"expires_in":   expiresIn,
	})
}

func Logout(c *gin.Context) {
	refreshToken, _ := c.Cookie("refresh_token")
	if refreshToken != "" {
		tokenHash := HashToken(refreshToken)
		ctx := c.Request.Context()
		_, _ = utils.DB.Exec(ctx, "DELETE FROM user_sessions WHERE refresh_token_hash = $1", tokenHash)
	}

	setRefreshCookie(c, "", -1)
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

func GetMe(c *gin.Context) {
	userID := GetUserID(c)
	ctx := c.Request.Context()

	var id string
	var username string
	var name *string
	var email *string
	var avatarURL *string

	err := utils.DB.QueryRow(ctx, `
		SELECT id::text, github_username, github_name, email, avatar_url
		FROM users
		WHERE id = $1::uuid
	`, userID).Scan(&id, &username, &name, &email, &avatarURL)

	if err != nil {
		RespondError(c, http.StatusNotFound, "NOT_FOUND", "User not found", "")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       id,
		"username": username,
		"name":     name,
		"email":    email,
		"avatar":   avatarURL,
	})
}

// DevLogin allows instant test authentication in development
func DevLogin(c *gin.Context) {
	if strings.ToLower(config.GetEnv("DEV_AUTH_BYPASS")) != "true" {
		RespondError(c, http.StatusForbidden, "FORBIDDEN", "Dev login is disabled", "")
		return
	}

	ctx := c.Request.Context()
	var userID string
	devName := "Dev Operator"
	devEmail := "dev@stayup.local"
	devAvatar := "https://avatars.githubusercontent.com/u/9919?v=4"

	err := utils.DB.QueryRow(ctx, `
		INSERT INTO users (github_id, github_username, github_name, email, avatar_url, updated_at, last_login_at)
		VALUES (99999999, 'dev_operator', $1, $2, $3, NOW(), NOW())
		ON CONFLICT (github_id) DO UPDATE
		SET last_login_at = NOW()
		RETURNING id::text
	`, devName, devEmail, devAvatar).Scan(&userID)

	if err != nil {
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), "")
		return
	}

	refreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate token", "")
		return
	}

	refreshTokenHash := HashToken(refreshToken)
	expiresAt := time.Now().UTC().Add(5 * 24 * time.Hour)

	_, err = utils.DB.Exec(ctx, `
		INSERT INTO user_sessions (user_id, refresh_token_hash, expires_at, ip_address, user_agent, created_at, last_used_at)
		VALUES ($1::uuid, $2, $3, $4, $5, NOW(), NOW())
	`, userID, refreshTokenHash, expiresAt, c.ClientIP(), c.GetHeader("User-Agent"))

	if err != nil {
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), "")
		return
	}

	setRefreshCookie(c, refreshToken, 5*86400)

	jwtSecret := config.GetEnv("JWT_SECRET")
	accessToken, expiresIn, err := utils.GenerateAccessToken(userID, jwtSecret)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to sign token", "")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
		"expires_in":   expiresIn,
		"user": gin.H{
			"id":       userID,
			"username": "dev_operator",
			"name":     devName,
			"email":    devEmail,
			"avatar":   devAvatar,
		},
	})
}

// Unused dummy to satisfy unused imports
var _ = net.ParseIP
var _ = strconv.Itoa
