package controller

import (
	"context"
	"fmt"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/config/constants"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/utils/httperror"
)

func (impl *ObjectFileControllerImpl) DeleteByID(ctx context.Context, id primitive.ObjectID) error {
	// Extract from our session the following data.
	tenantID, _ := ctx.Value(constants.SessionUserTenantID).(primitive.ObjectID)

	// Update the database.
	objectFile, err := impl.GetByID(ctx, id)
	if err != nil {
		impl.Logger.Error("database get by id error", slog.Any("error", err))
		return err
	}
	if objectFile == nil {
		impl.Logger.Error("database returns nothing from get by id")
		return httperror.NewForBadRequestWithSingleField("message", fmt.Sprintf("object file does not exist for id: %s", id.Hex()))
	}
	if tenantID != objectFile.TenantID {
		impl.Logger.Error("forbidden")
		return httperror.NewForForbiddenWithSingleField("message", "you do not belong to this tenant")
	}

	// Proceed to delete the physical files from AWS object.
	if err := impl.ObjectStorage.DeleteByKeys(ctx, []string{objectFile.ObjectKey}); err != nil {
		impl.Logger.Warn("object delete by keys error", slog.Any("error", err))
		// Do not return an error, simply continue this function as there might
		// be a case were the file was removed on the object bucket by ourselves
		// or some other reason.
	}
	impl.Logger.Debug("deleted from remote object storage", slog.String("object_file_id", id.Hex()))

	if err := impl.ObjectFileStorer.DeleteByID(ctx, objectFile.ID); err != nil {
		impl.Logger.Error("database delete by id error", slog.Any("error", err))
		return err
	}
	impl.Logger.Debug("deleted from database", slog.String("object_file_id", id.Hex()))

	return nil
}
