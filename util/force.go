package util

func WithoutErr[T any](v T, _ error) T {
	return v
}
