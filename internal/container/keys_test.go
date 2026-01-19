package container

import "testing"

func TestKey_SetAndGet(t *testing.T) {
	c := &Container{
		components: make(map[string]any),
	}

	type TestService struct {
		Name string
	}

	key := Key[*TestService]("test.service")
	svc := &TestService{Name: "test"}

	// Set using typed key
	key.Set(c, svc)

	// Get using typed key
	result := key.MustGet(c)
	if result.Name != "test" {
		t.Errorf("Key.MustGet() = %v, want %v", result.Name, "test")
	}
}

func TestKey_Get_NotFound(t *testing.T) {
	c := &Container{
		components: make(map[string]any),
	}

	key := Key[*string]("nonexistent.key")

	result, ok := key.Get(c)
	if ok {
		t.Error("Key.Get() should return ok=false for nonexistent key")
	}
	if result != nil {
		t.Error("Key.Get() should return zero value for nonexistent key")
	}
}

func TestKey_MustGet_Panics(t *testing.T) {
	c := &Container{
		components: make(map[string]any),
	}

	key := Key[*string]("nonexistent.key")

	defer func() {
		r := recover()
		if r == nil {
			t.Error("Key.MustGet() should panic for nonexistent key")
		}
		// Check that the panic message contains the helpful hint
		msg, ok := r.(string)
		if ok && len(msg) > 0 {
			if !contains(msg, "ensure domain is registered") {
				t.Errorf("Panic message should contain helpful hint, got: %s", msg)
			}
		}
	}()

	key.MustGet(c)
}

func TestKey_String(t *testing.T) {
	key := Key[*string]("test.key")

	if key.String() != "test.key" {
		t.Errorf("Key.String() = %v, want %v", key.String(), "test.key")
	}
}

func TestKey_TypeSafety(t *testing.T) {
	c := &Container{
		components: make(map[string]any),
	}

	// Define typed keys for different types
	stringKey := Key[string]("string.key")
	intKey := Key[int]("int.key")

	// Set values
	stringKey.Set(c, "hello")
	intKey.Set(c, 42)

	// Get values with correct types
	strVal := stringKey.MustGet(c)
	intVal := intKey.MustGet(c)

	if strVal != "hello" {
		t.Errorf("string value = %v, want %v", strVal, "hello")
	}
	if intVal != 42 {
		t.Errorf("int value = %v, want %v", intVal, 42)
	}
}

// testGreeter is used for testing interface type assertions
type testGreeter struct{}

func (g testGreeter) Greet() string {
	return "Hello!"
}

// Greeter interface for testing
type testGreeterInterface interface {
	Greet() string
}

func TestMustGetAs(t *testing.T) {
	c := &Container{
		components: make(map[string]any),
	}

	// Store concrete type
	c.Set("greeter", testGreeter{})

	// Get as interface using MustGetAs
	greeter := MustGetAs[testGreeterInterface](c, "greeter")
	if greeter.Greet() != "Hello!" {
		t.Errorf("MustGetAs() returned wrong implementation")
	}
}

func TestMustGetAs_Panics_NotFound(t *testing.T) {
	c := &Container{
		components: make(map[string]any),
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("MustGetAs() should panic for nonexistent key")
		}
	}()

	MustGetAs[string](c, "nonexistent.key")
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
