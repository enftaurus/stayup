package main

import (
	"github.com/enftaurus/stayup/app/api"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine) {
	// CORS Middleware
	router.Use(api.CORSMiddleware())

	// Health check
	router.GET("/health", api.Health)

	// Authentication routes
	auth := router.Group("/auth")
	{
		auth.GET("/url", api.GithubURL)
		auth.GET("/github/callback", api.GIthubCallback)
		auth.POST("/refresh", api.Refresh)
		auth.POST("/logout", api.Logout)
		auth.POST("/dev-login", api.DevLogin)
	}

	// Protected API routes
	apiGroup := router.Group("/api")
	apiGroup.Use(api.AuthRequired())
	{
		// User profile
		apiGroup.GET("/me", api.GetMe)

		// Dashboard summary
		apiGroup.GET("/dashboard/summary", api.GetDashboardSummary)

		// Incidents list
		apiGroup.GET("/incidents", api.ListIncidents)

		// Monitors CRUD & operations
		apiGroup.POST("/monitors", api.CreateMonitor)
		apiGroup.GET("/monitors", api.ListMonitors)
		apiGroup.GET("/monitors/:id", api.GetMonitor)
		apiGroup.PATCH("/monitors/:id", api.UpdateMonitor)
		apiGroup.DELETE("/monitors/:id", api.DeleteMonitor)
		apiGroup.POST("/monitors/:id/pause", api.PauseMonitor)
		apiGroup.POST("/monitors/:id/resume", api.ResumeMonitor)
		apiGroup.GET("/monitors/:id/checks", api.GetMonitorChecks)
		apiGroup.GET("/monitors/:id/checks/recent", api.GetRecentChecks)
		apiGroup.GET("/monitors/:id/incidents", api.GetMonitorIncidents)
	}
}
