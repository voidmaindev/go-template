package api

import (
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/common/filter"
	"github.com/voidmaindev/go-template/internal/domain/city"
	"github.com/voidmaindev/go-template/internal/domain/country"
	"github.com/voidmaindev/go-template/internal/domain/document"
	"github.com/voidmaindev/go-template/internal/domain/item"
	"github.com/voidmaindev/go-template/internal/domain/user"
)

// ================================
// Token Response Converters
// ================================

func toTokenResponse(t *user.TokenResponse) TokenResponse {
	return TokenResponse{
		AccessToken:  ptr(t.AccessToken),
		RefreshToken: ptr(t.RefreshToken),
		ExpiresAt:    ptr(time.Unix(t.ExpiresAt, 0)),
		User:         toUserResponsePtr(t.User),
	}
}

// ================================
// User Response Converters
// ================================

func toUserResponse(u *user.User) UserResponse {
	role := User
	if u.Role == "admin" {
		role = Admin
	}
	email := openapi_types.Email(u.Email)
	return UserResponse{
		Id:        ptr(int64(u.ID)),
		Email:     &email,
		Name:      ptr(u.Name),
		Role:      &role,
		CreatedAt: ptr(u.CreatedAt),
		UpdatedAt: ptr(u.UpdatedAt),
	}
}

func toUserResponsePtr(u *user.UserResponse) *UserResponse {
	if u == nil {
		return nil
	}
	role := User
	if u.Role == "admin" {
		role = Admin
	}
	email := openapi_types.Email(u.Email)
	return &UserResponse{
		Id:        ptr(int64(u.ID)),
		Email:     &email,
		Name:      ptr(u.Name),
		Role:      &role,
		CreatedAt: ptr(u.CreatedAt),
		UpdatedAt: ptr(u.UpdatedAt),
	}
}

func toUserListResponse(result *common.FilteredResult[user.User], params *filter.Params) UserListResponse {
	users := make([]UserResponse, len(result.Data))
	for i, u := range result.Data {
		users[i] = toUserResponse(&u)
	}

	totalPages := int(result.Total) / params.Limit
	if int(result.Total)%params.Limit > 0 {
		totalPages++
	}

	return UserListResponse{
		Data:       &users,
		Total:      ptr(result.Total),
		Page:       ptr(params.Page),
		PageSize:   ptr(params.Limit),
		TotalPages: ptr(totalPages),
		HasMore:    ptr(params.Page < totalPages),
	}
}

// ================================
// Item Response Converters
// ================================

func toItemResponse(i *item.Item) ItemResponse {
	return ItemResponse{
		Id:          ptr(int64(i.ID)),
		Name:        ptr(i.Name),
		Description: ptr(i.Description),
		Price:       ptr(int(i.Price)),
		CreatedAt:   ptr(i.CreatedAt),
		UpdatedAt:   ptr(i.UpdatedAt),
	}
}

func toItemResponsePtr(i *item.Item) *ItemResponse {
	if i == nil || i.ID == 0 {
		return nil
	}
	r := toItemResponse(i)
	return &r
}

func toItemListResponse(result *common.FilteredResult[item.Item], params *filter.Params) ItemListResponse {
	items := make([]ItemResponse, len(result.Data))
	for i, it := range result.Data {
		items[i] = toItemResponse(&it)
	}

	totalPages := int(result.Total) / params.Limit
	if int(result.Total)%params.Limit > 0 {
		totalPages++
	}

	return ItemListResponse{
		Data:       &items,
		Total:      ptr(result.Total),
		Page:       ptr(params.Page),
		PageSize:   ptr(params.Limit),
		TotalPages: ptr(totalPages),
		HasMore:    ptr(params.Page < totalPages),
	}
}

// ================================
// Country Response Converters
// ================================

func toCountryResponse(c *country.Country) CountryResponse {
	return CountryResponse{
		Id:        ptr(int64(c.ID)),
		Name:      ptr(c.Name),
		Code:      ptr(c.Code),
		CreatedAt: ptr(c.CreatedAt),
		UpdatedAt: ptr(c.UpdatedAt),
	}
}

func toCountryResponsePtr(c *country.Country) *CountryResponse {
	if c == nil || c.ID == 0 {
		return nil
	}
	r := toCountryResponse(c)
	return &r
}

func toCountryListResponse(result *common.FilteredResult[country.Country], params *filter.Params) CountryListResponse {
	countries := make([]CountryResponse, len(result.Data))
	for i, c := range result.Data {
		countries[i] = toCountryResponse(&c)
	}

	totalPages := int(result.Total) / params.Limit
	if int(result.Total)%params.Limit > 0 {
		totalPages++
	}

	return CountryListResponse{
		Data:       &countries,
		Total:      ptr(result.Total),
		Page:       ptr(params.Page),
		PageSize:   ptr(params.Limit),
		TotalPages: ptr(totalPages),
		HasMore:    ptr(params.Page < totalPages),
	}
}

// ================================
// City Response Converters
// ================================

func toCityResponse(c *city.City) CityResponse {
	resp := CityResponse{
		Id:        ptr(int64(c.ID)),
		Name:      ptr(c.Name),
		CountryId: ptr(int64(c.CountryID)),
		CreatedAt: ptr(c.CreatedAt),
		UpdatedAt: ptr(c.UpdatedAt),
	}

	// Include country if loaded
	if c.Country.ID != 0 {
		resp.Country = toCountryResponsePtr(&c.Country)
	}

	return resp
}

func toCityResponsePtr(c *city.City) *CityResponse {
	if c == nil || c.ID == 0 {
		return nil
	}
	r := toCityResponse(c)
	return &r
}

func toCityListResponse(result *common.FilteredResult[city.City], params *filter.Params) CityListResponse {
	cities := make([]CityResponse, len(result.Data))
	for i, c := range result.Data {
		cities[i] = toCityResponse(&c)
	}

	totalPages := int(result.Total) / params.Limit
	if int(result.Total)%params.Limit > 0 {
		totalPages++
	}

	return CityListResponse{
		Data:       &cities,
		Total:      ptr(result.Total),
		Page:       ptr(params.Page),
		PageSize:   ptr(params.Limit),
		TotalPages: ptr(totalPages),
		HasMore:    ptr(params.Page < totalPages),
	}
}

func toCityListFromPaginated(result *common.PaginatedResult[city.City]) CityListResponse {
	cities := make([]CityResponse, len(result.Data))
	for i, c := range result.Data {
		cities[i] = toCityResponse(&c)
	}

	return CityListResponse{
		Data:       &cities,
		Total:      ptr(result.Total),
		Page:       ptr(result.Page),
		PageSize:   ptr(result.PageSize),
		TotalPages: ptr(result.TotalPages),
		HasMore:    ptr(result.HasMore),
	}
}

// ================================
// Document Response Converters
// ================================

func toDocumentResponse(d *document.Document) DocumentResponse {
	resp := DocumentResponse{
		Id:           ptr(int64(d.ID)),
		Code:         ptr(d.Code),
		CityId:       ptr(int64(d.CityID)),
		DocumentDate: ptr(openapi_types.Date{Time: d.DocumentDate}),
		TotalAmount:  ptr(int(d.TotalAmount)),
		CreatedAt:    ptr(d.CreatedAt),
		UpdatedAt:    ptr(d.UpdatedAt),
	}

	// Include city if loaded
	if d.City.ID != 0 {
		resp.City = toCityResponsePtr(&d.City)
	}

	// Include items if loaded
	if len(d.Items) > 0 {
		items := make([]DocumentItemResponse, len(d.Items))
		for i, item := range d.Items {
			items[i] = toDocumentItemResponse(&item)
		}
		resp.Items = &items
	}

	return resp
}

func toDocumentListResponse(result *common.FilteredResult[document.Document], params *filter.Params) DocumentListResponse {
	docs := make([]DocumentResponse, len(result.Data))
	for i, d := range result.Data {
		docs[i] = toDocumentResponse(&d)
	}

	totalPages := int(result.Total) / params.Limit
	if int(result.Total)%params.Limit > 0 {
		totalPages++
	}

	return DocumentListResponse{
		Data:       &docs,
		Total:      ptr(result.Total),
		Page:       ptr(params.Page),
		PageSize:   ptr(params.Limit),
		TotalPages: ptr(totalPages),
		HasMore:    ptr(params.Page < totalPages),
	}
}

func toDocumentItemResponse(di *document.DocumentItem) DocumentItemResponse {
	resp := DocumentItemResponse{
		Id:         ptr(int64(di.ID)),
		DocumentId: ptr(int64(di.DocumentID)),
		ItemId:     ptr(int64(di.ItemID)),
		Quantity:   ptr(di.Quantity),
		Price:      ptr(int(di.Price)),
		LineTotal:  ptr(int(di.GetLineTotal())),
	}

	// Include item if loaded
	if di.Item.ID != 0 {
		resp.Item = toItemResponsePtr(&di.Item)
	}

	return resp
}
