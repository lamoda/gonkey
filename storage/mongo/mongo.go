package mongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
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

func (c *Client) InsertDocuments(database string, collection string, documents []map[string]interface{}) ([]map[string]interface{}, error) {
	cl := c.Client.Database(database).Collection(collection)
	insertResult, err := cl.InsertMany(context.Background(), sliceToInterfaces(documents))
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": bson.M{"$in": insertResult.InsertedIDs}}
	cursor, err := cl.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}

	var result = make([]map[string]interface{}, len(insertResult.InsertedIDs))
	if err = cursor.All(context.Background(), &result); err != nil {
		return nil, err
	}

	result = sortedByInserted(insertResult.InsertedIDs, result)
	return result, err
}

func sortedByInserted(ids []interface{}, documents []map[string]interface{}) []map[string]interface{} {
	idMap := make(map[interface{}]map[string]interface{})
	result := make([]map[string]interface{}, len(ids))

	for _, doc := range documents {
		idMap[doc["_id"]] = doc
	}

	for i, id := range ids {
		result[i] = idMap[id]
	}

	return result
}

func sliceToInterfaces(elements []map[string]interface{}) []interface{} {
	var result []interface{}
	for _, el := range elements {
		result = append(result, el)
	}
	return result
}
