package httptransport

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"

	objectfile_c "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/objectfile/controller"
	sub_s "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/objectfile/datastore"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/utils/httperror"
)

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request, id string) {
	ctx := r.Context()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	m, err := h.Controller.GetByID(ctx, objectID)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	MarshalDetailResponse(m, w)
}

func MarshalDetailResponse(res *sub_s.ObjectFile, w http.ResponseWriter) {
	if err := json.NewEncoder(w).Encode(&res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetPresignedURLByID(w http.ResponseWriter, r *http.Request, id string) {
	ctx := r.Context()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	res, err := h.Controller.GetPresignedURLByID(ctx, objectID)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	MarshalPresignedURLResponse(res, w)
}

func MarshalPresignedURLResponse(res *objectfile_c.PresignedURLResponseIDO, w http.ResponseWriter) {
	if err := json.NewEncoder(w).Encode(&res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetContentByID(w http.ResponseWriter, r *http.Request, id string) {
	ctx := r.Context()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		httperror.ResponseError(w, errors.New("invalid object ID")) // Generic error message
		return
	}

	// Retrieve the file data from the object storage.
	content, filename, contentType, err := h.Controller.GetContent(ctx, objectID)
	if err != nil {
		log.Println("Error retrieving file content:", err)                // Log the error
		httperror.ResponseError(w, errors.New("failed to retrieve file")) // Generic error message
		return
	}

	// Set content headers for file download.
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Type", contentType)

	// Optionally, set the Content-Length header if known.
	// w.Header().Set("Content-Length", strconv.Itoa(len(content)))

	// Write the file content to the response body.
	if _, err := w.Write(content); err != nil {
		log.Println("Error writing file content to response:", err)                        // Log the error
		httperror.ResponseError(w, errors.New("failed to write file content to response")) // Generic error message
		return
	}
}
