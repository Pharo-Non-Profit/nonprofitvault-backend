package controller

import (
	"context"
	"fmt"
	"time"

	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"

	shareablelink_s "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/shareablelink/datastore"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/utils/httperror"
)

func (c *ShareableLinkControllerImpl) GetByID(ctx context.Context, id primitive.ObjectID) (*shareablelink_s.ShareableLink, error) {
	// Retrieve from our database the record for the specific id.
	m, err := c.ShareableLinkStorer.GetByID(ctx, id)
	if err != nil {
		c.Logger.Error("database get by id error", slog.Any("error", err))
		return nil, err
	}
	return m, err
}

type PublicShareableLinkResponseIDO struct {
	ExpiryDate             time.Time          `bson:"expiry_date" json:"expiry_date"`
	ExpiresIn              uint64             `bson:"expires_in,omitempty" json:"expires_in,omitempty"`
	SmartFolderID          primitive.ObjectID `bson:"smart_folder_id" json:"smart_folder_id"`
	SmartFolderName        string             `bson:"smart_folder_name" json:"smart_folder_name"`
	SmartFolderCategory    uint64             `bson:"smart_folder_category,omitempty" json:"smart_folder_category,omitempty"`
	SmartFolderSubCategory uint64             `bson:"smart_folder_sub_category,omitempty" json:"smart_folder_sub_category,omitempty"`
	SmartFolderDescription string             `bson:"smart_folder_description" json:"smart_folder_description"`
	ID                     primitive.ObjectID `bson:"_id" json:"id"`
	CreatedAt              time.Time          `bson:"created_at" json:"created_at"`
	CreatedByUserID        primitive.ObjectID `bson:"created_by_user_id" json:"created_by_user_id,omitempty"`
	CreatedByUserName      string             `bson:"created_by_user_name" json:"created_by_user_name"`
	CreatedFromIPAddress   string             `bson:"created_from_ip_address" json:"created_from_ip_address"`
	ModifiedAt             time.Time          `bson:"modified_at" json:"modified_at"`
	ModifiedByUserID       primitive.ObjectID `bson:"modified_by_user_id" json:"modified_by_user_id,omitempty"`
	ModifiedByUserName     string             `bson:"modified_by_user_name" json:"modified_by_user_name"`
	ModifiedFromIPAddress  string             `bson:"modified_from_ip_address" json:"modified_from_ip_address"`
	TenantID               primitive.ObjectID `bson:"tenant_id" json:"tenant_id"`
	TenantName             string             `bson:"tenant_name" json:"tenant_name"`
}

func (c *ShareableLinkControllerImpl) PublicGetByID(ctx context.Context, id primitive.ObjectID) (*PublicShareableLinkResponseIDO, error) {
	// Retrieve from our database the record for the specific id.
	sl, err := c.ShareableLinkStorer.GetByID(ctx, id)
	if err != nil {
		c.Logger.Error("failed getting shareable link by id",
			slog.Any("error", err))
		return nil, err
	}
	if sl == nil {
		return nil, httperror.NewForBadRequestWithSingleField("id", fmt.Sprintf("shareable link does not exist for id: %s", id.Hex()))
	}
	res := &PublicShareableLinkResponseIDO{
		ExpiryDate:             sl.ExpiryDate,
		ExpiresIn:              sl.ExpiresIn,
		SmartFolderID:          sl.SmartFolderID,
		SmartFolderName:        sl.SmartFolderName,
		SmartFolderCategory:    sl.SmartFolderCategory,
		SmartFolderSubCategory: sl.SmartFolderSubCategory,
		SmartFolderDescription: sl.SmartFolderDescription,
		ID:                     sl.ID,
		CreatedAt:              sl.CreatedAt,
		ModifiedAt:             sl.ModifiedAt,
		TenantID:               sl.TenantID,
		TenantName:             sl.TenantName,
	}
	return res, nil
}
