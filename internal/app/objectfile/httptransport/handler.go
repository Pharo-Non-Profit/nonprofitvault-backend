package httptransport

import (
	"log/slog"

	objectfile_c "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/objectfile/controller"
)

// Handler Creates http request handler
type Handler struct {
	Logger     *slog.Logger
	Controller objectfile_c.ObjectFileController
}

// NewHandler Constructor
func NewHandler(loggerp *slog.Logger, c objectfile_c.ObjectFileController) *Handler {
	return &Handler{
		Logger:     loggerp,
		Controller: c,
	}
}
