// Package api provides the OpenAPI server implementation.
package api

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/common"
	commonerrors "github.com/voidmaindev/go-template/internal/common/errors"
	"github.com/voidmaindev/go-template/internal/common/filter"
	"github.com/voidmaindev/go-template/internal/container"
	"github.com/voidmaindev/go-template/internal/domain/city"
	"github.com/voidmaindev/go-template/internal/domain/country"
	"github.com/voidmaindev/go-template/internal/domain/document"
	"github.com/voidmaindev/go-template/internal/domain/item"
	"github.com/voidmaindev/go-template/internal/domain/rbac"
	"github.com/voidmaindev/go-template/internal/domain/user"
	"github.com/voidmaindev/go-template/internal/middleware"
)

// Server implements the StrictServerInterface.
type Server struct {
	container       *container.Container
	userService     user.Service
	itemService     item.Service
	countryService  country.Service
	cityService     city.Service
	documentService document.Service
	rbacService     rbac.Service
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
		rbacService:     container.MustGetTyped[rbac.Service](c, rbac.ServiceKey),
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
		if errors.Is(err, user.ErrEmailExists) {
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
		if errors.Is(err, user.ErrInvalidCredentials) || commonerrors.IsUnauthorized(err) {
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
// Note: To fully prevent refresh token reuse attacks, update the OpenAPI spec
// to accept a refresh_token in the request body.
func (s *Server) Logout(ctx context.Context, request LogoutRequestObject) (LogoutResponseObject, error) {
	// Get token from fiber context - need to extract from ctx
	fiberCtx := getFiberContext(ctx)
	if fiberCtx == nil {
		return Logout401JSONResponse{
			Error:   ptr("unauthorized"),
			Message: ptr("no context"),
		}, nil
	}

	accessToken := middleware.GetTokenFromContext(fiberCtx)
	if accessToken == "" {
		return Logout401JSONResponse{
			Error:   ptr("unauthorized"),
			Message: ptr("no token provided"),
		}, nil
	}

	// Try to extract refresh token from raw body if present
	var refreshToken string
	if fiberCtx.Body() != nil {
		var body struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := fiberCtx.BodyParser(&body); err == nil && body.RefreshToken != "" {
			refreshToken = body.RefreshToken
		}
	}

	if err := s.userService.Logout(ctx, accessToken, refreshToken); err != nil {
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
		if errors.Is(err, user.ErrTokenInvalid) || errors.Is(err, user.ErrTokenBlacklisted) || commonerrors.IsUnauthorized(err) {
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

	// Authorization is handled by RBAC middleware at route level

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

	// Authorization: self-access allowed, admin access handled by RBAC middleware at route level
	// For admin viewing other users, the route should have RBAC middleware with user:read permission
	_ = currentUserID // Unused - authorization handled by RBAC middleware

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

	// Authorization: self-delete allowed, admin delete handled by RBAC middleware at route level
	// For admin deleting other users, the route should have RBAC middleware with user:delete permission
	if uint(request.Id) != currentUserID {
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
// Authorization is handled by RBAC middleware at route level.
func (s *Server) CreateItem(ctx context.Context, request CreateItemRequestObject) (CreateItemResponseObject, error) {
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
	// Authorization is handled by RBAC middleware at route level
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
	// Authorization is handled by RBAC middleware at route level
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
	// Authorization is handled by RBAC middleware at route level
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
// Authorization is handled by RBAC middleware at route level.
func (s *Server) UpdateCountry(ctx context.Context, request UpdateCountryRequestObject) (UpdateCountryResponseObject, error) {
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
	// Authorization is handled by RBAC middleware at route level
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
	// Authorization is handled by RBAC middleware at route level
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
// Authorization is handled by RBAC middleware at route level.
func (s *Server) UpdateCity(ctx context.Context, request UpdateCityRequestObject) (UpdateCityResponseObject, error) {
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
	// Authorization is handled by RBAC middleware at route level
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
	pagination := common.PaginationFromOptional(
		request.Params.Page,
		request.Params.PageSize,
		nil, nil, // Sort/Order not in API spec for this endpoint
	)

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
// Authorization is handled by RBAC middleware at route level.
func (s *Server) CreateDocument(ctx context.Context, request CreateDocumentRequestObject) (CreateDocumentResponseObject, error) {
	var docItems []document.CreateDocumentItemRequest
	if request.Body.Items != nil {
		docItems = make([]document.CreateDocumentItemRequest, 0, len(*request.Body.Items))
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
// Authorization is handled by RBAC middleware at route level.
func (s *Server) UpdateDocument(ctx context.Context, request UpdateDocumentRequestObject) (UpdateDocumentResponseObject, error) {
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
// Authorization is handled by RBAC middleware at route level.
func (s *Server) DeleteDocument(ctx context.Context, request DeleteDocumentRequestObject) (DeleteDocumentResponseObject, error) {
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
// Authorization is handled by RBAC middleware at route level.
func (s *Server) AddDocumentItem(ctx context.Context, request AddDocumentItemRequestObject) (AddDocumentItemResponseObject, error) {
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
// Authorization is handled by RBAC middleware at route level.
func (s *Server) UpdateDocumentItem(ctx context.Context, request UpdateDocumentItemRequestObject) (UpdateDocumentItemResponseObject, error) {
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
// Authorization is handled by RBAC middleware at route level.
func (s *Server) DeleteDocumentItem(ctx context.Context, request DeleteDocumentItemRequestObject) (DeleteDocumentItemResponseObject, error) {
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
	checks := make(map[string]CheckResult)
	now := time.Now()

	// Check database connection
	dbStatus := CheckResultStatusHealthy
	var dbError *string
	sqlDB, err := s.container.DB.DB()
	if err != nil {
		dbStatus = CheckResultStatusUnhealthy
		dbError = ptr("unhealthy")
		slog.Error("health check: database connection failed", "error", err)
	} else if err := sqlDB.Ping(); err != nil {
		dbStatus = CheckResultStatusUnhealthy
		dbError = ptr("unhealthy")
		slog.Error("health check: database ping failed", "error", err)
	}
	checks["database"] = CheckResult{
		Status:    &dbStatus,
		Error:     dbError,
		Timestamp: &now,
	}

	// Check Redis connection
	redisStatus := CheckResultStatusHealthy
	var redisError *string
	if s.container.Redis != nil {
		if err := s.container.Redis.Ping(ctx).Err(); err != nil {
			redisStatus = CheckResultStatusUnhealthy
			redisError = ptr("unhealthy")
			slog.Error("health check: redis ping failed", "error", err)
		}
	}
	checks["redis"] = CheckResult{
		Status:    &redisStatus,
		Error:     redisError,
		Timestamp: &now,
	}

	// Determine overall status
	overallStatus := HealthResponseStatusHealthy
	if dbStatus == CheckResultStatusUnhealthy || redisStatus == CheckResultStatusUnhealthy {
		overallStatus = HealthResponseStatusUnhealthy
		return HealthCheck503JSONResponse{
			Status:    &overallStatus,
			Checks:    &checks,
			Timestamp: &now,
		}, nil
	}

	return HealthCheck200JSONResponse{
		Status:    &overallStatus,
		Checks:    &checks,
		Timestamp: &now,
	}, nil
}

// GetMetrics implements StrictServerInterface.
func (s *Server) GetMetrics(ctx context.Context, request GetMetricsRequestObject) (GetMetricsResponseObject, error) {
	// Metrics are handled by the Prometheus handler middleware, not here
	// This is a placeholder that shouldn't be called directly
	return GetMetrics200TextResponse("# Metrics endpoint handled by Prometheus middleware"), nil
}

// LivenessCheck implements StrictServerInterface.
func (s *Server) LivenessCheck(ctx context.Context, request LivenessCheckRequestObject) (LivenessCheckResponseObject, error) {
	status := Healthy
	return LivenessCheck200JSONResponse{
		Status: &status,
	}, nil
}

// ReadinessCheck implements StrictServerInterface.
func (s *Server) ReadinessCheck(ctx context.Context, request ReadinessCheckRequestObject) (ReadinessCheckResponseObject, error) {
	checks := make(map[string]CheckResult)
	now := time.Now()

	// Check database connection
	dbStatus := CheckResultStatusHealthy
	var dbError *string
	sqlDB, err := s.container.DB.DB()
	if err != nil {
		dbStatus = CheckResultStatusUnhealthy
		dbError = ptr("unhealthy")
		slog.Error("readiness check: database connection failed", "error", err)
	} else if err := sqlDB.Ping(); err != nil {
		dbStatus = CheckResultStatusUnhealthy
		dbError = ptr("unhealthy")
		slog.Error("readiness check: database ping failed", "error", err)
	}
	checks["database"] = CheckResult{
		Status:    &dbStatus,
		Error:     dbError,
		Timestamp: &now,
	}

	// Check Redis connection
	redisStatus := CheckResultStatusHealthy
	var redisError *string
	if s.container.Redis != nil {
		if err := s.container.Redis.Ping(ctx).Err(); err != nil {
			redisStatus = CheckResultStatusUnhealthy
			redisError = ptr("unhealthy")
			slog.Error("readiness check: redis ping failed", "error", err)
		}
	}
	checks["redis"] = CheckResult{
		Status:    &redisStatus,
		Error:     redisError,
		Timestamp: &now,
	}

	// Determine overall status
	overallStatus := HealthResponseStatusHealthy
	if dbStatus == CheckResultStatusUnhealthy || redisStatus == CheckResultStatusUnhealthy {
		overallStatus = HealthResponseStatusUnhealthy
		return ReadinessCheck503JSONResponse{
			Status:    &overallStatus,
			Checks:    &checks,
			Timestamp: &now,
		}, nil
	}

	return ReadinessCheck200JSONResponse{
		Status:    &overallStatus,
		Checks:    &checks,
		Timestamp: &now,
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

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
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

// ================================
// RBAC Endpoints
// ================================

// ListActions implements StrictServerInterface.
func (s *Server) ListActions(ctx context.Context, request ListActionsRequestObject) (ListActionsResponseObject, error) {
	actions := s.rbacService.GetActions(ctx)
	return ListActions200JSONResponse{Actions: &actions}, nil
}

// ListDomains implements StrictServerInterface.
func (s *Server) ListDomains(ctx context.Context, request ListDomainsRequestObject) (ListDomainsResponseObject, error) {
	domains := s.rbacService.GetDomains(ctx)
	apiDomains := make([]DomainResponse, len(domains))
	for i, d := range domains {
		apiDomains[i] = DomainResponse{
			Name:        ptr(d.Name),
			IsProtected: ptr(d.IsProtected),
		}
	}
	return ListDomains200JSONResponse{Domains: &apiDomains}, nil
}

// ListRoles implements StrictServerInterface.
// Authorization is handled by RBAC middleware at route level.
func (s *Server) ListRoles(ctx context.Context, request ListRolesRequestObject) (ListRolesResponseObject, error) {
	params := buildFilterParams(request.Params.Page, request.Params.PageSize, nil, nil)
	result, err := s.rbacService.ListRoles(ctx, params)
	if err != nil {
		return ListRoles401JSONResponse{
			Error:   ptr("internal_error"),
			Message: ptr(err.Error()),
		}, nil
	}

	roles := make([]RoleResponse, len(result.Data))
	for i, r := range result.Data {
		roles[i] = RoleResponse{
			Code:        ptr(r.Code),
			Name:        ptr(r.Name),
			Description: ptr(r.Description),
			IsSystem:    ptr(r.IsSystem),
			CreatedAt:   ptr(r.CreatedAt),
		}
	}

	totalPages := common.CalculateTotalPages(result.Total, params.Limit)
	return ListRoles200JSONResponse{
		Data:       &roles,
		Total:      ptr(result.Total),
		Page:       ptr(params.Page),
		PageSize:   ptr(params.Limit),
		TotalPages: ptr(totalPages),
		HasMore:    ptr(params.Page < totalPages),
	}, nil
}

// CreateRole implements StrictServerInterface.
// Authorization is handled by RBAC middleware at route level.
func (s *Server) CreateRole(ctx context.Context, request CreateRoleRequestObject) (CreateRoleResponseObject, error) {
	perms := make([]rbac.PermissionInput, 0)
	if request.Body.Permissions != nil {
		for _, p := range *request.Body.Permissions {
			actions := make([]string, len(p.Actions))
			for i, a := range p.Actions {
				actions[i] = string(a)
			}
			perms = append(perms, rbac.PermissionInput{
				Domain:  p.Domain,
				Actions: actions,
			})
		}
	}

	req := &rbac.CreateRoleRequest{
		Code:        request.Body.Code,
		Name:        request.Body.Name,
		Description: deref(request.Body.Description),
		Permissions: perms,
	}

	role, err := s.rbacService.CreateRole(ctx, req)
	if err != nil {
		if errors.Is(err, rbac.ErrRoleCodeExists) {
			return CreateRole409JSONResponse{
				Error:   ptr("conflict"),
				Message: ptr("Role code already exists"),
			}, nil
		}
		return CreateRole400JSONResponse{
			Error:   ptr("bad_request"),
			Message: ptr(err.Error()),
		}, nil
	}

	return CreateRole201JSONResponse{
		Code:        ptr(role.Code),
		Name:        ptr(role.Name),
		Description: ptr(role.Description),
		IsSystem:    ptr(role.IsSystem),
		CreatedAt:   ptr(role.CreatedAt),
	}, nil
}

// GetRole implements StrictServerInterface.
// Authorization is handled by RBAC middleware at route level.
func (s *Server) GetRole(ctx context.Context, request GetRoleRequestObject) (GetRoleResponseObject, error) {
	roleWithPerms, err := s.rbacService.GetRoleByCode(ctx, request.Code)
	if err != nil {
		if errors.Is(err, rbac.ErrRoleNotFound) {
			return GetRole404JSONResponse{
				Error:   ptr("not_found"),
				Message: ptr("Role not found"),
			}, nil
		}
		return GetRole401JSONResponse{
			Error:   ptr("internal_error"),
			Message: ptr(err.Error()),
		}, nil
	}

	perms := make([]PermissionResponse, len(roleWithPerms.Permissions))
	for i, p := range roleWithPerms.Permissions {
		perms[i] = PermissionResponse{
			Domain:  ptr(p.Domain),
			Actions: ptr(p.Actions),
		}
	}

	return GetRole200JSONResponse{
		Code:        ptr(roleWithPerms.Role.Code),
		Name:        ptr(roleWithPerms.Role.Name),
		Description: ptr(roleWithPerms.Role.Description),
		IsSystem:    ptr(roleWithPerms.Role.IsSystem),
		CreatedAt:   ptr(roleWithPerms.Role.CreatedAt),
		Permissions: &perms,
	}, nil
}

// DeleteRole implements StrictServerInterface.
// Authorization is handled by RBAC middleware at route level.
func (s *Server) DeleteRole(ctx context.Context, request DeleteRoleRequestObject) (DeleteRoleResponseObject, error) {
	err := s.rbacService.DeleteRole(ctx, request.Code)
	if err != nil {
		if errors.Is(err, rbac.ErrRoleNotFound) {
			return DeleteRole404JSONResponse{
				Error:   ptr("not_found"),
				Message: ptr("Role not found"),
			}, nil
		}
		if errors.Is(err, rbac.ErrSystemRoleCannotBeDeleted) {
			return DeleteRole403JSONResponse{
				Error:   ptr("forbidden"),
				Message: ptr("Cannot delete system role"),
			}, nil
		}
		return DeleteRole401JSONResponse{
			Error:   ptr("internal_error"),
			Message: ptr(err.Error()),
		}, nil
	}

	return DeleteRole204Response{}, nil
}

// UpdateRolePermissions implements StrictServerInterface.
// Authorization is handled by RBAC middleware at route level.
func (s *Server) UpdateRolePermissions(ctx context.Context, request UpdateRolePermissionsRequestObject) (UpdateRolePermissionsResponseObject, error) {
	perms := make([]rbac.PermissionInput, len(request.Body.Permissions))
	for i, p := range request.Body.Permissions {
		actions := make([]string, len(p.Actions))
		for j, a := range p.Actions {
			actions[j] = string(a)
		}
		perms[i] = rbac.PermissionInput{
			Domain:  p.Domain,
			Actions: actions,
		}
	}

	req := &rbac.UpdateRolePermissionsRequest{
		Permissions: perms,
	}

	roleWithPerms, err := s.rbacService.UpdateRolePermissions(ctx, request.Code, req)
	if err != nil {
		if errors.Is(err, rbac.ErrRoleNotFound) {
			return UpdateRolePermissions404JSONResponse{
				Error:   ptr("not_found"),
				Message: ptr("Role not found"),
			}, nil
		}
		if errors.Is(err, rbac.ErrSystemRoleCannotBeModified) {
			return UpdateRolePermissions403JSONResponse{
				Error:   ptr("forbidden"),
				Message: ptr("Cannot modify system role"),
			}, nil
		}
		return UpdateRolePermissions400JSONResponse{
			Error:   ptr("bad_request"),
			Message: ptr(err.Error()),
		}, nil
	}

	permResponses := make([]PermissionResponse, len(roleWithPerms.Permissions))
	for i, p := range roleWithPerms.Permissions {
		permResponses[i] = PermissionResponse{
			Domain:  ptr(p.Domain),
			Actions: ptr(p.Actions),
		}
	}

	return UpdateRolePermissions200JSONResponse{
		Code:        ptr(roleWithPerms.Role.Code),
		Name:        ptr(roleWithPerms.Role.Name),
		Description: ptr(roleWithPerms.Role.Description),
		IsSystem:    ptr(roleWithPerms.Role.IsSystem),
		CreatedAt:   ptr(roleWithPerms.Role.CreatedAt),
		Permissions: &permResponses,
	}, nil
}

// GetUserRoles implements StrictServerInterface.
func (s *Server) GetUserRoles(ctx context.Context, request GetUserRolesRequestObject) (GetUserRolesResponseObject, error) {
	fiberCtx := getFiberContext(ctx)
	if fiberCtx == nil {
		return GetUserRoles401JSONResponse{}, nil
	}

	// Self-access allowed; admin access to other users' roles handled by RBAC middleware
	// For admin viewing other users' roles, the route should have RBAC middleware with rbac:read permission
	currentUserID, _ := middleware.GetUserIDFromContext(fiberCtx)
	_ = currentUserID // Authorization handled by RBAC middleware at route level

	roles, err := s.rbacService.GetUserRoles(ctx, uint(request.Id))
	if err != nil {
		return GetUserRoles401JSONResponse{
			Error:   ptr("internal_error"),
			Message: ptr(err.Error()),
		}, nil
	}

	apiRoles := make([]UserRoleResponse, len(roles))
	for i, r := range roles {
		apiRoles[i] = UserRoleResponse{
			Code: ptr(r.Code),
			Name: ptr(r.Name),
		}
	}

	return GetUserRoles200JSONResponse{
		UserId: ptr(request.Id),
		Roles:  &apiRoles,
	}, nil
}

// AssignRoleToUser implements StrictServerInterface.
// Authorization is handled by RBAC middleware at route level.
func (s *Server) AssignRoleToUser(ctx context.Context, request AssignRoleToUserRequestObject) (AssignRoleToUserResponseObject, error) {
	err := s.rbacService.AssignRole(ctx, uint(request.Id), request.Body.RoleCode)
	if err != nil {
		if errors.Is(err, rbac.ErrRoleNotFound) {
			return AssignRoleToUser404JSONResponse{
				Error:   ptr("not_found"),
				Message: ptr("Role not found"),
			}, nil
		}
		if errors.Is(err, rbac.ErrRoleAlreadyAssigned) {
			return AssignRoleToUser400JSONResponse{
				Error:   ptr("bad_request"),
				Message: ptr("Role already assigned"),
			}, nil
		}
		return AssignRoleToUser400JSONResponse{
			Error:   ptr("bad_request"),
			Message: ptr(err.Error()),
		}, nil
	}

	// Return updated roles
	roles, err := s.rbacService.GetUserRoles(ctx, uint(request.Id))
	if err != nil {
		return AssignRoleToUser400JSONResponse{
			Error:   ptr("internal_error"),
			Message: ptr(err.Error()),
		}, nil
	}

	apiRoles := make([]UserRoleResponse, len(roles))
	for i, r := range roles {
		apiRoles[i] = UserRoleResponse{
			Code: ptr(r.Code),
			Name: ptr(r.Name),
		}
	}

	return AssignRoleToUser200JSONResponse{
		UserId: ptr(request.Id),
		Roles:  &apiRoles,
	}, nil
}

// RemoveRoleFromUser implements StrictServerInterface.
// Authorization is handled by RBAC middleware at route level.
func (s *Server) RemoveRoleFromUser(ctx context.Context, request RemoveRoleFromUserRequestObject) (RemoveRoleFromUserResponseObject, error) {
	err := s.rbacService.RemoveRole(ctx, uint(request.Id), request.Code)
	if err != nil {
		if errors.Is(err, rbac.ErrRoleNotFound) || errors.Is(err, rbac.ErrRoleNotAssigned) {
			return RemoveRoleFromUser404JSONResponse{
				Error:   ptr("not_found"),
				Message: ptr("Role assignment not found"),
			}, nil
		}
		return RemoveRoleFromUser403JSONResponse{
			Error:   ptr("forbidden"),
			Message: ptr(err.Error()),
		}, nil
	}

	return RemoveRoleFromUser204Response{}, nil
}

// Ensure Server implements StrictServerInterface at compile time.
var _ StrictServerInterface = (*Server)(nil)
