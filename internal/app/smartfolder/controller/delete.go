package controller

import (
	"context"

	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (impl *SmartFolderControllerImpl) DeleteByID(ctx context.Context, sfid primitive.ObjectID) error {
	// STEP 1: Lookup the record or error.
	smartfolder, err := impl.GetByID(ctx, sfid)
	if err != nil {
		impl.Logger.Error("database get by id error", slog.Any("error", err))
		return err
	}
	if smartfolder == nil {
		impl.Logger.Error("database returns nothing from get by id")
		return err
	}

	// STEP 2: Get all the files that were uploaded to our object store and del.
	keys, err := impl.ObjectFileStorer.ListObjectKeysBySmartFolderID(ctx, sfid)
	if err != nil {
		impl.Logger.Error("failed getting object keys by smart folder id", slog.Any("error", err))
		return err
	}
	if err := impl.ObjectStorage.DeleteByKeys(ctx, keys); err != nil {
		impl.Logger.Error("failed deleting object from object store", slog.Any("error", err))
		return err
	}

	// Step 3: Delete all the object files related.
	if err := impl.ObjectFileStorer.DeleteBySmartFolderID(ctx, sfid); err != nil {
		impl.Logger.Error("failed deleting related object files", slog.Any("error", err))
		return err
	}

	// STEP 4: Delete from database.
	if err := impl.SmartFolderStorer.DeleteByID(ctx, sfid); err != nil {
		impl.Logger.Error("database delete by id error", slog.Any("error", err))
		return err
	}
	return nil
}
