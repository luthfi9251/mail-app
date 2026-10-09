// Command import-v1 copies users, mailboxes and mailbox assignments from the
// v1 app database (Better Auth schema) into this app's database.
//
// It is idempotent: users are matched by email, mailboxes by address and
// assignments by (user, mailbox), so existing records are kept untouched and
// re-running only fills in what is missing. Emails are not imported.
//
//	cd server
//	go run ./cmd/import-v1                 # dry run: print the plan
//	go run ./cmd/import-v1 -apply          # write to the database
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"email-app/models"
)

type v1User struct {
	ID        bson.ObjectID `bson:"_id"`
	Email     string        `bson:"email"`
	Name      string        `bson:"name"`
	Role      string        `bson:"role"`
	Banned    bool          `bson:"banned"`
	CreatedAt time.Time     `bson:"createdAt"`
}

type v1Account struct {
	UserID     bson.ObjectID `bson:"userId"`
	ProviderID string        `bson:"providerId"`
	Password   string        `bson:"password"`
}

type v1Mailbox struct {
	ID           bson.ObjectID `bson:"_id"`
	EmailAddress string        `bson:"emailAddress"`
	CreatedAt    time.Time     `bson:"createdAt"`
}

type v1UserMailbox struct {
	UserID    any           `bson:"userId"` // stored as a string in v1, ObjectID in places
	MailboxID bson.ObjectID `bson:"mailboxId"`
	Role      string        `bson:"role"`
	CreatedAt time.Time     `bson:"createdAt"`
}

func main() {
	source := flag.String("source", "mailbox-email", "v1 database name")
	target := flag.String("target", "email-app", "target database name")
	skip := flag.String("skip", "tmp-sync-test@mailbox.app", "comma-separated v1 user emails to leave out")
	apply := flag.Bool("apply", false, "write changes (default is a dry run)")
	flag.Parse()

	godotenv.Load()
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("MONGODB_URI not set (run from the server/ directory)")
	}
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	defer client.Disconnect(ctx)

	src := client.Database(*source)
	dst := client.Database(*target)
	skipped := map[string]bool{}
	for _, e := range strings.Split(*skip, ",") {
		if e = strings.ToLower(strings.TrimSpace(e)); e != "" {
			skipped[e] = true
		}
	}

	mode := "DRY RUN (use -apply to write)"
	if *apply {
		mode = "APPLY"
	}
	fmt.Printf("Importing %s -> %s  [%s]\n", *source, *target, mode)

	var users []v1User
	findAll(ctx, src.Collection("user"), &users)
	var accounts []v1Account
	findAll(ctx, src.Collection("account"), &accounts)
	var mailboxes []v1Mailbox
	findAll(ctx, src.Collection("mailboxes"), &mailboxes)
	var assignments []v1UserMailbox
	findAll(ctx, src.Collection("usermailboxes"), &assignments)

	passwords := map[bson.ObjectID]string{}
	for _, a := range accounts {
		if a.ProviderID == "credential" && a.Password != "" {
			passwords[a.UserID] = a.Password
		}
	}

	// --- users ---
	fmt.Println("\nUsers:")
	userMap := map[string]bson.ObjectID{} // v1 user id hex -> target user id
	usersCol := dst.Collection("users")
	for _, u := range users {
		email := strings.TrimSpace(u.Email)
		switch {
		case skipped[strings.ToLower(email)]:
			fmt.Printf("  skip    %s (excluded)\n", email)
			continue
		case u.Banned:
			fmt.Printf("  skip    %s (banned in v1)\n", email)
			continue
		}

		var existing models.User
		if err := usersCol.FindOne(ctx, bson.M{"email": email}).Decode(&existing); err == nil {
			userMap[u.ID.Hex()] = existing.ID
			fmt.Printf("  keep    %s (already exists, left unchanged)\n", email)
			continue
		}

		pw, ok := passwords[u.ID]
		if !ok {
			fmt.Printf("  skip    %s (no password credential in v1)\n", email)
			continue
		}
		role := u.Role
		if !models.ValidUserRole(role) {
			role = models.RoleUser
		}
		userMap[u.ID.Hex()] = u.ID
		fmt.Printf("  create  %s (%s, role %s)\n", email, u.Name, role)
		if *apply {
			_, err := usersCol.InsertOne(ctx, models.User{
				ID:        u.ID,
				Email:     email,
				Name:      u.Name,
				Password:  pw, // Better Auth scrypt hash, upgraded to bcrypt on first login
				Role:      role,
				CreatedAt: millis(u.CreatedAt),
			})
			if err != nil {
				log.Fatalf("insert user %s: %v", email, err)
			}
		}
	}

	// --- mailboxes ---
	fmt.Println("\nMailboxes:")
	mailboxMap := map[bson.ObjectID]bson.ObjectID{} // v1 mailbox id -> target mailbox id
	mbCol := dst.Collection("mailboxes")
	domain := strings.ToLower(os.Getenv("MAILBOX_DOMAIN"))
	for _, m := range mailboxes {
		addr := strings.ToLower(strings.TrimSpace(m.EmailAddress))
		var existing models.Mailbox
		if err := mbCol.FindOne(ctx, bson.M{"emailAddress": addr}).Decode(&existing); err == nil {
			mailboxMap[m.ID] = existing.ID
			fmt.Printf("  keep    %s (already exists)\n", addr)
			continue
		}
		note := ""
		if domain != "" && !strings.HasSuffix(addr, "@"+domain) {
			note = "  [warning: not on MAILBOX_DOMAIN " + domain + "]"
		}
		mailboxMap[m.ID] = m.ID
		fmt.Printf("  create  %s%s\n", addr, note)
		if *apply {
			_, err := mbCol.InsertOne(ctx, models.Mailbox{ID: m.ID, EmailAddress: addr, CreatedAt: millis(m.CreatedAt)})
			if err != nil {
				log.Fatalf("insert mailbox %s: %v", addr, err)
			}
		}
	}

	// --- assignments ---
	fmt.Println("\nMailbox access:")
	umCol := dst.Collection("usermailboxes")
	emailByV1 := map[string]string{}
	for _, u := range users {
		emailByV1[u.ID.Hex()] = u.Email
	}
	addrByV1 := map[bson.ObjectID]string{}
	for _, m := range mailboxes {
		addrByV1[m.ID] = m.EmailAddress
	}
	for _, a := range assignments {
		v1UserID := idHex(a.UserID)
		userID, okU := userMap[v1UserID]
		mailboxID, okM := mailboxMap[a.MailboxID]
		label := fmt.Sprintf("%s -> %s", orID(emailByV1[v1UserID], v1UserID), orID(addrByV1[a.MailboxID], a.MailboxID.Hex()))
		if !okU || !okM {
			fmt.Printf("  skip    %s (user or mailbox not imported)\n", label)
			continue
		}
		// v1 had an "admin" mailbox role; the closest here is "owner".
		role := "member"
		if a.Role == "owner" || a.Role == "admin" {
			role = "owner"
		}
		n, err := umCol.CountDocuments(ctx, bson.M{"userId": userID, "mailboxId": mailboxID})
		if err != nil {
			log.Fatal(err)
		}
		if n > 0 {
			fmt.Printf("  keep    %s (already assigned)\n", label)
			continue
		}
		fmt.Printf("  create  %s (%s)\n", label, role)
		if *apply {
			_, err := umCol.InsertOne(ctx, models.UserMailbox{
				UserID:    userID,
				MailboxID: mailboxID,
				Role:      role,
				CreatedAt: millis(a.CreatedAt),
			})
			if err != nil {
				log.Fatalf("insert assignment %s: %v", label, err)
			}
		}
	}

	if !*apply {
		fmt.Println("\nDry run only, nothing was written. Re-run with -apply to import.")
	} else {
		fmt.Println("\nDone.")
	}
}

func findAll(ctx context.Context, col *mongo.Collection, out any) {
	cur, err := col.Find(ctx, bson.M{})
	if err != nil {
		log.Fatalf("read %s: %v", col.Name(), err)
	}
	if err := cur.All(ctx, out); err != nil {
		log.Fatalf("decode %s: %v", col.Name(), err)
	}
}

func idHex(v any) string {
	switch id := v.(type) {
	case bson.ObjectID:
		return id.Hex()
	case string:
		return id
	}
	return ""
}

func orID(name, id string) string {
	if name != "" {
		return name
	}
	return id
}

func millis(t time.Time) int64 {
	if t.IsZero() {
		return time.Now().UnixMilli()
	}
	return t.UnixMilli()
}
