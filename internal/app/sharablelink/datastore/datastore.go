package datastore

import (
	"context"
	"log"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	c "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/config"
)

const (
	StatusActive   = 1
	StatusArchived = 2

	CategoryUnspecified      = 1
	CategoryGovernmentCanada = 2
)

type SharableLink struct {
	ExpiryDate             time.Time          `bson:"expiry_date" json:"expiry_date"`
	ExpiresIn              uint64             `bson:"expires_in,omitempty" json:"expires_in,omitempty"`
	SmartFolderID          primitive.ObjectID `bson:"smart_folder_id" json:"smart_folder_id"`
	SmartFolderName        string             `bson:"smart_folder_name" json:"smart_folder_name"`
	SmartFolderCategory    uint64             `bson:"smart_folder_category,omitempty" json:"smart_folder_category,omitempty"`
	SmartFolderSubCategory uint64             `bson:"smart_folder_sub_category,omitempty" json:"smart_folder_sub_category,omitempty"`
	SmartFolderDescription string             `bson:"smart_folder_description" json:"smart_folder_description"`
	ID                     primitive.ObjectID `bson:"_id" json:"id"`
	Status                 int8               `bson:"status" json:"status"`
	PublicID               uint64             `bson:"public_id" json:"public_id"`
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

type SharableLinkListResult struct {
	Results     []*SharableLink    `json:"results"`
	NextCursor  primitive.ObjectID `json:"next_cursor"`
	HasNextPage bool               `json:"has_next_page"`
}

type SharableLinkAsSelectOption struct {
	Value primitive.ObjectID `bson:"_id" json:"value"` // Extract from the database `_id` field and output through API as `value`.
	Label string             `bson:"text" json:"label"`
}

// SharableLinkStorer Interface for user.
type SharableLinkStorer interface {
	Create(ctx context.Context, m *SharableLink) error
	CreateOrGetByID(ctx context.Context, hh *SharableLink) (*SharableLink, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*SharableLink, error)
	GetByPublicID(ctx context.Context, oldID uint64) (*SharableLink, error)
	GetByText(ctx context.Context, text string) (*SharableLink, error)
	GetLatestByTenantID(ctx context.Context, tenantID primitive.ObjectID) (*SharableLink, error)
	CheckIfExistsByEmail(ctx context.Context, email string) (bool, error)
	UpdateByID(ctx context.Context, m *SharableLink) error
	ListByFilter(ctx context.Context, f *SharableLinkPaginationListFilter) (*SharableLinkPaginationListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *SharableLinkPaginationListFilter) ([]*SharableLinkAsSelectOption, error)
	ListByTenantID(ctx context.Context, tid primitive.ObjectID) (*SharableLinkPaginationListResult, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
}

type SharableLinkStorerImpl struct {
	Logger     *slog.Logger
	DbClient   *mongo.Client
	Collection *mongo.Collection
}

func NewDatastore(appCfg *c.Conf, loggerp *slog.Logger, client *mongo.Client) SharableLinkStorer {
	// ctx := context.Background()
	uc := client.Database(appCfg.DB.Name).Collection("sharable_links")

	_, err := uc.Indexes().CreateMany(context.TODO(), []mongo.IndexModel{
		{Keys: bson.D{{Key: "tenant_id", Value: 1}}},
		{Keys: bson.D{{Key: "public_id", Value: -1}}},
		{Keys: bson.D{{Key: "created_at", Value: -1}}},
		{Keys: bson.D{{Key: "expiry_date", Value: -1}}},
		{Keys: bson.D{{Key: "status", Value: 1}}},
		// {Keys: bson.D{
		// 	{"name", "text"},
		// }},
	})
	if err != nil {
		// It is important that we crash the app on startup to meet the
		// requirements of `google/wire` framework.
		log.Fatal(err)
	}

	s := &SharableLinkStorerImpl{
		Logger:     loggerp,
		DbClient:   client,
		Collection: uc,
	}
	return s
}
