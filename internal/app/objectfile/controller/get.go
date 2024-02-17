package controller

import (
	"context"
	"io/ioutil"
	"log/slog"
	"mime"
	"path/filepath"
	"time"

	domain "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/objectfile/datastore"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (c *ObjectFileControllerImpl) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.ObjectFile, error) {
	// // Extract from our session the following data.
	// userObjectFileID := ctx.Value(constants.SessionUserObjectFileID).(primitive.ObjectID)
	// userRole := ctx.Value(constants.SessionUserRole).(int8)
	//
	// If user is not administrator nor belongs to the objectfile then error.
	// if userRole != user_d.UserRoleRoot && id != userObjectFileID {
	// 	c.Logger.Error("authenticated user is not staff role nor belongs to the objectfile error",
	// 		slog.Any("userRole", userRole),
	// 		slog.Any("userObjectFileID", userObjectFileID))
	// 	return nil, httperror.NewForForbiddenWithSingleField("message", "you do not belong to this objectfile")
	// }

	// Retrieve from our database the record for the specific id.
	m, err := c.ObjectFileStorer.GetByID(ctx, id)
	if err != nil {
		c.Logger.Error("database get by id error",
			slog.String("object_file_id", id.Hex()),
			slog.Any("error", err))
		return nil, err
	}

	// // Generate the URL.
	// fileURL, err := c.ObjectStorage.GetPresignedURL(ctx, m.ObjectKey, 5*time.Minute)
	// if err != nil {
	// 	c.Logger.Error("object failed get presigned url error", slog.Any("error", err))
	// 	return nil, err
	// }
	//
	// m.ObjectURL = fileURL
	return m, err
}

type PresignedURLResponseIDO struct {
	PresignedURL string `bson:"presigned_url" json:"presigned_url"`
}

func (c *ObjectFileControllerImpl) GetPresignedURLByID(ctx context.Context, id primitive.ObjectID) (*PresignedURLResponseIDO, error) {
	// // Extract from our session the following data.
	// userObjectFileID := ctx.Value(constants.SessionUserObjectFileID).(primitive.ObjectID)
	// userRole := ctx.Value(constants.SessionUserRole).(int8)
	//
	// If user is not administrator nor belongs to the objectfile then error.
	// if userRole != user_d.UserRoleRoot && id != userObjectFileID {
	// 	c.Logger.Error("authenticated user is not staff role nor belongs to the objectfile error",
	// 		slog.Any("userRole", userRole),
	// 		slog.Any("userObjectFileID", userObjectFileID))
	// 	return nil, httperror.NewForForbiddenWithSingleField("message", "you do not belong to this objectfile")
	// }

	// Retrieve from our database the record for the specific id.
	m, err := c.ObjectFileStorer.GetByID(ctx, id)
	if err != nil {
		c.Logger.Error("database get by id error", slog.Any("error", err))
		return nil, err
	}

	// Generate the URL.
	fileURL, err := c.ObjectStorage.GetDownloadablePresignedURL(ctx, m.ObjectKey, 15*time.Minute)
	if err != nil {
		c.Logger.Error("object failed get presigned url error",
			slog.String("object_file_id", id.Hex()),
			slog.Any("error", err))
		return nil, err
	}

	c.Logger.Debug("generated presigned url", slog.Any("presigned_url", fileURL))

	// Return the new URL.
	return &PresignedURLResponseIDO{PresignedURL: fileURL}, nil
}

func (c *ObjectFileControllerImpl) GetContent(ctx context.Context, id primitive.ObjectID) ([]byte, string, string, error) {
	// Retrieve from our database the record for the specific id.
	m, err := c.ObjectFileStorer.GetByID(ctx, id)
	if err != nil {
		c.Logger.Error("database get by id error", slog.Any("error", err))
		return nil, "", "", err
	}

	// Generate the URL.
	reader, err := c.ObjectStorage.GetBinaryData(ctx, m.ObjectKey)
	if err != nil {
		return nil, "", "", err
	}
	defer reader.Close()

	// Read the contents of the file.
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, "", "", err
	}

	// Determine the filename and content type from the object key.
	// This assumes the object key contains the filename with its extension.
	filename := filepath.Base(m.ObjectKey)
	contentType := mime.TypeByExtension(filepath.Ext(filename))
	if contentType == "" {
		// Default content type if not found.
		contentType = "application/octet-stream"
	}

	return content, filename, contentType, nil
}
