package httptransport

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ljubushkin/container-management-service/internal/repository/inmemory"
	"github.com/ljubushkin/container-management-service/internal/usecase"
)

func newTestHandler() *Handler {
	containerRepo := inmemory.NewContainerRepo()
	typeRepo := inmemory.NewContainerTypeRepo()
	warehouseRepo := inmemory.NewWarehouseRepo()

	service := usecase.NewService(containerRepo, typeRepo, warehouseRepo)
	return NewHandler(service)
}

func TestHandler_Health_Success(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	h.Health(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	resp := decodeResponse[HealthResponse](t, rr)

	if resp.Status != "ok" {
		t.Fatalf("expected status ok, got %s", resp.Status)
	}
}

func decodeResponse[T any](t *testing.T, rr *httptest.ResponseRecorder) T {
	t.Helper()

	var v T
	if err := json.Unmarshal(rr.Body.Bytes(), &v); err != nil {
		t.Fatalf("unmarshal response: %v, body=%s", err, rr.Body.String())
	}

	return v
}

func createTestContainer(t *testing.T, h *Handler, containerType string) ContainerResponse {
	t.Helper()

	reqBody := CreateContainerRequest{
		Type: containerType,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("marshal request body: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/containers", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	h.CreateContainer(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusCreated, rr.Code, rr.Body.String())
	}

	return decodeResponse[ContainerResponse](t, rr)
}

func firstWarehouseID(t *testing.T, h *Handler) string {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/warehouses", nil)
	rr := httptest.NewRecorder()

	h.ListWarehouses(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	resp := decodeResponse[[]WarehouseResponse](t, rr)
	if len(resp) == 0 {
		t.Fatal("expected non-empty warehouses list")
	}

	return resp[0].ID
}

func TestHandler_CreateContainer_Success(t *testing.T) {
	h := newTestHandler()

	reqBody := CreateContainerRequest{
		Type: "EURO_PALLET",
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("marshal request body: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/containers", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	h.CreateContainer(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusCreated, rr.Code, rr.Body.String())
	}

	resp := decodeResponse[ContainerResponse](t, rr)

	if resp.ID == "" {
		t.Fatal("expected non-empty id")
	}
	if resp.Type != "EURO_PALLET" {
		t.Fatalf("expected type EURO_PALLET, got %s", resp.Type)
	}
	if resp.Status != "valid" {
		t.Fatalf("expected status valid, got %s", resp.Status)
	}
	if resp.CreatedAt == "" {
		t.Fatal("expected non-empty created_at")
	}
	if resp.WarehouseID != nil {
		t.Fatalf("expected warehouse_id to be nil, got %v", *resp.WarehouseID)
	}
}

func TestHandler_CreateContainer_InvalidJSON(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodPost, "/containers", bytes.NewBufferString("{invalid-json"))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	h.CreateContainer(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}

	resp := decodeResponse[ErrorResponse](t, rr)

	if resp.Error.Code != "INVALID_TYPE" {
		t.Fatalf("expected error code INVALID_TYPE, got %s", resp.Error.Code)
	}
	if resp.Error.Message != "invalid request" {
		t.Fatalf("expected message %q, got %q", "invalid request", resp.Error.Message)
	}
}

func TestHandler_CreateContainer_ValidationFailed(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodPost, "/containers", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	h.CreateContainer(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}

	resp := decodeResponse[ErrorResponse](t, rr)

	if resp.Error.Code != "INVALID_TYPE" {
		t.Fatalf("expected error code INVALID_TYPE, got %s", resp.Error.Code)
	}
	if resp.Error.Message != "validation failed" {
		t.Fatalf("expected message %q, got %q", "validation failed", resp.Error.Message)
	}
}

func TestHandler_CreateContainer_InvalidType(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodPost, "/containers", bytes.NewBufferString(`{"type":"UNKNOWN"}`))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	h.CreateContainer(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}

	resp := decodeResponse[ErrorResponse](t, rr)

	if resp.Error.Code != "INVALID_TYPE" {
		t.Fatalf("expected error code INVALID_TYPE, got %s", resp.Error.Code)
	}
}

func TestHandler_CreateContainerBatch_Success(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodPost, "/containers/batch", bytes.NewBufferString(`{"type":"EURO_PALLET","count":3}`))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	h.CreateContainerBatch(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusCreated, rr.Code, rr.Body.String())
	}

	resp := decodeResponse[[]ContainerResponse](t, rr)

	if len(resp) != 3 {
		t.Fatalf("expected 3 containers, got %d", len(resp))
	}

	for i, c := range resp {
		if c.ID == "" {
			t.Fatalf("container %d: expected non-empty id", i)
		}
		if c.Type != "EURO_PALLET" {
			t.Fatalf("container %d: expected type EURO_PALLET, got %s", i, c.Type)
		}
		if c.Status != "valid" {
			t.Fatalf("container %d: expected status valid, got %s", i, c.Status)
		}
	}
}

func TestHandler_CreateContainerBatch_InvalidJSON(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodPost, "/containers/batch", bytes.NewBufferString("{invalid-json"))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	h.CreateContainerBatch(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}

	resp := decodeResponse[ErrorResponse](t, rr)

	if resp.Error.Code != "INVALID_TYPE" {
		t.Fatalf("expected error code INVALID_TYPE, got %s", resp.Error.Code)
	}
	if resp.Error.Message != "invalid request" {
		t.Fatalf("expected message %q, got %q", "invalid request", resp.Error.Message)
	}
}

func TestHandler_CreateContainerBatch_ValidationFailed(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodPost, "/containers/batch", bytes.NewBufferString(`{"type":"EURO_PALLET","count":0}`))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	h.CreateContainerBatch(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}

	resp := decodeResponse[ErrorResponse](t, rr)

	if resp.Error.Code != "INVALID_TYPE" {
		t.Fatalf("expected error code INVALID_TYPE, got %s", resp.Error.Code)
	}
	if resp.Error.Message != "validation failed" {
		t.Fatalf("expected message %q, got %q", "validation failed", resp.Error.Message)
	}
}

func TestHandler_GetByID_Success(t *testing.T) {
	h := newTestHandler()

	created := createTestContainer(t, h, "EURO_PALLET")

	req := httptest.NewRequest(http.MethodGet, "/containers/get?id="+created.ID, nil)
	rr := httptest.NewRecorder()

	h.GetByID(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	resp := decodeResponse[ContainerResponse](t, rr)

	if resp.ID != created.ID {
		t.Fatalf("expected id %s, got %s", created.ID, resp.ID)
	}
	if resp.Type != "EURO_PALLET" {
		t.Fatalf("expected type EURO_PALLET, got %s", resp.Type)
	}
}

func TestHandler_GetByID_NotFound(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/containers/get?id=unknown-id", nil)
	rr := httptest.NewRecorder()

	h.GetByID(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusNotFound, rr.Code, rr.Body.String())
	}

	resp := decodeResponse[ErrorResponse](t, rr)

	if resp.Error.Code != "NOT_FOUND" {
		t.Fatalf("expected error code NOT_FOUND, got %s", resp.Error.Code)
	}
}

func TestHandler_AssignWarehouse_Success(t *testing.T) {
	h := newTestHandler()

	created := createTestContainer(t, h, "EURO_PALLET")
	warehouseID := firstWarehouseID(t, h)

	req := httptest.NewRequest(
		http.MethodPost,
		"/containers/assign?id="+created.ID,
		bytes.NewBufferString(`{"warehouse_id":"`+warehouseID+`"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	h.AssignWarehouse(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusNoContent, rr.Code, rr.Body.String())
	}

	getReq := httptest.NewRequest(http.MethodGet, "/containers/get?id="+created.ID, nil)
	getRR := httptest.NewRecorder()

	h.GetByID(getRR, getReq)

	if getRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, getRR.Code, getRR.Body.String())
	}

	updated := decodeResponse[ContainerResponse](t, getRR)
	if updated.WarehouseID == nil {
		t.Fatal("expected warehouse_id to be assigned")
	}
	if *updated.WarehouseID != warehouseID {
		t.Fatalf("expected warehouse_id %s, got %s", warehouseID, *updated.WarehouseID)
	}
}

func TestHandler_AssignWarehouse_InvalidJSON(t *testing.T) {
	h := newTestHandler()

	created := createTestContainer(t, h, "EURO_PALLET")

	req := httptest.NewRequest(
		http.MethodPost,
		"/containers/assign?id="+created.ID,
		bytes.NewBufferString("{invalid-json"),
	)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	h.AssignWarehouse(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}

	resp := decodeResponse[ErrorResponse](t, rr)

	if resp.Error.Code != "INVALID_WAREHOUSE" {
		t.Fatalf("expected error code INVALID_WAREHOUSE, got %s", resp.Error.Code)
	}
	if resp.Error.Message != "invalid request" {
		t.Fatalf("expected message %q, got %q", "invalid request", resp.Error.Message)
	}
}

func TestHandler_AssignWarehouse_ValidationFailed(t *testing.T) {
	h := newTestHandler()

	created := createTestContainer(t, h, "EURO_PALLET")

	req := httptest.NewRequest(
		http.MethodPost,
		"/containers/assign?id="+created.ID,
		bytes.NewBufferString(`{}`),
	)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	h.AssignWarehouse(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}

	resp := decodeResponse[ErrorResponse](t, rr)

	// Сейчас в handler здесь именно CodeInvalidType.
	if resp.Error.Code != "INVALID_TYPE" {
		t.Fatalf("expected error code INVALID_TYPE, got %s", resp.Error.Code)
	}
	if resp.Error.Message != "validation failed" {
		t.Fatalf("expected message %q, got %q", "validation failed", resp.Error.Message)
	}
}

func TestHandler_AssignWarehouse_InvalidWarehouse(t *testing.T) {
	h := newTestHandler()

	created := createTestContainer(t, h, "EURO_PALLET")

	req := httptest.NewRequest(
		http.MethodPost,
		"/containers/assign?id="+created.ID,
		bytes.NewBufferString(`{"warehouse_id":"unknown-warehouse"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	h.AssignWarehouse(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}

	resp := decodeResponse[ErrorResponse](t, rr)

	if resp.Error.Code != "INVALID_WAREHOUSE" {
		t.Fatalf("expected error code INVALID_WAREHOUSE, got %s", resp.Error.Code)
	}
}

func TestHandler_ListContainers_Success(t *testing.T) {
	h := newTestHandler()

	createTestContainer(t, h, "EURO_PALLET")
	createTestContainer(t, h, "EURO_PALLET")
	createTestContainer(t, h, "BOX")

	req := httptest.NewRequest(http.MethodGet, "/containers?type=EURO_PALLET&limit=10&offset=0", nil)
	rr := httptest.NewRecorder()

	h.ListContainers(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	resp := decodeResponse[ListContainersResponse](t, rr)

	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 containers, got %d", len(resp.Data))
	}
	if resp.Meta.Limit != 10 {
		t.Fatalf("expected meta.limit 10, got %d", resp.Meta.Limit)
	}
	if resp.Meta.Offset != 0 {
		t.Fatalf("expected meta.offset 0, got %d", resp.Meta.Offset)
	}
	if resp.Meta.Count != 2 {
		t.Fatalf("expected meta.count 2, got %d", resp.Meta.Count)
	}

	for _, c := range resp.Data {
		if c.Type != "EURO_PALLET" {
			t.Fatalf("expected only EURO_PALLET containers, got %s", c.Type)
		}
	}
}

func TestHandler_ListContainers_InvalidStatus(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/containers?status=bad-status", nil)
	rr := httptest.NewRecorder()

	h.ListContainers(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}

	resp := decodeResponse[ErrorResponse](t, rr)

	if resp.Error.Code != "INVALID_STATUS" {
		t.Fatalf("expected error code INVALID_STATUS, got %s", resp.Error.Code)
	}
	if resp.Error.Message != "invalid status" {
		t.Fatalf("expected message %q, got %q", "invalid status", resp.Error.Message)
	}
}

func TestHandler_ListContainers_InvalidLimit(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/containers?limit=abc", nil)
	rr := httptest.NewRecorder()

	h.ListContainers(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}

	resp := decodeResponse[ErrorResponse](t, rr)

	if resp.Error.Code != "INVALID_TYPE" {
		t.Fatalf("expected error code INVALID_TYPE, got %s", resp.Error.Code)
	}
	if resp.Error.Message != "invalid limit" {
		t.Fatalf("expected message %q, got %q", "invalid limit", resp.Error.Message)
	}
}

func TestHandler_ListContainers_InvalidOffset(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/containers?offset=abc", nil)
	rr := httptest.NewRecorder()

	h.ListContainers(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}

	resp := decodeResponse[ErrorResponse](t, rr)

	if resp.Error.Code != "INVALID_TYPE" {
		t.Fatalf("expected error code INVALID_TYPE, got %s", resp.Error.Code)
	}
	if resp.Error.Message != "invalid offset" {
		t.Fatalf("expected message %q, got %q", "invalid offset", resp.Error.Message)
	}
}

func TestHandler_ListWarehouses_Success(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/warehouses", nil)
	rr := httptest.NewRecorder()

	h.ListWarehouses(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	resp := decodeResponse[[]WarehouseResponse](t, rr)

	if len(resp) == 0 {
		t.Fatal("expected non-empty warehouses list")
	}
	if resp[0].ID == "" {
		t.Fatal("expected non-empty warehouse id")
	}
	if resp[0].Name == "" {
		t.Fatal("expected non-empty warehouse name")
	}
}

func TestHandler_ListContainerTypes_Success(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(http.MethodGet, "/types", nil)
	rr := httptest.NewRecorder()

	h.ListContainerTypes(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	resp := decodeResponse[[]ContainerTypeResponse](t, rr)

	if len(resp) == 0 {
		t.Fatal("expected non-empty container types list")
	}
	if resp[0].Code == "" {
		t.Fatal("expected non-empty type code")
	}
	if resp[0].Name == "" {
		t.Fatal("expected non-empty type name")
	}
}
