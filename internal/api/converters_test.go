package api

import (
	"testing"
	"time"

	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/common/filter"
	"github.com/voidmaindev/go-template/internal/domain/city"
	"github.com/voidmaindev/go-template/internal/domain/country"
	"github.com/voidmaindev/go-template/internal/domain/document"
	"github.com/voidmaindev/go-template/internal/domain/item"
	"github.com/voidmaindev/go-template/internal/domain/user"
)

func TestToUserResponse(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		user     *user.User
		wantRole UserResponseRole
	}{
		{
			name: "user role",
			user: &user.User{
				BaseModel: common.BaseModel{ID: 1, CreatedAt: now, UpdatedAt: now},
				Email:     "test@example.com",
				Name:      "Test User",
				Role:      "user",
			},
			wantRole: UserResponseRoleUser,
		},
		{
			name: "admin role",
			user: &user.User{
				BaseModel: common.BaseModel{ID: 2, CreatedAt: now, UpdatedAt: now},
				Email:     "admin@example.com",
				Name:      "Admin User",
				Role:      "admin",
			},
			wantRole: UserResponseRoleAdmin,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := toUserResponse(tt.user)

			if *resp.Id != int64(tt.user.ID) {
				t.Errorf("ID = %d, want %d", *resp.Id, tt.user.ID)
			}
			if string(*resp.Email) != tt.user.Email {
				t.Errorf("Email = %s, want %s", string(*resp.Email), tt.user.Email)
			}
			if *resp.Name != tt.user.Name {
				t.Errorf("Name = %s, want %s", *resp.Name, tt.user.Name)
			}
			if *resp.Role != tt.wantRole {
				t.Errorf("Role = %v, want %v", *resp.Role, tt.wantRole)
			}
		})
	}
}

func TestToUserListResponse(t *testing.T) {
	now := time.Now()
	users := []user.User{
		{BaseModel: common.BaseModel{ID: 1, CreatedAt: now, UpdatedAt: now}, Email: "u1@test.com", Name: "User 1", Role: "user"},
		{BaseModel: common.BaseModel{ID: 2, CreatedAt: now, UpdatedAt: now}, Email: "u2@test.com", Name: "User 2", Role: "admin"},
	}

	result := &common.FilteredResult[user.User]{
		Data:  users,
		Total: 25,
	}
	params := &filter.Params{Page: 1, Limit: 10}

	resp := toUserListResponse(result, params)

	if len(*resp.Data) != 2 {
		t.Errorf("Data length = %d, want 2", len(*resp.Data))
	}
	if *resp.Total != 25 {
		t.Errorf("Total = %d, want 25", *resp.Total)
	}
	if *resp.Page != 1 {
		t.Errorf("Page = %d, want 1", *resp.Page)
	}
	if *resp.PageSize != 10 {
		t.Errorf("PageSize = %d, want 10", *resp.PageSize)
	}
	if *resp.TotalPages != 3 {
		t.Errorf("TotalPages = %d, want 3", *resp.TotalPages)
	}
	if *resp.HasMore != true {
		t.Errorf("HasMore = %v, want true", *resp.HasMore)
	}
}

func TestToItemResponse(t *testing.T) {
	now := time.Now()
	i := &item.Item{
		BaseModel:   common.BaseModel{ID: 1, CreatedAt: now, UpdatedAt: now},
		Name:        "Test Item",
		Description: "A test item",
		Price:       1999,
	}

	resp := toItemResponse(i)

	if *resp.Id != 1 {
		t.Errorf("ID = %d, want 1", *resp.Id)
	}
	if *resp.Name != "Test Item" {
		t.Errorf("Name = %s, want Test Item", *resp.Name)
	}
	if *resp.Description != "A test item" {
		t.Errorf("Description = %s, want A test item", *resp.Description)
	}
	if *resp.Price != 1999 {
		t.Errorf("Price = %d, want 1999", *resp.Price)
	}
}

func TestToItemResponsePtr(t *testing.T) {
	now := time.Now()

	t.Run("nil item", func(t *testing.T) {
		resp := toItemResponsePtr(nil)
		if resp != nil {
			t.Error("expected nil response for nil item")
		}
	})

	t.Run("zero ID item", func(t *testing.T) {
		i := &item.Item{BaseModel: common.BaseModel{ID: 0}}
		resp := toItemResponsePtr(i)
		if resp != nil {
			t.Error("expected nil response for zero ID item")
		}
	})

	t.Run("valid item", func(t *testing.T) {
		i := &item.Item{
			BaseModel: common.BaseModel{ID: 1, CreatedAt: now, UpdatedAt: now},
			Name:      "Test",
		}
		resp := toItemResponsePtr(i)
		if resp == nil {
			t.Error("expected non-nil response for valid item")
		}
	})
}

func TestToCountryResponse(t *testing.T) {
	now := time.Now()
	c := &country.Country{
		BaseModel: common.BaseModel{ID: 1, CreatedAt: now, UpdatedAt: now},
		Name:      "Germany",
		Code:      "DEU",
	}

	resp := toCountryResponse(c)

	if *resp.Id != 1 {
		t.Errorf("ID = %d, want 1", *resp.Id)
	}
	if *resp.Name != "Germany" {
		t.Errorf("Name = %s, want Germany", *resp.Name)
	}
	if *resp.Code != "DEU" {
		t.Errorf("Code = %s, want DEU", *resp.Code)
	}
}

func TestToCityResponse(t *testing.T) {
	now := time.Now()

	t.Run("without country", func(t *testing.T) {
		c := &city.City{
			BaseModel: common.BaseModel{ID: 1, CreatedAt: now, UpdatedAt: now},
			Name:      "Berlin",
			CountryID: 1,
		}

		resp := toCityResponse(c)

		if *resp.Id != 1 {
			t.Errorf("ID = %d, want 1", *resp.Id)
		}
		if *resp.Name != "Berlin" {
			t.Errorf("Name = %s, want Berlin", *resp.Name)
		}
		if *resp.CountryId != 1 {
			t.Errorf("CountryId = %d, want 1", *resp.CountryId)
		}
		if resp.Country != nil {
			t.Error("expected nil Country when not loaded")
		}
	})

	t.Run("with country", func(t *testing.T) {
		c := &city.City{
			BaseModel: common.BaseModel{ID: 1, CreatedAt: now, UpdatedAt: now},
			Name:      "Berlin",
			CountryID: 1,
			Country: country.Country{
				BaseModel: common.BaseModel{ID: 1, CreatedAt: now, UpdatedAt: now},
				Name:      "Germany",
				Code:      "DEU",
			},
		}

		resp := toCityResponse(c)

		if resp.Country == nil {
			t.Fatal("expected Country to be loaded")
		}
		if *resp.Country.Name != "Germany" {
			t.Errorf("Country.Name = %s, want Germany", *resp.Country.Name)
		}
	})
}

func TestToDocumentResponse(t *testing.T) {
	now := time.Now()
	docDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	t.Run("without relations", func(t *testing.T) {
		d := &document.Document{
			BaseModel:    common.BaseModel{ID: 1, CreatedAt: now, UpdatedAt: now},
			Code:         "DOC-001",
			CityID:       1,
			DocumentDate: docDate,
			TotalAmount:  5000,
		}

		resp := toDocumentResponse(d)

		if *resp.Id != 1 {
			t.Errorf("ID = %d, want 1", *resp.Id)
		}
		if *resp.Code != "DOC-001" {
			t.Errorf("Code = %s, want DOC-001", *resp.Code)
		}
		if *resp.TotalAmount != 5000 {
			t.Errorf("TotalAmount = %d, want 5000", *resp.TotalAmount)
		}
		if resp.City != nil {
			t.Error("expected nil City when not loaded")
		}
		if resp.Items != nil {
			t.Error("expected nil Items when not loaded")
		}
	})

	t.Run("with city and items", func(t *testing.T) {
		d := &document.Document{
			BaseModel:    common.BaseModel{ID: 1, CreatedAt: now, UpdatedAt: now},
			Code:         "DOC-001",
			CityID:       1,
			DocumentDate: docDate,
			TotalAmount:  5000,
			City: city.City{
				BaseModel: common.BaseModel{ID: 1, CreatedAt: now, UpdatedAt: now},
				Name:      "Berlin",
			},
			Items: []document.DocumentItem{
				{
					BaseModel:  common.BaseModel{ID: 1},
					DocumentID: 1,
					ItemID:     1,
					Quantity:   2,
					Price:      1000,
				},
			},
		}

		resp := toDocumentResponse(d)

		if resp.City == nil {
			t.Fatal("expected City to be loaded")
		}
		if *resp.City.Name != "Berlin" {
			t.Errorf("City.Name = %s, want Berlin", *resp.City.Name)
		}
		if resp.Items == nil || len(*resp.Items) != 1 {
			t.Fatal("expected 1 item")
		}
	})
}

func TestToDocumentItemResponse(t *testing.T) {
	now := time.Now()

	t.Run("without item", func(t *testing.T) {
		di := &document.DocumentItem{
			BaseModel:  common.BaseModel{ID: 1},
			DocumentID: 1,
			ItemID:     10,
			Quantity:   3,
			Price:      500,
		}

		resp := toDocumentItemResponse(di)

		if *resp.Id != 1 {
			t.Errorf("ID = %d, want 1", *resp.Id)
		}
		if *resp.DocumentId != 1 {
			t.Errorf("DocumentId = %d, want 1", *resp.DocumentId)
		}
		if *resp.ItemId != 10 {
			t.Errorf("ItemId = %d, want 10", *resp.ItemId)
		}
		if *resp.Quantity != 3 {
			t.Errorf("Quantity = %d, want 3", *resp.Quantity)
		}
		if *resp.Price != 500 {
			t.Errorf("Price = %d, want 500", *resp.Price)
		}
		if *resp.LineTotal != 1500 {
			t.Errorf("LineTotal = %d, want 1500", *resp.LineTotal)
		}
		if resp.Item != nil {
			t.Error("expected nil Item when not loaded")
		}
	})

	t.Run("with item", func(t *testing.T) {
		di := &document.DocumentItem{
			BaseModel:  common.BaseModel{ID: 1},
			DocumentID: 1,
			ItemID:     10,
			Quantity:   3,
			Price:      500,
			Item: item.Item{
				BaseModel: common.BaseModel{ID: 10, CreatedAt: now, UpdatedAt: now},
				Name:      "Widget",
			},
		}

		resp := toDocumentItemResponse(di)

		if resp.Item == nil {
			t.Fatal("expected Item to be loaded")
		}
		if *resp.Item.Name != "Widget" {
			t.Errorf("Item.Name = %s, want Widget", *resp.Item.Name)
		}
	})
}

func TestPaginationCalculation(t *testing.T) {
	tests := []struct {
		name           string
		total          int64
		limit          int
		page           int
		wantTotalPages int
		wantHasMore    bool
	}{
		{"exact division", 20, 10, 1, 2, true},
		{"remainder", 25, 10, 1, 3, true},
		{"last page", 25, 10, 3, 3, false},
		{"single page", 5, 10, 1, 1, false},
		{"empty", 0, 10, 1, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &common.FilteredResult[user.User]{
				Data:  make([]user.User, 0),
				Total: tt.total,
			}
			params := &filter.Params{Page: tt.page, Limit: tt.limit}

			resp := toUserListResponse(result, params)

			if *resp.TotalPages != tt.wantTotalPages {
				t.Errorf("TotalPages = %d, want %d", *resp.TotalPages, tt.wantTotalPages)
			}
			if *resp.HasMore != tt.wantHasMore {
				t.Errorf("HasMore = %v, want %v", *resp.HasMore, tt.wantHasMore)
			}
		})
	}
}
