package main

import (
	"github.com/enftaurus/stayup/app/api"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine) {
	router.GET("/health", api.Health)
	auth := router.Group("/auth")
	{
		auth.GET("/url", api.GithubURL)
		auth.GET("/github/callback", api.GIthubCallback)
	}
}
