package controller

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"log/slog"

	smartfolder_s "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/smartfolder/datastore"
)

func (c *SmartFolderControllerImpl) GetByID(ctx context.Context, id primitive.ObjectID) (*smartfolder_s.SmartFolder, error) {
	// Retrieve from our database the record for the specific id.
	m, err := c.SmartFolderStorer.GetByID(ctx, id)
	if err != nil {
		c.Logger.Error("database get by id error", slog.Any("error", err))
		return nil, err
	}
	return m, err
}
