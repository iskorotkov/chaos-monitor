package mongo

import (
	"context"
	"errors"
	"fmt"
	"github.com/iskorotkov/chaos-monitor/storage"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"time"
)

var (
	ConnectionError     = errors.New("couldn't connect to MongoDB")
	InsertError         = errors.New("couldn't add element to collection")
	ReadElementError    = errors.New("couldn't read or deserialize collection element")
	ReadCollectionError = errors.New("couldn't read collection till the end")
)

type Mongo struct {
	client *mongo.Client
}

func (m Mongo) PutSnapshot(snapshot storage.Snapshot) error {
	collection := m.client.Database("chaos-framework").Collection("pods")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, snapshot)
	if err != nil {
		fmt.Println(err)
		return InsertError
	}

	return nil
}

func (m Mongo) GetSnapshots() ([]storage.Snapshot, error) {
	collection := m.client.Database("chaos-framework").Collection("pods")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cur, err := collection.Find(ctx, bson.D{})
	if err != nil {
		fmt.Println(err)
		return nil, ConnectionError
	}
	defer func() {
		_ = cur.Close(ctx)
	}()

	snapshots := make([]storage.Snapshot, cur.RemainingBatchLength())
	for cur.Next(ctx) {
		var snapshot storage.Snapshot

		err := cur.Decode(&snapshot)
		if err != nil {
			fmt.Println(err)
			return nil, ReadElementError
		}

		snapshots = append(snapshots, snapshot)
	}

	if err := cur.Err(); err != nil {
		fmt.Println(err)
		return nil, ReadCollectionError
	}

	return snapshots, nil
}

func Connect(host string, port int) (storage.Storage, error) {
	uri := fmt.Sprintf("mongodb://%s:%d", host, port)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println(err)
		return nil, ConnectionError
	}

	ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		fmt.Println(err)
		return nil, ConnectionError
	}

	return Mongo{
		client: client,
	}, nil
}
