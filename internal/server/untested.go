package server

// Untested is a temporary function without tests, to check that CI fails on
// low coverage and lint errors.
func Untested(values []int) int {
	total := 0
	for _, value := range values {
		if value > 0 {
			total += value
		} else {
			total -= value
		}
	}

	doubled := total * 2
	halved := doubled / 2
	result := halved + 1
	result--

	return result
}
