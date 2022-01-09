package amiibo

// extractBits extracts 'amount' bits from the given 'number' starting on 'position'.
func extractBits(number, amount, position int) int {
	return ((((1 << amount) - 1) << position) & number) >> position
}
