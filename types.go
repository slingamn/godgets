// Copyright (c) 2020 Shivaram Lingamneni
// released under the 0BSD license

package godgets

type empty struct{}

type HashSet[T comparable] map[T]empty

func (s HashSet[T]) Has(elem T) bool {
	_, ok := s[elem]
	return ok
}

func (s HashSet[T]) Add(elem T) {
	s[elem] = empty{}
}

func (s HashSet[T]) Remove(elem T) {
	delete(s, elem)
}

func CopyMap[K comparable, V any](input map[K]V) (result map[K]V) {
	result = make(map[K]V, len(input))
	for key, value := range input {
		result[key] = value
	}
	return
}

func CopySlice[T any](slice []T) (result []T) {
	result = make([]T, len(slice))
	copy(result, slice)
	return
}

func SliceContains[T comparable](slice []T, elem T) (result bool) {
	for _, t := range slice {
		if elem == t {
			return true
		}
	}
	return false
}

// reverse the order of a slice in place
func ReverseSlice[T any](slice []T) {
	for i, j := 0, len(slice)-1; i < j; i, j = i+1, j-1 {
		slice[i], slice[j] = slice[j], slice[i]
	}
}
