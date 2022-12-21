package db

import (
	"context"
	"fmt"
	"log"
	"time"
	"user-service/mongo"
)

var (
	MongoDB mongo.Database
)

func InitMongoDatabase() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbHost := "mongo.host"
	dbPort := "mongo.port"
	dbUser := "mongo.user"
	dbPass := "mongo.pass"
	dbName := "mongo.name"
	mongodbURI := fmt.Sprintf("mongodb://%s:%s@%s:%s/%s", dbUser, dbPass, dbHost, dbPort, dbName)
	if dbUser == "" || dbPass == "" {
		mongodbURI = fmt.Sprintf("mongodb://%s:%s/%s", dbHost, dbPort, dbName)
	}

	client, err := mongo.NewClient(mongodbURI)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx)
	if err != nil {
		log.Fatal(err)
	}

	MongoDB = client.Database("")
}
