package handlers

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"

	"email-app/config"
	"email-app/database"
	"email-app/middleware"
	"email-app/models"
)

var (
	// Per-account limits complement the per-IP limit on the routes, so a
	// distributed guesser still can't hammer a single account.
	loginAccountLimiter   = middleware.NewLimiter(10, 15*time.Minute)
	changePasswordLimiter = middleware.NewLimiter(10, 15*time.Minute)

	// Compared against when the email is unknown, so a miss costs the same
	// bcrypt time as a hit and response timing doesn't reveal valid emails.
	dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-password-for-timing"), bcrypt.DefaultCost)
)

const minPasswordLen = 8

func validatePassword(p string) error {
	if len(p) < minPasswordLen {
		return errors.New("password must be at least 8 characters")
	}
	if len(p) > 72 { // bcrypt only uses the first 72 bytes
		return errors.New("password must be at most 72 bytes")
	}
	return nil
}

func validEmail(s string) bool {
	a, err := mail.ParseAddress(s)
	return err == nil && a.Address == s
}

func Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email and password required"})
		return
	}

	accountKey := strings.ToLower(strings.TrimSpace(req.Email))
	if !loginAccountLimiter.Allow(accountKey) {
		middleware.TooManyRequests(c)
		return
	}

	col := database.GetCollection("users")
	var user models.User
	err := col.FindOne(c.Request.Context(), bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		bcrypt.CompareHashAndPassword(dummyHash, []byte(req.Password))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	ok, legacy := checkPassword(user.Password, req.Password)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	loginAccountLimiter.Reset(accountKey)

	// Upgrade an imported v1 (scrypt) hash to bcrypt now that we know the password.
	if legacy {
		if hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost); err == nil {
			_, _ = col.UpdateByID(c.Request.Context(), user.ID, bson.M{"$set": bson.M{"password": string(hash)}})
		}
	}

	if err := middleware.IssueToken(c, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":    user.ID.Hex(),
			"email": user.Email,
			"name":  user.Name,
			"role":  user.Role,
		},
	})
}

func SetupStatus(c *gin.Context) {
	count, err := database.GetCollection("users").CountDocuments(c.Request.Context(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check setup status"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"needsSetup": count == 0, "setupEnabled": config.SetupToken != ""})
}

func Setup(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required"`
		Name     string `json:"name" binding:"required"`
		Password string `json:"password" binding:"required"`
		Token    string `json:"setupToken"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email, name and password required"})
		return
	}

	// The first admin can only be created by someone who knows SETUP_TOKEN,
	// so a freshly deployed instance can't be claimed by a stranger.
	if config.SetupToken == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "setup is disabled: set SETUP_TOKEN on the server"})
		return
	}
	if subtle.ConstantTimeCompare([]byte(req.Token), []byte(config.SetupToken)) != 1 {
		c.JSON(http.StatusForbidden, gin.H{"error": "invalid setup token"})
		return
	}
	if !validEmail(req.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid email address"})
		return
	}
	if err := validatePassword(req.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	col := database.GetCollection("users")
	count, err := col.CountDocuments(c.Request.Context(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check setup status"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "setup already completed"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// ponytail: count+insert race can create two admins on concurrent first setups; unique email index blocks same email, use a transaction if that matters.
	result, err := col.InsertOne(c.Request.Context(), models.User{
		Email:     req.Email,
		Name:      req.Name,
		Password:  string(hash),
		Role:      models.RoleAdmin,
		CreatedAt: time.Now().UnixMilli(),
	})
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
		return
	}

	// Use the admin's own address if it is on the mailbox domain, otherwise
	// fall back to <local-part>@<mailbox domain>.
	address := req.Email
	parts := strings.SplitN(req.Email, "@", 2)
	if parts[0] != "" && (len(parts) != 2 || parts[1] != config.MailboxDomain) {
		address = parts[0] + "@" + config.MailboxDomain
	}
	adminID := result.InsertedID.(bson.ObjectID)
	if err := createOwnedMailbox(c.Request.Context(), adminID, address); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "admin created but failed to create mailbox"})
		return
	}
	if err := assignAllMailboxesToUser(c.Request.Context(), adminID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "admin created but failed to assign mailboxes"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true})
}

func Logout(c *gin.Context) {
	middleware.ClearToken(c)
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

func ChangePassword(c *gin.Context) {
	var req struct {
		CurrentPassword string `json:"currentPassword" binding:"required"`
		NewPassword     string `json:"newPassword" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "currentPassword and newPassword required"})
		return
	}

	if err := validatePassword(req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userId")
	if !changePasswordLimiter.Allow(userID.(string)) {
		middleware.TooManyRequests(c)
		return
	}
	col := database.GetCollection("users")

	var user models.User
	err := col.FindOne(c.Request.Context(), bson.M{"_id": MustObjectID(userID.(string))}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if ok, _ := checkPassword(user.Password, req.CurrentPassword); !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "current password is incorrect"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// Bumping tokenVersion signs out every other session; this one gets a fresh token.
	_, err = col.UpdateByID(c.Request.Context(), user.ID, bson.M{
		"$set": bson.M{"password": string(hash)},
		"$inc": bson.M{"tokenVersion": 1},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
		return
	}
	changePasswordLimiter.Reset(userID.(string))
	user.TokenVersion++
	if err := middleware.IssueToken(c, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "password changed, please sign in again"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func MustObjectID(hex string) bson.ObjectID {
	id, _ := bson.ObjectIDFromHex(hex)
	return id
}
