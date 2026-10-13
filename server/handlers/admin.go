package handlers

import (
	"context"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"golang.org/x/crypto/bcrypt"

	"email-app/config"
	"email-app/database"
	"email-app/models"
)

var localPartRe = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9._+-]{0,62}[a-z0-9])?$`)

// parseID reads an ObjectID path/body value, answering 400 when it is malformed.
func parseID(c *gin.Context, hex string) (bson.ObjectID, bool) {
	id, err := bson.ObjectIDFromHex(hex)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return id, false
	}
	return id, true
}

func isSelf(c *gin.Context, id bson.ObjectID) bool {
	userID, _ := c.Get("userId")
	return userID == id.Hex()
}

// createOwnedMailbox inserts a mailbox, assigns it to the user as owner and
// gives every admin access to it.
func createOwnedMailbox(ctx context.Context, userID bson.ObjectID, address string) error {
	now := time.Now().UnixMilli()
	mb, err := database.GetCollection("mailboxes").InsertOne(ctx, models.Mailbox{
		EmailAddress: address,
		CreatedAt:    now,
	})
	if err != nil {
		return err
	}
	mailboxID := mb.InsertedID.(bson.ObjectID)
	_, err = database.GetCollection("usermailboxes").InsertOne(ctx, models.UserMailbox{
		UserID:    userID,
		MailboxID: mailboxID,
		Role:      "owner",
		CreatedAt: now,
	})
	if err != nil {
		return err
	}
	return assignAdminsToMailbox(ctx, mailboxID)
}

// grantMemberAccess gives every user access to every mailbox as a member,
// skipping pairs that already have an assignment (whatever its role).
func grantMemberAccess(ctx context.Context, userIDs, mailboxIDs []bson.ObjectID) error {
	if len(userIDs) == 0 || len(mailboxIDs) == 0 {
		return nil
	}
	col := database.GetCollection("usermailboxes")
	cursor, err := col.Find(ctx, bson.M{"userId": bson.M{"$in": userIDs}, "mailboxId": bson.M{"$in": mailboxIDs}})
	if err != nil {
		return err
	}
	var existing []models.UserMailbox
	if err := cursor.All(ctx, &existing); err != nil {
		return err
	}
	type pair struct{ user, mailbox bson.ObjectID }
	have := make(map[pair]bool, len(existing))
	for _, a := range existing {
		have[pair{a.UserID, a.MailboxID}] = true
	}

	now := time.Now().UnixMilli()
	var docs []any
	for _, u := range userIDs {
		for _, m := range mailboxIDs {
			if !have[pair{u, m}] {
				docs = append(docs, models.UserMailbox{UserID: u, MailboxID: m, Role: "member", CreatedAt: now})
			}
		}
	}
	if len(docs) == 0 {
		return nil
	}
	_, err = col.InsertMany(ctx, docs)
	return err
}

// findIDs returns the _id of every document in the collection matching filter.
func findIDs(ctx context.Context, collection string, filter bson.M) ([]bson.ObjectID, error) {
	cursor, err := database.GetCollection(collection).Find(ctx, filter, options.Find().SetProjection(bson.M{"_id": 1}))
	if err != nil {
		return nil, err
	}
	var docs []struct {
		ID bson.ObjectID `bson:"_id"`
	}
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	ids := make([]bson.ObjectID, len(docs))
	for i, d := range docs {
		ids[i] = d.ID
	}
	return ids, nil
}

// assignAllMailboxesToUser gives an admin access to every existing mailbox.
func assignAllMailboxesToUser(ctx context.Context, userID bson.ObjectID) error {
	mailboxIDs, err := findIDs(ctx, "mailboxes", bson.M{})
	if err != nil {
		return err
	}
	return grantMemberAccess(ctx, []bson.ObjectID{userID}, mailboxIDs)
}

// assignAdminsToMailbox gives every admin access to a new mailbox.
func assignAdminsToMailbox(ctx context.Context, mailboxID bson.ObjectID) error {
	adminIDs, err := findIDs(ctx, "users", bson.M{"role": models.RoleAdmin})
	if err != nil {
		return err
	}
	return grantMemberAccess(ctx, adminIDs, []bson.ObjectID{mailboxID})
}

func GetAdminConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"mailboxDomain": config.MailboxDomain})
}

func ListUsers(c *gin.Context) {
	col := database.GetCollection("users")
	cursor, err := col.Find(c.Request.Context(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch users"})
		return
	}
	defer cursor.Close(c.Request.Context())

	var users []models.User
	if err := cursor.All(c.Request.Context(), &users); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse users"})
		return
	}

	umCol := database.GetCollection("usermailboxes")
	result := []gin.H{}
	for _, u := range users {
		count, _ := umCol.CountDocuments(c.Request.Context(), bson.M{"userId": u.ID})
		result = append(result, gin.H{
			"id":           u.ID.Hex(),
			"email":        u.Email,
			"name":         u.Name,
			"role":         u.Role,
			"mailboxCount": count,
			"createdAt":    u.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, result)
}

func CreateUser(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required"`
		Name     string `json:"name" binding:"required"`
		Password string `json:"password" binding:"required"`
		Role     string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, email, password required"})
		return
	}

	if req.Role == "" {
		req.Role = models.RoleUser
	}
	if !models.ValidUserRole(req.Role) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
		return
	}
	if !validEmail(req.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid email address"})
		return
	}
	if err := validatePassword(req.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	col := database.GetCollection("users")
	result, err := col.InsertOne(c.Request.Context(), models.User{
		Email:     req.Email,
		Name:      req.Name,
		Password:  string(hash),
		Role:      req.Role,
		CreatedAt: time.Now().UnixMilli(),
	})
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
		return
	}

	userID := result.InsertedID.(bson.ObjectID)

	parts := strings.SplitN(req.Email, "@", 2)
	if len(parts) == 2 && parts[1] == config.MailboxDomain {
		createOwnedMailbox(c.Request.Context(), userID, req.Email)
	}
	if req.Role == models.RoleAdmin {
		if err := assignAllMailboxesToUser(c.Request.Context(), userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user created but failed to assign mailboxes"})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{"success": true})
}

func DeleteUser(c *gin.Context) {
	id, ok := parseID(c, c.Param("id"))
	if !ok {
		return
	}
	if isSelf(c, id) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "you cannot delete your own account"})
		return
	}

	col := database.GetCollection("users")
	_, err := col.DeleteOne(c.Request.Context(), bson.M{"_id": id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}
	_, _ = database.GetCollection("usermailboxes").DeleteMany(c.Request.Context(), bson.M{"userId": id})
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func SetUserRole(c *gin.Context) {
	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role required"})
		return
	}
	if !models.ValidUserRole(req.Role) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
		return
	}
	id, ok := parseID(c, c.Param("id"))
	if !ok {
		return
	}
	// Admins can't demote themselves, so there is always at least one admin left.
	if isSelf(c, id) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "you cannot change your own role"})
		return
	}

	col := database.GetCollection("users")
	res, err := col.UpdateByID(c.Request.Context(), id, bson.M{"$set": bson.M{"role": req.Role}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update role"})
		return
	}
	if res.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	// A new admin gets every mailbox; revoking admin keeps the assignments.
	if req.Role == models.RoleAdmin {
		if err := assignAllMailboxesToUser(c.Request.Context(), id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "role updated but failed to assign mailboxes"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func ListAllMailboxes(c *gin.Context) {
	col := database.GetCollection("mailboxes")
	cursor, err := col.Find(c.Request.Context(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch mailboxes"})
		return
	}
	defer cursor.Close(c.Request.Context())

	var mailboxes []models.Mailbox
	if err := cursor.All(c.Request.Context(), &mailboxes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse mailboxes"})
		return
	}

	umCol := database.GetCollection("usermailboxes")
	result := []gin.H{}
	for _, mb := range mailboxes {
		count, _ := umCol.CountDocuments(c.Request.Context(), bson.M{"mailboxId": mb.ID})
		result = append(result, gin.H{
			"id":           mb.ID.Hex(),
			"emailAddress": mb.EmailAddress,
			"userCount":    count,
			"createdAt":    mb.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, result)
}

func CreateMailbox(c *gin.Context) {
	var req struct {
		LocalPart string `json:"localPart" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "localPart required"})
		return
	}

	localPart := strings.ToLower(strings.TrimSpace(req.LocalPart))
	if !localPartRe.MatchString(localPart) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid local part: use letters, digits, . _ + -"})
		return
	}

	col := database.GetCollection("mailboxes")
	res, err := col.InsertOne(c.Request.Context(), models.Mailbox{
		EmailAddress: localPart + "@" + config.MailboxDomain,
		CreatedAt:    time.Now().UnixMilli(),
	})
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "mailbox already exists"})
		return
	}
	if err := assignAdminsToMailbox(c.Request.Context(), res.InsertedID.(bson.ObjectID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "mailbox created but failed to assign admins"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true})
}

func DeleteMailbox(c *gin.Context) {
	id, ok := parseID(c, c.Param("id"))
	if !ok {
		return
	}
	emailCol := database.GetCollection("emails")
	count, err := emailCol.CountDocuments(c.Request.Context(), bson.M{"mailboxId": id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete mailbox"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete mailbox with existing emails"})
		return
	}

	col := database.GetCollection("mailboxes")
	_, err = col.DeleteOne(c.Request.Context(), bson.M{"_id": id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete mailbox"})
		return
	}
	_, _ = database.GetCollection("usermailboxes").DeleteMany(c.Request.Context(), bson.M{"mailboxId": id})
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func CreateAssignment(c *gin.Context) {
	var req struct {
		UserID    string `json:"userId" binding:"required"`
		MailboxID string `json:"mailboxId" binding:"required"`
		Role      string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId and mailboxId required"})
		return
	}
	if req.Role == "" {
		req.Role = "member"
	}
	if req.Role != "member" && req.Role != "owner" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
		return
	}
	userID, ok := parseID(c, req.UserID)
	if !ok {
		return
	}
	mailboxID, ok := parseID(c, req.MailboxID)
	if !ok {
		return
	}
	ctx := c.Request.Context()
	if n, _ := database.GetCollection("users").CountDocuments(ctx, bson.M{"_id": userID}); n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	if n, _ := database.GetCollection("mailboxes").CountDocuments(ctx, bson.M{"_id": mailboxID}); n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "mailbox not found"})
		return
	}

	col := database.GetCollection("usermailboxes")
	// The unique (userId, mailboxId) index also enforces this; checking first
	// gives a clear error even if index creation failed at startup.
	if n, _ := col.CountDocuments(ctx, bson.M{"userId": userID, "mailboxId": mailboxID}); n > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "user already has access to this mailbox"})
		return
	}
	res, err := col.InsertOne(ctx, models.UserMailbox{
		UserID:    userID,
		MailboxID: mailboxID,
		Role:      req.Role,
		CreatedAt: time.Now().UnixMilli(),
	})
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "user already has access to this mailbox"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "id": res.InsertedID.(bson.ObjectID).Hex()})
}

func DeleteAssignment(c *gin.Context) {
	id, ok := parseID(c, c.Param("id"))
	if !ok {
		return
	}
	col := database.GetCollection("usermailboxes")
	_, err := col.DeleteOne(c.Request.Context(), bson.M{"_id": id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete assignment"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ListMailboxAssignments returns who can access a mailbox, with user details.
func ListMailboxAssignments(c *gin.Context) {
	mailboxID, ok := parseID(c, c.Param("id"))
	if !ok {
		return
	}
	ctx := c.Request.Context()
	cursor, err := database.GetCollection("usermailboxes").Find(ctx, bson.M{"mailboxId": mailboxID},
		options.Find().SetSort(bson.D{{Key: "createdAt", Value: 1}}))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch assignments"})
		return
	}
	var assignments []models.UserMailbox
	if err := cursor.All(ctx, &assignments); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse assignments"})
		return
	}

	userIDs := make([]bson.ObjectID, len(assignments))
	for i, a := range assignments {
		userIDs[i] = a.UserID
	}
	users := map[bson.ObjectID]models.User{}
	if len(userIDs) > 0 {
		uc, err := database.GetCollection("users").Find(ctx, bson.M{"_id": bson.M{"$in": userIDs}})
		if err == nil {
			var list []models.User
			if uc.All(ctx, &list) == nil {
				for _, u := range list {
					users[u.ID] = u
				}
			}
		}
	}

	result := []gin.H{}
	for _, a := range assignments {
		u, ok := users[a.UserID]
		if !ok {
			continue // user was deleted
		}
		result = append(result, gin.H{
			"id":     a.ID.Hex(),
			"userId": u.ID.Hex(),
			"name":   u.Name,
			"email":  u.Email,
			"role":   a.Role,
		})
	}
	c.JSON(http.StatusOK, result)
}
