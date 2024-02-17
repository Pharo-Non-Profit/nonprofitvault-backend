package controller

import (
	"context"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	mg "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/adapter/emailer/mailgun"
	object_storage "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/adapter/storage/object"
	domain "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/tenant/datastore"
	org_d "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/tenant/datastore"
	tenant_s "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/tenant/datastore"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/config"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/provider/kmutex"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/provider/uuid"
)

// TenantController Interface for Tenant business logic controller.
type TenantController interface {
	Create(ctx context.Context, m *domain.Tenant) (*domain.Tenant, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Tenant, error)
	UpdateByID(ctx context.Context, m *domain.Tenant) (*domain.Tenant, error)
	ListByFilter(ctx context.Context, f *domain.TenantListFilter) (*domain.TenantListResult, error)
	ListAsSelectOptionByFilter(ctx context.Context, f *domain.TenantListFilter) ([]*domain.TenantAsSelectOption, error)
	DeleteByID(ctx context.Context, id primitive.ObjectID) error
	CreateComment(ctx context.Context, customerID primitive.ObjectID, content string) (*org_d.Tenant, error)
}

type TenantControllerImpl struct {
	Config        *config.Conf
	Logger        *slog.Logger
	UUID          uuid.Provider
	Kmutex        kmutex.Provider
	ObjectStorage object_storage.ObjectStorager
	Emailer       mg.Emailer
	DbClient      *mongo.Client
	TenantStorer  tenant_s.TenantStorer
}

func NewController(
	appCfg *config.Conf,
	loggerp *slog.Logger,
	uuidp uuid.Provider,
	kmux kmutex.Provider,
	object object_storage.ObjectStorager,
	emailer mg.Emailer,
	client *mongo.Client,
	org_storer tenant_s.TenantStorer,
) TenantController {
	s := &TenantControllerImpl{
		Config:        appCfg,
		Logger:        loggerp,
		UUID:          uuidp,
		Kmutex:        kmux,
		ObjectStorage: object,
		Emailer:       emailer,
		DbClient:      client,
		TenantStorer:  org_storer,
	}
	s.Logger.Debug("Tenant controller initialization started...")
	s.Logger.Debug("Tenant controller initialized")
	return s
}
