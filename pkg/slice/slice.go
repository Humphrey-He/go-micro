package slice

import "slices"

// Contains returns true if the slice contains the element.
func Contains[E comparable](slice []E, element E) bool {
	return slices.Contains(slice, element)
}

// Unique removes duplicate elements from the slice.
func Unique[E comparable](slice []E) []E {
	seen := make(map[E]struct{}, len(slice))
	result := make([]E, 0, len(slice))
	for _, v := range slice {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}

// Filter returns a new slice containing elements that satisfy the predicate.
func Filter[E any](slice []E, predicate func(E) bool) []E {
	return slices.DeleteFunc(slices.Clone(slice), func(e E) bool {
		return !predicate(e)
	})
}

// Map transforms each element of the slice.
func Map[E any, R any](slice []E, transform func(E) R) []R {
	result := make([]R, len(slice))
	for i, v := range slice {
		result[i] = transform(v)
	}
	return result
}

// Chunk splits a slice into chunks of the specified size.
func Chunk[E any](slice []E, size int) [][]E {
	if size <= 0 {
		return nil
	}
	result := make([][]E, 0, (len(slice)+size-1)/size)
	for i := 0; i < len(slice); i += size {
		end := i + size
		if end > len(slice) {
			end = len(slice)
		}
		result = append(result, slices.Clone(slice[i:end]))
	}
	return result
}

// Flatten flattens a slice of slices into a single slice.
func Flatten[E any](slices [][]E) []E {
	result := make([]E, 0)
	for _, slice := range slices {
		result = append(result, slice...)
	}
	return result
}

// GroupBy groups elements by a key function.
func GroupBy[E any, K comparable](slice []E, keyFunc func(E) K) map[K][]E {
	result := make(map[K][]E)
	for _, v := range slice {
		k := keyFunc(v)
		result[k] = append(result[k], v)
	}
	return result
}

// Difference returns elements in a that are not in b.
func Difference[E comparable](a, b []E) []E {
	bSet := make(map[E]struct{}, len(b))
	for _, v := range b {
		bSet[v] = struct{}{}
	}
	result := make([]E, 0)
	for _, v := range a {
		if _, ok := bSet[v]; !ok {
			result = append(result, v)
		}
	}
	return result
}

// Intersection returns elements common to both slices.
func Intersection[E comparable](a, b []E) []E {
	bSet := make(map[E]struct{}, len(b))
	for _, v := range b {
		bSet[v] = struct{}{}
	}
	result := make([]E, 0)
	for _, v := range a {
		if _, ok := bSet[v]; ok {
			result = append(result, v)
		}
	}
	return result
}

// Reverse reverses the slice in place.
func Reverse[E any](slice []E) []E {
	slices.Reverse(slice)
	return slice
}

// Sort sorts the slice using the provided compare function (return negative if a < b, zero if a == b, positive if a > b).
func Sort[E any](slice []E, compare func(a, b E) int) []E {
	slices.SortFunc(slice, func(a, b E) int {
		return compare(a, b)
	})
	return slice
}

// Sorted returns true if the slice is already sorted.
func Sorted[E any](slice []E, compare func(a, b E) int) bool {
	return slices.IsSortedFunc(slice, func(a, b E) int {
		return compare(a, b)
	})
}

// Min returns the minimum element.
func Min[E any](slice []E, less func(E, E) bool) (E, bool) {
	if len(slice) == 0 {
		var zero E
		return zero, false
	}
	min := slice[0]
	for i := 1; i < len(slice); i++ {
		if less(slice[i], min) {
			min = slice[i]
		}
	}
	return min, true
}

// Max returns the maximum element.
func Max[E any](slice []E, less func(E, E) bool) (E, bool) {
	if len(slice) == 0 {
		var zero E
		return zero, false
	}
	max := slice[0]
	for i := 1; i < len(slice); i++ {
		if less(max, slice[i]) {
			max = slice[i]
		}
	}
	return max, true
}

// Count counts elements that satisfy the predicate.
func Count[E any](slice []E, predicate func(E) bool) int {
	count := 0
	for _, v := range slice {
		if predicate(v) {
			count++
		}
	}
	return count
}

// Any returns true if any element satisfies the predicate.
func Any[E any](slice []E, predicate func(E) bool) bool {
	for _, v := range slice {
		if predicate(v) {
			return true
		}
	}
	return false
}

// All returns true if all elements satisfy the predicate.
func All[E any](slice []E, predicate func(E) bool) bool {
	for _, v := range slice {
		if !predicate(v) {
			return false
		}
	}
	return true
}

// None returns true if no element satisfies the predicate.
func None[E any](slice []E, predicate func(E) bool) bool {
	return !Any(slice, predicate)
}

// First returns the first element that satisfies the predicate.
func First[E any](slice []E, predicate func(E) bool) (E, bool) {
	for _, v := range slice {
		if predicate(v) {
			return v, true
		}
	}
	var zero E
	return zero, false
}

// Drop drops the first n elements.
func Drop[E any](slice []E, n int) []E {
	if n <= 0 {
		return slices.Clone(slice)
	}
	if n >= len(slice) {
		return []E{}
	}
	return slices.Clone(slice[n:])
}

// Take takes the first n elements.
func Take[E any](slice []E, n int) []E {
	if n <= 0 {
		return []E{}
	}
	if n >= len(slice) {
		return slices.Clone(slice)
	}
	return slices.Clone(slice[:n])
}
