package controller

import (
	"context"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	object_storage "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/adapter/storage/object"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/adapter/templatedemailer"
	objectfile_s "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/objectfile/datastore"
	shareablelink_s "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/shareablelink/datastore"
	smartfolder_s "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/smartfolder/datastore"
	user_s "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/user/datastore"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/config"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/provider/kmutex"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/provider/password"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/provider/uuid"
)

// ShareableLinkController Interface for shareablelink business logic controller.
type ShareableLinkController interface {
	Create(ctx context.Context, requestData *ShareableLinkCreateRequestIDO) (*shareablelink_s.ShareableLink, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*shareablelink_s.ShareableLink, error)
	PublicGetByID(ctx context.Context, id primitive.ObjectID) (*PublicShareableLinkResponseIDO, error)
	// UpdateByID(ctx context.Context, requestData *ShareableLinkUpdateRequestIDO) (*shareablelink_s.ShareableLink, error)
	// ListByFilter(ctx context.Context, f *shareablelink_s.ShareableLinkPaginationListFilter) (*shareablelink_s.ShareableLinkPaginationListResult, error)
	// ListAsSelectOptionByFilter(ctx context.Context, f *shareablelink_s.ShareableLinkPaginationListFilter) ([]*shareablelink_s.ShareableLinkAsSelectOption, error)
	// PublicListAsSelectOptionByFilter(ctx context.Context, f *shareablelink_s.ShareableLinkPaginationListFilter) ([]*shareablelink_s.ShareableLinkAsSelectOption, error)
	// ArchiveByID(ctx context.Context, id primitive.ObjectID) (*shareablelink_s.ShareableLink, error)
	// DeleteByID(ctx context.Context, id primitive.ObjectID) error
}

type ShareableLinkControllerImpl struct {
	Config             *config.Conf
	Logger             *slog.Logger
	UUID               uuid.Provider
	ObjectStorage      object_storage.ObjectStorager
	Password           password.Provider
	Kmutex             kmutex.Provider
	DbClient           *mongo.Client
	UserStorer         user_s.UserStorer
	ShareableLinkStorer shareablelink_s.ShareableLinkStorer
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
	shareablelink_s shareablelink_s.ShareableLinkStorer,
	smartfolder_s smartfolder_s.SmartFolderStorer,
	obj_storer objectfile_s.ObjectFileStorer,
) ShareableLinkController {
	s := &ShareableLinkControllerImpl{
		Config:             appCfg,
		Logger:             loggerp,
		UUID:               uuidp,
		ObjectStorage:      object,
		Password:           passwordp,
		Kmutex:             kmux,
		TemplatedEmailer:   temailer,
		DbClient:           client,
		UserStorer:         usr_storer,
		ShareableLinkStorer: shareablelink_s,
		SmartFolderStorer:  smartfolder_s,
		ObjectFileStorer:   obj_storer,
	}
	s.Logger.Debug("shareablelink controller initialization started...")
	s.Logger.Debug("shareablelink controller initialized")
	return s
}
