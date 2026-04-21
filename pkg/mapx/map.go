package mapx

// Keys returns all keys from the map.
func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Values returns all values from the map.
func Values[K comparable, V any](m map[K]V) []V {
	values := make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

// Merge merges multiple maps into one.
func Merge[K comparable, V any](maps ...map[K]V) map[K]V {
	result := make(map[K]V)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

// Clone creates a shallow copy of the map.
func Clone[K comparable, V any](m map[K]V) map[K]V {
	result := make(map[K]V, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}

// Filter keeps only entries that satisfy the predicate.
func Filter[K comparable, V any](m map[K]V, predicate func(K, V) bool) map[K]V {
	result := make(map[K]V)
	for k, v := range m {
		if predicate(k, v) {
			result[k] = v
		}
	}
	return result
}

// MapValues transforms all values in the map.
func MapValues[K comparable, V any, R any](m map[K]V, transform func(V) R) map[K]R {
	result := make(map[K]R, len(m))
	for k, v := range m {
		result[k] = transform(v)
	}
	return result
}

// MapKeys transforms all keys in the map.
func MapKeys[K comparable, V any, R comparable](m map[K]V, transform func(K) R) map[R]V {
	result := make(map[R]V, len(m))
	for k, v := range m {
		result[transform(k)] = v
	}
	return result
}

// Invert swaps keys and values.
func Invert[K comparable, V comparable](m map[K]V) map[V]K {
	result := make(map[V]K, len(m))
	for k, v := range m {
		result[v] = k
	}
	return result
}

// Pick extracts entries with specified keys.
func Pick[K comparable, V any](m map[K]V, keys []K) map[K]V {
	result := make(map[K]V)
	for _, k := range keys {
		if v, ok := m[k]; ok {
			result[k] = v
		}
	}
	return result
}

// Omit excludes entries with specified keys.
func Omit[K comparable, V any](m map[K]V, keys []K) map[K]V {
	exclude := make(map[K]bool)
	for _, k := range keys {
		exclude[k] = true
	}
	result := make(map[K]V)
	for k, v := range m {
		if !exclude[k] {
			result[k] = v
		}
	}
	return result
}
