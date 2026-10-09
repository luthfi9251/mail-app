package handlers

import (
	"net/http"
	"os"
	"strings"
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

func SetupStatus(c *gin.Context) {
	count, err := database.GetCollection("users").CountDocuments(c.Request.Context(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check setup status"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"needsSetup": count == 0})
}

func Setup(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required"`
		Name     string `json:"name" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email, name and password required"})
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
		Role:      "admin",
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
	if parts[0] != "" && (len(parts) != 2 || parts[1] != mailboxDomain()) {
		address = parts[0] + "@" + mailboxDomain()
	}
	if err := createOwnedMailbox(c.Request.Context(), result.InsertedID.(bson.ObjectID), address); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "admin created but failed to create mailbox"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true})
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

func ChangePassword(c *gin.Context) {
	var req struct {
		CurrentPassword string `json:"currentPassword" binding:"required"`
		NewPassword     string `json:"newPassword" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "currentPassword and newPassword required"})
		return
	}

	userID, _ := c.Get("userId")
	col := database.GetCollection("users")

	var user models.User
	err := col.FindOne(c.Request.Context(), bson.M{"_id": MustObjectID(userID.(string))}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "current password is incorrect"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	_, err = col.UpdateByID(c.Request.Context(), user.ID, bson.M{"$set": bson.M{"password": string(hash)}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func MustObjectID(hex string) bson.ObjectID {
	id, _ := bson.ObjectIDFromHex(hex)
	return id
}
