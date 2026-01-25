// Package ptr provides generic pointer utility functions.
package ptr

// To returns a pointer to the given value.
func To[T any](v T) *T {
	return &v
}

// Deref returns the value pointed to, or the zero value if nil.
func Deref[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

// DerefOr returns the value pointed to, or the default if nil.
func DerefOr[T any](p *T, def T) T {
	if p == nil {
		return def
	}
	return *p
}
