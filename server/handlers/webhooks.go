package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"

	"email-app/config"
	"email-app/database"
	"email-app/models"
)

const webhookTolerance = 5 * time.Minute

// verifySvixSignature checks a Resend webhook the way Svix signs it:
// base64(HMAC-SHA256(secret, "<id>.<timestamp>.<body>")) in "v1,<sig>" entries.
func verifySvixSignature(secret string, h http.Header, body []byte, now time.Time) error {
	id, ts, sigs := h.Get("svix-id"), h.Get("svix-timestamp"), h.Get("svix-signature")
	if id == "" || ts == "" || sigs == "" {
		return errors.New("missing signature headers")
	}
	sec, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return errors.New("bad timestamp")
	}
	if d := now.Sub(time.Unix(sec, 0)); d > webhookTolerance || d < -webhookTolerance {
		return errors.New("timestamp outside tolerance")
	}
	key, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(secret, "whsec_"))
	if err != nil {
		return errors.New("bad webhook secret")
	}
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(id + "." + ts + "."))
	mac.Write(body)
	expected := mac.Sum(nil)
	for _, entry := range strings.Fields(sigs) {
		version, sig, ok := strings.Cut(entry, ",")
		if !ok || version != "v1" {
			continue
		}
		if got, err := base64.StdEncoding.DecodeString(sig); err == nil && hmac.Equal(got, expected) {
			return nil
		}
	}
	return errors.New("signature mismatch")
}

func ResendWebhook(c *gin.Context) {
	if config.WebhookSecret == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "webhook disabled"})
		return
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "payload too large"})
		return
	}
	if err := verifySvixSignature(config.WebhookSecret, c.Request.Header, body, time.Now()); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
		return
	}

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
	if err := json.Unmarshal(body, &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	mbCol := database.GetCollection("mailboxes")
	var mailbox models.Mailbox
	err = mbCol.FindOne(c.Request.Context(), bson.M{"emailAddress": payload.To}).Decode(&mailbox)
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

	threadID := resolveThreadID(c.Request.Context(), mailbox.ID, messageID, references, inReplyTo)
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

// resolveThreadID picks the thread for a message within one mailbox. Lookups
// are always scoped to mailboxID so header values chosen by an outside sender
// can never attach a message to (or reveal) another mailbox's thread. A
// message with no reply headers starts its own thread.
func resolveThreadID(ctx context.Context, mailboxID bson.ObjectID, messageID string, references []string, inReplyTo *string) string {
	if inReplyTo != nil && *inReplyTo != "" {
		var original models.Email
		err := database.GetCollection("emails").FindOne(ctx, bson.M{"mailboxId": mailboxID, "messageId": *inReplyTo}).Decode(&original)
		if err == nil {
			return original.ThreadID
		}
		return *inReplyTo
	}
	if len(references) > 0 {
		return references[0]
	}
	return messageID
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
