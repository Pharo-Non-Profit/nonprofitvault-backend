package httptransport

import (
	"log/slog"

	sharablelink_c "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/sharablelink/controller"
)

// Handler Creates http request handler
type Handler struct {
	Logger     *slog.Logger
	Controller sharablelink_c.SharableLinkController
}

// NewHandler Constructor
func NewHandler(loggerp *slog.Logger, c sharablelink_c.SharableLinkController) *Handler {
	return &Handler{
		Logger:     loggerp,
		Controller: c,
	}
}
