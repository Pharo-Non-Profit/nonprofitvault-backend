package controller

import (
	"context"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	smartfolder_s "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/smartfolder/datastore"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/config/constants"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/utils/httperror"
)

type SmartFolderCreateRequestIDO struct {
	Name        string `bson:"name" json:"name"`
	Description string `bson:"description" json:"description"`
	Category    uint64 `bson:"category,omitempty" json:"category,omitempty"`
	SubCategory uint64 `bson:"sub_category,omitempty" json:"sub_category,omitempty"`
	SortNumber  int8   `bson:"sort_number" json:"sort_number"`
}

func (impl *SmartFolderControllerImpl) validateCreateRequest(ctx context.Context, dirtyData *SmartFolderCreateRequestIDO) error {
	e := make(map[string]string)

	if dirtyData.Name == "" {
		e["name"] = "missing value"
	}
	if dirtyData.Category == 0 {
		e["category"] = "missing value"
	}
	if dirtyData.SubCategory == 0 {
		e["sub_category"] = "missing value"
	}
	if dirtyData.SortNumber == 0 {
		e["sort_number"] = "missing value"
	}

	if len(e) != 0 {
		return httperror.NewForBadRequest(&e)
	}
	return nil
}

func (impl *SmartFolderControllerImpl) Create(ctx context.Context, requestData *SmartFolderCreateRequestIDO) (*smartfolder_s.SmartFolder, error) {
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

	if err := impl.validateCreateRequest(ctx, requestData); err != nil {
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

		hh := &smartfolder_s.SmartFolder{}

		// Add defaults.
		hh.TenantID = tid
		hh.ID = primitive.NewObjectID()
		hh.CreatedAt = time.Now()
		hh.CreatedByUserID = userID
		hh.CreatedByUserName = userName
		hh.CreatedFromIPAddress = ipAddress
		hh.ModifiedAt = time.Now()
		hh.ModifiedByUserID = userID
		hh.ModifiedByUserName = userName
		hh.ModifiedFromIPAddress = ipAddress

		// Add base.
		hh.Name = requestData.Name
		hh.Description = requestData.Description
		hh.Category = requestData.Category
		hh.SubCategory = requestData.SubCategory
		hh.SortNumber = requestData.SortNumber
		hh.Status = smartfolder_s.StatusActive

		// Save to our database.
		if err := impl.SmartFolderStorer.Create(sessCtx, hh); err != nil {
			impl.Logger.Error("database create error", slog.Any("error", err))
			return nil, err
		}

		////
		//// Exit our transaction successfully.
		////

		return hh, nil
	}

	// Start a transaction
	result, err := session.WithTransaction(ctx, transactionFunc)
	if err != nil {
		impl.Logger.Error("session failed error",
			slog.Any("error", err))
		return nil, err
	}

	return result.(*smartfolder_s.SmartFolder), nil
}
