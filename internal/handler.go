package internal

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(httprate.LimitByIP(100, time.Minute))
	r.Post("/api/viewings", h.createViewing)
	r.Post("/api/viewings/query", h.listViewings)
	r.Get("/api/viewings/{id}", h.getViewing)
	r.Put("/api/viewings", h.bulkUpdate)
	r.Get("/swagger/*", httpSwagger.WrapHandler)
	return r
}

// CreateViewing godoc
// @Summary      Create a viewing
// @Tags         viewings
// @Accept       json
// @Produce      json
// @Param        body body CreateViewingRequest true "Viewing details"
// @Success      201 {object} CreateViewingResponse
// @Failure      400 {object} map[string]string
// @Failure      409 {object} map[string]string
// @Failure      422 {object} map[string]string
// @Router       /api/viewings [post]
func (h *Handler) createViewing(w http.ResponseWriter, r *http.Request) {
	var req CreateViewingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	id, err := h.svc.CreateViewing(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrMissingField):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, ErrPastDate):
			writeError(w, http.StatusUnprocessableEntity, err.Error())
		case errors.Is(err, ErrConflict):
			writeError(w, http.StatusConflict, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	var resp CreateViewingResponse
	resp.Data.ID = id
	writeJSON(w, http.StatusCreated, resp)
}

// GetViewing godoc
// @Summary      Get a viewing by ID
// @Tags         viewings
// @Produce      json
// @Param        id path int true "Viewing ID"
// @Success      200 {object} map[string]Viewing
// @Failure      400 {object} map[string]string
// @Failure      404 {object} map[string]string
// @Router       /api/viewings/{id} [get]
func (h *Handler) getViewing(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	v, err := h.svc.GetViewing(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "not found")
		} else {
			writeError(w, http.StatusInternalServerError, "internal error")
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": v})
}

// ListViewings godoc
// @Summary      List viewings
// @Tags         viewings
// @Accept       json
// @Produce      json
// @Param        body body ListViewingsRequest false "Filters and pagination"
// @Success      200 {object} ListViewingsResponse
// @Failure      500 {object} map[string]string
// @Router       /api/viewings/query [post]
func (h *Handler) listViewings(w http.ResponseWriter, r *http.Request) {
	var req ListViewingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	rows, hasMore, nextCursor, err := h.svc.ListViewings(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, ListViewingsResponse{
		Data:       rows,
		HasMore:    hasMore,
		NextCursor: nextCursor,
	})
}

// BulkUpdate godoc
// @Summary      Bulk update viewings (cancel, complete, update notes)
// @Tags         viewings
// @Accept       json
// @Produce      json
// @Param        body body BulkUpdateRequest true "Bulk action"
// @Success      200 {object} map[string]any
// @Failure      400 {object} map[string]string
// @Failure      422 {object} map[string]string
// @Router       /api/viewings [put]
func (h *Handler) bulkUpdate(w http.ResponseWriter, r *http.Request) {
	var req BulkUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if err := h.svc.BulkUpdate(r.Context(), req); err != nil {
		switch {
		case errors.Is(err, ErrInvalidStatus):
			writeError(w, http.StatusUnprocessableEntity, err.Error())
		case errors.Is(err, ErrMissingField):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, ErrNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		case errors.Is(err, ErrInvalidAction):
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": nil})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
