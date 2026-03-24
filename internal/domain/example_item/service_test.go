package example_item

import (
	"context"
	"testing"
	"time"

	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/common/errors"
	"github.com/voidmaindev/go-template/internal/common/filter"
	"gorm.io/gorm"
)

// mockRepository implements the Repository interface for testing
type mockRepository struct {
	items     map[uint]*Item
	nameIndex map[string]*Item
	nextID    uint
	findErr   error
	createErr error
	updateErr error
	deleteErr error
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		items:     make(map[uint]*Item),
		nameIndex: make(map[string]*Item),
		nextID:    1,
	}
}

func (m *mockRepository) Create(ctx context.Context, entity *Item) error {
	if m.createErr != nil {
		return m.createErr
	}
	entity.ID = m.nextID
	entity.CreatedAt = time.Now()
	entity.UpdatedAt = time.Now()
	m.nextID++
	m.items[entity.ID] = entity
	m.nameIndex[entity.Name] = entity
	return nil
}

func (m *mockRepository) CreateBatch(ctx context.Context, entities []Item, batchSize int) error {
	for i := range entities {
		if err := m.Create(ctx, &entities[i]); err != nil {
			return err
		}
	}
	return nil
}

func (m *mockRepository) FindByID(ctx context.Context, id uint) (*Item, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	item, ok := m.items[id]
	if !ok {
		return nil, errors.NotFound("repository", "entity")
	}
	return item, nil
}

func (m *mockRepository) FindAll(ctx context.Context, pagination *common.Pagination) ([]Item, int64, error) {
	var items []Item
	for _, item := range m.items {
		items = append(items, *item)
	}
	return items, int64(len(items)), nil
}

func (m *mockRepository) FindByCondition(ctx context.Context, condition map[string]any, pagination *common.Pagination) ([]Item, int64, error) {
	return m.FindAll(ctx, pagination)
}

func (m *mockRepository) FindAllFiltered(ctx context.Context, params *filter.Params) ([]Item, int64, error) {
	return m.FindAll(ctx, nil)
}

func (m *mockRepository) FindOne(ctx context.Context, condition map[string]any) (*Item, error) {
	if name, ok := condition["name"].(string); ok {
		if item, exists := m.nameIndex[name]; exists {
			return item, nil
		}
	}
	return nil, errors.NotFound("repository", "entity")
}

func (m *mockRepository) Update(ctx context.Context, entity *Item) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, ok := m.items[entity.ID]; !ok {
		return errors.NotFound("repository", "entity")
	}
	entity.UpdatedAt = time.Now()
	m.items[entity.ID] = entity
	m.nameIndex[entity.Name] = entity
	return nil
}

func (m *mockRepository) UpdateFields(ctx context.Context, id uint, fields map[string]any) error {
	return m.updateErr
}

func (m *mockRepository) Delete(ctx context.Context, id uint) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	item, ok := m.items[id]
	if !ok {
		return errors.NotFound("repository", "entity")
	}
	delete(m.items, id)
	delete(m.nameIndex, item.Name)
	return nil
}

func (m *mockRepository) HardDelete(ctx context.Context, id uint) error {
	return m.Delete(ctx, id)
}

func (m *mockRepository) Exists(ctx context.Context, condition map[string]any) (bool, error) {
	if name, ok := condition["name"].(string); ok {
		_, exists := m.nameIndex[name]
		return exists, nil
	}
	return false, nil
}

func (m *mockRepository) Count(ctx context.Context, condition map[string]any) (int64, error) {
	return int64(len(m.items)), nil
}

func (m *mockRepository) WithTx(tx *gorm.DB) common.Repository[Item] {
	return m
}

func (m *mockRepository) WithPreload(preloads ...string) common.Repository[Item] {
	return m
}

func (m *mockRepository) Transaction(ctx context.Context, fn func(txRepo common.Repository[Item]) error) error {
	return fn(m)
}

func (m *mockRepository) FindByName(ctx context.Context, name string) (*Item, error) {
	item, ok := m.nameIndex[name]
	if !ok {
		return nil, errors.NotFound("repository", "entity")
	}
	return item, nil
}

// Helper to seed test data
func (m *mockRepository) seed(name string, description string, price int64) *Item {
	item := &Item{
		Name:        name,
		Description: description,
		Price:       price,
	}
	m.Create(context.Background(), item)
	return item
}

func TestService_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		repo := newMockRepository()
		svc := NewService(repo)

		req := &CreateItemRequest{
			Name:        "Test Item",
			Description: "A test item",
			Price:       1999,
		}

		item, err := svc.Create(ctx, req)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		if item.ID == 0 {
			t.Error("item.ID should be set")
		}
		if item.Name != req.Name {
			t.Errorf("Name = %s, want %s", item.Name, req.Name)
		}
		if item.Description != req.Description {
			t.Errorf("Description = %s, want %s", item.Description, req.Description)
		}
		if item.Price != req.Price {
			t.Errorf("Price = %d, want %d", item.Price, req.Price)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		repo := newMockRepository()
		repo.createErr = errors.Internal("test", nil)
		svc := NewService(repo)

		req := &CreateItemRequest{
			Name:  "Test Item",
			Price: 1999,
		}

		_, err := svc.Create(ctx, req)
		if err == nil {
			t.Error("expected error")
		}
	})
}

func TestService_GetByID(t *testing.T) {
	ctx := context.Background()

	t.Run("existing item", func(t *testing.T) {
		repo := newMockRepository()
		item := repo.seed("Test Item", "Description", 1999)
		svc := NewService(repo)

		found, err := svc.GetByID(ctx, item.ID)
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}

		if found.ID != item.ID {
			t.Errorf("ID = %d, want %d", found.ID, item.ID)
		}
		if found.Name != item.Name {
			t.Errorf("Name = %s, want %s", found.Name, item.Name)
		}
	})

	t.Run("non-existent item", func(t *testing.T) {
		repo := newMockRepository()
		svc := NewService(repo)

		_, err := svc.GetByID(ctx, 999)
		if !errors.IsNotFound(err) {
			t.Errorf("GetByID() error = %v, want not found error", err)
		}
	})
}

func TestService_Update(t *testing.T) {
	ctx := context.Background()

	t.Run("successful update", func(t *testing.T) {
		repo := newMockRepository()
		item := repo.seed("Original Name", "Original Description", 1000)
		svc := NewService(repo)

		newName := "Updated Name"
		newPrice := int64(2000)
		req := &UpdateItemRequest{
			Name:  &newName,
			Price: &newPrice,
		}

		updated, err := svc.Update(ctx, item.ID, req)
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		if updated.Name != newName {
			t.Errorf("Name = %s, want %s", updated.Name, newName)
		}
		if updated.Price != newPrice {
			t.Errorf("Price = %d, want %d", updated.Price, newPrice)
		}
	})

	t.Run("partial update", func(t *testing.T) {
		repo := newMockRepository()
		item := repo.seed("Original Name", "Original Description", 1000)
		svc := NewService(repo)

		newDescription := "New Description"
		req := &UpdateItemRequest{
			Description: &newDescription,
		}

		updated, err := svc.Update(ctx, item.ID, req)
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		if updated.Name != item.Name {
			t.Errorf("Name = %s, want %s (unchanged)", updated.Name, item.Name)
		}
		if updated.Description != newDescription {
			t.Errorf("Description = %s, want %s", updated.Description, newDescription)
		}
	})

	t.Run("non-existent item", func(t *testing.T) {
		repo := newMockRepository()
		svc := NewService(repo)

		newName := "Updated"
		req := &UpdateItemRequest{
			Name: &newName,
		}

		_, err := svc.Update(ctx, 999, req)
		if !errors.IsNotFound(err) {
			t.Errorf("Update() error = %v, want not found error", err)
		}
	})
}

func TestService_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("successful deletion", func(t *testing.T) {
		repo := newMockRepository()
		item := repo.seed("Test Item", "Description", 1999)
		svc := NewService(repo)

		err := svc.Delete(ctx, item.ID)
		if err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		// Verify item is deleted
		_, err = svc.GetByID(ctx, item.ID)
		if !errors.IsNotFound(err) {
			t.Error("item should be deleted")
		}
	})

	t.Run("non-existent item", func(t *testing.T) {
		repo := newMockRepository()
		svc := NewService(repo)

		err := svc.Delete(ctx, 999)
		if !errors.IsNotFound(err) {
			t.Errorf("Delete() error = %v, want not found error", err)
		}
	})
}

func TestService_List(t *testing.T) {
	ctx := context.Background()

	t.Run("list with items", func(t *testing.T) {
		repo := newMockRepository()
		repo.seed("Item 1", "Description 1", 100)
		repo.seed("Item 2", "Description 2", 200)
		repo.seed("Item 3", "Description 3", 300)
		svc := NewService(repo)

		pagination := &common.Pagination{Page: 1, PageSize: 10}
		result, err := svc.List(ctx, pagination)
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}

		if result.Total != 3 {
			t.Errorf("Total = %d, want 3", result.Total)
		}
		if len(result.Data) != 3 {
			t.Errorf("len(Data) = %d, want 3", len(result.Data))
		}
	})

	t.Run("empty list", func(t *testing.T) {
		repo := newMockRepository()
		svc := NewService(repo)

		pagination := &common.Pagination{Page: 1, PageSize: 10}
		result, err := svc.List(ctx, pagination)
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}

		if result.Total != 0 {
			t.Errorf("Total = %d, want 0", result.Total)
		}
	})
}

func TestService_ListFiltered(t *testing.T) {
	ctx := context.Background()

	t.Run("list with filter params", func(t *testing.T) {
		repo := newMockRepository()
		repo.seed("Item 1", "Description 1", 100)
		repo.seed("Item 2", "Description 2", 200)
		svc := NewService(repo)

		params := &filter.Params{
			Page:  1,
			Limit: 10,
		}

		result, err := svc.ListFiltered(ctx, params)
		if err != nil {
			t.Fatalf("ListFiltered() error = %v", err)
		}

		if result.Total != 2 {
			t.Errorf("Total = %d, want 2", result.Total)
		}
	})
}

func TestItem_ToResponse(t *testing.T) {
	now := time.Now()
	item := &Item{
		Name:        "Test Item",
		Description: "Test Description",
		Price:       1999,
	}
	item.ID = 1
	item.CreatedAt = now
	item.UpdatedAt = now

	resp := item.ToResponse()

	if resp.ID != item.ID {
		t.Errorf("ID = %d, want %d", resp.ID, item.ID)
	}
	if resp.Name != item.Name {
		t.Errorf("Name = %s, want %s", resp.Name, item.Name)
	}
	if resp.Description != item.Description {
		t.Errorf("Description = %s, want %s", resp.Description, item.Description)
	}
	if resp.Price != item.Price {
		t.Errorf("Price = %d, want %d", resp.Price, item.Price)
	}
}

func TestItem_TableName(t *testing.T) {
	item := Item{}
	if item.TableName() != "items" {
		t.Errorf("TableName() = %s, want items", item.TableName())
	}
}

func TestItem_FilterConfig(t *testing.T) {
	item := Item{}
	config := item.FilterConfig()

	if config.TableName != "items" {
		t.Errorf("TableName = %s, want items", config.TableName)
	}

	expectedFields := []string{"id", "name", "description", "price", "created_at", "updated_at"}
	for _, field := range expectedFields {
		if _, ok := config.Fields[field]; !ok {
			t.Errorf("Expected field %s not found in FilterConfig", field)
		}
	}
}

func TestErrors(t *testing.T) {
	t.Run("ErrItemNotFound is not found error", func(t *testing.T) {
		if !errors.IsNotFound(ErrItemNotFound) {
			t.Error("ErrItemNotFound should be a not found error")
		}
	})

	t.Run("ErrItemNameExists is already exists error", func(t *testing.T) {
		if !errors.IsAlreadyExists(ErrItemNameExists) {
			t.Error("ErrItemNameExists should be an already exists error")
		}
	})
}

func TestMockRepository_CRUD(t *testing.T) {
	repo := newMockRepository()
	ctx := context.Background()

	// Create
	item := &Item{
		Name:        "Test",
		Description: "Test Desc",
		Price:       1000,
	}
	err := repo.Create(ctx, item)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if item.ID == 0 {
		t.Error("ID should be set after create")
	}

	// Read
	found, err := repo.FindByID(ctx, item.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if found.Name != "Test" {
		t.Errorf("Name = %s, want Test", found.Name)
	}

	// FindByName
	byName, err := repo.FindByName(ctx, "Test")
	if err != nil {
		t.Fatalf("FindByName() error = %v", err)
	}
	if byName.ID != item.ID {
		t.Errorf("ID = %d, want %d", byName.ID, item.ID)
	}

	// Update
	found.Name = "Updated"
	err = repo.Update(ctx, found)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Verify update
	found, _ = repo.FindByID(ctx, item.ID)
	if found.Name != "Updated" {
		t.Errorf("Name = %s, want Updated", found.Name)
	}

	// Delete
	err = repo.Delete(ctx, item.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify delete
	_, err = repo.FindByID(ctx, item.ID)
	if !errors.IsNotFound(err) {
		t.Error("Item should be deleted")
	}
}

func TestMockRepository_Errors(t *testing.T) {
	repo := newMockRepository()
	ctx := context.Background()

	t.Run("create error", func(t *testing.T) {
		repo.createErr = errors.Internal("test", nil)
		err := repo.Create(ctx, &Item{Name: "test"})
		if err == nil {
			t.Error("Expected error")
		}
		repo.createErr = nil
	})

	t.Run("find error", func(t *testing.T) {
		repo.findErr = errors.Internal("test", nil)
		_, err := repo.FindByID(ctx, 1)
		if err == nil {
			t.Error("Expected error")
		}
		repo.findErr = nil
	})

	t.Run("update error", func(t *testing.T) {
		repo.seed("Test", "Desc", 100)
		repo.updateErr = errors.Internal("test", nil)
		err := repo.Update(ctx, &Item{Name: "test"})
		if err == nil {
			t.Error("Expected error")
		}
		repo.updateErr = nil
	})

	t.Run("delete error", func(t *testing.T) {
		repo.deleteErr = errors.Internal("test", nil)
		err := repo.Delete(ctx, 1)
		if err == nil {
			t.Error("Expected error")
		}
		repo.deleteErr = nil
	})
}
