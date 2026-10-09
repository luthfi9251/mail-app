package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"email-app/config"
	"email-app/database"
	"email-app/handlers"
	"email-app/middleware"
)

func main() {
	godotenv.Load()
	config.Load()

	database.Connect()
	database.CreateIndexes()

	r := gin.Default()
	if err := r.SetTrustedProxies(config.TrustedProxy); err != nil {
		log.Fatalf("TRUSTED_PROXIES: %v", err)
	}

	r.Use(middleware.SecurityHeaders(), middleware.CORS())
	r.Use(serveSPA(config.StaticDir))

	loginLimiter := middleware.NewLimiter(20, 15*time.Minute)
	setupLimiter := middleware.NewLimiter(10, 15*time.Minute)

	// Webhook is not in use yet; receiving goes through SyncMailbox. To enable,
	// uncomment and set RESEND_WEBHOOK_SECRET. It is called by Resend, not the
	// browser: authenticated by its signature, outside the RequireJSON (CSRF) middleware.
	// r.POST("/api/webhooks/resend", middleware.LimitBody(), handlers.ResendWebhook)

	api := r.Group("/api")
	api.Use(middleware.LimitBody(), middleware.RequireJSON())
	{
		api.GET("/auth/setup-status", handlers.SetupStatus)
		api.POST("/auth/setup", setupLimiter.PerIP(), handlers.Setup)

		api.POST("/auth/login", loginLimiter.PerIP(), handlers.Login)
		api.POST("/auth/logout", handlers.Logout)

		auth := api.Group("")
		auth.Use(middleware.AuthRequired())
		{
			auth.GET("/auth/me", handlers.Me)
			auth.POST("/auth/change-password", handlers.ChangePassword)
			auth.GET("/mailboxes", handlers.ListMailboxes)
			auth.GET("/mailboxes/:id", handlers.GetMailbox)
			auth.POST("/mailboxes/:id/sync", handlers.SyncMailbox)
			auth.GET("/emails", handlers.ListEmails)
			auth.GET("/emails/:id/thread", handlers.GetThread)
			auth.PATCH("/emails/:id/star", handlers.ToggleStar)
			auth.POST("/emails/send", handlers.SendEmail)
			admin := auth.Group("/admin")
			admin.Use(middleware.AdminRequired())
			{
				admin.GET("/config", handlers.GetAdminConfig)
				admin.GET("/users", handlers.ListUsers)
				admin.POST("/users", handlers.CreateUser)
				admin.DELETE("/users/:id", handlers.DeleteUser)
				admin.PATCH("/users/:id/role", handlers.SetUserRole)
				admin.GET("/mailboxes", handlers.ListAllMailboxes)
				admin.POST("/mailboxes", handlers.CreateMailbox)
				admin.DELETE("/mailboxes/:id", handlers.DeleteMailbox)
				admin.GET("/mailboxes/:id/assignments", handlers.ListMailboxAssignments)
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

// serveSPA serves the built web app for every non-API path, falling back to
// index.html for client-side routes. http.Dir confines lookups to dir.
func serveSPA(dir string) gin.HandlerFunc {
	fs := http.Dir(dir)
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
			c.Next()
			return
		}
		path := c.Request.URL.Path
		if f, err := fs.Open(path); err == nil {
			info, statErr := f.Stat()
			f.Close()
			if statErr == nil && !info.IsDir() {
				c.FileFromFS(path, fs)
				c.Abort()
				return
			}
		}
		c.FileFromFS("/", fs)
		c.Abort()
	}
}
