package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	godotenv.Load()
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("MONGODB_URI not set")
	}

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	db := client.Database("email-app")
	users := db.Collection("users")

	password, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)

	_, err = users.InsertOne(context.Background(), bson.M{
		"email":    "admin@email.app",
		"name":     "Admin",
		"password": string(password),
		"role":     "admin",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Admin user created: admin@email.app / admin123")
}
