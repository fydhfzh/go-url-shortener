package db

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UrlCollection struct {
	ctx        context.Context
	collection *mongo.Collection
}

type UrlDoc struct {
	ID        primitive.ObjectID `bson:"_id"`
	UrlCode   string             `bson:"urlCode"`
	LongUrl   string             `bson:"longUrl"`
	ShortUrl  string             `bson:"shortUrl"`
	CreatedAt time.Time          `bson:"createdAt"`
	ExpiresAt time.Time          `bson:"expiresAt"`
}

var collection *mongo.Collection
var ctx = context.TODO()

func InitDB() *UrlCollection {
	dbURI := os.Getenv("DB_URI")
	dbName := os.Getenv("DB_NAME")

	clientOptions := options.Client().ApplyURI(dbURI)
	client, err := mongo.Connect(ctx, clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)

	if err != nil {
		log.Fatal(err)
	}

	collection = client.Database(dbName).Collection("urls")
	log.Print("DB Connected")

	urlCollection := &UrlCollection{
		ctx:        ctx,
		collection: collection,
	}

	return urlCollection
}

func (u *UrlCollection) FindOne(urlCode string) (bson.M, error) {
	var result bson.M

	err := u.collection.FindOne(u.ctx, bson.D{{"urlCode", urlCode}}).Decode(&result)

	return result, err
}

func (u *UrlCollection) InsertOne(urlDoc *UrlDoc) error {
	_, err := u.collection.InsertOne(u.ctx, urlDoc)

	return err
}
