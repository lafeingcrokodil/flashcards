package math

// Percent computes a percentage rounded down to the nearest integer.
// Returns 0 if the denominator is 0.
func Percent(numerator, denominator int) int {
	if denominator == 0 {
		return 0
	}
	return 100 * numerator / denominator
}
