package models

import "go.mongodb.org/mongo-driver/v2/bson"

type User struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Email     string        `bson:"email" json:"email"`
	Name      string        `bson:"name" json:"name"`
	Password  string        `bson:"password" json:"-"`
	Role      string        `bson:"role" json:"role"`
	CreatedAt int64         `bson:"createdAt" json:"createdAt"`
}
