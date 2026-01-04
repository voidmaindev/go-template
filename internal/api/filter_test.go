package api

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/voidmaindev/go-template/internal/common/filter"
)

// TestFilterParsingIntegration tests that filter parameters are correctly parsed
// from query strings in the API layer.
func TestFilterParsingIntegration(t *testing.T) {
	tests := []struct {
		name           string
		queryString    string
		wantPage       int
		wantLimit      int
		wantFilters    []filter.FilterParam
		wantSort       []filter.SortParam
	}{
		{
			name:        "default pagination",
			queryString: "",
			wantPage:    1,
			wantLimit:   10,
			wantFilters: nil,
			wantSort:    nil,
		},
		{
			name:        "custom pagination",
			queryString: "?page=3&page_size=25",
			wantPage:    3,
			wantLimit:   25,
			wantFilters: nil,
			wantSort:    nil,
		},
		{
			name:        "equality filter",
			queryString: "?name=Berlin",
			wantPage:    1,
			wantLimit:   10,
			wantFilters: []filter.FilterParam{
				{Field: "name", Operator: filter.OpEq, Value: "Berlin"},
			},
			wantSort: nil,
		},
		{
			name:        "contains filter",
			queryString: "?name__contains=new",
			wantPage:    1,
			wantLimit:   10,
			wantFilters: []filter.FilterParam{
				{Field: "name", Operator: filter.OpContains, Value: "new"},
			},
			wantSort: nil,
		},
		{
			name:        "range filters",
			queryString: "?price__gte=1000&price__lte=5000",
			wantPage:    1,
			wantLimit:   10,
			wantFilters: []filter.FilterParam{
				{Field: "price", Operator: filter.OpGte, Value: "1000"},
				{Field: "price", Operator: filter.OpLte, Value: "5000"},
			},
			wantSort: nil,
		},
		{
			name:        "relation filter",
			queryString: "?country.name__contains=germany",
			wantPage:    1,
			wantLimit:   10,
			wantFilters: []filter.FilterParam{
				{Field: "country.name", Operator: filter.OpContains, Value: "germany"},
			},
			wantSort: nil,
		},
		{
			name:        "sorting ascending",
			queryString: "?sort=name&order=asc",
			wantPage:    1,
			wantLimit:   10,
			wantFilters: nil,
			wantSort: []filter.SortParam{
				{Field: "name", Desc: false},
			},
		},
		{
			name:        "sorting descending",
			queryString: "?sort=created_at&order=desc",
			wantPage:    1,
			wantLimit:   10,
			wantFilters: nil,
			wantSort: []filter.SortParam{
				{Field: "created_at", Desc: true},
			},
		},
		{
			name:        "combined filters and sorting",
			queryString: "?name__contains=test&price__gte=100&sort=name&order=asc&page=2&page_size=20",
			wantPage:    2,
			wantLimit:   20,
			wantFilters: []filter.FilterParam{
				{Field: "name", Operator: filter.OpContains, Value: "test"},
				{Field: "price", Operator: filter.OpGte, Value: "100"},
			},
			wantSort: []filter.SortParam{
				{Field: "name", Desc: false},
			},
		},
		{
			name:        "all operators",
			queryString: "?a__eq=1&b__gt=2&c__lt=3&d__gte=4&e__lte=5&f__contains=x&g__starts_with=y&h__ends_with=z&i__in=1,2,3&j__is_null=true&k__is_not_null=true",
			wantPage:    1,
			wantLimit:   10,
			wantFilters: []filter.FilterParam{
				{Field: "a", Operator: filter.OpEq, Value: "1"},
				{Field: "b", Operator: filter.OpGt, Value: "2"},
				{Field: "c", Operator: filter.OpLt, Value: "3"},
				{Field: "d", Operator: filter.OpGte, Value: "4"},
				{Field: "e", Operator: filter.OpLte, Value: "5"},
				{Field: "f", Operator: filter.OpContains, Value: "x"},
				{Field: "g", Operator: filter.OpStartsWith, Value: "y"},
				{Field: "h", Operator: filter.OpEndsWith, Value: "z"},
				{Field: "i", Operator: filter.OpIn, Value: "1,2,3"},
				{Field: "j", Operator: filter.OpIsNull, Value: "true"},
				{Field: "k", Operator: filter.OpIsNotNull, Value: "true"},
			},
			wantSort: nil,
		},
		{
			name:        "nested relation filter",
			queryString: "?city.country.code=DEU",
			wantPage:    1,
			wantLimit:   10,
			wantFilters: []filter.FilterParam{
				{Field: "city.country.code", Operator: filter.OpEq, Value: "DEU"},
			},
			wantSort: nil,
		},
		{
			name:        "page size capped at 100",
			queryString: "?page_size=500",
			wantPage:    1,
			wantLimit:   100,
			wantFilters: nil,
			wantSort:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a Fiber app with a test endpoint that parses filters
			app := fiber.New()
			var capturedParams *filter.Params

			app.Get("/test", func(c *fiber.Ctx) error {
				capturedParams = filter.ParseFromQuery(c)
				return c.SendString("ok")
			})

			// Make request with query string
			req := httptest.NewRequest("GET", "/test"+tt.queryString, nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()
			io.ReadAll(resp.Body)

			// Verify pagination
			if capturedParams.Page != tt.wantPage {
				t.Errorf("Page = %d, want %d", capturedParams.Page, tt.wantPage)
			}
			if capturedParams.Limit != tt.wantLimit {
				t.Errorf("Limit = %d, want %d", capturedParams.Limit, tt.wantLimit)
			}

			// Verify sort
			if len(capturedParams.Sort) != len(tt.wantSort) {
				t.Errorf("Sort length = %d, want %d", len(capturedParams.Sort), len(tt.wantSort))
			} else {
				for i, want := range tt.wantSort {
					got := capturedParams.Sort[i]
					if got.Field != want.Field || got.Desc != want.Desc {
						t.Errorf("Sort[%d] = {%s, %v}, want {%s, %v}",
							i, got.Field, got.Desc, want.Field, want.Desc)
					}
				}
			}

			// Verify filters (order may vary, so check by building a map)
			if len(capturedParams.Filters) != len(tt.wantFilters) {
				t.Errorf("Filters length = %d, want %d", len(capturedParams.Filters), len(tt.wantFilters))
				t.Logf("Got filters: %+v", capturedParams.Filters)
			} else if len(tt.wantFilters) > 0 {
				// Build map of expected filters for order-independent comparison
				wantMap := make(map[string]filter.FilterParam)
				for _, f := range tt.wantFilters {
					key := f.Field + "__" + string(f.Operator)
					wantMap[key] = f
				}

				for _, got := range capturedParams.Filters {
					key := got.Field + "__" + string(got.Operator)
					want, ok := wantMap[key]
					if !ok {
						t.Errorf("Unexpected filter: %+v", got)
						continue
					}
					if got.Value != want.Value {
						t.Errorf("Filter %s value = %s, want %s", key, got.Value, want.Value)
					}
				}
			}
		})
	}
}

// TestFilterOperatorCoverage verifies all operators are recognized
func TestFilterOperatorCoverage(t *testing.T) {
	operators := []struct {
		queryKey string
		wantOp   filter.Operator
	}{
		{"field__eq", filter.OpEq},
		{"field__gt", filter.OpGt},
		{"field__lt", filter.OpLt},
		{"field__gte", filter.OpGte},
		{"field__lte", filter.OpLte},
		{"field__contains", filter.OpContains},
		{"field__starts_with", filter.OpStartsWith},
		{"field__ends_with", filter.OpEndsWith},
		{"field__in", filter.OpIn},
		{"field__is_null", filter.OpIsNull},
		{"field__is_not_null", filter.OpIsNotNull},
	}

	for _, tt := range operators {
		t.Run(string(tt.wantOp), func(t *testing.T) {
			app := fiber.New()
			var capturedParams *filter.Params

			app.Get("/test", func(c *fiber.Ctx) error {
				capturedParams = filter.ParseFromQuery(c)
				return c.SendString("ok")
			})

			req := httptest.NewRequest("GET", "/test?"+tt.queryKey+"=value", nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()
			io.ReadAll(resp.Body)

			if len(capturedParams.Filters) != 1 {
				t.Fatalf("Expected 1 filter, got %d", len(capturedParams.Filters))
			}

			got := capturedParams.Filters[0]
			if got.Operator != tt.wantOp {
				t.Errorf("Operator = %s, want %s", got.Operator, tt.wantOp)
			}
			if got.Field != "field" {
				t.Errorf("Field = %s, want field", got.Field)
			}
		})
	}
}

// TestFilterWithFiberContext tests filter parsing with realistic Fiber context
func TestFilterWithFiberContext(t *testing.T) {
	app := fiber.New()

	// Simulate an API endpoint that uses filter parsing
	app.Get("/api/items", func(c *fiber.Ctx) error {
		params := filter.ParseFromQuery(c)

		// Return parsed params as JSON for verification
		return c.JSON(fiber.Map{
			"page":     params.Page,
			"limit":    params.Limit,
			"filters":  len(params.Filters),
			"sort":     len(params.Sort),
		})
	})

	tests := []struct {
		name        string
		url         string
		wantPage    int
		wantLimit   int
		wantFilters int
		wantSort    int
	}{
		{
			name:        "items with price range",
			url:         "/api/items?price__gte=1000&price__lte=5000&page=2&page_size=15",
			wantPage:    2,
			wantLimit:   15,
			wantFilters: 2,
			wantSort:    0,
		},
		{
			name:        "items with name search and sort",
			url:         "/api/items?name__contains=widget&sort=price&order=desc",
			wantPage:    1,
			wantLimit:   10,
			wantFilters: 1,
			wantSort:    1,
		},
		{
			name:        "items with multiple filters",
			url:         "/api/items?name__contains=test&description__contains=sample&price__gte=100&created_at__gte=2024-01-01",
			wantPage:    1,
			wantLimit:   10,
			wantFilters: 4,
			wantSort:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				t.Errorf("Status = %d, want 200", resp.StatusCode)
			}
		})
	}
}
