package main

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

func main()  {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://123.57.194.18:27017"))
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	collection := client.Database("my_db").Collection("my_collection")
	var res bson.M
	//可以
	err = collection.FindOne(context.TODO(), bson.D{{"uid",1000}}).Decode(&res)

	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(res)
}
