package httptransport

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	smartfolder_c "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/smartfolder/controller"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/utils/httperror"
)

func UnmarshalGenerateSharableLinkRequest(ctx context.Context, r *http.Request) (*smartfolder_c.GenerateSharableLinkRequestIDO, error) {
	// Initialize our array which will store all the results from the remote server.
	var requestData smartfolder_c.GenerateSharableLinkRequestIDO

	defer r.Body.Close()

	// Read the JSON string and convert it into our golang stuct else we need
	// to send a `400 Bad Request` errror message back to the client,
	err := json.NewDecoder(r.Body).Decode(&requestData) // [1]
	if err != nil {
		log.Println(err)
		return nil, httperror.NewForSingleField(http.StatusBadRequest, "non_field_error", "payload structure is wrong")
	}
	return &requestData, nil
}

func (h *Handler) GenerateSharableLink(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	data, err := UnmarshalGenerateSharableLinkRequest(ctx, r)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	res, err := h.Controller.GenerateSharableLink(ctx, data)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	if err := json.NewEncoder(w).Encode(&res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
