package item

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/common"
	commonerrors "github.com/voidmaindev/go-template/internal/common/errors"
	"github.com/voidmaindev/go-template/internal/common/filter"
)

// mockHandlerService implements Service for handler testing
type mockHandlerService struct {
	createResponse *Item
	createErr      error
	getByIDResp    *Item
	getByIDErr     error
	updateResp     *Item
	updateErr      error
	deleteErr      error
	listResp       *common.PaginatedResult[Item]
	listErr        error
}

func (m *mockHandlerService) Create(ctx context.Context, req *CreateItemRequest) (*Item, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	return m.createResponse, nil
}

func (m *mockHandlerService) GetByID(ctx context.Context, id uint) (*Item, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.getByIDResp, nil
}

func (m *mockHandlerService) Update(ctx context.Context, id uint, req *UpdateItemRequest) (*Item, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	return m.updateResp, nil
}

func (m *mockHandlerService) Delete(ctx context.Context, id uint) error {
	return m.deleteErr
}

func (m *mockHandlerService) List(ctx context.Context, pagination *common.Pagination) (*common.PaginatedResult[Item], error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.listResp, nil
}

func (m *mockHandlerService) ListFiltered(ctx context.Context, params *filter.Params) (*common.PaginatedResult[Item], error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.listResp, nil
}

func testItem() *Item {
	return &Item{
		BaseModel:   common.BaseModel{ID: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		Name:        "Test Item",
		Description: "A test item",
		Price:       1999,
	}
}

func doHandlerReq(t *testing.T, app *fiber.App, method, path string, body any) (int, map[string]any) {
	t.Helper()

	var reqBody *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = bytes.NewReader(b)
	} else {
		reqBody = bytes.NewReader(nil)
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(resp.Body)
	var result map[string]any
	if len(b) > 0 {
		json.Unmarshal(b, &result)
	}
	return resp.StatusCode, result
}

// ================================
// Create tests
// ================================

func TestHandler_Create_Success(t *testing.T) {
	svc := &mockHandlerService{createResponse: testItem()}
	handler := NewHandler(svc)

	app := fiber.New()
	app.Post("/items", handler.Create)

	status, body := doHandlerReq(t, app, http.MethodPost, "/items", map[string]any{
		"name":        "Test Item",
		"description": "A test item",
		"price":       1999,
	})

	if status != http.StatusCreated {
		t.Errorf("expected 201, got %d", status)
	}
	if body["success"] != true {
		t.Error("expected success=true")
	}
}

func TestHandler_Create_InvalidJSON(t *testing.T) {
	svc := &mockHandlerService{}
	handler := NewHandler(svc)

	app := fiber.New()
	app.Post("/items", handler.Create)

	req := httptest.NewRequest(http.MethodPost, "/items", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", resp.StatusCode)
	}
}

func TestHandler_Create_ValidationFailure(t *testing.T) {
	svc := &mockHandlerService{}
	handler := NewHandler(svc)

	app := fiber.New()
	app.Post("/items", handler.Create)

	// Missing required "name"
	status, _ := doHandlerReq(t, app, http.MethodPost, "/items", map[string]any{
		"price": 100,
	})

	if status != http.StatusBadRequest {
		t.Errorf("expected 400 for missing required field, got %d", status)
	}
}

func TestHandler_Create_ServiceError(t *testing.T) {
	svc := &mockHandlerService{
		createErr: commonerrors.Internal("item", nil),
	}
	handler := NewHandler(svc)

	app := fiber.New()
	app.Post("/items", handler.Create)

	status, _ := doHandlerReq(t, app, http.MethodPost, "/items", map[string]any{
		"name": "Test",
	})

	if status != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", status)
	}
}

// ================================
// GetByID tests
// ================================

func TestHandler_GetByID_Success(t *testing.T) {
	svc := &mockHandlerService{getByIDResp: testItem()}
	handler := NewHandler(svc)

	app := fiber.New()
	app.Get("/items/:id", handler.GetByID)

	status, body := doHandlerReq(t, app, http.MethodGet, "/items/1", nil)

	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
	if body["success"] != true {
		t.Error("expected success=true")
	}
}

func TestHandler_GetByID_InvalidID(t *testing.T) {
	svc := &mockHandlerService{}
	handler := NewHandler(svc)

	app := fiber.New()
	app.Get("/items/:id", handler.GetByID)

	status, body := doHandlerReq(t, app, http.MethodGet, "/items/abc", nil)

	if status != http.StatusBadRequest {
		t.Errorf("expected 400 for non-numeric ID, got %d", status)
	}
	if errMap, ok := body["error"].(map[string]any); ok {
		if errMap["message"] != "invalid item ID" {
			t.Errorf("unexpected error message: %v", errMap["message"])
		}
	}
}

func TestHandler_GetByID_NotFound(t *testing.T) {
	svc := &mockHandlerService{getByIDErr: ErrItemNotFound}
	handler := NewHandler(svc)

	app := fiber.New()
	app.Get("/items/:id", handler.GetByID)

	status, body := doHandlerReq(t, app, http.MethodGet, "/items/999", nil)

	if status != http.StatusNotFound {
		t.Errorf("expected 404, got %d", status)
	}
	if errMap, ok := body["error"].(map[string]any); ok {
		if errMap["message"] != "item not found" {
			t.Errorf("unexpected error message: %v", errMap["message"])
		}
	}
}

// ================================
// Update tests
// ================================

func TestHandler_Update_Success(t *testing.T) {
	updated := testItem()
	updated.Name = "Updated"
	svc := &mockHandlerService{updateResp: updated}
	handler := NewHandler(svc)

	app := fiber.New()
	app.Put("/items/:id", handler.Update)

	name := "Updated"
	status, body := doHandlerReq(t, app, http.MethodPut, "/items/1", map[string]any{
		"name": name,
	})

	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
	if body["success"] != true {
		t.Error("expected success=true")
	}
}

func TestHandler_Update_InvalidID(t *testing.T) {
	handler := NewHandler(&mockHandlerService{})

	app := fiber.New()
	app.Put("/items/:id", handler.Update)

	status, _ := doHandlerReq(t, app, http.MethodPut, "/items/xyz", map[string]any{
		"name": "test",
	})

	if status != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", status)
	}
}

func TestHandler_Update_InvalidBody(t *testing.T) {
	handler := NewHandler(&mockHandlerService{})

	app := fiber.New()
	app.Put("/items/:id", handler.Update)

	req := httptest.NewRequest(http.MethodPut, "/items/1", bytes.NewReader([]byte("bad json")))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req, -1)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandler_Update_NotFound(t *testing.T) {
	svc := &mockHandlerService{updateErr: ErrItemNotFound}
	handler := NewHandler(svc)

	app := fiber.New()
	app.Put("/items/:id", handler.Update)

	status, _ := doHandlerReq(t, app, http.MethodPut, "/items/999", map[string]any{
		"name": "test",
	})

	if status != http.StatusNotFound {
		t.Errorf("expected 404 via HandleError, got %d", status)
	}
}

// ================================
// Delete tests
// ================================

func TestHandler_Delete_Success(t *testing.T) {
	handler := NewHandler(&mockHandlerService{})

	app := fiber.New()
	app.Delete("/items/:id", handler.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/items/1", nil)
	resp, _ := app.Test(req, -1)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected 204, got %d", resp.StatusCode)
	}
}

func TestHandler_Delete_InvalidID(t *testing.T) {
	handler := NewHandler(&mockHandlerService{})

	app := fiber.New()
	app.Delete("/items/:id", handler.Delete)

	status, _ := doHandlerReq(t, app, http.MethodDelete, "/items/abc", nil)

	if status != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", status)
	}
}

func TestHandler_Delete_NotFound(t *testing.T) {
	svc := &mockHandlerService{deleteErr: ErrItemNotFound}
	handler := NewHandler(svc)

	app := fiber.New()
	app.Delete("/items/:id", handler.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/items/999", nil)
	resp, _ := app.Test(req, -1)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 via HandleError, got %d", resp.StatusCode)
	}
}

// ================================
// List tests
// ================================

func TestHandler_List_Success(t *testing.T) {
	items := []Item{*testItem()}
	svc := &mockHandlerService{
		listResp: &common.PaginatedResult[Item]{
			Data:       items,
			Total:      1,
			Page:       1,
			PageSize:   10,
			TotalPages: 1,
			HasMore:    false,
		},
	}
	handler := NewHandler(svc)

	app := fiber.New()
	app.Get("/items", handler.List)

	status, body := doHandlerReq(t, app, http.MethodGet, "/items?page=1&page_size=10", nil)

	if status != http.StatusOK {
		t.Errorf("expected 200, got %d", status)
	}
	if body["success"] != true {
		t.Error("expected success=true")
	}

	// Verify flat pagination structure (not double-wrapped)
	data, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatal("expected data to be an object")
	}

	// These fields should exist at the top level of data (flat pagination)
	if _, ok := data["data"]; !ok {
		t.Error("expected data.data (items array)")
	}
	if _, ok := data["total"]; !ok {
		t.Error("expected data.total")
	}
	if _, ok := data["page"]; !ok {
		t.Error("expected data.page")
	}
	if _, ok := data["page_size"]; !ok {
		t.Error("expected data.page_size")
	}
	if _, ok := data["total_pages"]; !ok {
		t.Error("expected data.total_pages")
	}

	// Verify items are in data.data, not double-nested
	innerData, ok := data["data"].([]any)
	if !ok {
		t.Fatal("expected data.data to be an array")
	}
	if len(innerData) != 1 {
		t.Errorf("expected 1 item, got %d", len(innerData))
	}

	// Verify the items are DTOs (have "id", "name" keys), not raw models
	firstItem, ok := innerData[0].(map[string]any)
	if !ok {
		t.Fatal("expected items to be objects")
	}
	if _, ok := firstItem["name"]; !ok {
		t.Error("expected item to have 'name' field (DTO conversion)")
	}
}

func TestHandler_List_ServiceError(t *testing.T) {
	svc := &mockHandlerService{
		listErr: commonerrors.Internal("item", nil),
	}
	handler := NewHandler(svc)

	app := fiber.New()
	app.Get("/items", handler.List)

	status, _ := doHandlerReq(t, app, http.MethodGet, "/items", nil)

	if status != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", status)
	}
}

// ================================
// HandleError routing tests — verify errors.Is() removal is safe
// ================================

func TestHandler_HandleError_RoutesNotFoundCorrectly(t *testing.T) {
	// This test verifies that after removing manual errors.Is(err, ErrItemNotFound)
	// checks, HandleError still routes NotFound errors to 404
	svc := &mockHandlerService{getByIDErr: ErrItemNotFound}
	handler := NewHandler(svc)

	app := fiber.New()
	app.Get("/items/:id", handler.GetByID)

	status, body := doHandlerReq(t, app, http.MethodGet, "/items/1", nil)

	if status != http.StatusNotFound {
		t.Fatalf("HandleError should route ErrItemNotFound to 404, got %d", status)
	}

	errMap, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatal("expected error in response")
	}
	if errMap["code"] != "NOT_FOUND" {
		t.Errorf("expected NOT_FOUND code, got %v", errMap["code"])
	}
	if errMap["domain"] != "item" {
		t.Errorf("expected domain 'item', got %v", errMap["domain"])
	}
	if errMap["message"] != "item not found" {
		t.Errorf("expected 'item not found', got %v", errMap["message"])
	}
}

func TestHandler_HandleError_RoutesInternalCorrectly(t *testing.T) {
	svc := &mockHandlerService{
		deleteErr: commonerrors.Internal("item", nil),
	}
	handler := NewHandler(svc)

	app := fiber.New()
	app.Delete("/items/:id", handler.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/items/1", nil)
	resp, _ := app.Test(req, -1)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", resp.StatusCode)
	}
}

func TestHandler_HandleError_RoutesAlreadyExistsTo409(t *testing.T) {
	svc := &mockHandlerService{
		createErr: ErrItemNameExists,
	}
	handler := NewHandler(svc)

	app := fiber.New()
	app.Post("/items", handler.Create)

	status, body := doHandlerReq(t, app, http.MethodPost, "/items", map[string]any{
		"name": "Duplicate",
	})

	if status != http.StatusConflict {
		t.Errorf("expected 409 for AlreadyExists, got %d", status)
	}
	if errMap, ok := body["error"].(map[string]any); ok {
		if errMap["code"] != "ALREADY_EXISTS" {
			t.Errorf("expected ALREADY_EXISTS code, got %v", errMap["code"])
		}
	}
}
