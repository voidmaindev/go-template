// Package api provides the OpenAPI server implementation.
package api

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/common"
	"github.com/voidmaindev/go-template/internal/common/filter"
	"github.com/voidmaindev/go-template/internal/container"
	"github.com/voidmaindev/go-template/internal/domain/city"
	"github.com/voidmaindev/go-template/internal/domain/country"
	"github.com/voidmaindev/go-template/internal/domain/document"
	"github.com/voidmaindev/go-template/internal/domain/item"
	"github.com/voidmaindev/go-template/internal/domain/user"
	"github.com/voidmaindev/go-template/internal/middleware"
)

// Server implements the StrictServerInterface.
type Server struct {
	container   *container.Container
	userService user.Service
	itemService item.Service
	countryService country.Service
	cityService city.Service
	documentService document.Service
}

// NewServer creates a new API server with services from the container.
func NewServer(c *container.Container) *Server {
	return &Server{
		container:       c,
		userService:     container.MustGetTyped[user.Service](c, user.ServiceKey),
		itemService:     container.MustGetTyped[item.Service](c, item.ServiceKey),
		countryService:  container.MustGetTyped[country.Service](c, country.ServiceKey),
		cityService:     container.MustGetTyped[city.Service](c, city.ServiceKey),
		documentService: container.MustGetTyped[document.Service](c, document.ServiceKey),
	}
}

// ================================
// Auth Endpoints
// ================================

// Register implements StrictServerInterface.
func (s *Server) Register(ctx context.Context, request RegisterRequestObject) (RegisterResponseObject, error) {
	req := &user.RegisterRequest{
		Email:    string(request.Body.Email),
		Password: request.Body.Password,
		Name:     request.Body.Name,
	}

	response, err := s.userService.Register(ctx, req)
	if err != nil {
		if errors.Is(err, user.ErrEmailAlreadyExists) {
			return Register409JSONResponse{
				Error:   ptr("conflict"),
				Message: ptr("email already exists"),
			}, nil
		}
		return Register400JSONResponse{
			Error:   ptr("bad_request"),
			Message: ptr("registration failed"),
		}, nil
	}

	return Register201JSONResponse(toTokenResponse(response)), nil
}

// Login implements StrictServerInterface.
func (s *Server) Login(ctx context.Context, request LoginRequestObject) (LoginResponseObject, error) {
	req := &user.LoginRequest{
		Email:    string(request.Body.Email),
		Password: request.Body.Password,
	}

	response, err := s.userService.Login(ctx, req)
	if err != nil {
		if errors.Is(err, common.ErrInvalidCredentials) {
			return Login401JSONResponse{
				Error:   ptr("unauthorized"),
				Message: ptr("invalid email or password"),
			}, nil
		}
		return Login401JSONResponse{
			Error:   ptr("unauthorized"),
			Message: ptr("login failed"),
		}, nil
	}

	return Login200JSONResponse(toTokenResponse(response)), nil
}

// Logout implements StrictServerInterface.
func (s *Server) Logout(ctx context.Context, request LogoutRequestObject) (LogoutResponseObject, error) {
	// Get token from fiber context - need to extract from ctx
	fiberCtx := getFiberContext(ctx)
	if fiberCtx == nil {
		return Logout401JSONResponse{
			Error:   ptr("unauthorized"),
			Message: ptr("no context"),
		}, nil
	}

	token := middleware.GetTokenFromContext(fiberCtx)
	if token == "" {
		return Logout401JSONResponse{
			Error:   ptr("unauthorized"),
			Message: ptr("no token provided"),
		}, nil
	}

	if err := s.userService.Logout(ctx, token); err != nil {
		return Logout401JSONResponse{
			Error:   ptr("unauthorized"),
			Message: ptr("logout failed"),
		}, nil
	}

	return Logout200Response{}, nil
}

// RefreshToken implements StrictServerInterface.
func (s *Server) RefreshToken(ctx context.Context, request RefreshTokenRequestObject) (RefreshTokenResponseObject, error) {
	response, err := s.userService.RefreshToken(ctx, request.Body.RefreshToken)
	if err != nil {
		if errors.Is(err, common.ErrTokenInvalid) || errors.Is(err, common.ErrTokenBlacklisted) {
			return RefreshToken401JSONResponse{
				Error:   ptr("unauthorized"),
				Message: ptr("invalid or expired refresh token"),
			}, nil
		}
		return RefreshToken401JSONResponse{
			Error:   ptr("unauthorized"),
			Message: ptr("token refresh failed"),
		}, nil
	}

	return RefreshToken200JSONResponse(toTokenResponse(response)), nil
}

// ================================
// User Endpoints
// ================================

// GetCurrentUser implements StrictServerInterface.
func (s *Server) GetCurrentUser(ctx context.Context, request GetCurrentUserRequestObject) (GetCurrentUserResponseObject, error) {
	fiberCtx := getFiberContext(ctx)
	if fiberCtx == nil {
		return GetCurrentUser401JSONResponse{}, nil
	}

	userID, ok := middleware.GetUserIDFromContext(fiberCtx)
	if !ok {
		return GetCurrentUser401JSONResponse{}, nil
	}

	u, err := s.userService.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return GetCurrentUser401JSONResponse{}, nil
		}
		return GetCurrentUser401JSONResponse{}, nil
	}

	return GetCurrentUser200JSONResponse(toUserResponse(u)), nil
}

// UpdateCurrentUser implements StrictServerInterface.
func (s *Server) UpdateCurrentUser(ctx context.Context, request UpdateCurrentUserRequestObject) (UpdateCurrentUserResponseObject, error) {
	fiberCtx := getFiberContext(ctx)
	if fiberCtx == nil {
		return UpdateCurrentUser401JSONResponse{}, nil
	}

	userID, ok := middleware.GetUserIDFromContext(fiberCtx)
	if !ok {
		return UpdateCurrentUser401JSONResponse{}, nil
	}

	req := &user.UpdateUserRequest{
		Name: request.Body.Name,
	}

	u, err := s.userService.Update(ctx, userID, req)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return UpdateCurrentUser401JSONResponse{}, nil
		}
		return UpdateCurrentUser400JSONResponse{
			Error:   ptr("bad_request"),
			Message: ptr("update failed"),
		}, nil
	}

	return UpdateCurrentUser200JSONResponse(toUserResponse(u)), nil
}

// ChangePassword implements StrictServerInterface.
func (s *Server) ChangePassword(ctx context.Context, request ChangePasswordRequestObject) (ChangePasswordResponseObject, error) {
	// Add constant-time delay to prevent timing attacks
	start := time.Now()
	defer func() {
		elapsed := time.Since(start)
		const minResponseTime = 200 * time.Millisecond
		if elapsed < minResponseTime {
			time.Sleep(minResponseTime - elapsed)
		}
	}()

	fiberCtx := getFiberContext(ctx)
	if fiberCtx == nil {
		return ChangePassword401JSONResponse{}, nil
	}

	userID, ok := middleware.GetUserIDFromContext(fiberCtx)
	if !ok {
		return ChangePassword401JSONResponse{}, nil
	}

	req := &user.ChangePasswordRequest{
		CurrentPassword: request.Body.CurrentPassword,
		NewPassword:     request.Body.NewPassword,
	}

	if err := s.userService.ChangePassword(ctx, userID, req); err != nil {
		switch {
		case errors.Is(err, user.ErrUserNotFound), errors.Is(err, user.ErrInvalidPassword):
			return ChangePassword400JSONResponse{
				Error:   ptr("bad_request"),
				Message: ptr("current password is incorrect"),
			}, nil
		case errors.Is(err, user.ErrSamePassword):
			return ChangePassword400JSONResponse{
				Error:   ptr("bad_request"),
				Message: ptr("new password must be different"),
			}, nil
		default:
			return ChangePassword400JSONResponse{
				Error:   ptr("bad_request"),
				Message: ptr("password change failed"),
			}, nil
		}
	}

	return ChangePassword200Response{}, nil
}

// ListUsers implements StrictServerInterface.
func (s *Server) ListUsers(ctx context.Context, request ListUsersRequestObject) (ListUsersResponseObject, error) {
	fiberCtx := getFiberContext(ctx)
	if fiberCtx == nil {
		return ListUsers401JSONResponse{}, nil
	}

	if !middleware.IsAdmin(fiberCtx) {
		return ListUsers403JSONResponse{
			Error:   ptr("forbidden"),
			Message: ptr("admin access required"),
		}, nil
	}

	// Parse all query parameters for dynamic filtering
	params := filter.ParseFromQuery(fiberCtx)

	result, err := s.userService.ListFiltered(ctx, params)
	if err != nil {
		return ListUsers401JSONResponse{}, nil
	}

	return ListUsers200JSONResponse(toUserListResponse(result, params)), nil
}

// GetUser implements StrictServerInterface.
func (s *Server) GetUser(ctx context.Context, request GetUserRequestObject) (GetUserResponseObject, error) {
	fiberCtx := getFiberContext(ctx)
	if fiberCtx == nil {
		return GetUser401JSONResponse{}, nil
	}

	currentUserID, ok := middleware.GetUserIDFromContext(fiberCtx)
	if !ok {
		return GetUser401JSONResponse{}, nil
	}

	// Authorization: only allow self-view or admin
	if uint(request.Id) != currentUserID && !middleware.IsAdmin(fiberCtx) {
		return GetUser401JSONResponse{
			Error:   ptr("forbidden"),
			Message: ptr("cannot view other users"),
		}, nil
	}

	u, err := s.userService.GetByID(ctx, uint(request.Id))
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return GetUser404JSONResponse{
				Error:   ptr("not_found"),
				Message: ptr("user not found"),
			}, nil
		}
		return GetUser401JSONResponse{}, nil
	}

	return GetUser200JSONResponse(toUserResponse(u)), nil
}

// DeleteUser implements StrictServerInterface.
func (s *Server) DeleteUser(ctx context.Context, request DeleteUserRequestObject) (DeleteUserResponseObject, error) {
	fiberCtx := getFiberContext(ctx)
	if fiberCtx == nil {
		return DeleteUser401JSONResponse{}, nil
	}

	currentUserID, ok := middleware.GetUserIDFromContext(fiberCtx)
	if !ok {
		return DeleteUser401JSONResponse{}, nil
	}

	// Authorization: only allow self-delete or admin
	if uint(request.Id) != currentUserID && !middleware.IsAdmin(fiberCtx) {
		return DeleteUser403JSONResponse{
			Error:   ptr("forbidden"),
			Message: ptr("cannot delete other users"),
		}, nil
	}

	if err := s.userService.Delete(ctx, uint(request.Id)); err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return DeleteUser404JSONResponse{
				Error:   ptr("not_found"),
				Message: ptr("user not found"),
			}, nil
		}
		return DeleteUser401JSONResponse{}, nil
	}

	return DeleteUser204Response{}, nil
}

// ================================
// Item Endpoints
// ================================

// ListItems implements StrictServerInterface.
func (s *Server) ListItems(ctx context.Context, request ListItemsRequestObject) (ListItemsResponseObject, error) {
	fiberCtx := getFiberContext(ctx)
	if fiberCtx == nil {
		return ListItems401JSONResponse{}, nil
	}

	// Parse all query parameters for dynamic filtering
	params := filter.ParseFromQuery(fiberCtx)

	result, err := s.itemService.ListFiltered(ctx, params)
	if err != nil {
		return ListItems401JSONResponse{}, nil
	}

	return ListItems200JSONResponse(toItemListResponse(result, params)), nil
}

// CreateItem implements StrictServerInterface.
func (s *Server) CreateItem(ctx context.Context, request CreateItemRequestObject) (CreateItemResponseObject, error) {
	fiberCtx := getFiberContext(ctx)
	if fiberCtx != nil && !middleware.IsAdmin(fiberCtx) {
		return CreateItem403JSONResponse{
			Error:   ptr("forbidden"),
			Message: ptr("admin access required"),
		}, nil
	}

	req := &item.CreateItemRequest{
		Name:        request.Body.Name,
		Description: ptrToString(request.Body.Description),
		Price:       int64(request.Body.Price),
	}

	i, err := s.itemService.Create(ctx, req)
	if err != nil {
		return CreateItem400JSONResponse{
			Error:   ptr("bad_request"),
			Message: ptr("item creation failed"),
		}, nil
	}

	return CreateItem201JSONResponse(toItemResponse(i)), nil
}

// GetItem implements StrictServerInterface.
func (s *Server) GetItem(ctx context.Context, request GetItemRequestObject) (GetItemResponseObject, error) {
	i, err := s.itemService.GetByID(ctx, uint(request.Id))
	if err != nil {
		if errors.Is(err, item.ErrItemNotFound) {
			return GetItem404JSONResponse{
				Error:   ptr("not_found"),
				Message: ptr("item not found"),
			}, nil
		}
		return GetItem401JSONResponse{}, nil
	}

	return GetItem200JSONResponse(toItemResponse(i)), nil
}

// UpdateItem implements StrictServerInterface.
func (s *Server) UpdateItem(ctx context.Context, request UpdateItemRequestObject) (UpdateItemResponseObject, error) {
	fiberCtx := getFiberContext(ctx)
	if fiberCtx != nil && !middleware.IsAdmin(fiberCtx) {
		return UpdateItem403JSONResponse{
			Error:   ptr("forbidden"),
			Message: ptr("admin access required"),
		}, nil
	}

	req := &item.UpdateItemRequest{}
	if request.Body.Name != nil {
		req.Name = request.Body.Name
	}
	if request.Body.Description != nil {
		req.Description = request.Body.Description
	}
	if request.Body.Price != nil {
		p := int64(*request.Body.Price)
		req.Price = &p
	}

	i, err := s.itemService.Update(ctx, uint(request.Id), req)
	if err != nil {
		if errors.Is(err, item.ErrItemNotFound) {
			return UpdateItem404JSONResponse{
				Error:   ptr("not_found"),
				Message: ptr("item not found"),
			}, nil
		}
		return UpdateItem400JSONResponse{
			Error:   ptr("bad_request"),
			Message: ptr("update failed"),
		}, nil
	}

	return UpdateItem200JSONResponse(toItemResponse(i)), nil
}

// DeleteItem implements StrictServerInterface.
func (s *Server) DeleteItem(ctx context.Context, request DeleteItemRequestObject) (DeleteItemResponseObject, error) {
	fiberCtx := getFiberContext(ctx)
	if fiberCtx != nil && !middleware.IsAdmin(fiberCtx) {
		return DeleteItem403JSONResponse{
			Error:   ptr("forbidden"),
			Message: ptr("admin access required"),
		}, nil
	}

	if err := s.itemService.Delete(ctx, uint(request.Id)); err != nil {
		if errors.Is(err, item.ErrItemNotFound) {
			return DeleteItem404JSONResponse{
				Error:   ptr("not_found"),
				Message: ptr("item not found"),
			}, nil
		}
		return DeleteItem401JSONResponse{}, nil
	}

	return DeleteItem204Response{}, nil
}

// ================================
// Country Endpoints
// ================================

// ListCountries implements StrictServerInterface.
func (s *Server) ListCountries(ctx context.Context, request ListCountriesRequestObject) (ListCountriesResponseObject, error) {
	fiberCtx := getFiberContext(ctx)
	if fiberCtx == nil {
		return ListCountries401JSONResponse{}, nil
	}

	// Parse all query parameters for dynamic filtering
	params := filter.ParseFromQuery(fiberCtx)

	result, err := s.countryService.ListFiltered(ctx, params)
	if err != nil {
		return ListCountries401JSONResponse{}, nil
	}

	return ListCountries200JSONResponse(toCountryListResponse(result, params)), nil
}

// CreateCountry implements StrictServerInterface.
func (s *Server) CreateCountry(ctx context.Context, request CreateCountryRequestObject) (CreateCountryResponseObject, error) {
	fiberCtx := getFiberContext(ctx)
	if fiberCtx != nil && !middleware.IsAdmin(fiberCtx) {
		return CreateCountry403JSONResponse{
			Error:   ptr("forbidden"),
			Message: ptr("admin access required"),
		}, nil
	}

	req := &country.CreateCountryRequest{
		Name: request.Body.Name,
		Code: request.Body.Code,
	}

	c, err := s.countryService.Create(ctx, req)
	if err != nil {
		if errors.Is(err, country.ErrCountryCodeExists) {
			return CreateCountry400JSONResponse{
				Error:   ptr("conflict"),
				Message: ptr("country code already exists"),
			}, nil
		}
		return CreateCountry400JSONResponse{
			Error:   ptr("bad_request"),
			Message: ptr("country creation failed"),
		}, nil
	}

	return CreateCountry201JSONResponse(toCountryResponse(c)), nil
}

// GetCountry implements StrictServerInterface.
func (s *Server) GetCountry(ctx context.Context, request GetCountryRequestObject) (GetCountryResponseObject, error) {
	c, err := s.countryService.GetByID(ctx, uint(request.Id))
	if err != nil {
		if errors.Is(err, country.ErrCountryNotFound) {
			return GetCountry404JSONResponse{
				Error:   ptr("not_found"),
				Message: ptr("country not found"),
			}, nil
		}
		return GetCountry401JSONResponse{}, nil
	}

	return GetCountry200JSONResponse(toCountryResponse(c)), nil
}

// UpdateCountry implements StrictServerInterface.
func (s *Server) UpdateCountry(ctx context.Context, request UpdateCountryRequestObject) (UpdateCountryResponseObject, error) {
	fiberCtx := getFiberContext(ctx)
	if fiberCtx != nil && !middleware.IsAdmin(fiberCtx) {
		return UpdateCountry403JSONResponse{
			Error:   ptr("forbidden"),
			Message: ptr("admin access required"),
		}, nil
	}

	req := &country.UpdateCountryRequest{
		Name: request.Body.Name,
		Code: request.Body.Code,
	}

	c, err := s.countryService.Update(ctx, uint(request.Id), req)
	if err != nil {
		if errors.Is(err, country.ErrCountryNotFound) {
			return UpdateCountry404JSONResponse{
				Error:   ptr("not_found"),
				Message: ptr("country not found"),
			}, nil
		}
		if errors.Is(err, country.ErrCountryCodeExists) {
			return UpdateCountry400JSONResponse{
				Error:   ptr("conflict"),
				Message: ptr("country code already exists"),
			}, nil
		}
		return UpdateCountry400JSONResponse{
			Error:   ptr("bad_request"),
			Message: ptr("update failed"),
		}, nil
	}

	return UpdateCountry200JSONResponse(toCountryResponse(c)), nil
}

// DeleteCountry implements StrictServerInterface.
func (s *Server) DeleteCountry(ctx context.Context, request DeleteCountryRequestObject) (DeleteCountryResponseObject, error) {
	fiberCtx := getFiberContext(ctx)
	if fiberCtx != nil && !middleware.IsAdmin(fiberCtx) {
		return DeleteCountry403JSONResponse{
			Error:   ptr("forbidden"),
			Message: ptr("admin access required"),
		}, nil
	}

	if err := s.countryService.Delete(ctx, uint(request.Id)); err != nil {
		if errors.Is(err, country.ErrCountryNotFound) {
			return DeleteCountry404JSONResponse{
				Error:   ptr("not_found"),
				Message: ptr("country not found"),
			}, nil
		}
		return DeleteCountry401JSONResponse{}, nil
	}

	return DeleteCountry204Response{}, nil
}

// ================================
// City Endpoints
// ================================

// ListCities implements StrictServerInterface.
func (s *Server) ListCities(ctx context.Context, request ListCitiesRequestObject) (ListCitiesResponseObject, error) {
	fiberCtx := getFiberContext(ctx)
	if fiberCtx == nil {
		return ListCities401JSONResponse{}, nil
	}

	// Parse all query parameters for dynamic filtering
	params := filter.ParseFromQuery(fiberCtx)

	result, err := s.cityService.ListFiltered(ctx, params)
	if err != nil {
		return ListCities401JSONResponse{}, nil
	}

	return ListCities200JSONResponse(toCityListResponse(result, params)), nil
}

// CreateCity implements StrictServerInterface.
func (s *Server) CreateCity(ctx context.Context, request CreateCityRequestObject) (CreateCityResponseObject, error) {
	fiberCtx := getFiberContext(ctx)
	if fiberCtx != nil && !middleware.IsAdmin(fiberCtx) {
		return CreateCity403JSONResponse{
			Error:   ptr("forbidden"),
			Message: ptr("admin access required"),
		}, nil
	}

	req := &city.CreateCityRequest{
		Name:      request.Body.Name,
		CountryID: uint(request.Body.CountryId),
	}

	c, err := s.cityService.Create(ctx, req)
	if err != nil {
		if errors.Is(err, city.ErrCountryNotFound) {
			return CreateCity400JSONResponse{
				Error:   ptr("bad_request"),
				Message: ptr("country not found"),
			}, nil
		}
		return CreateCity400JSONResponse{
			Error:   ptr("bad_request"),
			Message: ptr("city creation failed"),
		}, nil
	}

	return CreateCity201JSONResponse(toCityResponse(c)), nil
}

// GetCity implements StrictServerInterface.
func (s *Server) GetCity(ctx context.Context, request GetCityRequestObject) (GetCityResponseObject, error) {
	c, err := s.cityService.GetByIDWithCountry(ctx, uint(request.Id))
	if err != nil {
		if errors.Is(err, city.ErrCityNotFound) {
			return GetCity404JSONResponse{
				Error:   ptr("not_found"),
				Message: ptr("city not found"),
			}, nil
		}
		return GetCity401JSONResponse{}, nil
	}

	return GetCity200JSONResponse(toCityResponse(c)), nil
}

// UpdateCity implements StrictServerInterface.
func (s *Server) UpdateCity(ctx context.Context, request UpdateCityRequestObject) (UpdateCityResponseObject, error) {
	fiberCtx := getFiberContext(ctx)
	if fiberCtx != nil && !middleware.IsAdmin(fiberCtx) {
		return UpdateCity403JSONResponse{
			Error:   ptr("forbidden"),
			Message: ptr("admin access required"),
		}, nil
	}

	req := &city.UpdateCityRequest{
		Name: request.Body.Name,
	}
	if request.Body.CountryId != nil {
		cid := uint(*request.Body.CountryId)
		req.CountryID = &cid
	}

	c, err := s.cityService.Update(ctx, uint(request.Id), req)
	if err != nil {
		if errors.Is(err, city.ErrCityNotFound) {
			return UpdateCity404JSONResponse{
				Error:   ptr("not_found"),
				Message: ptr("city not found"),
			}, nil
		}
		if errors.Is(err, city.ErrCountryNotFound) {
			return UpdateCity400JSONResponse{
				Error:   ptr("bad_request"),
				Message: ptr("country not found"),
			}, nil
		}
		return UpdateCity400JSONResponse{
			Error:   ptr("bad_request"),
			Message: ptr("update failed"),
		}, nil
	}

	return UpdateCity200JSONResponse(toCityResponse(c)), nil
}

// DeleteCity implements StrictServerInterface.
func (s *Server) DeleteCity(ctx context.Context, request DeleteCityRequestObject) (DeleteCityResponseObject, error) {
	fiberCtx := getFiberContext(ctx)
	if fiberCtx != nil && !middleware.IsAdmin(fiberCtx) {
		return DeleteCity403JSONResponse{
			Error:   ptr("forbidden"),
			Message: ptr("admin access required"),
		}, nil
	}

	if err := s.cityService.Delete(ctx, uint(request.Id)); err != nil {
		if errors.Is(err, city.ErrCityNotFound) {
			return DeleteCity404JSONResponse{
				Error:   ptr("not_found"),
				Message: ptr("city not found"),
			}, nil
		}
		return DeleteCity401JSONResponse{}, nil
	}

	return DeleteCity204Response{}, nil
}

// ListCitiesByCountry implements StrictServerInterface.
func (s *Server) ListCitiesByCountry(ctx context.Context, request ListCitiesByCountryRequestObject) (ListCitiesByCountryResponseObject, error) {
	pagination := &common.Pagination{
		Page:     getIntOrDefault(request.Params.Page, 1),
		PageSize: getIntOrDefault(request.Params.PageSize, 10),
	}

	result, err := s.cityService.ListByCountry(ctx, uint(request.CountryId), pagination)
	if err != nil {
		if errors.Is(err, city.ErrCountryNotFound) {
			return ListCitiesByCountry404JSONResponse{
				Error:   ptr("not_found"),
				Message: ptr("country not found"),
			}, nil
		}
		return ListCitiesByCountry401JSONResponse{}, nil
	}

	return ListCitiesByCountry200JSONResponse(toCityListFromPaginated(result)), nil
}

// ================================
// Document Endpoints
// ================================

// ListDocuments implements StrictServerInterface.
func (s *Server) ListDocuments(ctx context.Context, request ListDocumentsRequestObject) (ListDocumentsResponseObject, error) {
	fiberCtx := getFiberContext(ctx)
	if fiberCtx == nil {
		return ListDocuments401JSONResponse{}, nil
	}

	// Parse all query parameters for dynamic filtering
	params := filter.ParseFromQuery(fiberCtx)

	result, err := s.documentService.ListFiltered(ctx, params)
	if err != nil {
		return ListDocuments401JSONResponse{}, nil
	}

	return ListDocuments200JSONResponse(toDocumentListResponse(result, params)), nil
}

// CreateDocument implements StrictServerInterface.
func (s *Server) CreateDocument(ctx context.Context, request CreateDocumentRequestObject) (CreateDocumentResponseObject, error) {
	fiberCtx := getFiberContext(ctx)
	if fiberCtx != nil && !middleware.IsAdmin(fiberCtx) {
		return CreateDocument403JSONResponse{
			Error:   ptr("forbidden"),
			Message: ptr("admin access required"),
		}, nil
	}

	docItems := make([]document.CreateDocumentItemRequest, 0)
	if request.Body.Items != nil {
		for _, item := range *request.Body.Items {
			docItems = append(docItems, document.CreateDocumentItemRequest{
				ItemID:   uint(item.ItemId),
				Quantity: item.Quantity,
				Price:    int64(item.Price),
			})
		}
	}

	req := &document.CreateDocumentRequest{
		Code:         request.Body.Code,
		CityID:       uint(request.Body.CityId),
		DocumentDate: request.Body.DocumentDate.Time,
		Items:        docItems,
	}

	d, err := s.documentService.Create(ctx, req)
	if err != nil {
		if errors.Is(err, document.ErrDocumentCodeExists) {
			return CreateDocument400JSONResponse{
				Error:   ptr("conflict"),
				Message: ptr("document code already exists"),
			}, nil
		}
		if errors.Is(err, document.ErrCityNotFound) {
			return CreateDocument400JSONResponse{
				Error:   ptr("bad_request"),
				Message: ptr("city not found"),
			}, nil
		}
		if errors.Is(err, document.ErrItemNotFound) {
			return CreateDocument400JSONResponse{
				Error:   ptr("bad_request"),
				Message: ptr("item not found"),
			}, nil
		}
		return CreateDocument400JSONResponse{
			Error:   ptr("bad_request"),
			Message: ptr("document creation failed"),
		}, nil
	}

	return CreateDocument201JSONResponse(toDocumentResponse(d)), nil
}

// GetDocument implements StrictServerInterface.
func (s *Server) GetDocument(ctx context.Context, request GetDocumentRequestObject) (GetDocumentResponseObject, error) {
	d, err := s.documentService.GetByIDWithDetails(ctx, uint(request.Id))
	if err != nil {
		if errors.Is(err, document.ErrDocumentNotFound) {
			return GetDocument404JSONResponse{
				Error:   ptr("not_found"),
				Message: ptr("document not found"),
			}, nil
		}
		return GetDocument401JSONResponse{}, nil
	}

	return GetDocument200JSONResponse(toDocumentResponse(d)), nil
}

// UpdateDocument implements StrictServerInterface.
func (s *Server) UpdateDocument(ctx context.Context, request UpdateDocumentRequestObject) (UpdateDocumentResponseObject, error) {
	fiberCtx := getFiberContext(ctx)
	if fiberCtx != nil && !middleware.IsAdmin(fiberCtx) {
		return UpdateDocument403JSONResponse{
			Error:   ptr("forbidden"),
			Message: ptr("admin access required"),
		}, nil
	}

	req := &document.UpdateDocumentRequest{
		Code: request.Body.Code,
	}
	if request.Body.CityId != nil {
		cid := uint(*request.Body.CityId)
		req.CityID = &cid
	}
	if request.Body.DocumentDate != nil {
		t := request.Body.DocumentDate.Time
		req.DocumentDate = &t
	}

	d, err := s.documentService.Update(ctx, uint(request.Id), req)
	if err != nil {
		if errors.Is(err, document.ErrDocumentNotFound) {
			return UpdateDocument404JSONResponse{
				Error:   ptr("not_found"),
				Message: ptr("document not found"),
			}, nil
		}
		if errors.Is(err, document.ErrDocumentCodeExists) {
			return UpdateDocument400JSONResponse{
				Error:   ptr("conflict"),
				Message: ptr("document code already exists"),
			}, nil
		}
		if errors.Is(err, document.ErrCityNotFound) {
			return UpdateDocument400JSONResponse{
				Error:   ptr("bad_request"),
				Message: ptr("city not found"),
			}, nil
		}
		return UpdateDocument400JSONResponse{
			Error:   ptr("bad_request"),
			Message: ptr("update failed"),
		}, nil
	}

	return UpdateDocument200JSONResponse(toDocumentResponse(d)), nil
}

// DeleteDocument implements StrictServerInterface.
func (s *Server) DeleteDocument(ctx context.Context, request DeleteDocumentRequestObject) (DeleteDocumentResponseObject, error) {
	fiberCtx := getFiberContext(ctx)
	if fiberCtx != nil && !middleware.IsAdmin(fiberCtx) {
		return DeleteDocument403JSONResponse{
			Error:   ptr("forbidden"),
			Message: ptr("admin access required"),
		}, nil
	}

	if err := s.documentService.Delete(ctx, uint(request.Id)); err != nil {
		if errors.Is(err, document.ErrDocumentNotFound) {
			return DeleteDocument404JSONResponse{
				Error:   ptr("not_found"),
				Message: ptr("document not found"),
			}, nil
		}
		return DeleteDocument401JSONResponse{}, nil
	}

	return DeleteDocument204Response{}, nil
}

// AddDocumentItem implements StrictServerInterface.
func (s *Server) AddDocumentItem(ctx context.Context, request AddDocumentItemRequestObject) (AddDocumentItemResponseObject, error) {
	fiberCtx := getFiberContext(ctx)
	if fiberCtx != nil && !middleware.IsAdmin(fiberCtx) {
		return AddDocumentItem403JSONResponse{
			Error:   ptr("forbidden"),
			Message: ptr("admin access required"),
		}, nil
	}

	req := &document.AddDocumentItemRequest{
		ItemID:   uint(request.Body.ItemId),
		Quantity: request.Body.Quantity,
		Price:    int64(request.Body.Price),
	}

	item, err := s.documentService.AddItem(ctx, uint(request.Id), req)
	if err != nil {
		if errors.Is(err, document.ErrDocumentNotFound) {
			return AddDocumentItem404JSONResponse{
				Error:   ptr("not_found"),
				Message: ptr("document not found"),
			}, nil
		}
		if errors.Is(err, document.ErrItemNotFound) {
			return AddDocumentItem400JSONResponse{
				Error:   ptr("bad_request"),
				Message: ptr("item not found"),
			}, nil
		}
		return AddDocumentItem400JSONResponse{
			Error:   ptr("bad_request"),
			Message: ptr("add item failed"),
		}, nil
	}

	return AddDocumentItem201JSONResponse(toDocumentItemResponse(item)), nil
}

// UpdateDocumentItem implements StrictServerInterface.
func (s *Server) UpdateDocumentItem(ctx context.Context, request UpdateDocumentItemRequestObject) (UpdateDocumentItemResponseObject, error) {
	fiberCtx := getFiberContext(ctx)
	if fiberCtx != nil && !middleware.IsAdmin(fiberCtx) {
		return UpdateDocumentItem403JSONResponse{
			Error:   ptr("forbidden"),
			Message: ptr("admin access required"),
		}, nil
	}

	req := &document.UpdateDocumentItemRequest{}
	if request.Body.Quantity != nil {
		req.Quantity = request.Body.Quantity
	}
	if request.Body.Price != nil {
		p := int64(*request.Body.Price)
		req.Price = &p
	}

	item, err := s.documentService.UpdateItem(ctx, uint(request.Id), uint(request.ItemId), req)
	if err != nil {
		if errors.Is(err, document.ErrDocumentNotFound) {
			return UpdateDocumentItem404JSONResponse{
				Error:   ptr("not_found"),
				Message: ptr("document not found"),
			}, nil
		}
		if errors.Is(err, document.ErrDocumentItemNotFound) {
			return UpdateDocumentItem404JSONResponse{
				Error:   ptr("not_found"),
				Message: ptr("document item not found"),
			}, nil
		}
		return UpdateDocumentItem400JSONResponse{
			Error:   ptr("bad_request"),
			Message: ptr("update failed"),
		}, nil
	}

	return UpdateDocumentItem200JSONResponse(toDocumentItemResponse(item)), nil
}

// DeleteDocumentItem implements StrictServerInterface.
func (s *Server) DeleteDocumentItem(ctx context.Context, request DeleteDocumentItemRequestObject) (DeleteDocumentItemResponseObject, error) {
	fiberCtx := getFiberContext(ctx)
	if fiberCtx != nil && !middleware.IsAdmin(fiberCtx) {
		return DeleteDocumentItem403JSONResponse{
			Error:   ptr("forbidden"),
			Message: ptr("admin access required"),
		}, nil
	}

	if err := s.documentService.RemoveItem(ctx, uint(request.Id), uint(request.ItemId)); err != nil {
		if errors.Is(err, document.ErrDocumentNotFound) {
			return DeleteDocumentItem404JSONResponse{
				Error:   ptr("not_found"),
				Message: ptr("document not found"),
			}, nil
		}
		if errors.Is(err, document.ErrDocumentItemNotFound) {
			return DeleteDocumentItem404JSONResponse{
				Error:   ptr("not_found"),
				Message: ptr("document item not found"),
			}, nil
		}
		return DeleteDocumentItem401JSONResponse{}, nil
	}

	return DeleteDocumentItem204Response{}, nil
}

// ================================
// Health Endpoint
// ================================

// HealthCheck implements StrictServerInterface.
func (s *Server) HealthCheck(ctx context.Context, request HealthCheckRequestObject) (HealthCheckResponseObject, error) {
	// Check database connection
	sqlDB, err := s.container.DB.DB()
	if err != nil {
		return HealthCheck503JSONResponse{
			Status:   ptr("unhealthy"),
			Database: ptr("connection failed"),
		}, nil
	}
	if err := sqlDB.Ping(); err != nil {
		return HealthCheck503JSONResponse{
			Status:   ptr("unhealthy"),
			Database: ptr("ping failed"),
		}, nil
	}

	// Check Redis connection
	if s.container.Redis != nil {
		if err := s.container.Redis.Ping(ctx).Err(); err != nil {
			return HealthCheck503JSONResponse{
				Status:   ptr("unhealthy"),
				Database: ptr("ok"),
				Redis:    ptr("ping failed"),
			}, nil
		}
	}

	return HealthCheck200JSONResponse{
		Status:   ptr("healthy"),
		Database: ptr("ok"),
		Redis:    ptr("ok"),
	}, nil
}

// ================================
// Helper Functions
// ================================

// fiberContextKey is the key used to store fiber.Ctx in context.
type fiberContextKey struct{}

// getFiberContext extracts the fiber.Ctx from context.
func getFiberContext(ctx context.Context) *fiber.Ctx {
	if c, ok := ctx.Value(fiberContextKey{}).(*fiber.Ctx); ok {
		return c
	}
	return nil
}

// WithFiberContext adds the fiber.Ctx to the context.
func WithFiberContext(ctx context.Context, c *fiber.Ctx) context.Context {
	return context.WithValue(ctx, fiberContextKey{}, c)
}

// ptr returns a pointer to the given value.
func ptr[T any](v T) *T {
	return &v
}

func ptrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func intToString(i int) string {
	return strconv.Itoa(i)
}

func getIntOrDefault(p *int, def int) int {
	if p == nil {
		return def
	}
	return *p
}

func buildFilterParams(page, pageSize *int, sort, order *string) *filter.Params {
	params := &filter.Params{
		Page:    getIntOrDefault(page, 1),
		Limit:   getIntOrDefault(pageSize, 10),
		Filters: make([]filter.FilterParam, 0),
		Sort:    make([]filter.SortParam, 0),
	}

	if sort != nil && *sort != "" {
		desc := order != nil && *order == "desc"
		params.Sort = append(params.Sort, filter.SortParam{
			Field: *sort,
			Desc:  desc,
		})
	}

	return params
}

// Ensure Server implements StrictServerInterface at compile time.
var _ StrictServerInterface = (*Server)(nil)
