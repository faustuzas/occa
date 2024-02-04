package slices

func Map[I any, O any](input []I, transform func(I) O) []O {
	result := make([]O, len(input))
	for i := range result {
		result[i] = transform(input[i])
	}
	return result
}

func MapE[I any, O any](input []I, transform func(I) (O, error)) ([]O, error) {
	result := make([]O, len(input))
	for i := range result {
		entry, err := transform(input[i])
		if err != nil {
			return nil, err
		}
		result[i] = entry
	}
	return result, nil
}

func FilterMap[I any, O any](input []I, transform func(I) (O, bool)) []O {
	result := make([]O, 0, len(input))
	for _, in := range input {
		entry, ok := transform(in)
		if ok {
			result = append(result, entry)
		}
	}
	return result
}
