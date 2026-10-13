package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"email-app/database"
	"email-app/models"
)

func TestMailboxAssignmentsAddListRemove(t *testing.T) {
	r := newTestRouter(t)
	r.GET("/mailboxes/:id/assignments", ListMailboxAssignments)
	r.POST("/assignments", CreateAssignment)
	r.DELETE("/assignments/:id", DeleteAssignment)
	ctx := context.Background()

	u, _ := database.DB.Collection("users").InsertOne(ctx, models.User{Email: "a@example.com", Name: "Ann", Role: models.RoleUser})
	mb, _ := database.DB.Collection("mailboxes").InsertOne(ctx, models.Mailbox{EmailAddress: "team@example.com"})
	userID, mbID := u.InsertedID.(bson.ObjectID).Hex(), mb.InsertedID.(bson.ObjectID).Hex()

	list := func() []map[string]string {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/mailboxes/"+mbID+"/assignments", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("list: want 200, got %d %s", w.Code, w.Body.String())
		}
		var out []map[string]string
		json.Unmarshal(w.Body.Bytes(), &out)
		return out
	}

	if got := list(); len(got) != 0 {
		t.Fatalf("want no assignments, got %v", got)
	}
	w := postJSON(r, "/assignments", `{"userId":"`+userID+`","mailboxId":"`+mbID+`","role":"member"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("create: want 201, got %d %s", w.Code, w.Body.String())
	}
	if w := postJSON(r, "/assignments", `{"userId":"`+userID+`","mailboxId":"`+mbID+`"}`); w.Code != http.StatusConflict {
		t.Fatalf("duplicate: want 409, got %d", w.Code)
	}
	if w := postJSON(r, "/assignments", `{"userId":"`+userID+`","mailboxId":"`+mbID+`","role":"superuser"}`); w.Code != http.StatusBadRequest {
		t.Fatalf("bad role: want 400, got %d", w.Code)
	}

	got := list()
	if len(got) != 1 || got[0]["email"] != "a@example.com" || got[0]["name"] != "Ann" || got[0]["role"] != "member" {
		t.Fatalf("unexpected list: %v", got)
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/assignments/"+got[0]["id"], nil))
	if w.Code != http.StatusOK {
		t.Fatalf("delete: want 200, got %d", w.Code)
	}
	if got := list(); len(got) != 0 {
		t.Fatalf("want empty after delete, got %v", got)
	}
}

func TestAdminsGetEveryMailbox(t *testing.T) {
	r := newTestRouter(t)
	r.POST("/users", CreateUser)
	r.POST("/mailboxes", CreateMailbox)
	r.PATCH("/users/:id/role", SetUserRole)
	ctx := context.Background()
	users := database.DB.Collection("users")
	um := database.DB.Collection("usermailboxes")

	findUser := func(email string) bson.ObjectID {
		var u models.User
		if err := users.FindOne(ctx, bson.M{"email": email}).Decode(&u); err != nil {
			t.Fatalf("find %s: %v", email, err)
		}
		return u.ID
	}
	findMailbox := func(address string) bson.ObjectID {
		var mb models.Mailbox
		if err := database.DB.Collection("mailboxes").FindOne(ctx, bson.M{"emailAddress": address}).Decode(&mb); err != nil {
			t.Fatalf("find %s: %v", address, err)
		}
		return mb.ID
	}
	role := func(userID, mailboxID bson.ObjectID) string {
		var a models.UserMailbox
		if err := um.FindOne(ctx, bson.M{"userId": userID, "mailboxId": mailboxID}).Decode(&a); err != nil {
			return ""
		}
		return a.Role
	}

	// An existing mailbox, created before any admin exists.
	if w := postJSON(r, "/mailboxes", `{"localPart":"sales"}`); w.Code != http.StatusCreated {
		t.Fatalf("create mailbox: %d %s", w.Code, w.Body.String())
	}
	sales := findMailbox("sales@example.com")

	// A new admin gets every existing mailbox, and owns their own.
	if w := postJSON(r, "/users", `{"name":"Ann","email":"ann@example.com","password":"password123","role":"admin"}`); w.Code != http.StatusCreated {
		t.Fatalf("create admin: %d %s", w.Code, w.Body.String())
	}
	ann := findUser("ann@example.com")
	annBox := findMailbox("ann@example.com")
	if got := role(ann, sales); got != "member" {
		t.Fatalf("admin on existing mailbox: want member, got %q", got)
	}
	if got := role(ann, annBox); got != "owner" {
		t.Fatalf("admin on own mailbox: want owner, got %q", got)
	}

	// A regular user only gets their own mailbox; admins get it too.
	if w := postJSON(r, "/users", `{"name":"Bob","email":"bob@example.com","password":"password123"}`); w.Code != http.StatusCreated {
		t.Fatalf("create user: %d %s", w.Code, w.Body.String())
	}
	bob := findUser("bob@example.com")
	bobBox := findMailbox("bob@example.com")
	if got := role(bob, sales); got != "" {
		t.Fatalf("regular user on sales: want no access, got %q", got)
	}
	if got := role(ann, bobBox); got != "member" {
		t.Fatalf("admin on new user's mailbox: want member, got %q", got)
	}

	// A new mailbox goes to every admin, not to regular users.
	if w := postJSON(r, "/mailboxes", `{"localPart":"support"}`); w.Code != http.StatusCreated {
		t.Fatalf("create mailbox: %d %s", w.Code, w.Body.String())
	}
	support := findMailbox("support@example.com")
	if got := role(ann, support); got != "member" {
		t.Fatalf("admin on new mailbox: want member, got %q", got)
	}
	if got := role(bob, support); got != "" {
		t.Fatalf("regular user on new mailbox: want no access, got %q", got)
	}

	// Promoting a user to admin grants the rest, keeping their owner role.
	req := httptest.NewRequest(http.MethodPatch, "/users/"+bob.Hex()+"/role", bytes.NewBufferString(`{"role":"admin"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("promote: %d %s", w.Code, w.Body.String())
	}
	for _, mb := range []bson.ObjectID{sales, support, annBox} {
		if got := role(bob, mb); got != "member" {
			t.Fatalf("promoted admin: want member, got %q", got)
		}
	}
	if got := role(bob, bobBox); got != "owner" {
		t.Fatalf("promoted admin on own mailbox: want owner, got %q", got)
	}
}
