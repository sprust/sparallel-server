package connections

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"sync"
)

type Connections struct {
	ctx     context.Context
	mutex   sync.Mutex
	clients map[string]*mongo.Client
}

func NewConnections(ctx context.Context) *Connections {
	return &Connections{ctx: ctx}
}

func (c *Connections) Collection(connName string, dbName string, collName string) (*mongo.Collection, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	client, exists := c.clients[connName]

	if !exists {
		// TODO
		host := "TODO"
		port := 1234
		user := "TODO"
		password := "TODO"

		uri := fmt.Sprintf("mongodb://%s:%s@%s:%d", user, password, host, port)

		clientOptions := options.Client().ApplyURI(uri)

		var err error

		client, err = mongo.Connect(context.Background(), clientOptions)

		if err != nil {
			return nil, err
		}

		c.clients[connName] = client
	}

	return client.Database(dbName).Collection(collName), nil
}
