package main

import (
	"blogoconsumer/feed"
	"blogoconsumer/models"
	"context"
	"encoding/json"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/streadway/amqp"
)

var channelAmqp *amqp.Channel
var mongoClient *mongo.Client
var ctx context.Context

func init() {
	// database connection
	ctx = context.Background()
	mongoClient, _ = mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
}

func main() {
	// message queue connection
	amqpConnection, err := amqp.Dial(os.Getenv(
		"RABBITMQ_URI"))
	if err != nil {
		log.Fatal(err)
	}
	defer amqpConnection.Close()
	channelAmqp, _ := amqpConnection.Channel()
	defer channelAmqp.Close()
	forever := make(chan bool)
	msgs, err := channelAmqp.Consume(
		os.Getenv("RABBITMQ_QUEUE"),
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			var request models.RssFeedRequest
			json.Unmarshal(d.Body, &request)
			log.Println("RSS URL:", request.Url)
			entries, _ := feed.GetFeedEntries(request.Url)
			collection := mongoClient.Database(os.Getenv(
				"MONGO_DATABASE")).Collection("news")
			var maxNumOfNews int
			if len(entries) > 5 {
				maxNumOfNews = 5
			} else {
				maxNumOfNews = len(entries)
			}

			for _, entry := range entries[0:maxNumOfNews] {
				collection.InsertOne(ctx, bson.M{
					"title":       entry.Title,
					"description": entry.Description,
					"url":         entry.Link,
				})
			}
		}
	}()
	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
