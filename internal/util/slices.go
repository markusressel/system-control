package util

func FilterFunc[S ~[]E, E any](x S, filter func(e E) bool) []E {
	result := make([]E, 0)
	for _, v := range x {
		if filter(v) {
			result = append(result, v)
		}
	}
	return result
}
