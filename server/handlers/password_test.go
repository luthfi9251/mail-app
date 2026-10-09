package handlers

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"

	"email-app/config"
	"email-app/database"
	"email-app/models"
)

// Generated with better-auth 1.4.22 hashPassword("Rahasia-v1 pässwörd").
const betterAuthHash = "85cbd1833e0e47a6b5f21c5fc809e5b4:0558046c41597fd74d8012f1a48a2d1a6c9b85105ab4c1bea6ff1e819ea7af74f670b63a3004c36f6a81e7a96735ed5df3b86fc19891c6e12736f5ee4f8b7479"

func TestCheckPassword(t *testing.T) {
	if ok, legacy := checkPassword(betterAuthHash, "Rahasia-v1 pässwörd"); !ok || !legacy {
		t.Fatalf("legacy hash: want ok+legacy, got ok=%v legacy=%v", ok, legacy)
	}
	if ok, _ := checkPassword(betterAuthHash, "wrong"); ok {
		t.Fatal("legacy hash accepted wrong password")
	}

	b, _ := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.MinCost)
	if ok, legacy := checkPassword(string(b), "secret123"); !ok || legacy {
		t.Fatalf("bcrypt hash: want ok, not legacy, got ok=%v legacy=%v", ok, legacy)
	}
	if ok, _ := checkPassword(string(b), "wrong"); ok {
		t.Fatal("bcrypt hash accepted wrong password")
	}
}

// A user imported from v1 can sign in with their old password, and the hash
// is upgraded to bcrypt so the legacy path is only used once.
func TestLoginUpgradesLegacyHash(t *testing.T) {
	r := newTestRouter(t)
	config.JWTSecret = []byte("test-secret-test-secret-test-secret!")
	r.POST("/login", Login)

	users := database.DB.Collection("users")
	if _, err := users.InsertOne(context.Background(), models.User{
		Email: "v1@example.com", Name: "V1", Password: betterAuthHash, Role: models.RoleUser,
	}); err != nil {
		t.Fatal(err)
	}

	if w := postJSON(r, "/login", `{"email":"v1@example.com","password":"wrong"}`); w.Code != http.StatusUnauthorized {
		t.Fatalf("wrong password: want 401, got %d", w.Code)
	}
	if w := postJSON(r, "/login", `{"email":"v1@example.com","password":"Rahasia-v1 pässwörd"}`); w.Code != http.StatusOK {
		t.Fatalf("legacy login: want 200, got %d %s", w.Code, w.Body.String())
	}

	var u models.User
	if err := users.FindOne(context.Background(), bson.M{"email": "v1@example.com"}).Decode(&u); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(u.Password, "$2") {
		t.Fatalf("hash not upgraded to bcrypt: %.10s…", u.Password)
	}
	if w := postJSON(r, "/login", `{"email":"v1@example.com","password":"Rahasia-v1 pässwörd"}`); w.Code != http.StatusOK {
		t.Fatalf("login after upgrade: want 200, got %d", w.Code)
	}
}
