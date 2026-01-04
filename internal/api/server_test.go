package api

import (
	"context"
	"testing"
)

func TestPtr(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		s := "hello"
		p := ptr(s)
		if *p != s {
			t.Errorf("ptr() = %s, want %s", *p, s)
		}
	})

	t.Run("int", func(t *testing.T) {
		i := 42
		p := ptr(i)
		if *p != i {
			t.Errorf("ptr() = %d, want %d", *p, i)
		}
	})

	t.Run("int64", func(t *testing.T) {
		i := int64(100)
		p := ptr(i)
		if *p != i {
			t.Errorf("ptr() = %d, want %d", *p, i)
		}
	})

	t.Run("bool", func(t *testing.T) {
		b := true
		p := ptr(b)
		if *p != b {
			t.Errorf("ptr() = %v, want %v", *p, b)
		}
	})
}

func TestPtrToString(t *testing.T) {
	tests := []struct {
		name string
		s    *string
		want string
	}{
		{"nil", nil, ""},
		{"empty", ptr(""), ""},
		{"value", ptr("hello"), "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ptrToString(tt.s)
			if got != tt.want {
				t.Errorf("ptrToString() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestIntToString(t *testing.T) {
	tests := []struct {
		i    int
		want string
	}{
		{0, "0"},
		{42, "42"},
		{-1, "-1"},
		{1000, "1000"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := intToString(tt.i)
			if got != tt.want {
				t.Errorf("intToString(%d) = %s, want %s", tt.i, got, tt.want)
			}
		})
	}
}

func TestGetIntOrDefault(t *testing.T) {
	tests := []struct {
		name    string
		val     *int
		def     int
		want    int
	}{
		{"nil uses default", nil, 10, 10},
		{"value used", ptr(5), 10, 5},
		{"zero value used", ptr(0), 10, 0},
		{"negative value used", ptr(-1), 10, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getIntOrDefault(tt.val, tt.def)
			if got != tt.want {
				t.Errorf("getIntOrDefault() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestBuildFilterParams(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		params := buildFilterParams(nil, nil, nil, nil)

		if params.Page != 1 {
			t.Errorf("Page = %d, want 1", params.Page)
		}
		if params.Limit != 10 {
			t.Errorf("Limit = %d, want 10", params.Limit)
		}
		if len(params.Sort) != 0 {
			t.Errorf("Sort length = %d, want 0", len(params.Sort))
		}
	})

	t.Run("custom page and size", func(t *testing.T) {
		page := 3
		size := 25
		params := buildFilterParams(&page, &size, nil, nil)

		if params.Page != 3 {
			t.Errorf("Page = %d, want 3", params.Page)
		}
		if params.Limit != 25 {
			t.Errorf("Limit = %d, want 25", params.Limit)
		}
	})

	t.Run("sort ascending", func(t *testing.T) {
		sort := "name"
		order := "asc"
		params := buildFilterParams(nil, nil, &sort, &order)

		if len(params.Sort) != 1 {
			t.Fatalf("Sort length = %d, want 1", len(params.Sort))
		}
		if params.Sort[0].Field != "name" {
			t.Errorf("Sort field = %s, want name", params.Sort[0].Field)
		}
		if params.Sort[0].Desc != false {
			t.Errorf("Sort desc = %v, want false", params.Sort[0].Desc)
		}
	})

	t.Run("sort descending", func(t *testing.T) {
		sort := "created_at"
		order := "desc"
		params := buildFilterParams(nil, nil, &sort, &order)

		if len(params.Sort) != 1 {
			t.Fatalf("Sort length = %d, want 1", len(params.Sort))
		}
		if params.Sort[0].Field != "created_at" {
			t.Errorf("Sort field = %s, want created_at", params.Sort[0].Field)
		}
		if params.Sort[0].Desc != true {
			t.Errorf("Sort desc = %v, want true", params.Sort[0].Desc)
		}
	})

	t.Run("sort with nil order defaults to asc", func(t *testing.T) {
		sort := "name"
		params := buildFilterParams(nil, nil, &sort, nil)

		if len(params.Sort) != 1 {
			t.Fatalf("Sort length = %d, want 1", len(params.Sort))
		}
		if params.Sort[0].Desc != false {
			t.Errorf("Sort desc = %v, want false (default asc)", params.Sort[0].Desc)
		}
	})

	t.Run("empty sort string ignored", func(t *testing.T) {
		sort := ""
		params := buildFilterParams(nil, nil, &sort, nil)

		if len(params.Sort) != 0 {
			t.Errorf("Sort length = %d, want 0 for empty sort", len(params.Sort))
		}
	})
}

func TestFiberContext(t *testing.T) {
	t.Run("nil context returns nil", func(t *testing.T) {
		ctx := context.Background()
		fc := getFiberContext(ctx)
		if fc != nil {
			t.Error("expected nil fiber context from empty context")
		}
	})

	t.Run("WithFiberContext and getFiberContext roundtrip", func(t *testing.T) {
		// We can't easily test with a real fiber.Ctx without setting up Fiber,
		// but we can at least verify the nil case and type safety
		ctx := context.Background()

		// This would work with a real fiber.Ctx:
		// fiberCtx := &fiber.Ctx{} // not possible without app
		// ctx = WithFiberContext(ctx, fiberCtx)
		// if getFiberContext(ctx) != fiberCtx { ... }

		// For now, just verify nil handling
		ctx = WithFiberContext(ctx, nil)
		fc := getFiberContext(ctx)
		if fc != nil {
			t.Error("expected nil fiber context for nil input")
		}
	})
}
