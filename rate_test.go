package gofinance

import (
	"math"
	"testing"
)

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

func TestRateAnnualPercentage(t *testing.T) {
	tests := []struct {
		name           string
		apr            RateAnnualPercentage
		years          float64
		wantDF         float64
		wantEffAnnual  float64
		wantContinuous float64
	}{
		{
			name:           "nominal 5% APR, monthly, 3 years",
			apr:            RateAnnualPercentage{Value: 0.05, PeriodsPerYear: 12},
			years:          3,
			wantDF:         math.Pow(1+0.05/12, -12*3),
			wantEffAnnual:  math.Pow(1+0.05/12, 12) - 1,
			wantContinuous: math.Log(math.Pow(1+0.05/12, 12)),
		},
		{
			name:   "zero years must discount to 1",
			apr:    RateAnnualPercentage{Value: 0.08, PeriodsPerYear: 4},
			years:  0,
			wantDF: 1,
			// other two outputs do not depend on years, verify anyway:
			wantEffAnnual:  math.Pow(1+0.08/4, 4) - 1,
			wantContinuous: math.Log(math.Pow(1+0.08/4, 4)),
		},
		{
			name:           "negative APR (deflation) −2%, quarterly, 5 y",
			apr:            RateAnnualPercentage{Value: -0.02, PeriodsPerYear: 4},
			years:          5,
			wantDF:         math.Pow(1-0.02/4, -20),
			wantEffAnnual:  math.Pow(1-0.02/4, 4) - 1,
			wantContinuous: math.Log(math.Pow(1-0.02/4, 4)),
		},
		{
			name:           "extreme compounding frequency (≈continuous)",
			apr:            RateAnnualPercentage{Value: 0.07, PeriodsPerYear: 36500}, // 100× daily
			years:          2,
			wantDF:         math.Pow(1+0.07/36500, -73000),
			wantEffAnnual:  math.Pow(1+0.07/36500, 36500) - 1,
			wantContinuous: math.Log(math.Pow(1+0.07/36500, 36500)),
		},
		{
			name:           "division by zero periods — expect +Inf DF and NaN rates",
			apr:            RateAnnualPercentage{Value: 0.05, PeriodsPerYear: 0},
			years:          1,
			wantDF:         math.Inf(1), // 1 + r/0 → +Inf then ^(-0) = +Inf
			wantEffAnnual:  0,
			wantContinuous: math.Log(1),
		},
	}

	for _, tc := range tests {
		tc := tc // capture loop var
		// Run allows concurrency
		t.Run(tc.name, func(t *testing.T) {
			gotDF := tc.apr.DiscountFactor(tc.years)
			if math.IsNaN(tc.wantDF) && !math.IsNaN(gotDF) {
				t.Fatalf("DiscountFactor() got %v, want NaN", gotDF)
			} else if !almostEq(gotDF, tc.wantDF, epsilon) {
				t.Fatalf("DiscountFactor() got %v, want %v", gotDF, tc.wantDF)
			}

			gotEA := tc.apr.RateAnnualEffective()
			if math.IsNaN(tc.wantEffAnnual) && !math.IsNaN(gotEA) {
				t.Fatalf("RateAnnualEffective() got %v, want NaN", gotEA)
			} else if !almostEq(gotEA, tc.wantEffAnnual, epsilon) {
				t.Fatalf("RateAnnualEffective() got %v, want %v", gotEA, tc.wantEffAnnual)
			}

			gotC := tc.apr.RateContinuous()
			if math.IsNaN(tc.wantContinuous) && !math.IsNaN(gotC) {
				t.Fatalf("RateContinuous() got %v, want NaN", gotC)
			} else if !almostEq(gotC, tc.wantContinuous, epsilon) {
				t.Fatalf("RateContinuous() got %v, want %v", gotC, tc.wantContinuous)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// RateEffective
// -----------------------------------------------------------------------------

func TestRateEffective(t *testing.T) {
	tests := []struct {
		name           string
		reff           RateEffective
		years          float64
		wantDF         float64
		wantEffAnnual  float64
		wantContinuous float64
	}{
		{
			name:           "1% effective monthly, 10 y",
			reff:           RateEffective{Value: 0.01, PeriodsPerYear: 12},
			years:          10,
			wantDF:         math.Pow(1.01, -120),
			wantEffAnnual:  math.Pow(1.01, 12) - 1,
			wantContinuous: math.Log(math.Pow(1.01, 12)),
		},
		{
			name:           "zero years",
			reff:           RateEffective{Value: 0.03, PeriodsPerYear: 2},
			years:          0,
			wantDF:         1,
			wantEffAnnual:  math.Pow(1.03, 2) - 1,
			wantContinuous: math.Log(math.Pow(1.03, 2)),
		},
		{
			name:           "negative rate (deflation) −0.5% quarterly",
			reff:           RateEffective{Value: -0.005, PeriodsPerYear: 4},
			years:          4,
			wantDF:         math.Pow(1-0.005, -16),
			wantEffAnnual:  math.Pow(1-0.005, 4) - 1,
			wantContinuous: math.Log(math.Pow(1-0.005, 4)),
		},
		{
			name:           "periods=0 division by zero",
			reff:           RateEffective{Value: 0.02, PeriodsPerYear: 0},
			years:          1,
			wantDF:         1,           // exponent becomes -0*years = 0 → (1+r)^0 = 1
			wantEffAnnual:  0,           // (1+r)^0-1 = 0
			wantContinuous: math.Log(1), // 0
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.reff.DiscountFactor(tc.years); !almostEq(got, tc.wantDF, epsilon) {
				t.Fatalf("DiscountFactor() got %v, want %v", got, tc.wantDF)
			}
			if got := tc.reff.RateAnnualEffective(); !almostEq(got, tc.wantEffAnnual, epsilon) {
				t.Fatalf("RateAnnualEffective() got %v, want %v", got, tc.wantEffAnnual)
			}
			if got := tc.reff.RateContinuous(); !almostEq(got, tc.wantContinuous, epsilon) {
				t.Fatalf("RateContinuous() got %v, want %v", got, tc.wantContinuous)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// RateContinuous
// -----------------------------------------------------------------------------

func TestRateContinuous(t *testing.T) {
	tests := []struct {
		name           string
		rcont          RateContinuous
		years          float64
		wantDF         float64
		wantEffAnnual  float64
		wantContinuous float64
	}{
		{
			name:           "4% continuous, 7 y",
			rcont:          RateContinuous{Value: 0.04},
			years:          7,
			wantDF:         math.Exp(-0.04 * 7),
			wantEffAnnual:  math.Exp(0.04) - 1,
			wantContinuous: 0.04,
		},
		{
			name:           "zero years",
			rcont:          RateContinuous{Value: 0.04},
			years:          0,
			wantDF:         1,
			wantEffAnnual:  math.Exp(0.04) - 1,
			wantContinuous: 0.04,
		},
		{
			name:           "negative continuous rate (deflation)",
			rcont:          RateContinuous{Value: -0.01},
			years:          3,
			wantDF:         math.Exp(0.01 * 3), // minus minus
			wantEffAnnual:  math.Exp(-0.01) - 1,
			wantContinuous: -0.01,
		},
		{
			name:           "huge positive rate (overflow safe)",
			rcont:          RateContinuous{Value: 100},
			years:          0.01,
			wantDF:         math.Exp(-1), // exactly e^-1
			wantEffAnnual:  math.Exp(100) - 1,
			wantContinuous: 100,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.rcont.DiscountFactor(tc.years); !almostEq(got, tc.wantDF, epsilon) {
				t.Fatalf("DiscountFactor() got %v, want %v", got, tc.wantDF)
			}
			if got := tc.rcont.RateAnnualEffective(); !almostEq(got, tc.wantEffAnnual, epsilon) {
				t.Fatalf("RateAnnualEffective() got %v, want %v", got, tc.wantEffAnnual)
			}
			if got := tc.rcont.RateContinuous(); got != tc.wantContinuous {
				t.Fatalf("RateContinuous() got %v, want %v", got, tc.wantContinuous)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// Interface conformance smoke test
// -----------------------------------------------------------------------------

func TestRateInterfaceSatisfaction(t *testing.T) {
	var (
		_ Rate = RateAnnualPercentage{Value: 0.01, PeriodsPerYear: 12}
		_ Rate = RateEffective{Value: 0.01, PeriodsPerYear: 12}
		_ Rate = RateContinuous{Value: 0.01}
	)
}

// -----------------------------------------------------------------------------
// Benchmarks (optional but often handy for perf regressions)
// -----------------------------------------------------------------------------

func BenchmarkRateAnnualPercentageDiscountFactor(b *testing.B) {
	r := RateAnnualPercentage{Value: 0.05, PeriodsPerYear: 12}
	for i := 0; i < b.N; i++ {
		_ = r.DiscountFactor(5)
	}
}
