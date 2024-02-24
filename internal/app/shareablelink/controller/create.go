package controller

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	shareablelink_s "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/shareablelink/datastore"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/config/constants"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/utils/httperror"
)

type ShareableLinkCreateRequestIDO struct {
	SmartFolderID primitive.ObjectID `bson:"smart_folder_id" json:"smart_folder_id"`
	ExpiresIn     uint64             `bson:"expires_in,omitempty" json:"expires_in,omitempty"`
}

func (impl *ShareableLinkControllerImpl) validateCreateRequest(ctx context.Context, dirtyData *ShareableLinkCreateRequestIDO) error {
	e := make(map[string]string)

	if dirtyData.SmartFolderID.IsZero() {
		e["smart_folder_id"] = "missing value"
	}
	if dirtyData.ExpiresIn == 0 {
		e["expires_in"] = "missing value"
	}

	if len(e) != 0 {
		return httperror.NewForBadRequest(&e)
	}
	return nil
}

func (impl *ShareableLinkControllerImpl) Create(ctx context.Context, req *ShareableLinkCreateRequestIDO) (*shareablelink_s.ShareableLink, error) {
	//
	// Get variables from our user authenticated session.
	//

	tid, _ := ctx.Value(constants.SessionUserTenantID).(primitive.ObjectID)
	// role, _ := ctx.Value(constants.SessionUserRole).(int8)
	userID, _ := ctx.Value(constants.SessionUserID).(primitive.ObjectID)
	userName, _ := ctx.Value(constants.SessionUserName).(string)
	ipAddress, _ := ctx.Value(constants.SessionIPAddress).(string)

	// DEVELOPERS NOTE:
	// Every submission needs to have a unique `public id` (PID)
	// generated. The following needs to happen to generate the unique PID:
	// 1. Make the `Create` function be `atomic` and thus lock this function.
	// 2. Count total records in system (for particular tenant).
	// 3. Generate PID.
	// 4. Apply the PID to the record.
	// 5. Unlock this `Create` function to be usable again by other calls after
	//    the function successfully submits the record into our system.
	impl.Kmutex.Lockf("create-smart-folder-by-tenant-%s", tid.Hex())
	defer impl.Kmutex.Unlockf("create-smart-folder-by-tenant-%s", tid.Hex())

	//
	// Perform our validation and return validation error on any issues detected.
	//

	if err := impl.validateCreateRequest(ctx, req); err != nil {
		impl.Logger.Error("validation error", slog.Any("error", err))
		return nil, err
	}

	// switch role {
	// case u_s.UserRoleExecutive, u_s.UserRoleManagement, u_s.UserRoleFrontlineStaff:
	// 	break
	// default:
	// 	impl.Logger.Error("you do not have permission to create a client")
	// 	return nil, httperror.NewForForbiddenWithSingleField("message", "you do not have permission to create a client")
	// }

	////
	//// Start the transaction.
	////

	session, err := impl.DbClient.StartSession()
	if err != nil {
		impl.Logger.Error("start session error",
			slog.Any("error", err))
		return nil, err
	}
	defer session.EndSession(ctx)

	// Define a transaction function with a series of operations
	transactionFunc := func(sessCtx mongo.SessionContext) (interface{}, error) {
		sf, err := impl.SmartFolderStorer.GetByID(sessCtx, req.SmartFolderID)
		if err != nil {
			impl.Logger.Error("failed getting smart folder by id",
				slog.Any("error", err))
			return nil, err
		}
		if sf == nil {
			impl.Logger.Warn("smart folder does not exist",
				slog.Any("smart_folder_id", req.SmartFolderID))
			return nil, httperror.NewForSingleField(http.StatusBadRequest, "smart_folder_id", "smart folder does not exist")
		}

		sl := &shareablelink_s.ShareableLink{}

		// Add defaults.
		sl.TenantID = tid
		sl.ID = primitive.NewObjectID()
		sl.CreatedAt = time.Now()
		sl.CreatedByUserID = userID
		sl.CreatedByUserName = userName
		sl.CreatedFromIPAddress = ipAddress
		sl.ModifiedAt = time.Now()
		sl.ModifiedByUserID = userID
		sl.ModifiedByUserName = userName
		sl.ModifiedFromIPAddress = ipAddress

		// Add base.
		sl.ExpiryDate = time.Now().Add(time.Duration(req.ExpiresIn) * time.Hour)
		sl.ExpiresIn = req.ExpiresIn
		sl.SmartFolderID = sf.ID
		sl.SmartFolderName = sf.Name
		sl.SmartFolderCategory = sf.Category
		sl.SmartFolderSubCategory = sf.SubCategory
		sl.SmartFolderDescription = sf.Description
		sl.Status = shareablelink_s.StatusActive

		// Save to our database.
		if err := impl.ShareableLinkStorer.Create(sessCtx, sl); err != nil {
			impl.Logger.Error("failed creating shareable link", slog.Any("error", err))
			return nil, err
		}

		////
		//// Exit our transaction successfully.
		////

		return sl, nil
	}

	// Start a transaction
	result, err := session.WithTransaction(ctx, transactionFunc)
	if err != nil {
		impl.Logger.Error("session failed",
			slog.Any("error", err))
		return nil, err
	}

	return result.(*shareablelink_s.ShareableLink), nil
}
