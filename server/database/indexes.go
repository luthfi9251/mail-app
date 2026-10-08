package database

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func CreateIndexes() {
	userMailboxCol := GetCollection("usermailboxes")
	_, err := userMailboxCol.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    bson.D{{Key: "userId", Value: 1}, {Key: "mailboxId", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		log.Printf("Warning: failed to create usermailbox index: %v", err)
	}

	emailCol := GetCollection("emails")
	_, err = emailCol.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{Keys: bson.D{{Key: "mailboxId", Value: 1}, {Key: "createdAt", Value: -1}}},
		{Keys: bson.D{{Key: "mailboxId", Value: 1}, {Key: "direction", Value: 1}, {Key: "createdAt", Value: -1}}},
		{Keys: bson.D{{Key: "messageId", Value: 1}}, Options: options.Index().SetUnique(true)},
	})
	if err != nil {
		log.Printf("Warning: failed to create email indexes: %v", err)
	}
}
