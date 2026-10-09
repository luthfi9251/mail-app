package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"

	"email-app/database"
	"email-app/models"
)

func signSvix(secret []byte, id string, ts time.Time, body []byte) http.Header {
	tsStr := strconv.FormatInt(ts.Unix(), 10)
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(id + "." + tsStr + "."))
	mac.Write(body)
	h := http.Header{}
	h.Set("svix-id", id)
	h.Set("svix-timestamp", tsStr)
	h.Set("svix-signature", "v1,"+base64.StdEncoding.EncodeToString(mac.Sum(nil)))
	return h
}

func TestVerifySvixSignature(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	secret := "whsec_" + base64.StdEncoding.EncodeToString(key)
	body := []byte(`{"to":"a@example.com"}`)
	now := time.Now()

	if err := verifySvixSignature(secret, signSvix(key, "msg_1", now, body), body, now); err != nil {
		t.Fatalf("valid signature rejected: %v", err)
	}
	if err := verifySvixSignature(secret, signSvix(key, "msg_1", now, body), []byte(`{"to":"b@example.com"}`), now); err == nil {
		t.Fatal("tampered body accepted")
	}
	if err := verifySvixSignature(secret, signSvix([]byte("other-key"), "msg_1", now, body), body, now); err == nil {
		t.Fatal("wrong key accepted")
	}
	old := now.Add(-10 * time.Minute)
	if err := verifySvixSignature(secret, signSvix(key, "msg_1", old, body), body, now); err == nil {
		t.Fatal("stale timestamp accepted")
	}
	if err := verifySvixSignature(secret, http.Header{}, body, now); err == nil {
		t.Fatal("missing headers accepted")
	}
}

// A user must only ever see thread messages from mailboxes they can access,
// even when another mailbox has messages with the same threadId.
func TestGetThreadIsScopedToMailbox(t *testing.T) {
	newTestRouter(t)
	ctx := context.Background()

	userID := bson.NewObjectID()
	mine, theirs := bson.NewObjectID(), bson.NewObjectID()
	if _, err := database.DB.Collection("usermailboxes").InsertOne(ctx, models.UserMailbox{UserID: userID, MailboxID: mine, Role: "owner"}); err != nil {
		t.Fatal(err)
	}
	emails := database.DB.Collection("emails")
	res, err := emails.InsertOne(ctx, models.Email{MailboxID: mine, MessageID: "<a@x>", ThreadID: "shared", Subject: "Invoice"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := emails.InsertOne(ctx, models.Email{MailboxID: theirs, MessageID: "<b@x>", ThreadID: "shared", Subject: "Invoice"}); err != nil {
		t.Fatal(err)
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/emails/:id/thread", func(c *gin.Context) { c.Set("userId", userID.Hex()); c.Next() }, GetThread)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/emails/"+res.InsertedID.(bson.ObjectID).Hex()+"/thread", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d %s", w.Code, w.Body.String())
	}
	var body struct {
		Emails []struct {
			MailboxID string `json:"mailboxId"`
		} `json:"emails"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Emails) != 1 || body.Emails[0].MailboxID != mine.Hex() {
		t.Fatalf("thread leaked other mailbox messages: %s", w.Body.String())
	}
}
