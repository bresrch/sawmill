package sawmill

import "strconv"

// FastNumericBuffer provides optimized numeric to string conversion
type FastNumericBuffer struct {
	buf [24]byte // Enough for any int64
}

// FormatInt converts an integer to string using a pre-allocated buffer
func (f *FastNumericBuffer) FormatInt(n int) string {
	return strconv.Itoa(n)
}

// FormatInt64 converts an int64 to string using a pre-allocated buffer
func (f *FastNumericBuffer) FormatInt64(n int64) string {
	return strconv.FormatInt(n, 10)
}

// Global instance for reuse
var numericBuffer FastNumericBuffer

// FastItoa converts integer to string efficiently
func FastItoa(n int) string {
	return numericBuffer.FormatInt(n)
}

// FastInt64toa converts int64 to string efficiently
func FastInt64toa(n int64) string {
	return numericBuffer.FormatInt64(n)
}
