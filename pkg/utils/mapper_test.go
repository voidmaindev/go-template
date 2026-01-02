package utils

import (
	"testing"
)

type TestModel struct {
	ID          uint
	Name        string
	Email       string
	Description string
	Price       int64
	Active      bool
}

type TestUpdateDTO struct {
	Name        string
	Email       string
	Description string
	Price       int64
}

func TestUpdateModel_PartialUpdate(t *testing.T) {
	model := &TestModel{
		ID:          1,
		Name:        "Original Name",
		Email:       "original@example.com",
		Description: "Original Description",
		Price:       1000,
		Active:      true,
	}

	dto := &TestUpdateDTO{
		Name: "Updated Name",
		// Email, Description, Price left as zero values
	}

	err := UpdateModel(model, dto)
	if err != nil {
		t.Fatalf("UpdateModel() error = %v", err)
	}

	// Name should be updated
	if model.Name != "Updated Name" {
		t.Errorf("Name = %q, want %q", model.Name, "Updated Name")
	}

	// Other fields should remain unchanged (IgnoreEmpty)
	if model.Email != "original@example.com" {
		t.Errorf("Email = %q, want %q", model.Email, "original@example.com")
	}
	if model.Description != "Original Description" {
		t.Errorf("Description = %q, want %q", model.Description, "Original Description")
	}
	if model.Price != 1000 {
		t.Errorf("Price = %d, want %d", model.Price, 1000)
	}

	// ID should not be affected
	if model.ID != 1 {
		t.Errorf("ID = %d, want %d", model.ID, 1)
	}
}

func TestUpdateModel_FullUpdate(t *testing.T) {
	model := &TestModel{
		ID:          1,
		Name:        "Original Name",
		Email:       "original@example.com",
		Description: "Original Description",
		Price:       1000,
	}

	dto := &TestUpdateDTO{
		Name:        "New Name",
		Email:       "new@example.com",
		Description: "New Description",
		Price:       2000,
	}

	err := UpdateModel(model, dto)
	if err != nil {
		t.Fatalf("UpdateModel() error = %v", err)
	}

	if model.Name != "New Name" {
		t.Errorf("Name = %q, want %q", model.Name, "New Name")
	}
	if model.Email != "new@example.com" {
		t.Errorf("Email = %q, want %q", model.Email, "new@example.com")
	}
	if model.Description != "New Description" {
		t.Errorf("Description = %q, want %q", model.Description, "New Description")
	}
	if model.Price != 2000 {
		t.Errorf("Price = %d, want %d", model.Price, 2000)
	}
}

func TestUpdateModel_EmptyDTO(t *testing.T) {
	model := &TestModel{
		ID:          1,
		Name:        "Original Name",
		Email:       "original@example.com",
		Description: "Original Description",
		Price:       1000,
	}

	dto := &TestUpdateDTO{} // All zero values

	err := UpdateModel(model, dto)
	if err != nil {
		t.Fatalf("UpdateModel() error = %v", err)
	}

	// All fields should remain unchanged
	if model.Name != "Original Name" {
		t.Errorf("Name = %q, want %q", model.Name, "Original Name")
	}
	if model.Email != "original@example.com" {
		t.Errorf("Email = %q, want %q", model.Email, "original@example.com")
	}
}

func TestUpdateModel_NilDestination(t *testing.T) {
	dto := &TestUpdateDTO{Name: "Test"}

	// This might panic or return error depending on copier behavior
	// Just ensure it doesn't crash unexpectedly
	defer func() {
		if r := recover(); r != nil {
			// Panic is acceptable for nil destination
			t.Logf("Recovered from panic (expected for nil dest): %v", r)
		}
	}()

	var model *TestModel = nil
	_ = UpdateModel(model, dto)
}

type TestModelWithPointers struct {
	ID   uint
	Name *string
}

type TestDTOWithPointers struct {
	Name *string
}

func TestUpdateModel_WithPointers(t *testing.T) {
	originalName := "Original"
	model := &TestModelWithPointers{
		ID:   1,
		Name: &originalName,
	}

	newName := "Updated"
	dto := &TestDTOWithPointers{
		Name: &newName,
	}

	err := UpdateModel(model, dto)
	if err != nil {
		t.Fatalf("UpdateModel() error = %v", err)
	}

	if model.Name == nil {
		t.Fatal("Name should not be nil")
	}
	if *model.Name != "Updated" {
		t.Errorf("Name = %q, want %q", *model.Name, "Updated")
	}
}

func TestUpdateModel_NilPointerInDTO(t *testing.T) {
	originalName := "Original"
	model := &TestModelWithPointers{
		ID:   1,
		Name: &originalName,
	}

	dto := &TestDTOWithPointers{
		Name: nil, // nil pointer should not overwrite
	}

	err := UpdateModel(model, dto)
	if err != nil {
		t.Fatalf("UpdateModel() error = %v", err)
	}

	// Original value should be preserved
	if model.Name == nil {
		t.Fatal("Name should not be nil after update with nil")
	}
	if *model.Name != "Original" {
		t.Errorf("Name = %q, want %q", *model.Name, "Original")
	}
}
