package httptransport

import (
	"log/slog"

	shareablelink_c "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/shareablelink/controller"
)

// Handler Creates http request handler
type Handler struct {
	Logger     *slog.Logger
	Controller shareablelink_c.ShareableLinkController
}

// NewHandler Constructor
func NewHandler(loggerp *slog.Logger, c shareablelink_c.ShareableLinkController) *Handler {
	return &Handler{
		Logger:     loggerp,
		Controller: c,
	}
}
