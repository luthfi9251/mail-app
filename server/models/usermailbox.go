package models

import "go.mongodb.org/mongo-driver/v2/bson"

type UserMailbox struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    bson.ObjectID `bson:"userId" json:"userId"`
	MailboxID bson.ObjectID `bson:"mailboxId" json:"mailboxId"`
	Role      string        `bson:"role" json:"role"`
	CreatedAt int64         `bson:"createdAt" json:"createdAt"`
}
