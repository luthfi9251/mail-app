package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"email-app/database"
	"email-app/middleware"
	"email-app/models"
)

const (
	resendReceivingURL = "https://api.resend.com/emails/receiving"
	syncListPageSize   = 100
	syncMaxListPages   = 10 // newest 1000 received emails are scanned per sync
	syncMaxFetch       = 30 // bodies fetched per sync; the rest are picked up on the next refresh
)

// One sync per mailbox per window caps how hard a user can drive Resend API calls.
var syncLimiter = middleware.NewLimiter(1, 10*time.Second)

type resendAttachment struct {
	ID          string `json:"id"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Size        int    `json:"size"`
}

type resendReceived struct {
	ID          string             `json:"id"`
	From        string             `json:"from"`
	To          []string           `json:"to"`
	Cc          []string           `json:"cc"`
	Bcc         []string           `json:"bcc"`
	Subject     string             `json:"subject"`
	CreatedAt   string             `json:"created_at"`
	MessageID   string             `json:"message_id"`
	HTML        *string            `json:"html"`
	Text        *string            `json:"text"`
	Headers     map[string]any     `json:"headers"`
	Attachments []resendAttachment `json:"attachments"`
}

func resendGet(ctx context.Context, path string, out any) error {
	var lastErr error
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, resendReceivingURL+path, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+os.Getenv("RESEND_API_KEY"))
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		if res.StatusCode == http.StatusTooManyRequests {
			res.Body.Close()
			lastErr = fmt.Errorf("resend rate limited")
			select {
			case <-time.After(time.Duration(attempt+1) * 600 * time.Millisecond):
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		err = func() error {
			defer res.Body.Close()
			if res.StatusCode != http.StatusOK {
				return fmt.Errorf("resend returned %d", res.StatusCode)
			}
			return json.NewDecoder(res.Body).Decode(out)
		}()
		return err
	}
	return lastErr
}

func normalizeAddr(s string) string {
	if a, err := mail.ParseAddress(s); err == nil {
		return strings.ToLower(a.Address)
	}
	return strings.ToLower(strings.TrimSpace(s))
}

func addressedTo(address string, lists ...[]string) bool {
	for _, l := range lists {
		for _, a := range l {
			if normalizeAddr(a) == address {
				return true
			}
		}
	}
	return false
}

func parseResendTime(s string) int64 {
	for _, layout := range []string{time.RFC3339Nano, "2006-01-02 15:04:05.999999-07", "2006-01-02 15:04:05.999999Z07:00"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UnixMilli()
		}
	}
	return time.Now().UnixMilli()
}

func headerString(h map[string]any, key string) string {
	for k, v := range h {
		if strings.EqualFold(k, key) {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	return ""
}

// SyncMailbox pulls received emails for one mailbox from Resend into the DB.
func SyncMailbox(c *gin.Context) {
	mailboxID := c.Param("id")
	if !requireMailboxAccess(c, mailboxID) {
		return
	}
	if !syncLimiter.Allow(mailboxID) {
		c.JSON(http.StatusOK, gin.H{"imported": 0, "remaining": 0, "throttled": true})
		return
	}
	ctx := c.Request.Context()

	var mailbox models.Mailbox
	if err := database.GetCollection("mailboxes").FindOne(ctx, bson.M{"_id": MustObjectID(mailboxID)}).Decode(&mailbox); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "mailbox not found"})
		return
	}
	address := normalizeAddr(mailbox.EmailAddress)
	emailCol := database.GetCollection("emails")

	// 1. List received emails (newest first) and keep those addressed to this mailbox.
	var candidates []resendReceived
	after := ""
	for page := 0; page < syncMaxListPages; page++ {
		path := fmt.Sprintf("?limit=%d", syncListPageSize)
		if after != "" {
			path += "&after=" + after
		}
		var list struct {
			HasMore bool             `json:"has_more"`
			Data    []resendReceived `json:"data"`
		}
		if err := resendGet(ctx, path, &list); err != nil {
			log.Printf("sync: list received emails: %v", err)
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to fetch emails from Resend"})
			return
		}
		for _, e := range list.Data {
			if addressedTo(address, e.To, e.Cc, e.Bcc) {
				candidates = append(candidates, e)
			}
		}
		if !list.HasMore || len(list.Data) == 0 {
			break
		}
		after = list.Data[len(list.Data)-1].ID
	}

	// 2. Drop the ones already stored.
	ids := make([]string, len(candidates))
	for i, e := range candidates {
		ids[i] = e.ID
	}
	known := map[string]bool{}
	if len(ids) > 0 {
		cur, err := emailCol.Find(ctx, bson.M{"resendId": bson.M{"$in": ids}})
		if err == nil {
			var existing []models.Email
			if cur.All(ctx, &existing) == nil {
				for _, e := range existing {
					known[e.ResendID] = true
				}
			}
		}
	}
	var fresh []resendReceived
	for _, e := range candidates {
		if !known[e.ID] {
			fresh = append(fresh, e)
		}
	}

	// 3. Fetch the body of each new email and store it.
	imported, remaining := 0, 0
	for i, ref := range fresh {
		if i >= syncMaxFetch {
			remaining = len(fresh) - i
			break
		}
		var full resendReceived
		if err := resendGet(ctx, "/"+ref.ID, &full); err != nil {
			log.Printf("sync: retrieve %s: %v", ref.ID, err)
			remaining++
			continue
		}

		messageID := full.MessageID
		if messageID == "" {
			messageID = fmt.Sprintf("<%s@resend>", full.ID)
		}
		var inReplyTo *string
		if v := headerString(full.Headers, "in-reply-to"); v != "" {
			inReplyTo = &v
		}
		references := []string{}
		if v := headerString(full.Headers, "references"); v != "" {
			references = splitRefs(v)
		}
		threadID := resolveThreadID(ctx, mailbox.ID, messageID, references, inReplyTo)

		attachments := []models.Attachment{}
		for _, a := range full.Attachments {
			attachments = append(attachments, models.Attachment{ID: a.ID, Filename: a.Filename, Size: a.Size, ContentType: a.ContentType})
		}
		deref := func(s *string) string {
			if s == nil {
				return ""
			}
			return *s
		}

		_, err := emailCol.InsertOne(ctx, models.Email{
			MailboxID:   mailbox.ID,
			ResendID:    full.ID,
			From:        full.From,
			To:          mailbox.EmailAddress,
			Subject:     full.Subject,
			HTMLBody:    deref(full.HTML),
			TextBody:    deref(full.Text),
			MessageID:   messageID,
			InReplyTo:   inReplyTo,
			References:  references,
			ThreadID:    threadID,
			Direction:   "inbound",
			Attachments: attachments,
			CreatedAt:   parseResendTime(full.CreatedAt),
		})
		if err != nil {
			if !mongo.IsDuplicateKeyError(err) {
				log.Printf("sync: insert %s: %v", full.ID, err)
				remaining++
			}
			continue
		}
		imported++
	}

	c.JSON(http.StatusOK, gin.H{"imported": imported, "remaining": remaining})
}
