# ADR-012: Django-Style Filtering System

## Status

Accepted

## Context

List endpoints need flexible filtering capabilities:
- Filter by field values (exact match, ranges, contains)
- Sort by multiple fields
- Paginate results
- Filter across relations (e.g., cities by country name)

Requirements:
- Type-safe configuration (prevent invalid filter combinations)
- SQL injection prevention
- Consistent API across all domains
- Easy to add new filterable fields

## Decision

### Operator Syntax

Use Django-style double-underscore operators in query parameters:

```
?field__operator=value
```

Default operator is `eq` (equality):
```
?name=John        # Equivalent to ?name__eq=John
?status__in=active,pending
?price__gte=100
?created_at__gt=2024-01-01
```

### Supported Operators (11)

| Operator | SQL | Example |
|----------|-----|---------|
| `eq` | `= value` | `?status__eq=active` |
| `gt` | `> value` | `?price__gt=100` |
| `lt` | `< value` | `?price__lt=50` |
| `gte` | `>= value` | `?age__gte=18` |
| `lte` | `<= value` | `?age__lte=65` |
| `contains` | `ILIKE %value%` | `?name__contains=smith` |
| `starts_with` | `ILIKE value%` | `?email__starts_with=john` |
| `ends_with` | `ILIKE %value` | `?email__ends_with=@gmail.com` |
| `in` | `IN (...)` | `?status__in=active,pending,review` |
| `is_null` | `IS NULL` | `?deleted_at__is_null=true` |
| `is_not_null` | `IS NOT NULL` | `?email_verified_at__is_not_null=true` |

### FilterConfig Structure

Each filterable model defines its configuration:

```go
type FieldConfig struct {
    DBColumn   string     // Database column name (SQL injection safe)
    Type       FieldType  // TypeString, TypeNumber, TypeDate, TypeBool
    Operators  []Operator // Allowed operators for this field
    Sortable   bool       // Can be used for sorting
    Relation   string     // Related table name (e.g., "Country")
    RelationFK string     // Foreign key column for joins
}

type Config struct {
    TableName             string
    Fields                map[string]FieldConfig
    AllowedRelationFields map[string][]string  // Whitelist for relation fields
}
```

### Filterable Interface

Models implement `Filterable` to enable filtering:

```go
type Filterable interface {
    FilterConfig() Config
}
```

Example implementation:

```go
func (City) FilterConfig() filter.Config {
    return filter.Config{
        TableName: "cities",
        Fields: map[string]filter.FieldConfig{
            "id":         {DBColumn: "id", Type: filter.TypeNumber, Operators: filter.NumberOps, Sortable: true},
            "name":       {DBColumn: "name", Type: filter.TypeString, Operators: filter.StringOps, Sortable: true},
            "country_id": {DBColumn: "country_id", Type: filter.TypeNumber, Operators: filter.NumberOps, Sortable: true},
            "country":    {Relation: "Country", RelationFK: "country_id"},
            "created_at": {DBColumn: "created_at", Type: filter.TypeDate, Operators: filter.DateOps, Sortable: true},
        },
        AllowedRelationFields: map[string][]string{
            "country": {"id", "name", "code"},  // Explicit whitelist
        },
    }
}
```

### Pre-defined Operator Sets

Common operator combinations for each type:

```go
var (
    StringOps = []Operator{OpEq, OpContains, OpStartsWith, OpEndsWith, OpIn, OpIsNull, OpIsNotNull}
    NumberOps = []Operator{OpEq, OpGt, OpLt, OpGte, OpLte, OpIn, OpIsNull, OpIsNotNull}
    DateOps   = []Operator{OpEq, OpGt, OpLt, OpGte, OpLte, OpIsNull, OpIsNotNull}
    BoolOps   = []Operator{OpEq, OpIsNull, OpIsNotNull}
)
```

### Relation Filtering

Filter by related model fields using dot notation:

```
?country.name__contains=United
?country.code__eq=US
```

**Security**: Only explicitly whitelisted fields can be queried:

```go
AllowedRelationFields: map[string][]string{
    "country": {"id", "name", "code"},  // Only these fields allowed
}
```

### Query Building Flow

```
              Query Parameters
                    │
                    ▼
        ┌───────────────────────┐
        │ ParseFromQuery()      │
        │ Extract filters, sort,│
        │ pagination            │
        └───────────────────────┘
                    │
                    ▼
        ┌───────────────────────┐
        │ Validate operators    │
        │ against config        │──── Invalid ───▶ Error
        └───────────────────────┘
                    │
                    ▼
        ┌───────────────────────┐
        │ Build GORM query      │
        │ Apply filters         │
        │ Apply sorting         │
        │ Apply pagination      │
        └───────────────────────┘
                    │
                    ▼
               Execute Query
```

### Repository Integration

```go
func (r *BaseRepository[T]) FindAllFiltered(ctx context.Context, params *filter.Params) ([]T, int64, error) {
    var model T
    filterable, ok := any(model).(filter.Filterable)
    if !ok {
        return nil, 0, errors.New("model does not implement Filterable")
    }

    config := filterable.FilterConfig()

    // Count query (filters only, no pagination)
    var total int64
    countQuery := r.db.WithContext(ctx).Model(&model)
    countQuery = filter.ApplyFiltersOnly(countQuery, config, params)
    countQuery.Count(&total)

    // Data query (filters + sorting + pagination)
    var results []T
    query := r.db.WithContext(ctx).Model(&model)
    query = filter.Apply(query, config, params)
    query.Find(&results)

    return results, total, nil
}
```

### Handler Usage

```go
func (h *Handler) List(c *fiber.Ctx) error {
    pagination := common.PaginationFromQuery(c)
    filterParams := common.FilterParamsFromQuery(c, City{}.FilterConfig())

    cities, total, err := h.service.List(c.Context(),
        common.WithFilter(filterParams),
        common.WithPagination(pagination),
    )
    if err != nil {
        return common.HandleError(c, err)
    }

    return common.PaginatedResponse(c, cities, total, pagination)
}
```

### Pagination

Standard pagination parameters:

| Parameter | Default | Description |
|-----------|---------|-------------|
| `page` | 1 | Page number (1-indexed) |
| `page_size` | 20 | Items per page (max 100) |

### Sorting

Sorting parameters:

| Parameter | Default | Description |
|-----------|---------|-------------|
| `sort` | `id` | Field to sort by |
| `order` | `asc` | Sort direction (`asc` or `desc`) |

Relation sorting supported:
```
?sort=country.name&order=desc
```

## SQL Injection Prevention

1. **Column Whitelist**: Only `DBColumn` values from config are used in SQL
2. **Relation Whitelist**: Only `AllowedRelationFields` can be queried
3. **Parameterized Queries**: Values passed as GORM parameters, never interpolated
4. **No Dynamic Table Names**: Table names come from config, not user input

## Consequences

### Positive

- **Familiar Syntax**: Django-style operators are well-known
- **Type Safety**: Config prevents invalid operator/field combinations
- **SQL Safe**: Whitelist approach prevents injection
- **Extensible**: Easy to add new fields and operators
- **Consistent**: Same filtering works across all domains

### Negative

- **Config Boilerplate**: Each model needs FilterConfig implementation
- **Learning Curve**: Operators must be memorized or documented
- **Query Complexity**: Relation filters generate JOINs

### Neutral

- **Case Sensitivity**: String comparisons use ILIKE (case-insensitive)
- **AND Logic**: Multiple filters are combined with AND (no OR support)

## Related

- ADR-002: Generic Repository Pattern (repository integration)
- ADR-004: OpenAPI-First Design (API documentation)
