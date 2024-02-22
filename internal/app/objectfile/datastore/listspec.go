package datastore

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (impl ObjectFileStorerImpl) ListObjectKeysBySmartFolderID(ctx context.Context, sfid primitive.ObjectID) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	// Create the filter based on the cursor
	filter := bson.M{}
	filter["smart_folder_id"] = sfid

	// Define projection to include only the object_key field
	projection := bson.M{"object_key": 1}

	// Find documents matching the filter and projection
	cursor, err := impl.Collection.Find(ctx, filter, options.Find().SetProjection(projection))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Iterate over the cursor and extract the object keys
	var objectKeys []string
	for cursor.Next(ctx) {
		var obj ObjectFile
		if err := cursor.Decode(&obj); err != nil {
			return nil, err
		}
		objectKeys = append(objectKeys, obj.ObjectKey)
	}

	// Check for any errors during cursor iteration
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	// Return the list of object keys
	return objectKeys, nil
}
