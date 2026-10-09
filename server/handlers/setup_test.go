package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"email-app/config"
	"email-app/database"
)

func newTestRouter(t *testing.T) *gin.Engine {
	godotenv.Load("../.env")
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		t.Skip("MONGODB_URI not set")
	}

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatal(err)
	}
	database.DB = client.Database("email-app-test")
	config.SetupToken = "test-setup-token"
	config.MailboxDomain = "example.com"
	for _, name := range []string{"users", "mailboxes", "usermailboxes", "emails"} {
		if err := database.DB.Collection(name).Drop(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	database.CreateIndexes()
	t.Cleanup(func() {
		for _, name := range []string{"users", "mailboxes", "usermailboxes", "emails"} {
			_ = database.DB.Collection(name).Drop(context.Background())
		}
		_ = client.Disconnect(context.Background())
	})

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/setup-status", SetupStatus)
	r.POST("/setup", Setup)
	return r
}

func postJSON(r *gin.Engine, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestSetupFlow(t *testing.T) {
	r := newTestRouter(t)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/setup-status", nil))
	if got := w.Body.String(); got != `{"needsSetup":true,"setupEnabled":true}` {
		t.Fatalf("empty db: want needsSetup true, got %s", got)
	}

	w = postJSON(r, "/setup", `{"email":"admin@x.com","name":"Admin","password":"secret123","setupToken":"wrong"}`)
	if w.Code != http.StatusForbidden {
		t.Fatalf("wrong setup token: want 403, got %d %s", w.Code, w.Body.String())
	}

	w = postJSON(r, "/setup", `{"email":"admin@x.com","name":"Admin","password":"short","setupToken":"test-setup-token"}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("weak password: want 400, got %d %s", w.Code, w.Body.String())
	}

	w = postJSON(r, "/setup", `{"email":"admin@x.com","name":"Admin","password":"secret123","setupToken":"test-setup-token"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("create admin: want 201, got %d %s", w.Code, w.Body.String())
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/setup-status", nil))
	if got := w.Body.String(); got != `{"needsSetup":false,"setupEnabled":true}` {
		t.Fatalf("after setup: want needsSetup false, got %s", got)
	}

	w = postJSON(r, "/setup", `{"email":"other@x.com","name":"Other","password":"secret123","setupToken":"test-setup-token"}`)
	if w.Code != http.StatusConflict {
		t.Fatalf("second setup: want 409, got %d %s", w.Code, w.Body.String())
	}

	var user struct {
		Role string `bson:"role"`
	}
	err := database.DB.Collection("users").
		FindOne(context.Background(), bson.M{"email": "admin@x.com"}).Decode(&user)
	if err != nil {
		t.Fatal(err)
	}
	if user.Role != "admin" {
		t.Fatalf("created user role: want admin, got %s", user.Role)
	}

	// admin@x.com is not on the mailbox domain, so it falls back to admin@<domain>.
	want := "admin@" + config.MailboxDomain
	var mb struct {
		ID bson.ObjectID `bson:"_id"`
	}
	err = database.DB.Collection("mailboxes").
		FindOne(context.Background(), bson.M{"emailAddress": want}).Decode(&mb)
	if err != nil {
		t.Fatalf("mailbox %s not created: %v", want, err)
	}
	n, err := database.DB.Collection("usermailboxes").
		CountDocuments(context.Background(), bson.M{"mailboxId": mb.ID, "role": "owner"})
	if err != nil || n != 1 {
		t.Fatalf("owner assignment: want 1, got %d (%v)", n, err)
	}
}
