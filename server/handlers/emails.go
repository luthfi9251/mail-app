package handlers

import (
	"crypto/rand"
	"fmt"
	"math"
	"net/http"
	"net/mail"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/resend/resend-go/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"email-app/database"
	"email-app/models"
)

// ponytail: GetAttachment removed — resend-go/v2 v2.28.0 has no Receiving.Attachments.Get API. Add when the SDK supports it.

func requireMailboxAccess(c *gin.Context, mailboxID string) bool {
	userID, _ := c.Get("userId")
	umCol := database.GetCollection("usermailboxes")
	count, err := umCol.CountDocuments(c.Request.Context(), bson.M{
		"userId":    MustObjectID(userID.(string)),
		"mailboxId": MustObjectID(mailboxID),
	})
	if err != nil || count == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return false
	}
	return true
}

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

	if !requireMailboxAccess(c, mailboxID) {
		return
	}

	filter := bson.M{"mailboxId": MustObjectID(mailboxID)}
	if direction != "" {
		filter["direction"] = direction
	}
	if starred == "true" {
		filter["starred"] = true
	}
	if len(search) > 100 {
		search = search[:100]
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
	if err := cursor.All(c.Request.Context(), &emails); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse emails"})
		return
	}

	result := []gin.H{}
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

	if !requireMailboxAccess(c, email.MailboxID.Hex()) {
		return
	}

	cursor, err := col.Find(
		c.Request.Context(),
		bson.M{"threadId": email.ThreadID, "mailboxId": email.MailboxID},
		options.Find().SetSort(bson.D{{Key: "createdAt", Value: 1}}),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch thread"})
		return
	}
	defer cursor.Close(c.Request.Context())

	var emails []models.Email
	if err := cursor.All(c.Request.Context(), &emails); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse thread"})
		return
	}

	// Opening a thread marks every message in it as read.
	_, _ = col.UpdateMany(
		c.Request.Context(),
		bson.M{"threadId": email.ThreadID, "mailboxId": email.MailboxID, "read": false},
		bson.M{"$set": bson.M{"read": true}},
	)

	result := []gin.H{}
	for _, e := range emails {
		e.Read = true
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

	if !requireMailboxAccess(c, email.MailboxID.Hex()) {
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

	recipients, err := mail.ParseAddressList(req.To)
	if err != nil || len(recipients) == 0 || len(recipients) > maxRecipients {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("to must be 1-%d valid email addresses", maxRecipients)})
		return
	}
	to := make([]string, len(recipients))
	for i, r := range recipients {
		to[i] = r.Address
	}
	if strings.ContainsAny(req.Subject, "\r\n") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subject"})
		return
	}
	if req.InReplyTo != "" && !messageIDRe.MatchString(req.InReplyTo) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid inReplyTo"})
		return
	}

	if !requireMailboxAccess(c, req.MailboxID) {
		return
	}

	mbCol := database.GetCollection("mailboxes")
	var mailbox models.Mailbox
	err = mbCol.FindOne(c.Request.Context(), bson.M{"_id": MustObjectID(req.MailboxID)}).Decode(&mailbox)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "mailbox not found"})
		return
	}

	// Show the sender's name in the From header: "Name <mailbox@domain>".
	from := mailbox.EmailAddress
	if uid, ok := c.Get("userId"); ok {
		var user models.User
		err := database.GetCollection("users").FindOne(
			c.Request.Context(), bson.M{"_id": MustObjectID(uid.(string))},
		).Decode(&user)
		if err == nil && user.Name != "" {
			from = (&mail.Address{Name: user.Name, Address: mailbox.EmailAddress}).String()
		}
	}

	client := resend.NewClient(os.Getenv("RESEND_API_KEY"))
	messageID := fmt.Sprintf("<%s@email-app>", uuid())

	params := &resend.SendEmailRequest{
		From:    from,
		To:      to,
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

	emailCol := database.GetCollection("emails")
	var inReplyTo *string
	references := []string{}
	if req.InReplyTo != "" {
		inReplyTo = &req.InReplyTo
		references = append(references, req.InReplyTo)
	}
	threadID := resolveThreadID(c.Request.Context(), mailbox.ID, messageID, references, inReplyTo)

	email := models.Email{
		MailboxID:  MustObjectID(req.MailboxID),
		From:       mailbox.EmailAddress,
		To:         strings.Join(to, ", "),
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

const maxRecipients = 20

var messageIDRe = regexp.MustCompile(`^<[^<>\s]{1,250}>$`)

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
