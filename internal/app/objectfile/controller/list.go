package controller

import (
	"context"
	"time"

	"log/slog"

	domain "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/objectfile/datastore"
	user_d "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/user/datastore"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/config/constants"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/utils/httperror"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (c *ObjectFileControllerImpl) ListByFilter(ctx context.Context, f *domain.ObjectFileListFilter) (*domain.ObjectFileListResult, error) {
	// Extract from our session the following data.
	orgID := ctx.Value(constants.SessionUserTenantID).(primitive.ObjectID)
	userID := ctx.Value(constants.SessionUserID).(primitive.ObjectID)
	userRole := ctx.Value(constants.SessionUserRole).(int8)

	// Apply protection based on ownership and role.
	if userRole != user_d.UserRoleExecutive {
		f.TenantID = orgID // Force tenant tenancy restrictions.
	}

	c.Logger.Debug("fetching objectfiles now...", slog.Any("userID", userID))

	aa, err := c.ObjectFileStorer.ListByFilter(ctx, f)
	if err != nil {
		c.Logger.Error("database list by filter error", slog.Any("error", err))
		return nil, err
	}
	c.Logger.Debug("fetched objectfiles", slog.Any("aa", aa))

	for _, a := range aa.Results {
		// Generate the URL.
		fileURL, err := c.ObjectStorage.GetPresignedURL(ctx, a.ObjectKey, 5*time.Minute)
		if err != nil {
			c.Logger.Error("object failed get presigned url error", slog.Any("error", err))
			return nil, err
		}
		a.ObjectURL = fileURL
	}
	return aa, err
}

func (c *ObjectFileControllerImpl) ListAsSelectOptionByFilter(ctx context.Context, f *domain.ObjectFileListFilter) ([]*domain.ObjectFileAsSelectOption, error) {
	// Extract from our session the following data.
	userID := ctx.Value(constants.SessionUserID).(primitive.ObjectID)
	userRole := ctx.Value(constants.SessionUserRole).(int8)

	// Apply protection based on ownership and role.
	if userRole != user_d.UserRoleExecutive {
		c.Logger.Error("authenticated user is not staff role error",
			slog.Any("role", userRole),
			slog.Any("userID", userID))
		return nil, httperror.NewForForbiddenWithSingleField("message", "you role does not grant you access to this")
	}

	c.Logger.Debug("fetching objectfiles now...", slog.Any("userID", userID))

	m, err := c.ObjectFileStorer.ListAsSelectOptionByFilter(ctx, f)
	if err != nil {
		c.Logger.Error("database list by filter error", slog.Any("error", err))
		return nil, err
	}
	c.Logger.Debug("fetched objectfiles", slog.Any("m", m))
	return m, err
}
