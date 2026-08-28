package api

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/enftaurus/stayup/app/models"
	"github.com/enftaurus/stayup/app/services"
	"github.com/enftaurus/stayup/config"
	"github.com/enftaurus/stayup/utils"
	"github.com/gin-gonic/gin"
)

func GithubURL(c *gin.Context) {
	state := utils.GenerateRandomState()
	GitHubUrl, err := services.BuildUrl(config.GetEnv("GH_CLIENT"), config.GetEnv("REDIRECT_URL"), state)
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{
			"error": `unable to redirect to github`})
		return
	}

	c.SetCookie(
		"oauth_state",
		state,
		300,
		"/",
		"",
		false,
		true,
	)
	c.Redirect(http.StatusFound, GitHubUrl)
}

func GIthubCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	if code == "" || state == "" {
		c.JSON(400, gin.H{"error": "missing code or state parameter"})
		return
	}
	storedstate, err := c.Cookie("oauth_state")
	if err != nil {
		c.JSON(500, gin.H{"error": "missing state "})
		return
	}
	if storedstate != state {
		c.JSON(400, gin.H{"error": "invalid state"})
		return
	}
	c.SetCookie(
		"oauth_state",
		"",
		-1,
		"/",
		"",
		false,
		true,
	)

	data := url.Values{}
	data.Add("client_id", config.GetEnv("GH_CLIENT"))
	data.Add("client_secret", config.GetEnv("GH_SECRET"))
	data.Add("code", code)
	data.Add("redirect_uri", config.GetEnv("REDIRECT_URL"))
	req, err := http.NewRequest(
		http.MethodPost,
		"https://github.com/login/oauth/access_token",
		strings.NewReader(data.Encode()),
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(500, gin.H{"error": "unable to reach github servers"})
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{"error": resp.Status})
		return
	}
	var token GitHubTokenResponse
	err = json.NewDecoder(resp.Body).Decode(&token)
	if err != nil {
		c.JSON(resp.StatusCode, gin.H{"error": "unable to decode token"})
		return
	}
	req, err = http.NewRequest(http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		c.JSON(resp.StatusCode, gin.H{"error": err.Error()})
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(resp.StatusCode, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	var data1 models.GithubUser
	err = json.NewDecoder(resp.Body).Decode(&data1)
	if err != nil {
		c.JSON(resp.StatusCode, gin.H{"error": err.Error()})
		return
	}
	browser := c.GetHeader("User-Agent")
	metadata := models.Userip{
		Ip:        net.ParseIP(c.ClientIP()),
		UserAgent: &browser,
	}
	ctx := c.Request.Context()
	RefreshToken, err := services.UserAuth(data1, metadata, ctx)
	c.SetCookie(
		"refresh_token",
		RefreshToken,
		5*24*60*60,
		"/",
		"",
		false,
		true,
	)
	c.JSON(200, gin.H{"data": data1})

}
