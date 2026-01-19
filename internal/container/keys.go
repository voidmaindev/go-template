package container

import "fmt"

// Key is a typed container key that provides compile-time type safety.
// Instead of using string constants, domains define typed keys:
//
//	var HandlerKey = container.Key[*Handler]("user.handler")
//
// Usage:
//
//	handler := user.HandlerKey.MustGet(c)  // Returns *Handler directly
type Key[T any] string

// Get retrieves the component with compile-time type safety.
// Returns the zero value and false if not found or wrong type.
func (k Key[T]) Get(c *Container) (T, bool) {
	return GetTyped[T](c, string(k))
}

// MustGet retrieves the component or panics if not found.
// The panic message includes a hint about domain registration order.
func (k Key[T]) MustGet(c *Container) T {
	return MustGetTyped[T](c, string(k))
}

// Set stores the component in the container.
func (k Key[T]) Set(c *Container, value T) {
	c.Set(string(k), value)
}

// String returns the underlying string key.
func (k Key[T]) String() string {
	return string(k)
}

// MustGetAs retrieves a component and asserts it to a different type.
// Useful when storing as interface but retrieving as concrete type.
// Panics if the component is not found or cannot be type-asserted.
func MustGetAs[T any](c *Container, key string) T {
	comp, ok := c.components[key]
	if !ok {
		panic(fmt.Sprintf("component not found: %s (ensure domain is registered before dependent domains)", key))
	}
	typed, ok := comp.(T)
	if !ok {
		panic(fmt.Sprintf("component %s is not of expected type %T, got %T", key, *new(T), comp))
	}
	return typed
}
