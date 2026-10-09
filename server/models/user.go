package models

import "go.mongodb.org/mongo-driver/v2/bson"

const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

func ValidUserRole(r string) bool { return r == RoleAdmin || r == RoleUser }

type User struct {
	ID       bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Email    string        `bson:"email" json:"email"`
	Name     string        `bson:"name" json:"name"`
	Password string        `bson:"password" json:"-"`
	Role     string        `bson:"role" json:"role"`
	// TokenVersion is embedded in session tokens; bumping it revokes them all.
	TokenVersion int   `bson:"tokenVersion" json:"-"`
	CreatedAt    int64 `bson:"createdAt" json:"createdAt"`
}
