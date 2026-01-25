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
	"github.com/voidmaindev/go-template/pkg/ptr"
)

// ================================
// Token Response Converters
// ================================

func toTokenResponse(t *user.TokenResponse) TokenResponse {
	return TokenResponse{
		AccessToken:  ptr.To(t.AccessToken),
		RefreshToken: ptr.To(t.RefreshToken),
		ExpiresAt:    ptr.To(time.Unix(t.ExpiresAt, 0)),
		User:         toUserResponsePtr(t.User),
	}
}

// ================================
// User Response Converters
// ================================

func toUserResponse(u *user.User) UserResponse {
	email := openapi_types.Email(u.Email)
	return UserResponse{
		Id:        ptr.To(int64(u.ID)),
		Email:     &email,
		Name:      ptr.To(u.Name),
		CreatedAt: ptr.To(u.CreatedAt),
		UpdatedAt: ptr.To(u.UpdatedAt),
	}
}

func toUserResponsePtr(u *user.UserResponse) *UserResponse {
	if u == nil {
		return nil
	}
	email := openapi_types.Email(u.Email)
	return &UserResponse{
		Id:        ptr.To(int64(u.ID)),
		Email:     &email,
		Name:      ptr.To(u.Name),
		CreatedAt: ptr.To(u.CreatedAt),
		UpdatedAt: ptr.To(u.UpdatedAt),
	}
}

func toUserListResponse(result *common.FilteredResult[user.User], params *filter.Params) UserListResponse {
	users := make([]UserResponse, len(result.Data))
	for i, u := range result.Data {
		users[i] = toUserResponse(&u)
	}

	totalPages := common.CalculateTotalPages(result.Total, params.Limit)

	return UserListResponse{
		Data:       &users,
		Total:      ptr.To(result.Total),
		Page:       ptr.To(params.Page),
		PageSize:   ptr.To(params.Limit),
		TotalPages: ptr.To(totalPages),
		HasMore:    ptr.To(params.Page < totalPages),
	}
}

// ================================
// Item Response Converters
// ================================

func toItemResponse(i *item.Item) ItemResponse {
	return ItemResponse{
		Id:          ptr.To(int64(i.ID)),
		Name:        ptr.To(i.Name),
		Description: ptr.To(i.Description),
		Price:       ptr.To(int(i.Price)),
		CreatedAt:   ptr.To(i.CreatedAt),
		UpdatedAt:   ptr.To(i.UpdatedAt),
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

	totalPages := common.CalculateTotalPages(result.Total, params.Limit)

	return ItemListResponse{
		Data:       &items,
		Total:      ptr.To(result.Total),
		Page:       ptr.To(params.Page),
		PageSize:   ptr.To(params.Limit),
		TotalPages: ptr.To(totalPages),
		HasMore:    ptr.To(params.Page < totalPages),
	}
}

// ================================
// Country Response Converters
// ================================

func toCountryResponse(c *country.Country) CountryResponse {
	return CountryResponse{
		Id:        ptr.To(int64(c.ID)),
		Name:      ptr.To(c.Name),
		Code:      ptr.To(c.Code),
		CreatedAt: ptr.To(c.CreatedAt),
		UpdatedAt: ptr.To(c.UpdatedAt),
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

	totalPages := common.CalculateTotalPages(result.Total, params.Limit)

	return CountryListResponse{
		Data:       &countries,
		Total:      ptr.To(result.Total),
		Page:       ptr.To(params.Page),
		PageSize:   ptr.To(params.Limit),
		TotalPages: ptr.To(totalPages),
		HasMore:    ptr.To(params.Page < totalPages),
	}
}

// ================================
// City Response Converters
// ================================

func toCityResponse(c *city.City) CityResponse {
	resp := CityResponse{
		Id:        ptr.To(int64(c.ID)),
		Name:      ptr.To(c.Name),
		CountryId: ptr.To(int64(c.CountryID)),
		CreatedAt: ptr.To(c.CreatedAt),
		UpdatedAt: ptr.To(c.UpdatedAt),
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

	totalPages := common.CalculateTotalPages(result.Total, params.Limit)

	return CityListResponse{
		Data:       &cities,
		Total:      ptr.To(result.Total),
		Page:       ptr.To(params.Page),
		PageSize:   ptr.To(params.Limit),
		TotalPages: ptr.To(totalPages),
		HasMore:    ptr.To(params.Page < totalPages),
	}
}

func toCityListFromPaginated(result *common.PaginatedResult[city.City]) CityListResponse {
	cities := make([]CityResponse, len(result.Data))
	for i, c := range result.Data {
		cities[i] = toCityResponse(&c)
	}

	return CityListResponse{
		Data:       &cities,
		Total:      ptr.To(result.Total),
		Page:       ptr.To(result.Page),
		PageSize:   ptr.To(result.PageSize),
		TotalPages: ptr.To(result.TotalPages),
		HasMore:    ptr.To(result.HasMore),
	}
}

// ================================
// Document Response Converters
// ================================

func toDocumentResponse(d *document.Document) DocumentResponse {
	resp := DocumentResponse{
		Id:           ptr.To(int64(d.ID)),
		Code:         ptr.To(d.Code),
		CityId:       ptr.To(int64(d.CityID)),
		DocumentDate: ptr.To(openapi_types.Date{Time: d.DocumentDate}),
		TotalAmount:  ptr.To(int(d.TotalAmount)),
		CreatedAt:    ptr.To(d.CreatedAt),
		UpdatedAt:    ptr.To(d.UpdatedAt),
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

	totalPages := common.CalculateTotalPages(result.Total, params.Limit)

	return DocumentListResponse{
		Data:       &docs,
		Total:      ptr.To(result.Total),
		Page:       ptr.To(params.Page),
		PageSize:   ptr.To(params.Limit),
		TotalPages: ptr.To(totalPages),
		HasMore:    ptr.To(params.Page < totalPages),
	}
}

func toDocumentItemResponse(di *document.DocumentItem) DocumentItemResponse {
	resp := DocumentItemResponse{
		Id:         ptr.To(int64(di.ID)),
		DocumentId: ptr.To(int64(di.DocumentID)),
		ItemId:     ptr.To(int64(di.ItemID)),
		Quantity:   ptr.To(di.Quantity),
		Price:      ptr.To(int(di.Price)),
		LineTotal:  ptr.To(int(di.GetLineTotal())),
	}

	// Include item if loaded
	if di.Item.ID != 0 {
		resp.Item = toItemResponsePtr(&di.Item)
	}

	return resp
}
