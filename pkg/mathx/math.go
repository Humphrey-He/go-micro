package mathx

import (
	"math"
	"math/rand"
)

// Clamp clamps value to the range [min, max].
func Clamp[T int | int8 | int16 | int32 | int64 | float32 | float64](val, min, max T) T {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

// Clamp01 clamps value to the range [0, 1].
func Clamp01[T int | int8 | int16 | int32 | int64 | float32 | float64](val T) T {
	return Clamp(val, 0, 1)
}

// Min returns the minimum of two values.
func Min[T int | int8 | int16 | int32 | int64 | float32 | float64](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// Max returns the maximum of two values.
func Max[T int | int8 | int16 | int32 | int64 | float32 | float64](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// MinOf returns the minimum of multiple values.
func MinOf[T int | int8 | int16 | int32 | int64 | float32 | float64](vals ...T) T {
	if len(vals) == 0 {
		var zero T
		return zero
	}
	min := vals[0]
	for i := 1; i < len(vals); i++ {
		if vals[i] < min {
			min = vals[i]
		}
	}
	return min
}

// MaxOf returns the maximum of multiple values.
func MaxOf[T int | int8 | int16 | int32 | int64 | float32 | float64](vals ...T) T {
	if len(vals) == 0 {
		var zero T
		return zero
	}
	max := vals[0]
	for i := 1; i < len(vals); i++ {
		if vals[i] > max {
			max = vals[i]
		}
	}
	return max
}

// Abs returns the absolute value.
func Abs[T int | int8 | int16 | int32 | int64 | float32 | float64](val T) T {
	if val < 0 {
		return -val
	}
	return val
}

// Sign returns -1 if val < 0, 0 if val == 0, 1 if val > 0.
func Sign[T int | int8 | int16 | int32 | int64 | float32 | float64](val T) int {
	if val < 0 {
		return -1
	}
	if val > 0 {
		return 1
	}
	return 0
}

// Floor returns the largest integer <= val.
func Floor(val float64) float64 {
	return math.Floor(val)
}

// Ceil returns the smallest integer >= val.
func Ceil(val float64) float64 {
	return math.Ceil(val)
}

// Round returns the nearest integer, rounding half away from zero.
func Round(val float64) float64 {
	return math.Round(val)
}

// RoundTo returns val rounded to the given precision.
func RoundTo(val float64, precision int) float64 {
	ratio := math.Pow10(precision)
	return Round(val*ratio) / ratio
}

// FloorTo returns val floored to the given precision.
func FloorTo(val float64, precision int) float64 {
	ratio := math.Pow10(precision)
	return math.Floor(val*ratio) / ratio
}

// CeilTo returns val ceiled to the given precision.
func CeilTo(val float64, precision int) float64 {
	ratio := math.Pow10(precision)
	return math.Ceil(val*ratio) / ratio
}

// Pow returns x**y (x to the power of y).
func Pow(x, y float64) float64 {
	return math.Pow(x, y)
}

// Sqrt returns the square root.
func Sqrt(val float64) float64 {
	return math.Sqrt(val)
}

// Cbrt returns the cube root.
func Cbrt(val float64) float64 {
	return math.Cbrt(val)
}

// Exp returns e**val (e raised to the given power).
func Exp(val float64) float64 {
	return math.Exp(val)
}

// Exp2 returns 2**val (2 raised to the given power).
func Exp2(val float64) float64 {
	return math.Exp2(val)
}

// Log returns the natural logarithm.
func Log(val float64) float64 {
	return math.Log(val)
}

// Log10 returns the base 10 logarithm.
func Log10(val float64) float64 {
	return math.Log10(val)
}

// Log2 returns the base 2 logarithm.
func Log2(val float64) float64 {
	return math.Log2(val)
}

// Sin returns the sine.
func Sin(val float64) float64 {
	return math.Sin(val)
}

// Cos returns the cosine.
func Cos(val float64) float64 {
	return math.Cos(val)
}

// Tan returns the tangent.
func Tan(val float64) float64 {
	return math.Tan(val)
}

// Asin returns the arcsine.
func Asin(val float64) float64 {
	return math.Asin(val)
}

// Acos returns the arccosine.
func Acos(val float64) float64 {
	return math.Acos(val)
}

// Atan returns the arctangent.
func Atan(val float64) float64 {
	return math.Atan(val)
}

// Atan2 returns the arctangent of y/x.
func Atan2(y, x float64) float64 {
	return math.Atan2(y, x)
}

// Sinh returns the hyperbolic sine.
func Sinh(val float64) float64 {
	return math.Sinh(val)
}

// Cosh returns the hyperbolic cosine.
func Cosh(val float64) float64 {
	return math.Cosh(val)
}

// Tanh returns the hyperbolic tangent.
func Tanh(val float64) float64 {
	return math.Tanh(val)
}

// Sum returns the sum of all values.
func Sum[T int | int8 | int16 | int32 | int64 | float32 | float64](vals ...T) T {
	var sum T
	for _, v := range vals {
		sum += v
	}
	return sum
}

// Mean returns the average of all values.
func Mean[T int | int8 | int16 | int32 | int64 | float32 | float64](vals ...T) float64 {
	if len(vals) == 0 {
		return 0
	}
	var sum float64
	for _, v := range vals {
		sum += float64(v)
	}
	return sum / float64(len(vals))
}

// Median returns the median of all values.
func Median[T int | int8 | int16 | int32 | int64 | float32 | float64](vals ...T) float64 {
	if len(vals) == 0 {
		return 0
	}
	// Copy and sort
	sorted := make([]float64, len(vals))
	for i, v := range vals {
		sorted[i] = float64(v)
	}
	// Simple sort
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j] < sorted[i] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	return sorted[mid]
}

// Variance returns the variance of all values.
func Variance[T int | int8 | int16 | int32 | int64 | float32 | float64](vals ...T) float64 {
	if len(vals) == 0 {
		return 0
	}
	mean := Mean(vals...)
	var sum float64
	for _, v := range vals {
		d := float64(v) - mean
		sum += d * d
	}
	return sum / float64(len(vals))
}

// StdDev returns the standard deviation.
func StdDev[T int | int8 | int16 | int32 | int64 | float32 | float64](vals ...T) float64 {
	return math.Sqrt(Variance(vals...))
}

// Percentile returns the value at the given percentile (0-100).
func Percentile[T int | int8 | int16 | int32 | int64 | float32 | float64](p float64, vals ...T) float64 {
	if len(vals) == 0 || p < 0 || p > 100 {
		return 0
	}
	// Copy and sort
	sorted := make([]float64, len(vals))
	for i, v := range vals {
		sorted[i] = float64(v)
	}
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j] < sorted[i] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	idx := (p / 100) * float64(len(sorted)-1)
	lower := int(idx)
	frac := idx - float64(lower)
	if lower >= len(sorted)-1 {
		return sorted[len(sorted)-1]
	}
	return sorted[lower]*(1-frac) + sorted[lower+1]*frac
}

// IsInf checks if val is infinite.
func IsInf(val float64) bool {
	return math.IsInf(val, 0)
}

// IsNaN checks if val is NaN.
func IsNaN(val float64) bool {
	return math.IsNaN(val)
}

// IsFinite checks if val is finite (not Inf or NaN).
func IsFinite(val float64) bool {
	return !math.IsInf(val, 0) && !math.IsNaN(val)
}

// RandomInt returns a random integer in [min, max).
func RandomInt(min, max int) int {
	return rand.Intn(max-min) + min
}

// RandomFloat64 returns a random float64 in [0, 1).
func RandomFloat64() float64 {
	return rand.Float64()
}

// RandomIntn returns a random int in [0, n).
func RandomIntn(n int) int {
	return rand.Intn(n)
}

// IntToFloat converts int to float64.
func IntToFloat[T int | int8 | int16 | int32 | int64](val T) float64 {
	return float64(val)
}

// FloatToInt converts float64 to int.
func FloatToInt(val float64) int {
	return int(val)
}
