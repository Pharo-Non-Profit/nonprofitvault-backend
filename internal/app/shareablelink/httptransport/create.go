package httptransport

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	shareablelink_c "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/shareablelink/controller"
	shareablelink_s "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/shareablelink/datastore"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/utils/httperror"
)

func UnmarshalCreateRequest(ctx context.Context, r *http.Request) (*shareablelink_c.ShareableLinkCreateRequestIDO, error) {
	// Initialize our array which will store all the results from the remote server.
	var requestData shareablelink_c.ShareableLinkCreateRequestIDO

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

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	data, err := UnmarshalCreateRequest(ctx, r)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	res, err := h.Controller.Create(ctx, data)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	MarshalCreateResponse(res, w)
}

func MarshalCreateResponse(res *shareablelink_s.ShareableLink, w http.ResponseWriter) {
	if err := json.NewEncoder(w).Encode(&res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
