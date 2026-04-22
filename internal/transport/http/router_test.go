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

func newTestRouter() http.Handler {
	containerRepo := inmemory.NewContainerRepo()
	typeRepo := inmemory.NewContainerTypeRepo()
	warehouseRepo := inmemory.NewWarehouseRepo()

	service := usecase.NewService(containerRepo, typeRepo, warehouseRepo)
	handler := NewHandler(service)

	return NewRouter(handler)
}

func decodeRouterResponse[T any](t *testing.T, rr *httptest.ResponseRecorder) T {
	t.Helper()

	var v T
	if err := json.Unmarshal(rr.Body.Bytes(), &v); err != nil {
		t.Fatalf("unmarshal response: %v, body=%s", err, rr.Body.String())
	}

	return v
}

func TestRouter_Health_Success(t *testing.T) {
	router := newTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected %d got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	resp := decodeRouterResponse[HealthResponse](t, rr)

	if resp.Status != "ok" {
		t.Fatalf("expected status ok, got %s", resp.Status)
	}
}

func TestRouter_Health_MethodNotAllowed(t *testing.T) {
	router := newTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected %d got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func createContainerViaRouter(t *testing.T, router http.Handler, containerType string) ContainerResponse {
	t.Helper()

	req := httptest.NewRequest(
		http.MethodPost,
		"/containers",
		bytes.NewBufferString(`{"type":"`+containerType+`"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusCreated, rr.Code, rr.Body.String())
	}

	return decodeRouterResponse[ContainerResponse](t, rr)
}

func firstWarehouseIDViaRouter(t *testing.T, router http.Handler) string {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/warehouses", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	resp := decodeRouterResponse[[]WarehouseResponse](t, rr)
	if len(resp) == 0 {
		t.Fatal("expected non-empty warehouses list")
	}

	return resp[0].ID
}

func TestRouter_CreateContainer_Success(t *testing.T) {
	router := newTestRouter()

	req := httptest.NewRequest(
		http.MethodPost,
		"/containers",
		bytes.NewBufferString(`{"type":"EURO_PALLET"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected %d got %d body=%s", http.StatusCreated, rr.Code, rr.Body.String())
	}

	resp := decodeRouterResponse[ContainerResponse](t, rr)

	if resp.ID == "" {
		t.Fatal("expected non-empty id")
	}
	if resp.Type != "EURO_PALLET" {
		t.Fatalf("expected type EURO_PALLET, got %s", resp.Type)
	}
	if resp.Status != "valid" {
		t.Fatalf("expected status valid, got %s", resp.Status)
	}
}

func TestRouter_CreateContainer_MethodNotAllowed(t *testing.T) {
	router := newTestRouter()

	req := httptest.NewRequest(http.MethodPut, "/containers", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected %d got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestRouter_CreateContainerBatch_Success(t *testing.T) {
	router := newTestRouter()

	req := httptest.NewRequest(
		http.MethodPost,
		"/containers/batch",
		bytes.NewBufferString(`{"type":"EURO_PALLET","count":3}`),
	)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected %d got %d body=%s", http.StatusCreated, rr.Code, rr.Body.String())
	}

	resp := decodeRouterResponse[[]ContainerResponse](t, rr)

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
	}
}

func TestRouter_CreateContainerBatch_MethodNotAllowed(t *testing.T) {
	router := newTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/containers/batch", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected %d got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestRouter_GetByID_Success(t *testing.T) {
	router := newTestRouter()

	created := createContainerViaRouter(t, router, "EURO_PALLET")

	req := httptest.NewRequest(http.MethodGet, "/containers/get?id="+created.ID, nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected %d got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	resp := decodeRouterResponse[ContainerResponse](t, rr)

	if resp.ID != created.ID {
		t.Fatalf("expected id %s, got %s", created.ID, resp.ID)
	}
	if resp.Type != "EURO_PALLET" {
		t.Fatalf("expected type EURO_PALLET, got %s", resp.Type)
	}
}

func TestRouter_GetByID_NotFound(t *testing.T) {
	router := newTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/containers/get?id=unknown-id", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected %d got %d body=%s", http.StatusNotFound, rr.Code, rr.Body.String())
	}

	resp := decodeRouterResponse[ErrorResponse](t, rr)

	if resp.Error.Code != "NOT_FOUND" {
		t.Fatalf("expected error code NOT_FOUND, got %s", resp.Error.Code)
	}
}

func TestRouter_GetByID_MethodNotAllowed(t *testing.T) {
	router := newTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/containers/get?id=some-id", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected %d got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestRouter_AssignWarehouse_Success(t *testing.T) {
	router := newTestRouter()

	created := createContainerViaRouter(t, router, "EURO_PALLET")
	warehouseID := firstWarehouseIDViaRouter(t, router)

	assignReq := httptest.NewRequest(
		http.MethodPost,
		"/containers/assign?id="+created.ID,
		bytes.NewBufferString(`{"warehouse_id":"`+warehouseID+`"}`),
	)
	assignReq.Header.Set("Content-Type", "application/json")

	assignRR := httptest.NewRecorder()
	router.ServeHTTP(assignRR, assignReq)

	if assignRR.Code != http.StatusNoContent {
		t.Fatalf("expected %d got %d body=%s", http.StatusNoContent, assignRR.Code, assignRR.Body.String())
	}

	getReq := httptest.NewRequest(http.MethodGet, "/containers/get?id="+created.ID, nil)
	getRR := httptest.NewRecorder()
	router.ServeHTTP(getRR, getReq)

	if getRR.Code != http.StatusOK {
		t.Fatalf("expected %d got %d body=%s", http.StatusOK, getRR.Code, getRR.Body.String())
	}

	resp := decodeRouterResponse[ContainerResponse](t, getRR)

	if resp.WarehouseID == nil {
		t.Fatal("expected warehouse_id to be assigned")
	}
	if *resp.WarehouseID != warehouseID {
		t.Fatalf("expected warehouse_id %s, got %s", warehouseID, *resp.WarehouseID)
	}
}

func TestRouter_AssignWarehouse_InvalidWarehouse(t *testing.T) {
	router := newTestRouter()

	created := createContainerViaRouter(t, router, "EURO_PALLET")

	req := httptest.NewRequest(
		http.MethodPost,
		"/containers/assign?id="+created.ID,
		bytes.NewBufferString(`{"warehouse_id":"unknown-warehouse"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected %d got %d body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}

	resp := decodeRouterResponse[ErrorResponse](t, rr)

	if resp.Error.Code != "INVALID_WAREHOUSE" {
		t.Fatalf("expected error code INVALID_WAREHOUSE, got %s", resp.Error.Code)
	}
}

func TestRouter_AssignWarehouse_MethodNotAllowed(t *testing.T) {
	router := newTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/containers/assign?id=some-id", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected %d got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestRouter_ListContainers_Success(t *testing.T) {
	router := newTestRouter()

	createContainerViaRouter(t, router, "EURO_PALLET")
	createContainerViaRouter(t, router, "EURO_PALLET")
	createContainerViaRouter(t, router, "BOX")

	req := httptest.NewRequest(http.MethodGet, "/containers?type=EURO_PALLET&limit=10&offset=0", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected %d got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	resp := decodeRouterResponse[ListContainersResponse](t, rr)

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

func TestRouter_ListContainers_InvalidStatus(t *testing.T) {
	router := newTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/containers?status=bad-status", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected %d got %d body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}

	resp := decodeRouterResponse[ErrorResponse](t, rr)

	if resp.Error.Code != "INVALID_STATUS" {
		t.Fatalf("expected error code INVALID_STATUS, got %s", resp.Error.Code)
	}
}

func TestRouter_ListWarehouses_Success(t *testing.T) {
	router := newTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/warehouses", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected %d got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	resp := decodeRouterResponse[[]WarehouseResponse](t, rr)

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

func TestRouter_ListWarehouses_MethodNotAllowed(t *testing.T) {
	router := newTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/warehouses", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected %d got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestRouter_ListContainerTypes_Success(t *testing.T) {
	router := newTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/types", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected %d got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	resp := decodeRouterResponse[[]ContainerTypeResponse](t, rr)

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

func TestRouter_ListContainerTypes_MethodNotAllowed(t *testing.T) {
	router := newTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/types", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected %d got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}
