package main

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"email-app/database"
	"email-app/handlers"
	"email-app/middleware"
)

func main() {
	godotenv.Load()

	database.Connect()
	database.CreateIndexes()

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "http://localhost:5173")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.Use(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.Next()
			return
		}
		staticPath := "../react-spa/dist"
		filePath := staticPath + c.Request.URL.Path
		if _, err := os.Stat(filePath); err == nil {
			c.File(filePath)
			c.Abort()
			return
		}
		c.File(staticPath + "/index.html")
		c.Abort()
	})

	api := r.Group("/api")
	{
		api.POST("/auth/login", handlers.Login)
		api.POST("/auth/logout", handlers.Logout)

		api.POST("/webhooks/resend", handlers.ResendWebhook)

		auth := api.Group("")
		auth.Use(middleware.AuthRequired())
		{
			auth.GET("/auth/me", handlers.Me)
			auth.GET("/mailboxes", handlers.ListMailboxes)
			auth.GET("/mailboxes/:id", handlers.GetMailbox)
			auth.GET("/emails", handlers.ListEmails)
			auth.GET("/emails/:id/thread", handlers.GetThread)
			auth.PATCH("/emails/:id/star", handlers.ToggleStar)
			auth.POST("/emails/send", handlers.SendEmail)
			admin := auth.Group("/admin")
			admin.Use(middleware.AdminRequired())
			{
				admin.GET("/users", handlers.ListUsers)
				admin.POST("/users", handlers.CreateUser)
				admin.DELETE("/users/:id", handlers.DeleteUser)
				admin.PATCH("/users/:id/role", handlers.SetUserRole)
				admin.GET("/mailboxes", handlers.ListAllMailboxes)
				admin.POST("/mailboxes", handlers.CreateMailbox)
				admin.DELETE("/mailboxes/:id", handlers.DeleteMailbox)
				admin.POST("/assignments", handlers.CreateAssignment)
				admin.DELETE("/assignments/:id", handlers.DeleteAssignment)
			}
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
