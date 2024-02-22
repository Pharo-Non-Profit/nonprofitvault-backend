package controller

import (
	"context"
	"fmt"
	"mime/multipart"
	"time"

	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"

	a_d "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/objectfile/datastore"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/config/constants"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/utils/httperror"
)

type ObjectFileCreateRequestIDO struct {
	Name           string // Optional
	Description    string // Optional
	FileName       string
	FileType       string
	File           multipart.File
	SmartFolderID  primitive.ObjectID
	Classification uint64
}

func validateCreateRequest(dirtyData *ObjectFileCreateRequestIDO) error {
	e := make(map[string]string)

	if dirtyData.FileName == "" {
		e["file"] = "missing value"
	}
	if dirtyData.SmartFolderID.IsZero() {
		e["smart_folder_id"] = "missing value"
	}
	if dirtyData.Classification == 0 {
		e["classification"] = "missing value"
	}
	if len(e) != 0 {
		return httperror.NewForBadRequest(&e)
	}
	return nil
}

func (c *ObjectFileControllerImpl) Create(ctx context.Context, req *ObjectFileCreateRequestIDO) (*a_d.ObjectFile, error) {
	// Extract from our session the following data.
	orgID := ctx.Value(constants.SessionUserTenantID).(primitive.ObjectID)
	orgName := ctx.Value(constants.SessionUserTenantName).(string)
	userID := ctx.Value(constants.SessionUserID).(primitive.ObjectID)
	userName := ctx.Value(constants.SessionUserName).(string)

	if err := validateCreateRequest(req); err != nil {
		c.Logger.Warn("failed validation",
			slog.Any("payload", req),
			slog.Any("error", err),
		)
		return nil, err
	}

	sf, err := c.SmartFolderStorer.GetByID(ctx, req.SmartFolderID)
	if err != nil {
		c.Logger.Error("failed getting smart folder", slog.Any("error", err))
		return nil, err
	}

	// Generate the key of our upload.
	objectKey := fmt.Sprintf("ten_%v/cat_%d/subcat_%d/class_%d/%v", orgID.Hex(), sf.Category, sf.SubCategory, req.Classification, req.FileName)

	// For debugging purposes only.
	c.Logger.Debug("pre-upload meta",
		slog.String("file_name", req.FileName),
		slog.String("file_type", req.FileType),
		slog.String("object_key", objectKey),
		slog.String("name", req.Name),
		slog.String("description", req.Description),
		slog.Any("smart_folder_id", sf.ID),
		slog.Any("classification", req.Classification),
	)

	go func(file multipart.File, objkey string) {
		c.Logger.Debug("beginning private object file upload...")
		if err := c.ObjectStorage.UploadContentFromMulipart(context.Background(), objkey, file); err != nil {
			c.Logger.Error("private object file upload error", slog.Any("error", err))
			// Do not return an error, simply continue this function as there might
			// be a case were the file was removed on the object bucket by ourselves
			// or some other reason.
		}
		c.Logger.Debug("Finished private object file upload")
	}(req.File, objectKey)

	// Create our meta record in the database.
	res := &a_d.ObjectFile{
		TenantID:               orgID,
		TenantName:             orgName,
		ID:                     primitive.NewObjectID(),
		CreatedAt:              time.Now(),
		CreatedByUserName:      userName,
		CreatedByUserID:        userID,
		ModifiedAt:             time.Now(),
		ModifiedByUserName:     userName,
		ModifiedByUserID:       userID,
		Name:                   req.Name,
		Description:            req.Description,
		Filename:               req.FileName,
		ObjectKey:              objectKey,
		ObjectURL:              "",
		Status:                 a_d.StatusActive,
		SmartFolderID:          sf.ID,
		SmartFolderName:        sf.Name,
		SmartFolderCategory:    sf.Category,
		SmartFolderSubCategory: sf.SubCategory,
		Classification:         req.Classification,
	}

	if err := c.ObjectFileStorer.Create(ctx, res); err != nil {
		c.Logger.Error("objectfile create error", slog.Any("error", err))
		return nil, err
	}
	return res, nil
}
