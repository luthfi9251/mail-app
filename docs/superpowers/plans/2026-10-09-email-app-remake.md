# Email App Remake — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remake tripbooka email app with React (Vite) + Go (Gin) + MongoDB

**Architecture:** React SPA communicates with Go/Gin REST API via fetch. JWT auth via httpOnly cookie. Go serves built React static files in production. MongoDB via mongo-go-driver.

**Tech Stack:** React 19, Vite, Tailwind CSS, shadcn/ui, lucide-react, Tiptap, Go 1.22+, Gin, mongo-go-driver, golang-jwt, bcrypt, Resend SDK

**Spec:** `docs/superpowers/specs/2026-10-09-email-app-remake-design.md`

## Global Constraints

- All new code in `/home/luthfi9251/project/email-app/`
- Frontend in `react-spa/`, backend in `server/`
- JWT stored in httpOnly cookie named `token`
- Admin routes check `role == "admin"` server-side
- Loading, empty, error states required on all data-dependent pages
- MongoDB connection string from `MONGODB_URI` env var
- Resend API key from `RESEND_API_KEY` env var
- No user registration — admin creates users via admin panel or direct DB insert

---

## File Structure

```
email-app/
├── docker-compose.yml
├── react-spa/
│   ├── index.html
│   ├── package.json
│   ├── vite.config.ts
│   ├── tsconfig.json
│   ├── tsconfig.app.json
│   ├── tsconfig.node.json
│   ├── tailwind.config.ts
│   ├── postcss.config.js
│   ├── src/
│   │   ├── main.tsx
│   │   ├── App.tsx
│   │   ├── index.css
│   │   ├── lib/
│   │   │   ├── api.ts
│   │   │   └── utils.ts
│   │   ├── components/
│   │   │   ├── ui/button.tsx
│   │   │   ├── ui/input.tsx
│   │   │   ├── ui/card.tsx
│   │   │   ├── ui/avatar.tsx
│   │   │   ├── ui/badge.tsx
│   │   │   ├── ui/dialog.tsx
│   │   │   ├── ui/tabs.tsx
│   │   │   ├── ui/scroll-area.tsx
│   │   │   ├── ui/separator.tsx
│   │   │   ├── ui/skeleton.tsx
│   │   │   ├── ui/dropdown-menu.tsx
│   │   │   ├── sidebar.tsx
│   │   │   ├── email-list.tsx
│   │   │   ├── email-card.tsx
│   │   │   ├── thread-view.tsx
│   │   │   ├── compose-dialog.tsx
│   │   │   ├── rich-text-editor.tsx
│   │   │   ├── admin-users.tsx
│   │   │   ├── admin-mailboxes.tsx
│   │   │   ├── empty-state.tsx
│   │   │   └── logout-button.tsx
│   │   └── pages/
│   │       ├── LoginPage.tsx
│   │       ├── DashboardPage.tsx
│   │       ├── InboxPage.tsx
│   │       └── AdminPage.tsx
├── server/
│   ├── go.mod
│   ├── go.sum
│   ├── main.go
│   ├── .env.example
│   ├── database/
│   │   └── db.go
│   ├── models/
│   │   ├── user.go
│   │   ├── mailbox.go
│   │   ├── usermailbox.go
│   │   └── email.go
│   ├── middleware/
│   │   └── auth.go
│   └── handlers/
│       ├── auth.go
│       ├── mailboxes.go
│       ├── emails.go
│       ├── admin.go
│       └── webhooks.go
```

---

### Task 1: Project Scaffolding

**Files:**
- Create: `docker-compose.yml`
- Create: `server/go.mod`
- Create: `server/.env.example`
- Create: `react-spa/` (via Vite scaffold)
- Create: `react-spa/src/index.css`
- Create: `react-spa/src/lib/utils.ts`

- [ ] **Step 1: Create project directories**

```bash
mkdir -p /home/luthfi9251/project/email-app
cd /home/luthfi9251/project/email-app
mkdir -p server/database server/models server/middleware server/handlers
```

- [ ] **Step 2: Create docker-compose.yml**

```yaml
services:
  mongodb:
    image: mongo:7
    container_name: email-app-mongodb
    ports:
      - "27017:27017"
    volumes:
      - mongodb_data:/data/db
    environment:
      MONGO_INITDB_ROOT_USERNAME: user
      MONGO_INITDB_ROOT_PASSWORD: password

volumes:
  mongodb_data:
```

- [ ] **Step 3: Initialize Go module**

```bash
cd /home/luthfi9251/project/email-app/server
go mod init email-app
```

- [ ] **Step 4: Install Go dependencies**

```bash
cd /home/luthfi9251/project/email-app/server
go get github.com/gin-gonic/gin
go get go.mongodb.org/mongo-driver/v2
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/crypto
go get github.com/joho/godotenv
go get github.com/resend/resend-go/v2
```

- [ ] **Step 5: Create server/.env.example**

```
MONGODB_URI=mongodb://user:password@localhost:27017/email-app?authSource=admin
RESEND_API_KEY=re_xxxxxxxxxxxx
JWT_SECRET=change-this-to-a-random-secret
PORT=8080
```

- [ ] **Step 6: Scaffold React app with Vite**

```bash
cd /home/luthfi9251/project/email-app
npm create vite@latest react-spa -- --template react-ts
cd react-spa
npm install
npm install tailwindcss @tailwindcss/vite lucide-react class-variance-authority clsx tailwind-merge @radix-ui/react-avatar @radix-ui/react-dialog @radix-ui/react-dropdown-menu @radix-ui/react-tabs @radix-ui/react-scroll-area @radix-ui/react-separator @radix-ui/react-tooltip @radix-ui/react-slot
npm install react-router-dom
npm install @tiptap/react @tiptap/starter-kit @tiptap/extension-placeholder @tiptap/pm
```

- [ ] **Step 7: Configure Tailwind in vite.config.ts**

Read and overwrite `react-spa/vite.config.ts`:

```ts
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
})
```

- [ ] **Step 8: Create react-spa/src/index.css**

```css
@import "tailwindcss";

@custom-variant dark (&:is(.dark *));

:root {
  --background: oklch(1 0 0);
  --foreground: oklch(0.145 0 0);
  --card: oklch(1 0 0);
  --card-foreground: oklch(0.145 0 0);
  --primary: oklch(0.546 0.245 262.881);
  --primary-foreground: oklch(0.985 0 0);
  --secondary: oklch(0.965 0.001 286.375);
  --secondary-foreground: oklch(0.205 0.042 265.755);
  --muted: oklch(0.965 0.001 286.375);
  --muted-foreground: oklch(0.556 0.019 286.375);
  --accent: oklch(0.965 0.001 286.375);
  --accent-foreground: oklch(0.205 0.042 265.755);
  --destructive: oklch(0.577 0.245 27.325);
  --destructive-foreground: oklch(0.577 0.245 27.325);
  --border: oklch(0.922 0.004 286.375);
  --input: oklch(0.922 0.004 286.375);
  --ring: oklch(0.546 0.245 262.881);
  --radius: 0.5rem;
}

.dark {
  --background: oklch(0.145 0 0);
  --foreground: oklch(0.985 0 0);
  --card: oklch(0.145 0 0);
  --card-foreground: oklch(0.985 0 0);
  --primary: oklch(0.546 0.245 262.881);
  --primary-foreground: oklch(0.985 0 0);
  --secondary: oklch(0.269 0.015 286.375);
  --secondary-foreground: oklch(0.985 0 0);
  --muted: oklch(0.269 0.015 286.375);
  --muted-foreground: oklch(0.708 0.014 286.375);
  --accent: oklch(0.269 0.015 286.375);
  --accent-foreground: oklch(0.985 0 0);
  --destructive: oklch(0.577 0.245 27.325);
  --destructive-foreground: oklch(0.577 0.245 27.325);
  --border: oklch(0.269 0.015 286.375);
  --input: oklch(0.269 0.015 286.375);
  --ring: oklch(0.546 0.245 262.881);
}

* {
  border-color: var(--border);
}

body {
  background-color: var(--background);
  color: var(--foreground);
}
```

- [ ] **Step 9: Create react-spa/src/lib/utils.ts**

```ts
import { type ClassValue, clsx } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
```

- [ ] **Step 10: Create seed admin script**

Create `server/seed.go`:

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	godotenv.Load()
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("MONGODB_URI not set")
	}

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	db := client.Database("email-app")
	users := db.Collection("users")

	password, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)

	_, err = users.InsertOne(context.Background(), bson.M{
		"email":    "admin@email.app",
		"name":     "Admin",
		"password": string(password),
		"role":     "admin",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Admin user created: admin@email.app / admin123")
}
```

- [ ] **Step 11: Add seed script to docker-compose**

No change needed — run manually when needed:

```bash
cd /home/luthfi9251/project/email-app/server
go run seed.go
```

- [ ] **Step 12: Start MongoDB**

```bash
cd /home/luthfi9251/project/email-app
docker compose up -d
```

---

### Task 2: Database Connection & Models (Go)

**Files:**
- Create: `server/database/db.go`
- Create: `server/models/user.go`
- Create: `server/models/mailbox.go`
- Create: `server/models/usermailbox.go`
- Create: `server/models/email.go`

- [ ] **Step 1: Create database/db.go**

```go
package database

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var DB *mongo.Database

func Connect() {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("MONGODB_URI not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal(err)
	}

	DB = client.Database("email-app")
	log.Println("Connected to MongoDB")
}

func GetCollection(name string) *mongo.Collection {
	return DB.Collection(name)
}
```

- [ ] **Step 2: Create models/user.go**

```go
package models

import "go.mongodb.org/mongo-driver/v2/bson"

type User struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Email     string        `bson:"email" json:"email"`
	Name      string        `bson:"name" json:"name"`
	Password  string        `bson:"password" json:"-"`
	Role      string        `bson:"role" json:"role"`
	CreatedAt int64         `bson:"createdAt" json:"createdAt"`
}
```

- [ ] **Step 3: Create models/mailbox.go**

```go
package models

import "go.mongodb.org/mongo-driver/v2/bson"

type Mailbox struct {
	ID             bson.ObjectID `bson:"_id,omitempty" json:"id"`
	EmailAddress   string        `bson:"emailAddress" json:"emailAddress"`
	ResendDomainID string        `bson:"resendDomainId" json:"resendDomainId"`
	CreatedAt      int64         `bson:"createdAt" json:"createdAt"`
}
```

- [ ] **Step 4: Create models/usermailbox.go**

```go
package models

import "go.mongodb.org/mongo-driver/v2/bson"

type UserMailbox struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    bson.ObjectID `bson:"userId" json:"userId"`
	MailboxID bson.ObjectID `bson:"mailboxId" json:"mailboxId"`
	Role      string        `bson:"role" json:"role"`
	CreatedAt int64         `bson:"createdAt" json:"createdAt"`
}
```

- [ ] **Step 5: Create models/email.go**

```go
package models

import "go.mongodb.org/mongo-driver/v2/bson"

type Attachment struct {
	ID          string `bson:"id" json:"id"`
	Filename    string `bson:"filename" json:"filename"`
	Size        int    `bson:"size" json:"size"`
	ContentType string `bson:"content_type" json:"content_type"`
}

type Email struct {
	ID           bson.ObjectID `bson:"_id,omitempty" json:"id"`
	MailboxID    bson.ObjectID `bson:"mailboxId" json:"mailboxId"`
	From         string        `bson:"from" json:"from"`
	To           string        `bson:"to" json:"to"`
	Subject      string        `bson:"subject" json:"subject"`
	HTMLBody     string        `bson:"htmlBody" json:"htmlBody"`
	TextBody     string        `bson:"textBody" json:"textBody"`
	MessageID    string        `bson:"messageId" json:"messageId"`
	InReplyTo    *string       `bson:"inReplyTo" json:"inReplyTo"`
	References   []string      `bson:"references" json:"references"`
	ThreadID     string        `bson:"threadId" json:"threadId"`
	Direction    string        `bson:"direction" json:"direction"`
	Read         bool          `bson:"read" json:"read"`
	Starred      bool          `bson:"starred" json:"starred"`
	Attachments  []Attachment  `bson:"attachments" json:"attachments"`
	CreatedAt    int64         `bson:"createdAt" json:"createdAt"`
}
```

- [ ] **Step 6: Create indexes**

Create `server/database/indexes.go`:

```go
package database

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func CreateIndexes() {
	userMailboxCol := GetCollection("usermailboxes")
	_, err := userMailboxCol.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    bson.D{{Key: "userId", Value: 1}, {Key: "mailboxId", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		log.Printf("Warning: failed to create usermailbox index: %v", err)
	}

	emailCol := GetCollection("emails")
	_, err = emailCol.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{Keys: bson.D{{Key: "mailboxId", Value: 1}, {Key: "createdAt", Value: -1}}},
		{Keys: bson.D{{Key: "mailboxId", Value: 1}, {Key: "direction", Value: 1}, {Key: "createdAt", Value: -1}}},
		{Keys: bson.D{{Key: "messageId", Value: 1}}, Options: options.Index().SetUnique(true)},
	})
	if err != nil {
		log.Printf("Warning: failed to create email indexes: %v", err)
	}
}
```

---

### Task 3: JWT Auth Middleware & Login Handler

**Files:**
- Create: `server/middleware/auth.go`
- Create: `server/handlers/auth.go`

- [ ] **Step 1: Create middleware/auth.go**

```go
package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"userId"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		cookie, err := c.Cookie("token")
		if err == nil && cookie != "" {
			tokenString = cookie
		} else {
			auth := c.GetHeader("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				tokenString = strings.TrimPrefix(auth, "Bearer ")
			}
		}

		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set("userId", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}
```

- [ ] **Step 2: Create handlers/auth.go**

```go
package handlers

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"

	"email-app/database"
	"email-app/middleware"
	"email-app/models"
)

func Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email and password required"})
		return
	}

	col := database.GetCollection("users")
	var user models.User
	err := col.FindOne(c.Request.Context(), bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	claims := middleware.Claims{
		UserID: user.ID.Hex(),
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	c.SetCookie("token", tokenString, 7*24*3600, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"user": gin.H{
			"id":    user.ID.Hex(),
			"email": user.Email,
			"name":  user.Name,
			"role":  user.Role,
		},
	})
}

func Logout(c *gin.Context) {
	c.SetCookie("token", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func Me(c *gin.Context) {
	userID, _ := c.Get("userId")

	col := database.GetCollection("users")
	var user models.User
	err := col.FindOne(c.Request.Context(), bson.M{"_id": MustObjectID(userID.(string))}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":    user.ID.Hex(),
		"email": user.Email,
		"name":  user.Name,
		"role":  user.Role,
	})
}

func MustObjectID(hex string) bson.ObjectID {
	id, _ := bson.ObjectIDFromHex(hex)
	return id
}
```

---

### Task 4: Main Server Setup

**Files:**
- Create: `server/main.go`

- [ ] **Step 1: Create server/main.go**

```go
package main

import (
	"os"

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
			auth.GET("/emails/:id/attachments/:attachmentId", handlers.GetAttachment)

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
```

---

### Task 5: Mailbox & Email Handlers

**Files:**
- Create: `server/handlers/mailboxes.go`
- Create: `server/handlers/emails.go`

- [ ] **Step 1: Create handlers/mailboxes.go**

```go
package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"

	"email-app/database"
	"email-app/models"
)

func ListMailboxes(c *gin.Context) {
	userID, _ := c.Get("userId")

	umCol := database.GetCollection("usermailboxes")
	cursor, err := umCol.Find(c.Request.Context(), bson.M{"userId": MustObjectID(userID.(string))})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch mailboxes"})
		return
	}
	defer cursor.Close(c.Request.Context())

	var assignments []models.UserMailbox
	if err := cursor.All(c.Request.Context(), &assignments); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse mailboxes"})
		return
	}

	mbCol := database.GetCollection("mailboxes")
	var result []gin.H
	for _, a := range assignments {
		var mb models.Mailbox
		err := mbCol.FindOne(c.Request.Context(), bson.M{"_id": a.MailboxID}).Decode(&mb)
		if err != nil {
			continue
		}
		result = append(result, gin.H{
			"id":           mb.ID.Hex(),
			"emailAddress": mb.EmailAddress,
			"role":         a.Role,
		})
	}

	c.JSON(http.StatusOK, result)
}

func GetMailbox(c *gin.Context) {
	mbCol := database.GetCollection("mailboxes")
	var mb models.Mailbox
	err := mbCol.FindOne(c.Request.Context(), bson.M{"_id": MustObjectID(c.Param("id"))}).Decode(&mb)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "mailbox not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":           mb.ID.Hex(),
		"emailAddress": mb.EmailAddress,
	})
}
```

- [ ] **Step 2: Create handlers/emails.go**

```go
package handlers

import (
	"context"
	"crypto/rand"
	"fmt"
	"math"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/resend/resend-go/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"email-app/database"
	"email-app/models"
)

func ListEmails(c *gin.Context) {
	mailboxID := c.Query("mailboxId")
	direction := c.Query("direction")
	search := c.Query("search")
	starred := c.Query("starred")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	filter := bson.M{"mailboxId": MustObjectID(mailboxID)}
	if direction != "" {
		filter["direction"] = direction
	}
	if starred == "true" {
		filter["starred"] = true
	}
	if search != "" {
		re := regexp.QuoteMeta(search)
		filter["$or"] = bson.A{
			bson.M{"from": bson.M{"$regex": re, "$options": "i"}},
			bson.M{"to": bson.M{"$regex": re, "$options": "i"}},
			bson.M{"subject": bson.M{"$regex": re, "$options": "i"}},
			bson.M{"textBody": bson.M{"$regex": re, "$options": "i"}},
		}
	}

	col := database.GetCollection("emails")
	total, _ := col.CountDocuments(c.Request.Context(), filter)
	skip := (page - 1) * limit

	cursor, err := col.Find(
		c.Request.Context(),
		filter,
		options.Find().SetSkip(int64(skip)).SetLimit(int64(limit)).SetSort(bson.D{{Key: "createdAt", Value: -1}}),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch emails"})
		return
	}
	defer cursor.Close(c.Request.Context())

	var emails []models.Email
	cursor.All(c.Request.Context(), &emails)

	var result []gin.H
	for _, e := range emails {
		result = append(result, emailToJSON(e))
	}

	hasMore := int64(skip+limit) < total
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	c.JSON(http.StatusOK, gin.H{
		"emails":     result,
		"hasMore":    hasMore,
		"total":      total,
		"page":       page,
		"totalPages": totalPages,
	})
}

func GetThread(c *gin.Context) {
	id := c.Param("id")
	col := database.GetCollection("emails")

	var email models.Email
	err := col.FindOne(c.Request.Context(), bson.M{"_id": MustObjectID(id)}).Decode(&email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "email not found"})
		return
	}

	cursor, err := col.Find(
		c.Request.Context(),
		bson.M{"threadId": email.ThreadID},
		options.Find().SetSort(bson.D{{Key: "createdAt", Value: 1}}),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch thread"})
		return
	}
	defer cursor.Close(c.Request.Context())

	var emails []models.Email
	cursor.All(c.Request.Context(), &emails)

	var result []gin.H
	for _, e := range emails {
		result = append(result, emailToJSON(e))
	}

	c.JSON(http.StatusOK, gin.H{
		"threadId": email.ThreadID,
		"emails":   result,
	})
}

func ToggleStar(c *gin.Context) {
	var req struct {
		Starred bool `json:"starred"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	col := database.GetCollection("emails")
	var email models.Email
	err := col.FindOne(c.Request.Context(), bson.M{"_id": MustObjectID(c.Param("id"))}).Decode(&email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "email not found"})
		return
	}

	_, err = col.UpdateMany(
		c.Request.Context(),
		bson.M{"threadId": email.ThreadID, "mailboxId": email.MailboxID},
		bson.M{"$set": bson.M{"starred": req.Starred}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update star"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "starred": req.Starred})
}

func SendEmail(c *gin.Context) {
	var req struct {
		MailboxID string `json:"mailboxId" binding:"required"`
		To        string `json:"to" binding:"required"`
		Subject   string `json:"subject" binding:"required"`
		HTMLBody  string `json:"htmlBody"`
		TextBody  string `json:"textBody"`
		InReplyTo string `json:"inReplyTo"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	mbCol := database.GetCollection("mailboxes")
	var mailbox models.Mailbox
	err := mbCol.FindOne(c.Request.Context(), bson.M{"_id": MustObjectID(req.MailboxID)}).Decode(&mailbox)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "mailbox not found"})
		return
	}

	client := resend.NewClient(os.Getenv("RESEND_API_KEY"))
	messageID := fmt.Sprintf("<%s@email-app>", uuid())

	params := &resend.SendEmailRequest{
		From:    mailbox.EmailAddress,
		To:      []string{req.To},
		Subject: req.Subject,
		Html:    req.HTMLBody,
		Text:    req.TextBody,
	}

	if req.InReplyTo != "" {
		params.Headers = map[string]string{
			"In-Reply-To": req.InReplyTo,
			"References":  req.InReplyTo,
		}
	}

	_, err = client.Emails.Send(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send email"})
		return
	}

	threadID := uuid()
	if req.InReplyTo != "" {
		col := database.GetCollection("emails")
		var original models.Email
		err := col.FindOne(c.Request.Context(), bson.M{"messageId": req.InReplyTo}).Decode(&original)
		if err == nil {
			threadID = original.ThreadID
		}
	}

	emailCol := database.GetCollection("emails")
	var inReplyTo *string
	if req.InReplyTo != "" {
		inReplyTo = &req.InReplyTo
	}
	references := []string{}
	if req.InReplyTo != "" {
		references = append(references, req.InReplyTo)
	}

	email := models.Email{
		MailboxID:  MustObjectID(req.MailboxID),
		From:       mailbox.EmailAddress,
		To:         req.To,
		Subject:    req.Subject,
		HTMLBody:   req.HTMLBody,
		TextBody:   req.TextBody,
		MessageID:  messageID,
		InReplyTo:  inReplyTo,
		References: references,
		ThreadID:   threadID,
		Direction:  "outbound",
		Read:       true,
		Starred:    false,
		CreatedAt:  time.Now().UnixMilli(),
	}

	result, err := emailCol.InsertOne(c.Request.Context(), email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "id": result.InsertedID})
}

func GetAttachment(c *gin.Context) {
	client := resend.NewClient(os.Getenv("RESEND_API_KEY"))
	emailID := c.Param("id")
	attachmentID := c.Param("attachmentId")

	att, err := client.Emails.Receiving.Attachments.Get(c.Request.Context(), &resend.GetAttachmentRequest{
		EmailID:      emailID,
		AttachmentID: attachmentID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get attachment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"download_url": att.URL,
		"expires_at":   att.ExpiresAt,
	})
}

func emailToJSON(e models.Email) gin.H {
	return gin.H{
		"id":          e.ID.Hex(),
		"mailboxId":   e.MailboxID.Hex(),
		"from":        e.From,
		"to":          e.To,
		"subject":     e.Subject,
		"htmlBody":    e.HTMLBody,
		"textBody":    e.TextBody,
		"messageId":   e.MessageID,
		"inReplyTo":   e.InReplyTo,
		"references":  e.References,
		"threadId":    e.ThreadID,
		"direction":   e.Direction,
		"read":        e.Read,
		"starred":     e.Starred,
		"attachments": e.Attachments,
		"createdAt":   e.CreatedAt,
	}
}

func uuid() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
```

---

### Task 6: Admin & Webhook Handlers

**Files:**
- Create: `server/handlers/admin.go`
- Create: `server/handlers/webhooks.go`

- [ ] **Step 1: Create handlers/admin.go**

```go
package handlers

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"golang.org/x/crypto/bcrypt"

	"email-app/database"
	"email-app/models"
)

func ListUsers(c *gin.Context) {
	col := database.GetCollection("users")
	cursor, err := col.Find(c.Request.Context(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch users"})
		return
	}
	defer cursor.Close(c.Request.Context())

	var users []models.User
	cursor.All(c.Request.Context(), &users)

	// Count mailboxes per user
	umCol := database.GetCollection("usermailboxes")
	var result []gin.H
	for _, u := range users {
		count, _ := umCol.CountDocuments(c.Request.Context(), bson.M{"userId": u.ID})
		result = append(result, gin.H{
			"id":            u.ID.Hex(),
			"email":         u.Email,
			"name":          u.Name,
			"role":          u.Role,
			"mailboxCount":  count,
			"createdAt":     u.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, result)
}

func CreateUser(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required"`
		Name     string `json:"name" binding:"required"`
		Password string `json:"password" binding:"required"`
		Role     string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, email, password required"})
		return
	}

	if req.Role == "" {
		req.Role = "user"
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	col := database.GetCollection("users")
	_, err := col.InsertOne(c.Request.Context(), models.User{
		Email:     req.Email,
		Name:      req.Name,
		Password:  string(hash),
		Role:      req.Role,
		CreatedAt: time.Now().UnixMilli(),
	})
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true})
}

func DeleteUser(c *gin.Context) {
	col := database.GetCollection("users")
	_, err := col.DeleteOne(c.Request.Context(), bson.M{"_id": MustObjectID(c.Param("id"))})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func SetUserRole(c *gin.Context) {
	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role required"})
		return
	}

	col := database.GetCollection("users")
	_, err := col.UpdateByID(c.Request.Context(), MustObjectID(c.Param("id")), bson.M{"$set": bson.M{"role": req.Role}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update role"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func ListAllMailboxes(c *gin.Context) {
	col := database.GetCollection("mailboxes")
	cursor, err := col.Find(c.Request.Context(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch mailboxes"})
		return
	}
	defer cursor.Close(c.Request.Context())

	var mailboxes []models.Mailbox
	cursor.All(c.Request.Context(), &mailboxes)

	umCol := database.GetCollection("usermailboxes")
	var result []gin.H
	for _, mb := range mailboxes {
		count, _ := umCol.CountDocuments(c.Request.Context(), bson.M{"mailboxId": mb.ID})
		result = append(result, gin.H{
			"id":             mb.ID.Hex(),
			"emailAddress":   mb.EmailAddress,
			"userCount":      count,
			"createdAt":      mb.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, result)
}

func CreateMailbox(c *gin.Context) {
	var req struct {
		LocalPart string `json:"localPart" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "localPart required"})
		return
	}

	domain := os.Getenv("MAILBOX_DOMAIN")
	if domain == "" {
		domain = "tripbooka.com"
	}

	col := database.GetCollection("mailboxes")
	_, err := col.InsertOne(c.Request.Context(), models.Mailbox{
		EmailAddress: req.LocalPart + "@" + domain,
		CreatedAt:    time.Now().UnixMilli(),
	})
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "mailbox already exists"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true})
}

func DeleteMailbox(c *gin.Context) {
	emailCol := database.GetCollection("emails")
	count, _ := emailCol.CountDocuments(c.Request.Context(), bson.M{"mailboxId": MustObjectID(c.Param("id"))})
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete mailbox with existing emails"})
		return
	}

	col := database.GetCollection("mailboxes")
	_, err := col.DeleteOne(c.Request.Context(), bson.M{"_id": MustObjectID(c.Param("id"))})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete mailbox"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func CreateAssignment(c *gin.Context) {
	var req struct {
		UserID    string `json:"userId" binding:"required"`
		MailboxID string `json:"mailboxId" binding:"required"`
		Role      string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId and mailboxId required"})
		return
	}
	if req.Role == "" {
		req.Role = "member"
	}

	col := database.GetCollection("usermailboxes")
	_, err := col.InsertOne(c.Request.Context(), models.UserMailbox{
		UserID:    MustObjectID(req.UserID),
		MailboxID: MustObjectID(req.MailboxID),
		Role:      req.Role,
		CreatedAt: time.Now().UnixMilli(),
	})
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "assignment already exists"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true})
}

func DeleteAssignment(c *gin.Context) {
	col := database.GetCollection("usermailboxes")
	_, err := col.DeleteOne(c.Request.Context(), bson.M{"_id": MustObjectID(c.Param("id"))})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete assignment"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
```

- [ ] **Step 2: Create handlers/webhooks.go**

```go
package handlers

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"

	"email-app/database"
	"email-app/models"
)

func ResendWebhook(c *gin.Context) {
	var payload struct {
		From      string `json:"from"`
		To        string `json:"to"`
		Subject   string `json:"subject"`
		Html      string `json:"html"`
		Text      string `json:"text"`
		MessageID string `json:"messageId"`
		Headers   struct {
			InReplyTo  string `json:"in-reply-to"`
			References string `json:"references"`
		} `json:"headers"`
		Attachments []models.Attachment `json:"attachments"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	mbCol := database.GetCollection("mailboxes")
	var mailbox models.Mailbox
	err := mbCol.FindOne(c.Request.Context(), bson.M{"emailAddress": payload.To}).Decode(&mailbox)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "mailbox not found"})
		return
	}

	messageID := payload.MessageID
	if messageID == "" {
		messageID = fmt.Sprintf("<%s@email-app>", uuid())
	}

	var inReplyTo *string
	if payload.Headers.InReplyTo != "" {
		inReplyTo = &payload.Headers.InReplyTo
	}
	references := []string{}
	if payload.Headers.References != "" {
		references = splitRefs(payload.Headers.References)
	}

	threadID := generateThreadID(payload.Subject, references, inReplyTo)
	if payload.Attachments == nil {
		payload.Attachments = []models.Attachment{}
	}

	emailCol := database.GetCollection("emails")
	_, err = emailCol.InsertOne(c.Request.Context(), models.Email{
		MailboxID:   mailbox.ID,
		From:        payload.From,
		To:          payload.To,
		Subject:     payload.Subject,
		HTMLBody:    payload.Html,
		TextBody:    payload.Text,
		MessageID:   messageID,
		InReplyTo:   inReplyTo,
		References:  references,
		ThreadID:    threadID,
		Direction:   "inbound",
		Read:        false,
		Starred:     false,
		Attachments: payload.Attachments,
		CreatedAt:   time.Now().UnixMilli(),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func generateThreadID(subject string, references []string, inReplyTo *string) string {
	if inReplyTo != nil && *inReplyTo != "" {
		return *inReplyTo
	}
	if len(references) > 0 {
		return references[0]
	}
	return fmt.Sprintf("%x", md5.Sum([]byte(subject)))
}

func splitRefs(refs string) []string {
	var result []string
	current := ""
	for _, c := range refs {
		if c == ' ' || c == '\t' || c == '\n' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
```

---

### Task 7: React App Shell, Routing & Auth

**Files:**
- Create: `react-spa/src/lib/api.ts`
- Create: `react-spa/src/App.tsx`
- Create: `react-spa/src/main.tsx`
- Create: `react-spa/src/pages/LoginPage.tsx`

- [ ] **Step 1: Create src/lib/api.ts**

```ts
const BASE = "/api";

export async function api<T = unknown>(
  path: string,
  options?: RequestInit
): Promise<T> {
  const res = await fetch(BASE + path, {
    credentials: "include",
    headers: { "Content-Type": "application/json", ...options?.headers },
    ...options,
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: "request failed" }));
    throw new Error(body.error || `HTTP ${res.status}`);
  }

  return res.json();
}

export interface User {
  id: string;
  email: string;
  name: string;
  role: string;
}

export interface Mailbox {
  id: string;
  emailAddress: string;
  role: string;
}

export interface EmailItem {
  id: string;
  mailboxId: string;
  from: string;
  to: string;
  subject: string;
  htmlBody: string;
  textBody: string;
  messageId: string;
  inReplyTo: string | null;
  references: string[];
  threadId: string;
  direction: "inbound" | "outbound";
  read: boolean;
  starred: boolean;
  attachments: Attachment[];
  createdAt: number;
}

export interface Attachment {
  id: string;
  filename: string;
  size: number;
  content_type: string;
}

export interface PaginatedResponse {
  emails: EmailItem[];
  hasMore: boolean;
  total: number;
  page: number;
  totalPages: number;
}
```

- [ ] **Step 2: Create src/App.tsx**

```tsx
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import LoginPage from "./pages/LoginPage";
import DashboardPage from "./pages/DashboardPage";
import InboxPage from "./pages/InboxPage";
import AdminPage from "./pages/AdminPage";

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/dashboard" element={<DashboardPage />} />
        <Route path="/dashboard/:mailboxId" element={<InboxPage />} />
        <Route path="/dashboard/:mailboxId/:emailId" element={<InboxPage />} />
        <Route path="/admin" element={<AdminPage />} />
        <Route path="*" element={<Navigate to="/dashboard" replace />} />
      </Routes>
    </BrowserRouter>
  );
}
```

- [ ] **Step 3: Create src/main.tsx**

```tsx
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import App from "./App";
import "./index.css";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <App />
  </StrictMode>
);
```

- [ ] **Step 4: Create shadcn/ui Button component**

Create `react-spa/src/components/ui/button.tsx`:

```tsx
import * as React from "react"
import { Slot } from "@radix-ui/react-slot"
import { cn } from "../../lib/utils"

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "default" | "destructive" | "outline" | "secondary" | "ghost" | "link"
  size?: "default" | "sm" | "lg" | "icon"
  asChild?: boolean
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = "default", size = "default", asChild = false, ...props }, ref) => {
    const Comp = asChild ? Slot : "button"
    return (
      <Comp
        className={cn(
          "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0",
          {
            "bg-primary text-primary-foreground shadow hover:bg-primary/90": variant === "default",
            "bg-destructive text-destructive-foreground shadow-sm hover:bg-destructive/90": variant === "destructive",
            "border border-input bg-background shadow-sm hover:bg-accent hover:text-accent-foreground": variant === "outline",
            "bg-secondary text-secondary-foreground shadow-sm hover:bg-secondary/80": variant === "secondary",
            "hover:bg-accent hover:text-accent-foreground": variant === "ghost",
            "text-primary underline-offset-4 hover:underline": variant === "link",
          },
          {
            "h-9 px-4 py-2": size === "default",
            "h-8 rounded-md px-3 text-xs": size === "sm",
            "h-10 rounded-md px-8": size === "lg",
            "h-9 w-9": size === "icon",
          },
          className
        )}
        ref={ref}
        {...props}
      />
    )
  }
)
Button.displayName = "Button"

export { Button }
```

- [ ] **Step 5: Create shadcn/ui Input, Card, Avatar, Badge, Dialog, Tabs, ScrollArea, Separator, Skeleton**

Create `react-spa/src/components/ui/input.tsx`:

```tsx
import * as React from "react"
import { cn } from "../../lib/utils"

const Input = React.forwardRef<HTMLInputElement, React.InputHTMLAttributes<HTMLInputElement>>(
  ({ className, type, ...props }, ref) => {
    return (
      <input
        type={type}
        className={cn(
          "flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium file:text-foreground placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50",
          className
        )}
        ref={ref}
        {...props}
      />
    )
  }
)
Input.displayName = "Input"

export { Input }
```

Create `react-spa/src/components/ui/card.tsx`:

```tsx
import * as React from "react"
import { cn } from "../../lib/utils"

const Card = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div ref={ref} className={cn("rounded-xl border bg-card text-card-foreground shadow", className)} {...props} />
  )
)
Card.displayName = "Card"

const CardHeader = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div ref={ref} className={cn("flex flex-col space-y-1.5 p-6", className)} {...props} />
  )
)
CardHeader.displayName = "CardHeader"

const CardContent = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div ref={ref} className={cn("p-6 pt-0", className)} {...props} />
  )
)
CardContent.displayName = "CardContent"

export { Card, CardHeader, CardContent }
```

Create `react-spa/src/components/ui/avatar.tsx`:

```tsx
import * as React from "react"
import { cn } from "../../lib/utils"

const Avatar = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div ref={ref} className={cn("relative flex h-10 w-10 shrink-0 overflow-hidden rounded-full", className)} {...props} />
  )
)
Avatar.displayName = "Avatar"

const AvatarFallback = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div ref={ref} className={cn("flex h-full w-full items-center justify-center rounded-full bg-muted text-xs font-medium", className)} {...props} />
  )
)
AvatarFallback.displayName = "AvatarFallback"

export { Avatar, AvatarFallback }
```

Create `react-spa/src/components/ui/badge.tsx`:

```tsx
import * as React from "react"
import { cn } from "../../lib/utils"

const Badge = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement> & { variant?: "default" | "secondary" | "destructive" | "outline" }>(
  ({ className, variant = "default", ...props }, ref) => (
    <div
      ref={ref}
      className={cn(
        "inline-flex items-center rounded-md border px-2.5 py-0.5 text-xs font-semibold transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2",
        {
          "border-transparent bg-primary text-primary-foreground shadow": variant === "default",
          "border-transparent bg-secondary text-secondary-foreground": variant === "secondary",
          "border-transparent bg-destructive text-destructive-foreground shadow": variant === "destructive",
          "text-foreground": variant === "outline",
        },
        className
      )}
      {...props}
    />
  )
)
Badge.displayName = "Badge"

export { Badge }
```

Create `react-spa/src/components/ui/dialog.tsx`:

```tsx
import * as React from "react"
import * as DialogPrimitive from "@radix-ui/react-dialog"
import { X } from "lucide-react"
import { cn } from "../../lib/utils"

const Dialog = DialogPrimitive.Root
const DialogTrigger = DialogPrimitive.Trigger
const DialogClose = DialogPrimitive.Close
const DialogPortal = DialogPrimitive.Portal
const DialogOverlay = React.forwardRef<
  React.ComponentRef<typeof DialogPrimitive.Overlay>,
  React.ComponentPropsWithoutRef<typeof DialogPrimitive.Overlay>
>(({ className, ...props }, ref) => (
  <DialogPrimitive.Overlay
    ref={ref}
    className={cn(
      "fixed inset-0 z-50 bg-black/80 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0",
      className
    )}
    {...props}
  />
))
DialogOverlay.displayName = DialogPrimitive.Overlay.displayName

const DialogContent = React.forwardRef<
  React.ComponentRef<typeof DialogPrimitive.Content>,
  React.ComponentPropsWithoutRef<typeof DialogPrimitive.Content>
>(({ className, children, ...props }, ref) => (
  <DialogPortal>
    <DialogOverlay />
    <DialogPrimitive.Content
      ref={ref}
      className={cn(
        "fixed left-[50%] top-[50%] z-50 grid w-full max-w-lg translate-x-[-50%] translate-y-[-50%] gap-4 border bg-background p-6 shadow-lg duration-200 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[state=closed]:slide-out-to-left-1/2 data-[state=closed]:slide-out-to-top-[48%] data-[state=open]:slide-in-from-left-1/2 data-[state=open]:slide-in-from-top-[48%] sm:rounded-lg",
        className
      )}
      {...props}
    >
      {children}
      <DialogPrimitive.Close className="absolute right-4 top-4 rounded-sm opacity-70 ring-offset-background transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:pointer-events-none data-[state=open]:bg-accent data-[state=open]:text-muted-foreground">
        <X className="h-4 w-4" />
        <span className="sr-only">Close</span>
      </DialogPrimitive.Close>
    </DialogPrimitive.Content>
  </DialogPortal>
))
DialogContent.displayName = DialogPrimitive.Content.displayName

const DialogHeader = ({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) => (
  <div className={cn("flex flex-col space-y-1.5 text-center sm:text-left", className)} {...props} />
)
DialogHeader.displayName = "DialogHeader"

const DialogTitle = React.forwardRef<
  React.ComponentRef<typeof DialogPrimitive.Title>,
  React.ComponentPropsWithoutRef<typeof DialogPrimitive.Title>
>(({ className, ...props }, ref) => (
  <DialogPrimitive.Title ref={ref} className={cn("text-lg font-semibold leading-none tracking-tight", className)} {...props} />
))
DialogTitle.displayName = DialogPrimitive.Title.displayName

export { Dialog, DialogTrigger, DialogClose, DialogContent, DialogHeader, DialogTitle }
```

Create `react-spa/src/components/ui/tabs.tsx`:

```tsx
import * as React from "react"
import * as TabsPrimitive from "@radix-ui/react-tabs"
import { cn } from "../../lib/utils"

const Tabs = TabsPrimitive.Root
const TabsList = React.forwardRef<
  React.ComponentRef<typeof TabsPrimitive.List>,
  React.ComponentPropsWithoutRef<typeof TabsPrimitive.List>
>(({ className, ...props }, ref) => (
  <TabsPrimitive.List
    ref={ref}
    className={cn(
      "inline-flex h-9 items-center justify-center rounded-lg bg-muted p-1 text-muted-foreground",
      className
    )}
    {...props}
  />
))
TabsList.displayName = TabsPrimitive.List.displayName

const TabsTrigger = React.forwardRef<
  React.ComponentRef<typeof TabsPrimitive.Trigger>,
  React.ComponentPropsWithoutRef<typeof TabsPrimitive.Trigger>
>(({ className, ...props }, ref) => (
  <TabsPrimitive.Trigger
    ref={ref}
    className={cn(
      "inline-flex items-center justify-center whitespace-nowrap rounded-md px-3 py-1 text-sm font-medium ring-offset-background transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 data-[state=active]:bg-background data-[state=active]:text-foreground data-[state=active]:shadow",
      className
    )}
    {...props}
  />
))
TabsTrigger.displayName = TabsPrimitive.Trigger.displayName

const TabsContent = React.forwardRef<
  React.ComponentRef<typeof TabsPrimitive.Content>,
  React.ComponentPropsWithoutRef<typeof TabsPrimitive.Content>
>(({ className, ...props }, ref) => (
  <TabsPrimitive.Content
    ref={ref}
    className={cn("mt-2 ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2", className)}
    {...props}
  />
))
TabsContent.displayName = TabsPrimitive.Content.displayName

export { Tabs, TabsList, TabsTrigger, TabsContent }
```

Create `react-spa/src/components/ui/scroll-area.tsx`:

```tsx
import * as React from "react"
import * as ScrollAreaPrimitive from "@radix-ui/react-scroll-area"
import { cn } from "../../lib/utils"

const ScrollArea = React.forwardRef<
  React.ComponentRef<typeof ScrollAreaPrimitive.Root>,
  React.ComponentPropsWithoutRef<typeof ScrollAreaPrimitive.Root>
>(({ className, children, ...props }, ref) => (
  <ScrollAreaPrimitive.Root ref={ref} className={cn("relative overflow-hidden", className)} {...props}>
    <ScrollAreaPrimitive.Viewport className="h-full w-full rounded-[inherit]">
      {children}
    </ScrollAreaPrimitive.Viewport>
    <ScrollBar />
    <ScrollAreaPrimitive.Corner />
  </ScrollAreaPrimitive.Root>
))
ScrollArea.displayName = ScrollAreaPrimitive.Root.displayName

const ScrollBar = React.forwardRef<
  React.ComponentRef<typeof ScrollAreaPrimitive.ScrollAreaScrollbar>,
  React.ComponentPropsWithoutRef<typeof ScrollAreaPrimitive.ScrollAreaScrollbar>
>(({ className, orientation = "vertical", ...props }, ref) => (
  <ScrollAreaPrimitive.ScrollAreaScrollbar
    ref={ref}
    orientation={orientation}
    className={cn(
      "flex touch-none select-none transition-colors",
      orientation === "vertical" && "h-full w-2.5 border-l border-l-transparent p-[1px]",
      orientation === "horizontal" && "h-2.5 flex-col border-t border-t-transparent p-[1px]",
      className
    )}
    {...props}
  >
    <ScrollAreaPrimitive.ScrollAreaThumb className="relative flex-1 rounded-full bg-border" />
  </ScrollAreaPrimitive.ScrollAreaScrollbar>
))
ScrollBar.displayName = ScrollAreaPrimitive.ScrollAreaScrollbar.displayName

export { ScrollArea, ScrollBar }
```

Create `react-spa/src/components/ui/separator.tsx`:

```tsx
import * as React from "react"
import * as SeparatorPrimitive from "@radix-ui/react-separator"
import { cn } from "../../lib/utils"

const Separator = React.forwardRef<
  React.ComponentRef<typeof SeparatorPrimitive.Root>,
  React.ComponentPropsWithoutRef<typeof SeparatorPrimitive.Root>
>(({ className, orientation = "horizontal", decorative = true, ...props }, ref) => (
  <SeparatorPrimitive.Root
    ref={ref}
    decorative={decorative}
    orientation={orientation}
    className={cn(
      "shrink-0 bg-border",
      orientation === "horizontal" ? "h-[1px] w-full" : "h-full w-[1px]",
      className
    )}
    {...props}
  />
))
Separator.displayName = SeparatorPrimitive.Root.displayName

export { Separator }
```

Create `react-spa/src/components/ui/skeleton.tsx`:

```tsx
import { cn } from "../../lib/utils"

function Skeleton({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div className={cn("animate-pulse rounded-md bg-primary/10", className)} {...props} />
  )
}

export { Skeleton }
```

Create `react-spa/src/components/ui/dropdown-menu.tsx` (simplified — just the parts we need):

```tsx
import * as React from "react"
import * as DropdownMenuPrimitive from "@radix-ui/react-dropdown-menu"
import { cn } from "../../lib/utils"

const DropdownMenu = DropdownMenuPrimitive.Root
const DropdownMenuTrigger = DropdownMenuPrimitive.Trigger

const DropdownMenuContent = React.forwardRef<
  React.ComponentRef<typeof DropdownMenuPrimitive.Content>,
  React.ComponentPropsWithoutRef<typeof DropdownMenuPrimitive.Content>
>(({ className, sideOffset = 4, ...props }, ref) => (
  <DropdownMenuPrimitive.Portal>
    <DropdownMenuPrimitive.Content
      ref={ref}
      sideOffset={sideOffset}
      className={cn(
        "z-50 min-w-[8rem] overflow-hidden rounded-md border bg-popover p-1 text-popover-foreground shadow-md",
        className
      )}
      {...props}
    />
  </DropdownMenuPrimitive.Portal>
))
DropdownMenuContent.displayName = DropdownMenuPrimitive.Content.displayName

const DropdownMenuItem = React.forwardRef<
  React.ComponentRef<typeof DropdownMenuPrimitive.Item>,
  React.ComponentPropsWithoutRef<typeof DropdownMenuPrimitive.Item>
>(({ className, ...props }, ref) => (
  <DropdownMenuPrimitive.Item
    ref={ref}
    className={cn(
      "relative flex cursor-default select-none items-center gap-2 rounded-sm px-2 py-1.5 text-sm outline-none transition-colors focus:bg-accent focus:text-accent-foreground data-[disabled]:pointer-events-none data-[disabled]:opacity-50",
      className
    )}
    {...props}
  />
))
DropdownMenuItem.displayName = DropdownMenuPrimitive.Item.displayName

export { DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem }
```

- [ ] **Step 6: Create src/pages/LoginPage.tsx**

```tsx
import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { api } from "../lib/api";
import { Button } from "../components/ui/button";
import { Input } from "../components/ui/input";
import { Card, CardContent } from "../components/ui/card";
import { Mail } from "lucide-react";

export default function LoginPage() {
  const navigate = useNavigate();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");

    try {
      await api("/auth/login", {
        method: "POST",
        body: JSON.stringify({ email, password }),
      });
      navigate("/dashboard");
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-indigo-500 to-blue-600">
      <Card className="w-full max-w-sm">
        <CardContent className="pt-6">
          <div className="flex flex-col items-center mb-6">
            <Mail className="h-10 w-10 text-primary mb-2" />
            <h1 className="text-2xl font-bold">Email App</h1>
          </div>
          <form onSubmit={handleSubmit} className="space-y-4">
            {error && (
              <p className="text-sm text-destructive bg-destructive/10 p-2 rounded">{error}</p>
            )}
            <Input
              type="email"
              placeholder="Email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
            />
            <Input
              type="password"
              placeholder="Password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
            />
            <Button type="submit" className="w-full" disabled={loading}>
              {loading ? "Logging in..." : "Login"}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
```

---

### Task 8: Sidebar, Empty State, Logout Button

**Files:**
- Create: `react-spa/src/components/sidebar.tsx`
- Create: `react-spa/src/components/empty-state.tsx`
- Create: `react-spa/src/components/logout-button.tsx`

- [ ] **Step 1: Create components/empty-state.tsx**

```tsx
import { Mail } from "lucide-react";

export default function EmptyState({
  icon,
  title,
  description,
}: {
  icon?: React.ReactNode;
  title: string;
  description?: string;
}) {
  return (
    <div className="flex flex-col items-center justify-center h-full text-muted-foreground">
      <div className="mb-4 opacity-20">
        {icon || <Mail className="h-16 w-16" />}
      </div>
      <p className="text-lg font-medium">{title}</p>
      {description && <p className="text-sm mt-1">{description}</p>}
    </div>
  );
}
```

- [ ] **Step 2: Create components/logout-button.tsx**

```tsx
import { useNavigate } from "react-router-dom";
import { api } from "../lib/api";
import { Button } from "./ui/button";
import { LogOut } from "lucide-react";

export default function LogoutButton() {
  const navigate = useNavigate();

  async function handleLogout() {
    await api("/auth/logout", { method: "POST" });
    navigate("/login");
  }

  return (
    <Button variant="ghost" size="sm" onClick={handleLogout} className="w-full justify-start gap-2">
      <LogOut className="h-4 w-4" />
      Logout
    </Button>
  );
}
```

- [ ] **Step 3: Create components/sidebar.tsx**

```tsx
import { useEffect, useState } from "react";
import { Link, useParams, useNavigate } from "react-router-dom";
import { api, type Mailbox, type User } from "../lib/api";
import { Button } from "./ui/button";
import { ScrollArea } from "./ui/scroll-area";
import { Separator } from "./ui/separator";
import { Avatar, AvatarFallback } from "./ui/avatar";
import LogoutButton from "./logout-button";
import { Mail, Plus, Settings } from "lucide-react";

export default function Sidebar({ onCompose }: { onCompose: () => void }) {
  const { mailboxId } = useParams();
  const navigate = useNavigate();
  const [mailboxes, setMailboxes] = useState<Mailbox[]>([]);
  const [user, setUser] = useState<User | null>(null);

  useEffect(() => {
    api<Mailbox[]>("/mailboxes").then(setMailboxes).catch(() => navigate("/login"));
    api<User>("/auth/me").then(setUser);
  }, []);

  return (
    <div className="w-64 shrink-0 border-r bg-card flex flex-col h-screen">
      <div className="p-4">
        <Link to="/dashboard" className="flex items-center gap-2 font-bold text-lg">
          <Mail className="h-5 w-5 text-primary" />
          Email App
        </Link>
      </div>

      <div className="px-3 mb-2">
        <Button onClick={onCompose} className="w-full gap-2">
          <Plus className="h-4 w-4" />
          Compose
        </Button>
      </div>

      <Separator />

      <ScrollArea className="flex-1 px-3 py-2">
        <div className="space-y-1">
          {mailboxes.map((mb) => (
            <Link
              key={mb.id}
              to={`/dashboard/${mb.id}`}
              className={`flex items-center gap-3 px-3 py-2 rounded-md text-sm transition-colors ${
                mailboxId === mb.id
                  ? "bg-accent text-accent-foreground font-medium"
                  : "hover:bg-accent/50"
              }`}
            >
              <Mail className="h-4 w-4 shrink-0 text-muted-foreground" />
              <span className="truncate">{mb.emailAddress}</span>
            </Link>
          ))}
        </div>
      </ScrollArea>

      <Separator />

      <div className="p-3 space-y-2">
        {user?.role === "admin" && (
          <Link
            to="/admin"
            className="flex items-center gap-2 px-3 py-2 rounded-md text-sm hover:bg-accent/50 transition-colors"
          >
            <Settings className="h-4 w-4" />
            Admin
          </Link>
        )}
        <div className="flex items-center gap-3 px-3 py-2">
          <Avatar className="h-8 w-8">
            <AvatarFallback>
              {user?.name?.charAt(0)?.toUpperCase() || "?"}
            </AvatarFallback>
          </Avatar>
          <div className="flex-1 min-w-0">
            <p className="text-sm font-medium truncate">{user?.name}</p>
            <p className="text-xs text-muted-foreground truncate">{user?.email}</p>
          </div>
        </div>
        <LogoutButton />
      </div>
    </div>
  );
}
```

---

### Task 9: Dashboard & Inbox Pages

**Files:**
- Create: `react-spa/src/pages/DashboardPage.tsx`
- Create: `react-spa/src/pages/InboxPage.tsx`
- Create: `react-spa/src/components/email-list.tsx`
- Create: `react-spa/src/components/email-card.tsx`

- [ ] **Step 1: Create pages/DashboardPage.tsx**

```tsx
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { api, type Mailbox } from "../lib/api";
import Sidebar from "../components/sidebar";
import EmptyState from "../components/empty-state";
import { Skeleton } from "../components/ui/skeleton";

export default function DashboardPage() {
  const navigate = useNavigate();
  const [mailboxes, setMailboxes] = useState<Mailbox[] | null>(null);

  useEffect(() => {
    api<Mailbox[]>("/mailboxes")
      .then((list) => {
        setMailboxes(list);
        if (list.length > 0) {
          navigate(`/dashboard/${list[0].id}`, { replace: true });
        }
      })
      .catch(() => navigate("/login"));
  }, []);

  if (mailboxes === null) {
    return (
      <div className="flex h-screen">
        <div className="w-64 border-r p-4 space-y-3">
          <Skeleton className="h-8 w-full" />
          <Skeleton className="h-8 w-full" />
          <Skeleton className="h-8 w-full" />
        </div>
        <div className="flex-1 p-6 space-y-4">
          <Skeleton className="h-6 w-48" />
          <Skeleton className="h-20 w-full" />
          <Skeleton className="h-20 w-full" />
        </div>
      </div>
    );
  }

  if (mailboxes.length === 0) {
    return (
      <div className="flex h-screen">
        <Sidebar onCompose={() => {}} />
        <div className="flex-1 flex items-center justify-center">
          <EmptyState
            title="No mailboxes"
            description="Contact your administrator to get access."
          />
        </div>
      </div>
    );
  }

  return null;
}
```

- [ ] **Step 2: Create components/email-card.tsx**

```tsx
import { Link } from "react-router-dom";
import { cn } from "../lib/utils";
import { Avatar, AvatarFallback } from "./ui/avatar";
import { Star } from "lucide-react";
import type { EmailItem } from "../lib/api";

export default function EmailCard({
  email,
  mailboxId,
  selected,
  onStarToggle,
}: {
  email: EmailItem;
  mailboxId: string;
  selected?: boolean;
  onStarToggle?: () => void;
}) {
  const initials = (email.direction === "outbound" ? email.to : email.from)
    .charAt(0)
    .toUpperCase();

  return (
    <Link
      to={`/dashboard/${mailboxId}/${email.id}`}
      className={cn(
        "block border-b px-4 py-3 transition-colors hover:bg-accent/50",
        !email.read && "bg-primary/5",
        selected && "bg-accent"
      )}
    >
      <div className="flex items-start gap-3">
        <Avatar className="h-8 w-8 shrink-0 mt-0.5">
          <AvatarFallback>{initials}</AvatarFallback>
        </Avatar>
        <div className="flex-1 min-w-0">
          <div className="flex items-center justify-between gap-2">
            <span className={cn("text-sm truncate", !email.read && "font-semibold")}>
              {email.direction === "outbound" ? `To: ${email.to}` : email.from}
            </span>
            <span className="text-xs text-muted-foreground shrink-0">
              {formatDate(email.createdAt)}
            </span>
          </div>
          <p className={cn("text-sm truncate mt-0.5", !email.read && "font-medium")}>
            {email.subject}
          </p>
          <p className="text-xs text-muted-foreground truncate mt-0.5">
            {email.textBody || email.htmlBody?.replace(/<[^>]*>/g, "") || ""}
          </p>
        </div>
        <button
          onClick={(e) => {
            e.preventDefault();
            e.stopPropagation();
            onStarToggle?.();
          }}
          className="shrink-0 mt-1"
        >
          <Star
            className={cn(
              "h-4 w-4",
              email.starred
                ? "fill-yellow-400 text-yellow-400"
                : "text-muted-foreground opacity-0 group-hover:opacity-100"
            )}
          />
        </button>
      </div>
    </Link>
  );
}

function formatDate(ts: number) {
  const d = new Date(ts);
  const now = new Date();
  const diff = now.getTime() - d.getTime();
  if (diff < 86400000) return d.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
  if (diff < 604800000) return d.toLocaleDateString([], { weekday: "short" });
  return d.toLocaleDateString([], { month: "short", day: "numeric" });
}
```

- [ ] **Step 3: Create components/email-list.tsx**

```tsx
import { useState, useEffect, useRef, useCallback } from "react";
import { useParams } from "react-router-dom";
import { api, type EmailItem, type PaginatedResponse } from "../lib/api";
import EmailCard from "./email-card";
import { Input } from "./ui/input";
import { Button } from "./ui/button";
import { Skeleton } from "./ui/skeleton";
import { Search, RefreshCw } from "lucide-react";
import { cn } from "../lib/utils";

type Tab = "inbox" | "sent" | "starred";

export default function EmailList({ mailboxId }: { mailboxId: string }) {
  const { emailId } = useParams();
  const [emails, setEmails] = useState<EmailItem[]>([]);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(false);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [tab, setTab] = useState<Tab>("inbox");
  const sentinelRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const timer = setTimeout(() => setDebouncedSearch(search), 300);
    return () => clearTimeout(timer);
  }, [search]);

  async function fetchEmails(reset: boolean) {
    if (reset) { setLoading(true); } else { setLoadingMore(true); }

    const pageNum = reset ? 1 : page + 1;
    const params = new URLSearchParams({
      mailboxId,
      page: String(pageNum),
      limit: "20",
    });
    if (tab === "starred") {
      params.set("starred", "true");
    } else {
      params.set("direction", tab);
    }
    if (debouncedSearch) params.set("search", debouncedSearch);

    try {
      const data = await api<PaginatedResponse>(`/emails?${params}`);
      if (reset) {
        setEmails(data.emails);
        setPage(1);
      } else {
        setEmails((prev) => [...prev, ...data.emails]);
        setPage(pageNum);
      }
      setHasMore(data.hasMore);
    } catch {
      // ignore
    } finally {
      setLoading(false);
      setLoadingMore(false);
    }
  }

  useEffect(() => {
    fetchEmails(true);
  }, [mailboxId, tab, debouncedSearch]);

  useEffect(() => {
    if (!sentinelRef.current) return;
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && hasMore && !loadingMore) {
          fetchEmails(false);
        }
      },
      { threshold: 0.1 }
    );
    observer.observe(sentinelRef.current);
    return () => observer.disconnect();
  }, [hasMore, loadingMore, page]);

  async function handleStarToggle(email: EmailItem) {
    try {
      await api(`/emails/${email.id}/star`, {
        method: "PATCH",
        body: JSON.stringify({ starred: !email.starred }),
      });
      setEmails((prev) =>
        prev.map((e) => (e.id === email.id ? { ...e, starred: !e.starred } : e))
      );
    } catch {}
  }

  const tabs: { key: Tab; label: string }[] = [
    { key: "inbox", label: "Inbox" },
    { key: "sent", label: "Sent" },
    { key: "starred", label: "Starred" },
  ];

  return (
    <div className="w-96 shrink-0 border-r flex flex-col h-full bg-card">
      <div className="p-3 border-b space-y-2">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search emails..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-9"
          />
        </div>
        <div className="flex items-center justify-between">
          <div className="flex gap-1">
            {tabs.map((t) => (
              <button
                key={t.key}
                onClick={() => setTab(t.key)}
                className={cn(
                  "px-3 py-1 text-sm rounded-md transition-colors",
                  tab === t.key
                    ? "bg-primary text-primary-foreground"
                    : "hover:bg-accent"
                )}
              >
                {t.label}
              </button>
            ))}
          </div>
          <Button
            variant="ghost"
            size="icon"
            onClick={() => fetchEmails(true)}
            title="Refresh"
          >
            <RefreshCw className="h-4 w-4" />
          </Button>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto">
        {loading ? (
          <div className="p-4 space-y-3">
            {Array.from({ length: 5 }).map((_, i) => (
              <div key={i} className="flex gap-3">
                <Skeleton className="h-8 w-8 rounded-full" />
                <div className="flex-1 space-y-2">
                  <Skeleton className="h-4 w-3/4" />
                  <Skeleton className="h-3 w-1/2" />
                </div>
              </div>
            ))}
          </div>
        ) : emails.length === 0 ? (
          <div className="flex items-center justify-center h-full text-muted-foreground text-sm p-4 text-center">
            {debouncedSearch
              ? "No emails match your search."
              : tab === "starred"
              ? "No starred emails."
              : tab === "sent"
              ? "No sent emails yet."
              : "No emails yet."}
          </div>
        ) : (
          <div>
            {emails.map((email) => (
              <EmailCard
                key={email.id}
                email={email}
                mailboxId={mailboxId}
                selected={email.id === emailId}
                onStarToggle={() => handleStarToggle(email)}
              />
            ))}
            {hasMore && (
              <div ref={sentinelRef} className="py-4 text-center text-sm text-muted-foreground">
                {loadingMore ? "Loading more..." : "Scroll for more"}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
```

- [ ] **Step 4: Create pages/InboxPage.tsx**

```tsx
import { useState } from "react";
import { useParams } from "react-router-dom";
import Sidebar from "../components/sidebar";
import EmailList from "../components/email-list";
import ThreadView from "../components/thread-view";
import ComposeDialog from "../components/compose-dialog";
import EmptyState from "../components/empty-state";

export default function InboxPage() {
  const { mailboxId, emailId } = useParams();
  const [composeOpen, setComposeOpen] = useState(false);

  if (!mailboxId) return null;

  return (
    <div className="flex h-screen">
      <Sidebar onCompose={() => setComposeOpen(true)} />
      <EmailList mailboxId={mailboxId} />
      <div className="flex-1 flex flex-col h-full overflow-hidden">
        {emailId ? (
          <ThreadView mailboxId={mailboxId} emailId={emailId} />
        ) : (
          <EmptyState
            title="Select an email to read"
            description="Choose an email from the list to view its contents."
          />
        )}
      </div>
      <ComposeDialog
        mailboxId={mailboxId}
        open={composeOpen}
        onOpenChange={setComposeOpen}
      />
    </div>
  );
}
```

---

### Task 10: Thread View & Compose Dialog

**Files:**
- Create: `react-spa/src/components/thread-view.tsx`
- Create: `react-spa/src/components/compose-dialog.tsx`
- Create: `react-spa/src/components/rich-text-editor.tsx`

- [ ] **Step 1: Create components/rich-text-editor.tsx**

```tsx
import { useEditor, EditorContent } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import Placeholder from "@tiptap/extension-placeholder";
import { Button } from "./ui/button";
import { Bold, Italic, List, ListOrdered } from "lucide-react";

export default function RichTextEditor({
  value,
  onChange,
  placeholder,
}: {
  value: string;
  onChange: (html: string) => void;
  placeholder?: string;
}) {
  const editor = useEditor({
    extensions: [
      StarterKit,
      Placeholder.configure({ placeholder: placeholder || "Write something..." }),
    ],
    content: value,
    onUpdate: ({ editor }) => onChange(editor.getHTML()),
    editorProps: {
      attributes: {
        class: "prose prose-sm max-w-none min-h-[200px] px-3 py-2 focus:outline-none",
      },
    },
  });

  if (!editor) return null;

  return (
    <div className="border rounded-md">
      <div className="flex gap-1 px-2 py-1.5 border-b bg-muted/50">
        <Button
          variant="ghost"
          size="icon"
          onClick={() => editor.chain().focus().toggleBold().run()}
          data-active={editor.isActive("bold")}
          className="h-8 w-8 data-[active=true]:bg-accent"
        >
          <Bold className="h-4 w-4" />
        </Button>
        <Button
          variant="ghost"
          size="icon"
          onClick={() => editor.chain().focus().toggleItalic().run()}
          data-active={editor.isActive("italic")}
          className="h-8 w-8 data-[active=true]:bg-accent"
        >
          <Italic className="h-4 w-4" />
        </Button>
        <Button
          variant="ghost"
          size="icon"
          onClick={() => editor.chain().focus().toggleBulletList().run()}
          data-active={editor.isActive("bulletList")}
          className="h-8 w-8 data-[active=true]:bg-accent"
        >
          <List className="h-4 w-4" />
        </Button>
        <Button
          variant="ghost"
          size="icon"
          onClick={() => editor.chain().focus().toggleOrderedList().run()}
          data-active={editor.isActive("orderedList")}
          className="h-8 w-8 data-[active=true]:bg-accent"
        >
          <ListOrdered className="h-4 w-4" />
        </Button>
      </div>
      <EditorContent editor={editor} />
    </div>
  );
}
```

- [ ] **Step 2: Create components/compose-dialog.tsx**

```tsx
import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { api } from "../lib/api";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "./ui/dialog";
import { Input } from "./ui/input";
import { Button } from "./ui/button";
import RichTextEditor from "./rich-text-editor";
import { Send } from "lucide-react";

export default function ComposeDialog({
  mailboxId,
  open,
  onOpenChange,
  initialTo,
  initialSubject,
  inReplyTo,
}: {
  mailboxId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  initialTo?: string;
  initialSubject?: string;
  inReplyTo?: string;
}) {
  const navigate = useNavigate();
  const [to, setTo] = useState(initialTo || "");
  const [subject, setSubject] = useState(initialSubject || "");
  const [body, setBody] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");

    try {
      await api("/emails/send", {
        method: "POST",
        body: JSON.stringify({
          mailboxId,
          to,
          subject,
          htmlBody: body,
          textBody: body.replace(/<[^>]*>/g, ""),
          ...(inReplyTo ? { inReplyTo } : {}),
        }),
      });
      onOpenChange(false);
      navigate(`/dashboard/${mailboxId}`);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-xl">
        <DialogHeader>
          <DialogTitle>{inReplyTo ? "Reply" : "New Email"}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          {error && (
            <p className="text-sm text-destructive bg-destructive/10 p-2 rounded">{error}</p>
          )}
          <Input
            type="email"
            placeholder="To"
            value={to}
            onChange={(e) => setTo(e.target.value)}
            required
          />
          <Input
            type="text"
            placeholder="Subject"
            value={subject}
            onChange={(e) => setSubject(e.target.value)}
            required
          />
          <RichTextEditor value={body} onChange={setBody} placeholder="Write your message..." />
          <div className="flex justify-end gap-2">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={loading} className="gap-2">
              <Send className="h-4 w-4" />
              {loading ? "Sending..." : "Send"}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
```

- [ ] **Step 3: Create components/thread-view.tsx**

```tsx
import { useEffect, useState } from "react";
import { api, type EmailItem, type Attachment } from "../lib/api";
import { Avatar, AvatarFallback } from "./ui/avatar";
import { Badge } from "./ui/badge";
import { Separator } from "./ui/separator";
import { ScrollArea } from "./ui/scroll-area";
import { Button } from "./ui/button";
import ComposeDialog from "./compose-dialog";
import { Skeleton } from "./ui/skeleton";
import { ArrowUp, Paperclip, Reply } from "lucide-react";

export default function ThreadView({
  mailboxId,
  emailId,
}: {
  mailboxId: string;
  emailId: string;
}) {
  const [emails, setEmails] = useState<EmailItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [replyOpen, setReplyOpen] = useState(false);

  useEffect(() => {
    setLoading(true);
    api<{ emails: EmailItem[] }>(`/emails/${emailId}/thread`)
      .then((data) => setEmails(data.emails))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [emailId]);

  if (loading) {
    return (
      <div className="p-6 space-y-4 flex-1">
        <Skeleton className="h-6 w-1/2" />
        <Skeleton className="h-4 w-1/4" />
        <Skeleton className="h-32 w-full" />
        <Skeleton className="h-32 w-full" />
      </div>
    );
  }

  if (emails.length === 0) return null;

  const lastEmail = emails[emails.length - 1];

  function sanitizeHtml(html: string): string {
    return html
      .replace(/<script[^>]*>[\s\S]*?<\/script>/gi, "")
      .replace(/<iframe[^>]*>[\s\S]*?<\/iframe>/gi, "")
      .replace(/\son\w+="[^"]*"/gi, "")
      .replace(/\son\w+='[^']*'/gi, "");
  }

  function formatSize(bytes: number): string {
    if (bytes < 1024) return bytes + " B";
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + " KB";
    return (bytes / (1024 * 1024)).toFixed(1) + " MB";
  }

  return (
    <div className="flex-1 flex flex-col h-full">
      <ScrollArea className="flex-1">
        <div className="p-6">
          <h2 className="text-xl font-semibold mb-1">{emails[0].subject}</h2>
          <p className="text-sm text-muted-foreground mb-6">
            {emails.length} message{emails.length > 1 ? "s" : ""}
          </p>

          <div className="space-y-6">
            {emails.map((email, i) => (
              <div key={email.id}>
                <div className="flex items-start gap-3">
                  <Avatar className="h-8 w-8">
                    <AvatarFallback>
                      {(email.direction === "outbound" ? email.from : email.from).charAt(0).toUpperCase()}
                    </AvatarFallback>
                  </Avatar>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center justify-between gap-2">
                      <div className="flex items-center gap-2">
                        <span className="font-medium text-sm">{email.from}</span>
                        <span className="text-xs text-muted-foreground">to {email.to}</span>
                        {email.direction === "outbound" && (
                          <Badge variant="secondary" className="gap-1">
                            <ArrowUp className="h-3 w-3" />
                            Sent
                          </Badge>
                        )}
                      </div>
                      <span className="text-xs text-muted-foreground shrink-0">
                        {new Date(email.createdAt).toLocaleString()}
                      </span>
                    </div>
                    <div
                      className="mt-2 text-sm prose prose-sm max-w-none"
                      dangerouslySetInnerHTML={{
                        __html: email.htmlBody
                          ? sanitizeHtml(email.htmlBody)
                          : email.textBody?.replace(/\n/g, "<br>") || "",
                      }}
                    />

                    {email.attachments && email.attachments.length > 0 && (
                      <div className="mt-3 space-y-1">
                        {email.attachments.map((att: Attachment) => (
                          <div key={att.id} className="flex items-center gap-2 text-sm text-muted-foreground">
                            <Paperclip className="h-3 w-3" />
                            <span>{att.filename}</span>
                            <span className="text-xs">({formatSize(att.size)})</span>
                            <Button
                              variant="link"
                              size="sm"
                              className="h-auto p-0 text-xs"
                              onClick={async () => {
                                try {
                                  const data = await api<{ download_url: string }>(
                                    `/emails/${email.id}/attachments/${att.id}`
                                  );
                                  window.open(data.download_url, "_blank");
                                } catch {}
                              }}
                            >
                              Download
                            </Button>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                </div>
                {i < emails.length - 1 && <Separator className="my-4" />}
              </div>
            ))}
          </div>
        </div>
      </ScrollArea>

      <div className="border-t p-4">
        <Button
          variant="outline"
          className="gap-2"
          onClick={() => setReplyOpen(true)}
        >
          <Reply className="h-4 w-4" />
          Reply
        </Button>
      </div>

      <ComposeDialog
        mailboxId={mailboxId}
        open={replyOpen}
        onOpenChange={setReplyOpen}
        initialTo={
          lastEmail.direction === "inbound" ? lastEmail.from : lastEmail.to
        }
        initialSubject={
          lastEmail.subject.startsWith("Re:")
            ? lastEmail.subject
            : `Re: ${lastEmail.subject}`
        }
        inReplyTo={lastEmail.messageId}
      />
    </div>
  );
}
```

---

### Task 11: Admin Page

**Files:**
- Create: `react-spa/src/pages/AdminPage.tsx`
- Create: `react-spa/src/components/admin-users.tsx`
- Create: `react-spa/src/components/admin-mailboxes.tsx`

- [ ] **Step 1: Create components/admin-users.tsx**

```tsx
import { useState, useEffect } from "react";
import { api } from "../lib/api";
import { Button } from "./ui/button";
import { Input } from "./ui/input";
import { Badge } from "./ui/badge";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "./ui/dialog";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";
import { Plus, Trash2, Shield, ShieldOff } from "lucide-react";

interface AdminUser {
  id: string;
  email: string;
  name: string;
  role: string;
  mailboxCount: number;
}

export default function AdminUsers() {
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [createOpen, setCreateOpen] = useState(false);
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  useEffect(() => {
    api<AdminUser[]>("/admin/users").then(setUsers);
  }, []);

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    await api("/admin/users", {
      method: "POST",
      body: JSON.stringify({ name, email, password }),
    });
    setCreateOpen(false);
    setName("");
    setEmail("");
    setPassword("");
    api<AdminUser[]>("/admin/users").then(setUsers);
  }

  async function handleDelete(id: string) {
    await api(`/admin/users/${id}`, { method: "DELETE" });
    setUsers((prev) => prev.filter((u) => u.id !== id));
  }

  async function handleRole(id: string, role: string) {
    await api(`/admin/users/${id}/role`, {
      method: "PATCH",
      body: JSON.stringify({ role }),
    });
    setUsers((prev) => prev.map((u) => (u.id === id ? { ...u, role } : u)));
  }

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle>Users ({users.length})</CardTitle>
        <Button size="sm" onClick={() => setCreateOpen(true)} className="gap-1">
          <Plus className="h-4 w-4" /> Create User
        </Button>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          {users.map((u) => (
            <div key={u.id} className="flex items-center justify-between p-3 rounded-lg border">
              <div className="flex items-center gap-3">
                <div className="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center text-sm font-medium">
                  {u.name.charAt(0).toUpperCase()}
                </div>
                <div>
                  <p className="text-sm font-medium">{u.name}</p>
                  <p className="text-xs text-muted-foreground">{u.email}</p>
                </div>
              </div>
              <div className="flex items-center gap-2">
                <Badge variant={u.role === "admin" ? "default" : "secondary"}>
                  {u.role}
                </Badge>
                <span className="text-xs text-muted-foreground">{u.mailboxCount} mailboxes</span>
                {u.role === "admin" ? (
                  <Button variant="ghost" size="icon" onClick={() => handleRole(u.id, "user")}>
                    <ShieldOff className="h-4 w-4" />
                  </Button>
                ) : (
                  <Button variant="ghost" size="icon" onClick={() => handleRole(u.id, "admin")}>
                    <Shield className="h-4 w-4" />
                  </Button>
                )}
                <Button variant="ghost" size="icon" onClick={() => handleDelete(u.id)}>
                  <Trash2 className="h-4 w-4 text-destructive" />
                </Button>
              </div>
            </div>
          ))}
        </div>
      </CardContent>

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create User</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleCreate} className="space-y-4">
            <Input placeholder="Name" value={name} onChange={(e) => setName(e.target.value)} required />
            <Input type="email" placeholder="Email" value={email} onChange={(e) => setEmail(e.target.value)} required />
            <Input type="password" placeholder="Password" value={password} onChange={(e) => setPassword(e.target.value)} required />
            <Button type="submit" className="w-full">Create</Button>
          </form>
        </DialogContent>
      </Dialog>
    </Card>
  );
}
```

- [ ] **Step 2: Create components/admin-mailboxes.tsx**

```tsx
import { useState, useEffect } from "react";
import { api } from "../lib/api";
import { Button } from "./ui/button";
import { Input } from "./ui/input";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "./ui/dialog";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";
import { Plus, Trash2 } from "lucide-react";

interface AdminMailbox {
  id: string;
  emailAddress: string;
  userCount: number;
}

export default function AdminMailboxes() {
  const [mailboxes, setMailboxes] = useState<AdminMailbox[]>([]);
  const [createOpen, setCreateOpen] = useState(false);
  const [localPart, setLocalPart] = useState("");

  useEffect(() => {
    api<AdminMailbox[]>("/admin/mailboxes").then(setMailboxes);
  }, []);

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    await api("/admin/mailboxes", {
      method: "POST",
      body: JSON.stringify({ localPart }),
    });
    setCreateOpen(false);
    setLocalPart("");
    api<AdminMailbox[]>("/admin/mailboxes").then(setMailboxes);
  }

  async function handleDelete(id: string) {
    try {
      await api(`/admin/mailboxes/${id}`, { method: "DELETE" });
      setMailboxes((prev) => prev.filter((m) => m.id !== id));
    } catch (err: any) {
      alert(err.message);
    }
  }

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle>Mailboxes ({mailboxes.length})</CardTitle>
        <Button size="sm" onClick={() => setCreateOpen(true)} className="gap-1">
          <Plus className="h-4 w-4" /> Create Mailbox
        </Button>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          {mailboxes.map((mb) => (
            <div key={mb.id} className="flex items-center justify-between p-3 rounded-lg border">
              <div>
                <p className="text-sm font-medium">{mb.emailAddress}</p>
                <p className="text-xs text-muted-foreground">{mb.userCount} user(s)</p>
              </div>
              <Button variant="ghost" size="icon" onClick={() => handleDelete(mb.id)}>
                <Trash2 className="h-4 w-4 text-destructive" />
              </Button>
            </div>
          ))}
        </div>
      </CardContent>

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create Mailbox</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleCreate} className="space-y-4">
            <div className="flex items-center gap-2">
              <Input
                placeholder="local-part"
                value={localPart}
                onChange={(e) => setLocalPart(e.target.value)}
                required
                className="flex-1"
              />
              <span className="text-sm text-muted-foreground">@domain.com</span>
            </div>
            <Button type="submit" className="w-full">Create</Button>
          </form>
        </DialogContent>
      </Dialog>
    </Card>
  );
}
```

- [ ] **Step 3: Create pages/AdminPage.tsx**

```tsx
import { useState } from "react";
import Sidebar from "../components/sidebar";
import AdminUsers from "../components/admin-users";
import AdminMailboxes from "../components/admin-mailboxes";
import ComposeDialog from "../components/compose-dialog";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "../components/ui/tabs";

export default function AdminPage() {
  const [composeOpen, setComposeOpen] = useState(false);

  return (
    <div className="flex h-screen">
      <Sidebar onCompose={() => setComposeOpen(true)} />
      <div className="flex-1 p-6 overflow-y-auto">
        <h1 className="text-2xl font-bold mb-6">Admin</h1>
        <Tabs defaultValue="users">
          <TabsList>
            <TabsTrigger value="users">Users</TabsTrigger>
            <TabsTrigger value="mailboxes">Mailboxes</TabsTrigger>
          </TabsList>
          <TabsContent value="users"><AdminUsers /></TabsContent>
          <TabsContent value="mailboxes"><AdminMailboxes /></TabsContent>
        </Tabs>
      </div>
      <ComposeDialog
        mailboxId=""
        open={composeOpen}
        onOpenChange={setComposeOpen}
      />
    </div>
  );
}
```

---

### Task 12: Production Build Setup

**Files:**
- Modify: `server/main.go` (add static file serving)
- Modify: `react-spa/vite.config.ts` (set base path)

- [ ] **Step 1: Update vite.config.ts for production build**

No change needed — proxy config only applies in dev mode. Build output goes to `dist/`.

- [ ] **Step 2: Update server/main.go to serve static files**

Add this block after `r := gin.Default()` and before the API routes:

```go
// Serve static files in production
r.Use(func(c *gin.Context) {
	// Skip API routes
	if strings.HasPrefix(c.Request.URL.Path, "/api") {
		c.Next()
		return
	}
	// Try to serve the file
	staticPath := "../react-spa/dist"
	filePath := staticPath + c.Request.URL.Path
	if _, err := os.Stat(filePath); err == nil {
		c.File(filePath)
		c.Abort()
		return
	}
	// Fallback to index.html for SPA routing
	c.File(staticPath + "/index.html")
	c.Abort()
})
```

Add `"os"` and `"strings"` to the imports in main.go.

- [ ] **Step 3: Build React and test**

```bash
cd /home/luthfi9251/project/email-app/react-spa
npm run build

cd /home/luthfi9251/project/email-app/server
cp .env.example .env
# edit .env with real values
go run .
```

---

### Self-Review Checklist

- [ ] **Spec coverage:** Auth (Task 3, 7), mailbox list (Task 5, 8), email list with pagination/search/star (Task 5, 9), thread view (Task 10), compose/send/reply (Task 5, 10), admin panel (Task 6, 11), webhook (Task 6), rich text editor (Task 10) — all covered.
- [ ] **Placeholder scan:** No TBD, TODO, or placeholder steps.
- [ ] **Type consistency:** Go models use `bson.ObjectID` consistently. Email `messageId` is string unique. Thread grouping via `threadId` field. API responses match frontend types.
