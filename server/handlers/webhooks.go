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
