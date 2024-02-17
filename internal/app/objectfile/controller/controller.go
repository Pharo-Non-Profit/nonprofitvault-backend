package controller

import (
	"context"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	mg "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/adapter/emailer/mailgun"
	object_storage "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/adapter/storage/object"
	domain "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/objectfile/datastore"
	objectfile_s "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/objectfile/datastore"
	user_s "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/user/datastore"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/config"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/provider/uuid"
)

// ObjectFileController Interface for objectfile business logic controller.
type ObjectFileController interface {
	Create(ctx context.Context, req *ObjectFileCreateRequestIDO) (*domain.ObjectFile, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*domain.ObjectFile, error)
	GetPresignedURLByID(ctx context.Context, id primitive.ObjectID) (*PresignedURLResponseIDO, error)
	GetContent(ctx context.Context, id primitive.ObjectID) ([]byte, string, string, error)
	UpdateByID(ctx context.Context, ns *ObjectFileUpdateRequestIDO) (*domain.ObjectFile, error)
	ListByFilter(ctx context.Context, f *domain.ObjectFileListFilter) (*domain.ObjectFileListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *domain.ObjectFileListFilter) ([]*domain.ObjectFileAsSelectOption, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
}

type ObjectFileControllerImpl struct {
	Config           *config.Conf
	Logger           *slog.Logger
	UUID             uuid.Provider
	ObjectStorage    object_storage.ObjectStorager
	Emailer          mg.Emailer
	DbClient         *mongo.Client
	ObjectFileStorer objectfile_s.ObjectFileStorer
	UserStorer       user_s.UserStorer
}

func NewController(
	appCfg *config.Conf,
	loggerp *slog.Logger,
	uuidp uuid.Provider,
	object object_storage.ObjectStorager,
	client *mongo.Client,
	emailer mg.Emailer,
	org_storer objectfile_s.ObjectFileStorer,
	usr_storer user_s.UserStorer,
) ObjectFileController {
	s := &ObjectFileControllerImpl{
		Config:           appCfg,
		Logger:           loggerp,
		UUID:             uuidp,
		ObjectStorage:    object,
		Emailer:          emailer,
		DbClient:         client,
		ObjectFileStorer: org_storer,
		UserStorer:       usr_storer,
	}
	s.Logger.Debug("objectfile controller initialization started...")
	s.Logger.Debug("objectfile controller initialized")
	return s
}
