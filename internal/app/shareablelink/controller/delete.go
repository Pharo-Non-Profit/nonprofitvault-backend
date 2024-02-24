package controller

import (
	"context"

	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (impl *ShareableLinkControllerImpl) DeleteByID(ctx context.Context, sfid primitive.ObjectID) error {
	// STEP 1: Lookup the record or error.
	shareablelink, err := impl.GetByID(ctx, sfid)
	if err != nil {
		impl.Logger.Error("database get by id error", slog.Any("error", err))
		return err
	}
	if shareablelink == nil {
		impl.Logger.Error("database returns nothing from get by id")
		return err
	}

	// STEP 4: Delete from database.
	if err := impl.ShareableLinkStorer.DeleteByID(ctx, sfid); err != nil {
		impl.Logger.Error("database delete by id error", slog.Any("error", err))
		return err
	}
	return nil
}
