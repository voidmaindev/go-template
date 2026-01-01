package utils

import "github.com/jinzhu/copier"

// UpdateModel copies non-nil/non-zero fields from src (DTO) to dst (Model).
// This eliminates the need for manual field-by-field checks in service Update functions.
// Only fields that are set (non-nil pointers, non-zero values) will be copied.
func UpdateModel(dst, src any) error {
	return copier.CopyWithOption(dst, src, copier.Option{
		IgnoreEmpty: true,
		DeepCopy:    true,
	})
}
