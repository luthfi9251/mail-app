package handlers

import (
	"net/http"

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
	result := []gin.H{}
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
	if !requireMailboxAccess(c, c.Param("id")) {
		return
	}

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
