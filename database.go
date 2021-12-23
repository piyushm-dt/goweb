package db

import (
	"context"
	"fmt"
	"log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectDB() *mongo.Collection {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	//on docker options.Client().ApplyURI("mongodb://host.docker.internal:27017")
	
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
	fmt.Println("Connected to MongoDB!")

	collection := client.Database("webdb").Collection("collectionA")
	return collection
}
