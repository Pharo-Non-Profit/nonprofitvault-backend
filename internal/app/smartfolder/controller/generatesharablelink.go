package controller

import (
	"context"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/config/constants"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/utils/httperror"
)

type GenerateSharableLinkRequestIDO struct {
	SmartFolderID primitive.ObjectID `bson:"smart_folder_id" json:"smart_folder_id"`
	ExpiresIn     uint64             `bson:"expires_in,omitempty" json:"expires_in,omitempty"`
}

type GenerateSharableLinkResponseIDO struct {
	URL string `bson:"url,omitempty" json:"url,omitempty"`
}

func (impl *SmartFolderControllerImpl) validatGenerateSharableLinkRequest(ctx context.Context, dirtyData *GenerateSharableLinkRequestIDO) error {
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

func (impl *SmartFolderControllerImpl) GenerateSharableLink(ctx context.Context, requestData *GenerateSharableLinkRequestIDO) (*GenerateSharableLinkResponseIDO, error) {
	//
	// Get variables from our user authenticated session.
	//

	tid, _ := ctx.Value(constants.SessionUserTenantID).(primitive.ObjectID)
	// // role, _ := ctx.Value(constants.SessionUserRole).(int8)
	// userID, _ := ctx.Value(constants.SessionUserID).(primitive.ObjectID)
	// userName, _ := ctx.Value(constants.SessionUserName).(string)
	// ipAddress, _ := ctx.Value(constants.SessionIPAddress).(string)

	// DEVELOPERS NOTE:
	// Every submission needs to have a unique `public id` (PID)
	// generated. The following needs to happen to generate the unique PID:
	// 1. Make the `Create` function be `atomic` and thus lock this function.
	// 2. Count total records in system (for particular tenant).
	// 3. Generate PID.
	// 4. Apply the PID to the record.
	// 5. Unlock this `Create` function to be usable again by other calls after
	//    the function successfully submits the record into our system.
	impl.Kmutex.Lockf("generate-smart-folder-sharable-link-by-tenant-%s", tid.Hex())
	defer impl.Kmutex.Unlockf("generate-smart-folder-sharable-link-by-tenant-%s", tid.Hex())

	//
	// Perform our validation and return validation error on any issues detected.
	//

	if err := impl.validatGenerateSharableLinkRequest(ctx, requestData); err != nil {
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

		return "xxxyyyzzz", nil
	}

	// Start a transaction
	result, err := session.WithTransaction(ctx, transactionFunc)
	if err != nil {
		impl.Logger.Error("session failed error",
			slog.Any("error", err))
		return nil, err
	}

	res := &GenerateSharableLinkResponseIDO{
		URL: result.(string),
	}

	return res, nil
}
