package common

// MapSlice converts a slice of T to a slice of R using a mapper function.
func MapSlice[T any, R any](items []T, mapper func(*T) R) []R {
	if items == nil {
		return nil
	}
	result := make([]R, len(items))
	for i := range items {
		result[i] = mapper(&items[i])
	}
	return result
}
