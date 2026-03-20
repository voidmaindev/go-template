package container

import (
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestContainer_SetAndGet(t *testing.T) {
	c := &Container{
		components: make(map[string]any),
	}

	// Set a component
	c.Set("test.key", "test value")

	// Get it back
	value := c.Get("test.key")
	if value != "test value" {
		t.Errorf("Get() = %v, want %v", value, "test value")
	}
}

func TestContainer_Get_NotFound(t *testing.T) {
	c := &Container{
		components: make(map[string]any),
	}

	value := c.Get("nonexistent.key")
	if value != nil {
		t.Errorf("Get() = %v, want nil for nonexistent key", value)
	}
}

func TestContainer_MustGet(t *testing.T) {
	c := &Container{
		components: make(map[string]any),
	}

	c.Set("test.key", "test value")

	value := c.MustGet("test.key")
	if value != "test value" {
		t.Errorf("MustGet() = %v, want %v", value, "test value")
	}
}

func TestContainer_MustGet_Panics(t *testing.T) {
	c := &Container{
		components: make(map[string]any),
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("MustGet() should panic for nonexistent key")
		}
	}()

	c.MustGet("nonexistent.key")
}

func TestMustGetTyped(t *testing.T) {
	c := &Container{
		components: make(map[string]any),
	}

	type TestService struct {
		Name string
	}

	svc := &TestService{Name: "test"}
	c.Set("test.service", svc)

	// Get with correct type
	result := MustGetTyped[*TestService](c, "test.service")
	if result.Name != "test" {
		t.Errorf("MustGetTyped() = %v, want %v", result.Name, "test")
	}
}

func TestMustGetTyped_Panics_NotFound(t *testing.T) {
	c := &Container{
		components: make(map[string]any),
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("MustGetTyped() should panic for nonexistent key")
		}
	}()

	MustGetTyped[*string](c, "nonexistent.key")
}

func TestMustGetTyped_Panics_WrongType(t *testing.T) {
	c := &Container{
		components: make(map[string]any),
	}

	c.Set("test.key", "string value")

	defer func() {
		if r := recover(); r == nil {
			t.Error("MustGetTyped() should panic for wrong type")
		}
	}()

	// Try to get string as int pointer
	MustGetTyped[*int](c, "test.key")
}

func TestGetTyped(t *testing.T) {
	c := &Container{
		components: make(map[string]any),
	}

	type TestService struct {
		Name string
	}

	svc := &TestService{Name: "test"}
	c.Set("test.service", svc)

	// Get with correct type
	result, ok := GetTyped[*TestService](c, "test.service")
	if !ok {
		t.Error("GetTyped() should return ok=true for existing key")
	}
	if result.Name != "test" {
		t.Errorf("GetTyped() = %v, want %v", result.Name, "test")
	}
}

func TestGetTyped_NotFound(t *testing.T) {
	c := &Container{
		components: make(map[string]any),
	}

	result, ok := GetTyped[*string](c, "nonexistent.key")
	if ok {
		t.Error("GetTyped() should return ok=false for nonexistent key")
	}
	if result != nil {
		t.Error("GetTyped() should return zero value for nonexistent key")
	}
}

func TestGetTyped_WrongType(t *testing.T) {
	c := &Container{
		components: make(map[string]any),
	}

	c.Set("test.key", "string value")

	result, ok := GetTyped[*int](c, "test.key")
	if ok {
		t.Error("GetTyped() should return ok=false for wrong type")
	}
	if result != nil {
		t.Error("GetTyped() should return zero value for wrong type")
	}
}

func TestContainer_AddDomain(t *testing.T) {
	c := &Container{
		domains:    make([]Domain, 0),
		components: make(map[string]any),
	}

	mockDomain := &mockDomain{name: "test"}
	c.AddDomain(mockDomain)

	if len(c.domains) != 1 {
		t.Errorf("domains length = %d, want 1", len(c.domains))
	}
}

func TestContainer_GetAllModels(t *testing.T) {
	c := &Container{
		domains:    make([]Domain, 0),
		components: make(map[string]any),
	}

	type Model1 struct{}
	type Model2 struct{}

	c.AddDomain(&mockDomain{
		name:   "domain1",
		models: []any{&Model1{}},
	})
	c.AddDomain(&mockDomain{
		name:   "domain2",
		models: []any{&Model2{}},
	})

	models := c.GetAllModels()
	if len(models) != 2 {
		t.Errorf("GetAllModels() returned %d models, want 2", len(models))
	}
}

// mockDomain implements Domain interface for testing
type mockDomain struct {
	name       string
	models     []any
	registered bool
}

func (d *mockDomain) Name() string {
	return d.name
}

func (d *mockDomain) Models() []any {
	if d.models == nil {
		return []any{}
	}
	return d.models
}

func (d *mockDomain) Register(c *Container) {
	d.registered = true
}

func (d *mockDomain) Routes(api fiber.Router, c *Container) {
	// No-op for testing
}

func TestContainer_RegisterAll(t *testing.T) {
	c := &Container{
		domains:    make([]Domain, 0),
		components: make(map[string]any),
	}

	domain1 := &mockDomain{name: "domain1"}
	domain2 := &mockDomain{name: "domain2"}

	c.AddDomain(domain1)
	c.AddDomain(domain2)

	c.RegisterAll()

	if !domain1.registered {
		t.Error("domain1 should be registered")
	}
	if !domain2.registered {
		t.Error("domain2 should be registered")
	}
}

func TestContainer_Freeze_AfterRegisterAll(t *testing.T) {
	c := &Container{
		domains:    make([]Domain, 0),
		components: make(map[string]any),
	}

	// Set is allowed before RegisterAll
	c.Set("pre.key", "before freeze")
	if c.Get("pre.key") != "before freeze" {
		t.Error("Set before RegisterAll should work")
	}

	c.AddDomain(&mockDomain{name: "test"})
	c.RegisterAll()

	// Container should be frozen after RegisterAll
	if !c.frozen {
		t.Error("container should be frozen after RegisterAll")
	}

	// Set after freeze should panic
	defer func() {
		r := recover()
		if r == nil {
			t.Error("Set() after RegisterAll should panic")
		}
		msg, ok := r.(string)
		if !ok {
			t.Errorf("panic should be a string, got %T", r)
		}
		if !contains(msg, "frozen") {
			t.Errorf("panic message should mention 'frozen', got: %s", msg)
		}
	}()

	c.Set("post.key", "after freeze")
}

func TestContainer_Freeze_SetDuringRegistration(t *testing.T) {
	// Domains should be able to call Set() during Register()
	c := &Container{
		domains:    make([]Domain, 0),
		components: make(map[string]any),
	}

	settingDomain := &mockDomainWithSet{name: "setter"}
	c.AddDomain(settingDomain)

	// This should NOT panic — domains set components during registration
	c.RegisterAll()

	if c.Get("setter.component") != "registered" {
		t.Error("domain should be able to Set during Register")
	}
}

func TestContainer_Freeze_GetStillWorks(t *testing.T) {
	c := &Container{
		domains:    make([]Domain, 0),
		components: make(map[string]any),
	}

	c.Set("key", "value")
	c.AddDomain(&mockDomain{name: "test"})
	c.RegisterAll()

	// Get should still work after freeze
	if c.Get("key") != "value" {
		t.Error("Get should still work after freeze")
	}

	// MustGet should still work
	if c.MustGet("key") != "value" {
		t.Error("MustGet should still work after freeze")
	}

	// GetTyped should still work
	val, ok := GetTyped[string](c, "key")
	if !ok || val != "value" {
		t.Error("GetTyped should still work after freeze")
	}
}

func TestContainer_Freeze_TypedKeySetPanics(t *testing.T) {
	c := &Container{
		domains:    make([]Domain, 0),
		components: make(map[string]any),
	}

	c.AddDomain(&mockDomain{name: "test"})
	c.RegisterAll()

	key := Key[string]("new.key")

	defer func() {
		if r := recover(); r == nil {
			t.Error("Key.Set() after freeze should panic")
		}
	}()

	key.Set(c, "should panic")
}

// mockDomainWithSet calls Set during Register to verify it works before freeze
type mockDomainWithSet struct {
	name string
}

func (d *mockDomainWithSet) Name() string      { return d.name }
func (d *mockDomainWithSet) Models() []any      { return []any{} }
func (d *mockDomainWithSet) Routes(api fiber.Router, c *Container) {}
func (d *mockDomainWithSet) Register(c *Container) {
	c.Set("setter.component", "registered")
}

func TestContainer_MultipleTypes(t *testing.T) {
	c := &Container{
		components: make(map[string]any),
	}

	// Store different types
	c.Set("string.key", "string value")
	c.Set("int.key", 42)
	c.Set("bool.key", true)
	c.Set("slice.key", []string{"a", "b", "c"})

	// Retrieve with GetTyped
	strVal, ok := GetTyped[string](c, "string.key")
	if !ok || strVal != "string value" {
		t.Errorf("string value = %v, want %v", strVal, "string value")
	}

	intVal, ok := GetTyped[int](c, "int.key")
	if !ok || intVal != 42 {
		t.Errorf("int value = %v, want %v", intVal, 42)
	}

	boolVal, ok := GetTyped[bool](c, "bool.key")
	if !ok || boolVal != true {
		t.Errorf("bool value = %v, want %v", boolVal, true)
	}

	sliceVal, ok := GetTyped[[]string](c, "slice.key")
	if !ok || len(sliceVal) != 3 {
		t.Errorf("slice value = %v, want 3 elements", sliceVal)
	}
}
