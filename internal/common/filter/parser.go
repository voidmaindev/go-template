package filter

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// Reserved query parameter names that should not be treated as filters
var reservedParams = map[string]bool{
	"page":      true,
	"page_size": true,
	"sort":      true,
	"order":     true,
}

// ParseFromQuery extracts filter params from query string
// Syntax: ?field__operator=value (e.g., ?name__contains=Ber)
// Default operator is "eq" if not specified (?name=Berlin)
func ParseFromQuery(c *fiber.Ctx) *Params {
	params := &Params{
		Page:  c.QueryInt("page", 1),
		Limit: c.QueryInt("page_size", 10),
	}

	// Ensure valid pagination values
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 {
		params.Limit = 10
	}
	if params.Limit > 100 {
		params.Limit = 100 // Max limit to prevent abuse
	}

	// Parse sort
	if sort := c.Query("sort"); sort != "" {
		order := c.Query("order", "asc")
		params.Sort = append(params.Sort, SortParam{
			Field: sort,
			Desc:  strings.ToLower(order) == "desc",
		})
	}

	// Parse filters from all query params
	c.Context().QueryArgs().VisitAll(func(key, value []byte) {
		keyStr := string(key)
		if reservedParams[keyStr] {
			return
		}

		field, op := parseFieldOperator(keyStr)
		params.Filters = append(params.Filters, FilterParam{
			Field:    field,
			Operator: op,
			Value:    string(value),
		})
	})

	return params
}

// parseFieldOperator splits "field__operator" into parts
// Returns field name and operator (default: OpEq)
// Supports dot notation for relations: "example_country.name__contains"
func parseFieldOperator(key string) (string, Operator) {
	// Split by double underscore for operator
	parts := strings.Split(key, "__")
	if len(parts) == 1 {
		return parts[0], OpEq
	}

	// Last part after __ is the operator
	field := strings.Join(parts[:len(parts)-1], "__")
	op := Operator(parts[len(parts)-1])

	// Validate operator, default to eq if invalid
	if !op.IsValid() {
		return key, OpEq
	}

	return field, op
}

// ParseFromMap creates Params from a map (useful for testing or programmatic usage)
func ParseFromMap(m map[string]string) *Params {
	params := DefaultParams()

	for key, value := range m {
		switch key {
		case "page":
			// Parse page
			if v := parseIntSafe(value); v > 0 {
				params.Page = v
			}
		case "page_size":
			// Parse page size
			if v := parseIntSafe(value); v > 0 && v <= 100 {
				params.Limit = v
			}
		case "sort":
			// Parse sort (handled separately with order)
		case "order":
			// Skip, handled with sort
		default:
			field, op := parseFieldOperator(key)
			params.Filters = append(params.Filters, FilterParam{
				Field:    field,
				Operator: op,
				Value:    value,
			})
		}
	}

	// Handle sort separately to pair with order
	if sort, ok := m["sort"]; ok && sort != "" {
		order := m["order"]
		params.Sort = append(params.Sort, SortParam{
			Field: sort,
			Desc:  strings.ToLower(order) == "desc",
		})
	}

	return params
}

func parseIntSafe(s string) int {
	var result int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + int(c-'0')
		} else {
			return 0
		}
	}
	return result
}
