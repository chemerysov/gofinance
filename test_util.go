package gofinance

import "math"

// almostEq is a fuzzy equals.
// This relative test is stable across wide floating–point ranges.
// Math details: combines
//
// RelativeError = |a - b| / max(|a|, |b|)
//
// AbsoluteError = |a - b|
//
// side-stepping the issue when a, b are close to 0,
// so denominator of RelativeError explodes.
func almostEq(a, b, epsilon float64) bool {
	scale := math.Max(1, math.Max(math.Abs(a), math.Abs(b)))
	return math.Abs(a-b) <= epsilon*scale
}

// epsilon is large enough to hide last-bit noise but small enough to catch logic errors.
const epsilon = 1e-12
