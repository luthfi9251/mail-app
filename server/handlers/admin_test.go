package handlers

import (
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
