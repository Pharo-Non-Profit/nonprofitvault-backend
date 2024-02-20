package httptransport

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/utils/httperror"
)

func (h *Handler) GenerateOTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	res, err := h.Controller.GenerateOTP(ctx)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	if err := json.NewEncoder(w).Encode(&res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GenerateOTPAndQRCodePNGImage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	pngImage, err := h.Controller.GenerateOTPAndQRCodePNGImage(ctx)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}
	// Set the Content-Type header
	w.Header().Set("Content-Type", "image/png")

	// Serve the content
	http.ServeContent(w, r, "opt-qr-code.png", time.Now(), bytes.NewReader(pngImage))
}
