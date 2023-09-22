package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
)

type Client struct {
	*mongo.Client
}

func New(client *mongo.Client) *Client {
	return &Client{
		Client: client,
	}
}

func (c *Client) Truncate(database string, collection string) error {
	cl := c.Client.Database(database).Collection(collection)
	return cl.Drop(context.Background())
}

func (c *Client) InsertDocument(database string, collection string, document map[string]interface{}) error {
	cl := c.Client.Database(database).Collection(collection)
	_, err := cl.InsertOne(context.Background(), document)
	return err
}
