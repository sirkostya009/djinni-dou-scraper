package main

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
)

var subscriptions *mongo.Collection

// Subscription is a map of subscription url to a list of subscribers
type Subscription struct {
	Url         string   `json:"url"`
	Subscribers []int64  `json:"subscribers"`
	Data        []string `json:"data"`
}

func initMongo() {
	url := os.Getenv("MONGO_URL")
	if url == "" {
		url = "mongodb://localhost:27017"
	}
	db, err := mongo.Connect(context.Background(), options.Client().ApplyURI(url))
	if err != nil {
		panic(err)
	}
	subscriptions = db.Database("job-scraper").Collection("subscriptions")
}

func findByUrl(url string) (Subscription, error) {
	sub := Subscription{}
	return sub, subscriptions.FindOne(context.Background(), bson.M{"url": url}).Decode(&sub)
}

func updateSubscription(sub Subscription) (*mongo.UpdateResult, error) {
	return subscriptions.UpdateOne(context.Background(), bson.M{"url": sub.Url}, bson.M{"$set": sub}, options.Update().SetUpsert(true))
}

func deleteSubscription(url string) (*mongo.DeleteResult, error) {
	return subscriptions.DeleteOne(context.Background(), bson.M{"url": url})
}

func listSubscriptions(id int64) []Subscription {
	var subs []Subscription
	cursor, err := subscriptions.Find(context.Background(), bson.M{"subscribers": id})
	if err != nil {
		return nil
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var sub Subscription
		err = cursor.Decode(&sub)
		if err != nil {
			return nil
		}
		subs = append(subs, sub)
	}
	return subs
}

func deleteSubscriptionsByChatId(id int64) error {
	_, err := subscriptions.UpdateMany(context.Background(),
		bson.M{},
		bson.M{
			"$pull": bson.M{
				"subscribers": bson.M{
					"$in": []int64{id},
				},
			},
		},
	)
	_, err = subscriptions.DeleteMany(context.Background(), bson.M{"subscribers": bson.M{"$size": 0}})
	return err
}
