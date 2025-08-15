package client

type Named interface {
	GetName() string
}

func findByName[T Named](items []T, name string) (T, bool) {
	for _, item := range items {
		if item.GetName() == name {
			return item, true
		}
	}
	var zero T
	return zero, false
}
