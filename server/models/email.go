package models

import "go.mongodb.org/mongo-driver/v2/bson"

type Attachment struct {
	ID          string `bson:"id" json:"id"`
	Filename    string `bson:"filename" json:"filename"`
	Size        int    `bson:"size" json:"size"`
	ContentType string `bson:"content_type" json:"content_type"`
}

type Email struct {
	ID           bson.ObjectID `bson:"_id,omitempty" json:"id"`
	MailboxID    bson.ObjectID `bson:"mailboxId" json:"mailboxId"`
	From         string        `bson:"from" json:"from"`
	To           string        `bson:"to" json:"to"`
	Subject      string        `bson:"subject" json:"subject"`
	HTMLBody     string        `bson:"htmlBody" json:"htmlBody"`
	TextBody     string        `bson:"textBody" json:"textBody"`
	MessageID    string        `bson:"messageId" json:"messageId"`
	InReplyTo    *string       `bson:"inReplyTo" json:"inReplyTo"`
	References   []string      `bson:"references" json:"references"`
	ThreadID     string        `bson:"threadId" json:"threadId"`
	Direction    string        `bson:"direction" json:"direction"`
	Read         bool          `bson:"read" json:"read"`
	Starred      bool          `bson:"starred" json:"starred"`
	Attachments  []Attachment  `bson:"attachments" json:"attachments"`
	CreatedAt    int64         `bson:"createdAt" json:"createdAt"`
}
