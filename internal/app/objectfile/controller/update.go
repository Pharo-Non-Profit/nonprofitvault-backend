package controller

import (
	"context"
	"fmt"
	"mime/multipart"
	"time"

	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"

	domain "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/objectfile/datastore"
	user_d "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/user/datastore"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/config/constants"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/utils/httperror"
)

type ObjectFileUpdateRequestIDO struct {
	ID             primitive.ObjectID
	Name           string // Optional.
	Description    string // Optional.
	OwnershipID    primitive.ObjectID
	OwnershipType  int8
	FileName       string
	FileType       string
	File           multipart.File
	Category       uint64
	Classification uint64
}

func ValidateUpdateRequest(dirtyData *ObjectFileUpdateRequestIDO) error {
	e := make(map[string]string)

	if dirtyData.ID.IsZero() {
		e["id"] = "missing value"
	}
	if dirtyData.Category == 0 {
		e["category"] = "missing value"
	}
	if dirtyData.Classification == 0 {
		e["classification"] = "missing value"
	}
	if len(e) != 0 {
		return httperror.NewForBadRequest(&e)
	}
	return nil
}

func (c *ObjectFileControllerImpl) UpdateByID(ctx context.Context, req *ObjectFileUpdateRequestIDO) (*domain.ObjectFile, error) {
	// Extract from our session the following data.
	orgID := ctx.Value(constants.SessionUserTenantID).(primitive.ObjectID)
	// orgName := ctx.Value(constants.SessionUserTenantName).(string)
	// userID := ctx.Value(constants.SessionUserID).(primitive.ObjectID)
	// userName := ctx.Value(constants.SessionUserName).(string)

	if err := ValidateUpdateRequest(req); err != nil {
		return nil, err
	}

	// Fetch the original objectfile.
	os, err := c.ObjectFileStorer.GetByID(ctx, req.ID)
	if err != nil {
		c.Logger.Error("database get by id error",
			slog.Any("error", err),
			slog.Any("object_file_id", req.ID))
		return nil, err
	}
	if os == nil {
		c.Logger.Error("objectfile does not exist error",
			slog.Any("objectfile_id", req.ID))
		return nil, httperror.NewForBadRequestWithSingleField("message", "objectfile does not exist")
	}

	// Extract from our session the following data.
	userID := ctx.Value(constants.SessionUserID).(primitive.ObjectID)
	userTenantID := ctx.Value(constants.SessionUserTenantID).(primitive.ObjectID)
	userRole := ctx.Value(constants.SessionUserRole).(int8)
	userName := ctx.Value(constants.SessionUserName).(string)

	// If user is not administrator nor belongs to the objectfile then error.
	if userRole != user_d.UserRoleExecutive {
		c.Logger.Error("authenticated user is not staff role nor belongs to the objectfile error",
			slog.Any("userRole", userRole),
			slog.Any("userTenantID", userTenantID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "you do not belong to this objectfile")
	}

	// Update the file if the user uploaded a new file.
	if req.File != nil {
		// Proceed to delete the physical files from AWS object.
		if err := c.ObjectStorage.DeleteByKeys(ctx, []string{os.ObjectKey}); err != nil {
			c.Logger.Warn("object delete by keys error", slog.Any("error", err))
			// Do not return an error, simply continue this function as there might
			// be a case were the file was removed on the object bucket by ourselves
			// or some other reason.
		}

		// Generate the key of our upload.
		objectKey := fmt.Sprintf("ten_%v/cat_%d/class_%d/%v", orgID.Hex(), req.Category, req.Classification, req.FileName)

		go func(file multipart.File, objkey string) {
			c.Logger.Debug("beginning private object image upload...")
			if err := c.ObjectStorage.UploadContentFromMulipart(context.Background(), objkey, file); err != nil {
				c.Logger.Error("private object upload error", slog.Any("error", err))
				// Do not return an error, simply continue this function as there might
				// be a case were the file was removed on the object bucket by ourselves
				// or some other reason.
			}
			c.Logger.Debug("Finished private object image upload")
		}(req.File, objectKey)

		// Update file.
		os.ObjectKey = objectKey
		os.Filename = req.FileName

		c.Logger.Debug("pre-upload meta",
			slog.String("file_name", req.FileName),
			slog.String("file_type", req.FileType),
			slog.String("object_key", objectKey),
			slog.String("name", req.Name),
			slog.String("description", req.Description),
			slog.Any("category", req.Category),
			slog.Any("classification", req.Classification),
		)
	}

	// Modify our original objectfile.
	os.ModifiedAt = time.Now()
	os.ModifiedByUserID = userID
	os.ModifiedByUserName = userName
	os.Name = req.Name
	os.Description = req.Description
	os.Category = req.Category
	os.Classification = req.Classification

	// Save to the database the modified objectfile.
	if err := c.ObjectFileStorer.UpdateByID(ctx, os); err != nil {
		c.Logger.Error("database update by id error", slog.Any("error", err))
		return nil, err
	}

	// go func(org *domain.ObjectFile) {
	// 	c.updateObjectFileNameForAllUsers(ctx, org)
	// }(os)
	// go func(org *domain.ObjectFile) {
	// 	c.updateObjectFileNameForAllComicSubmissions(ctx, org)
	// }(os)

	return os, nil
}

// func (c *ObjectFileControllerImpl) updateObjectFileNameForAllUsers(ctx context.Context, ns *domain.ObjectFile) error {
// 	c.Logger.Debug("Beginning to update objectfile name for all uses")
// 	f := &user_d.UserListFilter{ObjectFileID: ns.ID}
// 	uu, err := c.UserStorer.ListByFilter(ctx, f)
// 	if err != nil {
// 		c.Logger.Error("database update by id error", slog.Any("error", err))
// 		return err
// 	}
// 	for _, u := range uu.Results {
// 		u.ObjectFileName = ns.Name
// 		log.Println("--->", u)
// 		// if err := c.UserStorer.UpdateByID(ctx, u); err != nil {
// 		// 	c.Logger.Error("database update by id error", slog.Any("error", err))
// 		// 	return err
// 		// }
// 	}
// 	return nil
// }
//
// func (c *ObjectFileControllerImpl) updateObjectFileNameForAllComicSubmissions(ctx context.Context, ns *domain.ObjectFile) error {
// 	c.Logger.Debug("Beginning to update objectfile name for all submissions")
// 	f := &domain.ComicSubmissionListFilter{ObjectFileID: ns.ID}
// 	uu, err := c.ComicSubmissionStorer.ListByFilter(ctx, f)
// 	if err != nil {
// 		c.Logger.Error("database update by id error", slog.Any("error", err))
// 		return err
// 	}
// 	for _, u := range uu.Results {
// 		u.ObjectFileName = ns.Name
// 		log.Println("--->", u)
// 		// if err := c.ComicSubmissionStorer.UpdateByID(ctx, u); err != nil {
// 		// 	c.Logger.Error("database update by id error", slog.Any("error", err))
// 		// 	return err
// 		// }
// 	}
// 	return nil
// }
