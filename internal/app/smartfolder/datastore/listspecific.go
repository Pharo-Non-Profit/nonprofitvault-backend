package datastore

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (impl SmartFolderStorerImpl) ListByTenantID(ctx context.Context, tid primitive.ObjectID) (*SmartFolderPaginationListResult, error) {
	f := &SmartFolderPaginationListFilter{
		Cursor:    "",
		PageSize:  1_000_000_000, // Unlimited
		SortField: "sort_number",
		SortOrder: 1,
		TenantID:  tid,
		Status:    StatusActive,
	}
	return impl.ListByFilter(ctx, f)
}
