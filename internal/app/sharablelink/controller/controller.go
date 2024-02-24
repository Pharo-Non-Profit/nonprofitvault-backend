package controller

import (
	"context"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	object_storage "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/adapter/storage/object"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/adapter/templatedemailer"
	objectfile_s "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/objectfile/datastore"
	sharablelink_s "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/sharablelink/datastore"
	smartfolder_s "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/smartfolder/datastore"
	user_s "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/user/datastore"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/config"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/provider/kmutex"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/provider/password"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/provider/uuid"
)

// SharableLinkController Interface for sharablelink business logic controller.
type SharableLinkController interface {
	Create(ctx context.Context, requestData *SharableLinkCreateRequestIDO) (*sharablelink_s.SharableLink, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*sharablelink_s.SharableLink, error)
	PublicGetByID(ctx context.Context, id primitive.ObjectID) (*PublicSharableLinkResponseIDO, error)
	// UpdateByID(ctx context.Context, requestData *SharableLinkUpdateRequestIDO) (*sharablelink_s.SharableLink, error)
	// ListByFilter(ctx context.Context, f *sharablelink_s.SharableLinkPaginationListFilter) (*sharablelink_s.SharableLinkPaginationListResult, error)
	// ListAsSelectOptionByFilter(ctx context.Context, f *sharablelink_s.SharableLinkPaginationListFilter) ([]*sharablelink_s.SharableLinkAsSelectOption, error)
	// PublicListAsSelectOptionByFilter(ctx context.Context, f *sharablelink_s.SharableLinkPaginationListFilter) ([]*sharablelink_s.SharableLinkAsSelectOption, error)
	// ArchiveByID(ctx context.Context, id primitive.ObjectID) (*sharablelink_s.SharableLink, error)
	// DeleteByID(ctx context.Context, id primitive.ObjectID) error
}

type SharableLinkControllerImpl struct {
	Config             *config.Conf
	Logger             *slog.Logger
	UUID               uuid.Provider
	ObjectStorage      object_storage.ObjectStorager
	Password           password.Provider
	Kmutex             kmutex.Provider
	DbClient           *mongo.Client
	UserStorer         user_s.UserStorer
	SharableLinkStorer sharablelink_s.SharableLinkStorer
	SmartFolderStorer  smartfolder_s.SmartFolderStorer
	ObjectFileStorer   objectfile_s.ObjectFileStorer
	TemplatedEmailer   templatedemailer.TemplatedEmailer
}

func NewController(
	appCfg *config.Conf,
	loggerp *slog.Logger,
	uuidp uuid.Provider,
	object object_storage.ObjectStorager,
	passwordp password.Provider,
	kmux kmutex.Provider,
	temailer templatedemailer.TemplatedEmailer,
	client *mongo.Client,
	usr_storer user_s.UserStorer,
	sharablelink_s sharablelink_s.SharableLinkStorer,
	smartfolder_s smartfolder_s.SmartFolderStorer,
	obj_storer objectfile_s.ObjectFileStorer,
) SharableLinkController {
	s := &SharableLinkControllerImpl{
		Config:             appCfg,
		Logger:             loggerp,
		UUID:               uuidp,
		ObjectStorage:      object,
		Password:           passwordp,
		Kmutex:             kmux,
		TemplatedEmailer:   temailer,
		DbClient:           client,
		UserStorer:         usr_storer,
		SharableLinkStorer: sharablelink_s,
		SmartFolderStorer:  smartfolder_s,
		ObjectFileStorer:   obj_storer,
	}
	s.Logger.Debug("sharablelink controller initialization started...")
	s.Logger.Debug("sharablelink controller initialized")
	return s
}
