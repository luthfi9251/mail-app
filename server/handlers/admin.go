package handlers

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
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

	umCol := database.GetCollection("usermailboxes")
	var result []gin.H
	for _, u := range users {
		count, _ := umCol.CountDocuments(c.Request.Context(), bson.M{"userId": u.ID})
		result = append(result, gin.H{
			"id":           u.ID.Hex(),
			"email":        u.Email,
			"name":         u.Name,
			"role":         u.Role,
			"mailboxCount": count,
			"createdAt":    u.CreatedAt,
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
			"id":           mb.ID.Hex(),
			"emailAddress": mb.EmailAddress,
			"userCount":    count,
			"createdAt":    mb.CreatedAt,
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
