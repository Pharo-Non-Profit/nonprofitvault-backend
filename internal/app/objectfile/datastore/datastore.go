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

type ObjectFile struct {
	TenantID               primitive.ObjectID `bson:"tenant_id,omitempty" json:"tenant_id,omitempty"`
	TenantName             string             `bson:"tenant_name" json:"tenant_name"`
	ID                     primitive.ObjectID `bson:"_id" json:"id"`
	CreatedAt              time.Time          `bson:"created_at,omitempty" json:"created_at,omitempty"`
	CreatedByUserName      string             `bson:"created_by_user_name" json:"-"` // Hidden from public.
	CreatedByUserID        primitive.ObjectID `bson:"created_by_user_id" json:"-"`   // Hidden from public.
	ModifiedAt             time.Time          `bson:"modified_at,omitempty" json:"modified_at,omitempty"`
	ModifiedByUserName     string             `bson:"modified_by_user_name" json:"-"` // Hidden from public.
	ModifiedByUserID       primitive.ObjectID `bson:"modified_by_user_id" json:"-"`   // Hidden from public.
	Name                   string             `bson:"name" json:"name"`
	Description            string             `bson:"description" json:"description"`
	Filename               string             `bson:"filename" json:"filename"`
	ObjectKey              string             `bson:"object_key" json:"-"` // Hidden from public.
	ObjectURL              string             `bson:"object_url" json:"-"` // Hidden from public.
	Status                 int8               `bson:"status" json:"status"`
	ContentType            int8               `bson:"content_type" json:"content_type"`
	Classification         uint64             `bson:"classification" json:"classification"`
	SmartFolderID          primitive.ObjectID `bson:"smart_folder_id" json:"smart_folder_id"`
	SmartFolderName        string             `bson:"smart_folder_name" json:"smart_folder_name"`
	SmartFolderCategory    uint64             `bson:"smart_folder_category,omitempty" json:"smart_folder_category,omitempty"`
	SmartFolderSubCategory uint64             `bson:"smart_folder_sub_category,omitempty" json:"smart_folder_sub_category,omitempty"`
}

type ObjectFileListFilter struct {
	// Pagination related.
	Cursor    primitive.ObjectID
	PageSize  int64
	SortField string
	SortOrder int8 // 1=ascending | -1=descending

	// Filter related.
	TenantID        primitive.ObjectID
	SmartFolderID   primitive.ObjectID
	UserID          primitive.ObjectID
	UserRole        int8
	ExcludeArchived bool
}

type ObjectFileListResult struct {
	Results     []*ObjectFile      `json:"results"`
	NextCursor  primitive.ObjectID `json:"next_cursor"`
	HasNextPage bool               `json:"has_next_page"`
}

type ObjectFileAsSelectOption struct {
	Value primitive.ObjectID `bson:"_id" json:"value"` // Extract from the database `_id` field and output through API as `value`.
	Label string             `bson:"name" json:"label"`
}

// ObjectFileStorer Interface for objectfile.
type ObjectFileStorer interface {
	Create(ctx context.Context, m *ObjectFile) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*ObjectFile, error)
	UpdateByID(ctx context.Context, m *ObjectFile) error
	ListByFilter(ctx context.Context, m *ObjectFileListFilter) (*ObjectFileListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *ObjectFileListFilter) ([]*ObjectFileAsSelectOption, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
	// //TODO: Add more...
}

type ObjectFileStorerImpl struct {
	Logger     *slog.Logger
	DbClient   *mongo.Client
	Collection *mongo.Collection
}

func NewDatastore(appCfg *c.Conf, loggerp *slog.Logger, client *mongo.Client) ObjectFileStorer {
	// ctx := context.Background()
	uc := client.Database(appCfg.DB.Name).Collection("object_files")

	//TODO: Improve indexes later...
	_, err := uc.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
		{Keys: bson.D{{Key: "tenant_id", Value: 1}}},
		{Keys: bson.D{{Key: "category", Value: -1}}},
		{Keys: bson.D{{Key: "classification", Value: -1}}},
		{Keys: bson.D{{Key: "status", Value: 1}}},
		{Keys: bson.D{
			{"tenant_name", "text"},
			{"name", "text"},
			{"description", "text"},
			{"filename", "text"},
		}},
	})
	if err != nil {
		// It is important that we crash the app on startup to meet the
		// requirements of `google/wire` framework.
		log.Fatal(err)
	}
	s := &ObjectFileStorerImpl{
		Logger:     loggerp,
		DbClient:   client,
		Collection: uc,
	}
	return s
}
