package util

type Option[T any] func(*T)

func ApplyOptions[T any](target *T, opts ...Option[T]) {
	for _, opt := range opts {
		opt(target)
	}
}
