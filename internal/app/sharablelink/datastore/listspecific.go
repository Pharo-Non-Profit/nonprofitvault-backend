package datastore

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (impl SharableLinkStorerImpl) ListByTenantID(ctx context.Context, tid primitive.ObjectID) (*SharableLinkPaginationListResult, error) {
	f := &SharableLinkPaginationListFilter{
		Cursor:    "",
		PageSize:  1_000_000_000, // Unlimited
		SortField: "sort_number",
		SortOrder: 1,
		TenantID:  tid,
		Status:    StatusActive,
	}
	return impl.ListByFilter(ctx, f)
}
