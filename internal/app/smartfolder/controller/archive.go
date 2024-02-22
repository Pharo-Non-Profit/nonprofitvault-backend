package controller

import (
	"context"

	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"

	smartfolder_s "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/smartfolder/datastore"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/utils/httperror"
)

func (impl *SmartFolderControllerImpl) ArchiveByID(ctx context.Context, id primitive.ObjectID) (*smartfolder_s.SmartFolder, error) {
	// // Extract from our session the following data.
	// userID := ctx.Value(constants.SessionUserID).(primitive.ObjectID)

	// Lookup the smartfolder in our database, else return a `400 Bad Request` error.
	ou, err := impl.SmartFolderStorer.GetByID(ctx, id)
	if err != nil {
		impl.Logger.Error("database error", slog.Any("err", err))
		return nil, err
	}
	if ou == nil {
		impl.Logger.Warn("smartfolder does not exist validation error")
		return nil, httperror.NewForBadRequestWithSingleField("id", "does not exist")
	}

	ou.Status = smartfolder_s.StatusArchived

	if err := impl.SmartFolderStorer.UpdateByID(ctx, ou); err != nil {
		impl.Logger.Error("smartfolder update by id error", slog.Any("error", err))
		return nil, err
	}
	return ou, nil
}
