package httptransport

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	a_c "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/objectfile/controller"
	a_s "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/app/objectfile/datastore"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/utils/httperror"
)

func (h *Handler) unmarshalCreateRequest(ctx context.Context, r *http.Request) (*a_c.ObjectFileCreateRequestIDO, error) {
	defer r.Body.Close()

	// Parse the multipart form data
	err := r.ParseMultipartForm(32 << 20) // Limit the maximum memory used for parsing to 32MB
	if err != nil {
		h.Logger.Error("failed parsing multipart form", slog.Any("error", err))
		return nil, err
	}

	// Get the values of form fields
	name := r.FormValue("name")
	description := r.FormValue("description")
	categoryStr := r.FormValue("category")
	category, _ := strconv.ParseInt(categoryStr, 10, 64)
	classificationStr := r.FormValue("classification")
	classification, _ := strconv.ParseInt(classificationStr, 10, 64)

	// Get the uploaded file from the request
	file, header, err := r.FormFile("file")
	if err != nil {
		h.Logger.Error("failed unmarshalling form file", slog.Any("error", err))
		// return nil, err, http.StatusInternalServerError
	}

	// Initialize our array which will store all the results from the remote server.
	requestData := &a_c.ObjectFileCreateRequestIDO{
		Name:           name,
		Description:    description,
		Category:       uint64(category),
		Classification: uint64(classification),
	}

	if header != nil {
		// Extract filename and filetype from the file header
		requestData.FileName = header.Filename
		requestData.FileType = header.Header.Get("Content-Type")
		requestData.File = file
	}
	return requestData, nil
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := h.unmarshalCreateRequest(ctx, r)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	res, err := h.Controller.Create(ctx, req)
	if err != nil {
		httperror.ResponseError(w, err)
		return
	}

	h.marshalCreateResponse(res, w)
}

func (h *Handler) marshalCreateResponse(res *a_s.ObjectFile, w http.ResponseWriter) {
	if err := json.NewEncoder(w).Encode(&res); err != nil {
		h.Logger.Error("failed encoding", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
