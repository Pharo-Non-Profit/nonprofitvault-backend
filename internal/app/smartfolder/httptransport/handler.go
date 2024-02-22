package httptransport

import (
	"log/slog"

	smartfolder_c "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/smartfolder/controller"
)

// Handler Creates http request handler
type Handler struct {
	Logger     *slog.Logger
	Controller smartfolder_c.SmartFolderController
}

// NewHandler Constructor
func NewHandler(loggerp *slog.Logger, c smartfolder_c.SmartFolderController) *Handler {
	return &Handler{
		Logger:     loggerp,
		Controller: c,
	}
}
