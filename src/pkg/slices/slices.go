package slices

func Map[I any, O any](input []I, transform func(I) O) []O {
	result := make([]O, len(input))
	for i := range result {
		result[i] = transform(input[i])
	}
	return result
}
