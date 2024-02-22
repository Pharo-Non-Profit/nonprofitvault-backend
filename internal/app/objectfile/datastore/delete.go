package datastore

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (impl ObjectFileStorerImpl) DeleteByID(ctx context.Context, id primitive.ObjectID) error {
	_, err := impl.Collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		log.Fatal("DeleteOne() ERROR:", err)
	}
	return nil
}

func (impl ObjectFileStorerImpl) DeleteBySmartFolderID(ctx context.Context, smartFolderID primitive.ObjectID) error {
	filter := bson.M{"smart_folder_id": smartFolderID}

	_, err := impl.Collection.DeleteMany(ctx, filter)
	if err != nil {
		return err
	}

	return nil
}
