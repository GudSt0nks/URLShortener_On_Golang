package mon

import (
	"context"
	"fmt"
	"urlShortener/iternal/storage"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Storage struct {
	DB     *mongo.Database
	Client *mongo.Client
}

type Document struct {
	Url   string `json:"url"`
	Alias string `json:"alias"`
}

func New(storagePath string) (*Storage, error) {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(storagePath).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}

	var result bson.M
	err = client.Database("admin").RunCommand(context.TODO(), bson.D{{Key: "ping", Value: 1}}).Decode(&result)
	if err != nil {
		return nil, err
	}

	dataBase := client.Database("urls")

	return &Storage{DB: dataBase, Client: client}, nil
}

func (s *Storage) SaveURL(UrlToSave string, alias string) error {
	op := "storage.mon.SaveURL"

	coll := s.DB.Collection("main")

	var result Document
	if err := coll.FindOne(context.TODO(), bson.D{{Key: "url", Value: UrlToSave}}).Decode(&result); err != mongo.ErrNoDocuments {
		return fmt.Errorf("%s: %w", op, storage.ErrURLExist)
	}

	_, err := coll.InsertOne(context.TODO(), Document{Url: UrlToSave, Alias: alias})

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	op := "storage.mon.GetURL"

	coll := s.DB.Collection("main")

	var result Document
	if coll.FindOne(context.TODO(), bson.D{{Key: "alias", Value: alias}}).Decode(&result) == mongo.ErrNoDocuments {
		return "", fmt.Errorf("%s: %w", op, storage.ErrURLNotFound)
	}

	return result.Url, nil
}

func (s *Storage) DeleteURL(alias string) error {
	op := "storage.mon.GetURL"

	coll := s.DB.Collection("main")

	_, err := coll.DeleteOne(context.TODO(), bson.D{{Key: "alias", Value: alias}})
	if err == mongo.ErrNoDocuments {
		return fmt.Errorf("%s: %w", op, storage.ErrURLNotFound)
	}

	return nil
}
