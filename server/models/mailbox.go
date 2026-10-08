package models

import "go.mongodb.org/mongo-driver/v2/bson"

type Mailbox struct {
	ID             bson.ObjectID `bson:"_id,omitempty" json:"id"`
	EmailAddress   string        `bson:"emailAddress" json:"emailAddress"`
	ResendDomainID string        `bson:"resendDomainId" json:"resendDomainId"`
	CreatedAt      int64         `bson:"createdAt" json:"createdAt"`
}
