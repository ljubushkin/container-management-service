package httptransport

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ljubushkin/container-management-service/internal/apperror"
	"github.com/ljubushkin/container-management-service/internal/domain"
	"github.com/ljubushkin/container-management-service/internal/usecase"
)

type Handler struct {
	service *usecase.Service
}

func NewHandler(s *usecase.Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) CreateContainer(w http.ResponseWriter, r *http.Request) {
	var req CreateContainerRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, apperror.New(
			apperror.CodeInvalidType,
			"invalid request",
			err,
		))
		return
	}

	c, err := h.service.CreateContainer(req.Type)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toResponse(c))
}

func (h *Handler) CreateContainerBatch(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var req CreateBatchRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, apperror.New(
			apperror.CodeInvalidType,
			"invalid request",
			err,
		))
		return
	}

	containers, err := h.service.CreateBatch(req.Type, req.Count)
	if err != nil {
		writeError(w, err)
		return
	}

	resp := make([]ContainerResponse, 0, len(containers))
	for _, c := range containers {
		resp = append(resp, toResponse(c))
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	c, err := h.service.GetByID(id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toResponse(c))
}

func (h *Handler) AssignWarehouse(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	var req AssignWarehouseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, apperror.New(
			apperror.CodeInvalidWarehouse,
			"invalid request",
			err,
		))
		return
	}

	if err := h.service.AssignWarehouse(id, req.WarehouseID); err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseFilter(r *http.Request) (domain.ContainerFilter, error) {
	q := r.URL.Query()

	var filter domain.ContainerFilter

	if v := q.Get("type"); v != "" {
		filter.TypeCode = &v
	}

	if v := q.Get("warehouse_id"); v != "" {
		filter.WarehouseID = &v
	}

	if v := q.Get("status"); v != "" {
		s := domain.Status(v)
		if !domain.IsValidStatus(s) {
			return filter, apperror.New(
				apperror.CodeInvalidStatus,
				"invalid status",
				nil,
			)
		}
		filter.Status = &s
	}

	if v := q.Get("limit"); v != "" {
		limit, err := strconv.Atoi(v)
		if err != nil {
			return filter, apperror.New(
				apperror.CodeInvalidType, // можно потом выделить CodeInvalidPagination
				"invalid limit",
				err,
			)
		}
		filter.Limit = limit
	}

	if v := q.Get("offset"); v != "" {
		offset, err := strconv.Atoi(v)
		if err != nil {
			return filter, apperror.New(
				apperror.CodeInvalidType,
				"invalid offset",
				err,
			)
		}
		filter.Offset = offset
	}

	return filter, nil
}

func (h *Handler) ListContainers(w http.ResponseWriter, r *http.Request) {
	filter, err := parseFilter(r)
	if err != nil {
		writeError(w, err)
		return
	}

	containers, err := h.service.List(filter)
	if err != nil {
		writeError(w, err)
		return
	}

	resp := make([]ContainerResponse, 0, len(containers))
	for _, c := range containers {
		resp = append(resp, toResponse(c))
	}

	writeJSON(w, http.StatusOK, ListContainersResponse{
		Data: resp,
		Meta: Meta{
			Limit:  filter.Limit,
			Offset: filter.Offset,
			Count:  len(resp),
		},
	})
}
