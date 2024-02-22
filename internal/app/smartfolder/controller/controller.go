package controller

import (
	"context"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	object_storage "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/adapter/storage/object"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/adapter/templatedemailer"
	objectfile_s "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/objectfile/datastore"
	smartfolder_s "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/smartfolder/datastore"
	user_s "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/user/datastore"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/config"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/provider/kmutex"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/provider/password"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/provider/uuid"
)

// SmartFolderController Interface for smartfolder business logic controller.
type SmartFolderController interface {
	Create(ctx context.Context, requestData *SmartFolderCreateRequestIDO) (*smartfolder_s.SmartFolder, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*smartfolder_s.SmartFolder, error)
	UpdateByID(ctx context.Context, requestData *SmartFolderUpdateRequestIDO) (*smartfolder_s.SmartFolder, error)
	ListByFilter(ctx context.Context, f *smartfolder_s.SmartFolderPaginationListFilter) (*smartfolder_s.SmartFolderPaginationListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *smartfolder_s.SmartFolderPaginationListFilter) ([]*smartfolder_s.SmartFolderAsSelectOption, error)
	PublicListAsSelectOptionByFilter(ctx context.Context, f *smartfolder_s.SmartFolderPaginationListFilter) ([]*smartfolder_s.SmartFolderAsSelectOption, error)
	ArchiveByID(ctx context.Context, id primitive.ObjectID) (*smartfolder_s.SmartFolder, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
}

type SmartFolderControllerImpl struct {
	Config            *config.Conf
	Logger            *slog.Logger
	UUID              uuid.Provider
	ObjectStorage     object_storage.ObjectStorager
	Password          password.Provider
	Kmutex            kmutex.Provider
	DbClient          *mongo.Client
	UserStorer        user_s.UserStorer
	SmartFolderStorer smartfolder_s.SmartFolderStorer
	ObjectFileStorer  objectfile_s.ObjectFileStorer
	TemplatedEmailer  templatedemailer.TemplatedEmailer
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
	smartfolder_s smartfolder_s.SmartFolderStorer,
	obj_storer objectfile_s.ObjectFileStorer,
) SmartFolderController {
	s := &SmartFolderControllerImpl{
		Config:            appCfg,
		Logger:            loggerp,
		UUID:              uuidp,
		ObjectStorage:     object,
		Password:          passwordp,
		Kmutex:            kmux,
		TemplatedEmailer:  temailer,
		DbClient:          client,
		UserStorer:        usr_storer,
		SmartFolderStorer: smartfolder_s,
		ObjectFileStorer:  obj_storer,
	}
	s.Logger.Debug("smartfolder controller initialization started...")
	s.Logger.Debug("smartfolder controller initialized")
	return s
}
