package gofinance

import (
	"math"
	"testing"
	"time"

	_ "unsafe" // required for go:linkname
	// for [IRR]
)

// anchor date reused in many tests
var anchor = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

// -----------------------------------------------------------------------------
// daysInYear
// -----------------------------------------------------------------------------

func TestDaysInYear(t *testing.T) {
	tests := []struct {
		year int
		want int
	}{
		{2020, 366}, // leap (divisible by 4, not by 100)
		{1900, 365}, // not leap (divisible by 100, not by 400)
		{2000, 366}, // leap (divisible by 400)
		{2025, 365}, // typical
	}
	for _, tc := range tests {
		if got := daysInYear(tc.year); got != tc.want {
			t.Fatalf("daysInYear(%d) = %d, want %d", tc.year, got, tc.want)
		}
	}
}

// -----------------------------------------------------------------------------
// yearsBetween
// -----------------------------------------------------------------------------

func TestYearsBetween(t *testing.T) {
	a := time.Date(2023, 3, 15, 0, 0, 0, 0, time.UTC)
	b := time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC)

	want := 2.9095890410958907 // worked out in README example
	if got := yearsBetween(a, b); !almostEq(got, want, 1e-12) {
		t.Errorf("yearsBetween example: got %.15f, want %.15f", got, want)
	}

	// symmetry & sign
	if got := yearsBetween(b, a); !almostEq(got, -want, 1e-12) {
		t.Errorf("yearsBetween symmetry failed, got %.15f, want %.15f", got, -want)
	}

	// identical dates
	if got := yearsBetween(a, a); got != 0 {
		t.Errorf("yearsBetween same date = %f, want 0", got)
	}

	// leap-year boundary (2020-01-01 → 2021-01-01 should be 1)
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)
	if got := yearsBetween(start, end); got != 1 {
		t.Errorf("yearsBetween leap-year: got %f, want 1", got)
	}
}

// -----------------------------------------------------------------------------
// CashFlow construction & YearsFrom
// -----------------------------------------------------------------------------

func TestNewCashFlowAndYearsFrom(t *testing.T) {
	cf, err := NewCashFlow(100, "2024-06") // month-granularity
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// YearsFrom with valuation before the flow should be positive
	val := time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC)
	if yrs := cf.YearsFrom(val); yrs <= 0 {
		t.Errorf("YearsFrom: expected positive, got %f", yrs)
	}

	// invalid time spec must error
	if _, err := NewCashFlow(50, "feb-31-2025"); err == nil {
		t.Error("expected error for invalid date, got nil")
	}

	// two-period form
	cf2, err := NewCashFlow(10, "2020-01-01", "2020-12-31")
	if err != nil {
		t.Fatalf("two-period form returned error: %v", err)
	}
	mid, _ := StringToTime("2020-01-01", "2020-12-31")
	if !cf2.Date.Equal(mid) {
		t.Errorf("two-period NewCashFlow date mismatch: got %v, want %v", cf2.Date, mid)
	}
}

// -----------------------------------------------------------------------------
// PresentValue & PresentValueNow
// -----------------------------------------------------------------------------

func TestPresentValue(t *testing.T) {
	// cash flow two years in the future
	cf := CashFlow{Value: 100, Date: anchor.AddDate(2, 0, 0)}
	r := RateAnnualContinuous{Value: 0.05} // 5 % continuous

	want := 100 * math.Exp(-0.05*2)
	if got := cf.PresentValue(r, anchor); !almostEq(got, want, epsilon) {
		t.Errorf("PresentValue got %.10f, want %.10f", got, want)
	}

	// PresentValueNow should equal PresentValue with valuationDate = now
	cfNow := CashFlow{Value: 50, Date: time.Now().UTC()}
	gotNow := cfNow.PresentValueNow(r)
	wantNow := cfNow.PresentValue(r, time.Now().UTC())
	if !almostEq(gotNow, wantNow, 1e-6) { // milliseconds delay acceptable
		t.Errorf("PresentValueNow mismatch: got %f, want %f", gotNow, wantNow)
	}
}

// -----------------------------------------------------------------------------
// CashFlows.Sort
// -----------------------------------------------------------------------------

func TestCashFlowsSort(t *testing.T) {
	d1 := anchor.AddDate(0, 0, 10)
	d2 := anchor.AddDate(0, 0, 5)
	d3 := anchor

	cfs := CashFlows{
		{Value: 3, Date: d1},
		{Value: 2, Date: d2},
		{Value: 1, Date: d3},
	}
	cfs.Sort()

	if !(cfs[0].Date.Equal(d3) && cfs[1].Date.Equal(d2) && cfs[2].Date.Equal(d1)) {
		t.Errorf("CashFlows.Sort order incorrect: %+v", cfs)
	}
}

// -----------------------------------------------------------------------------
// NPV
// -----------------------------------------------------------------------------

func TestNPV(t *testing.T) {
	r := RateAnnualContinuous{Value: 0.10} // 10 % continuous

	cfs := CashFlows{
		{-1000, anchor},
		{400, anchor.AddDate(1, 0, 0)},
		{400, anchor.AddDate(2, 0, 0)},
		{400, anchor.AddDate(3, 0, 0)},
	}

	// brute-force expected value
	exp := -1000 +
		400*math.Exp(-0.10*1) +
		400*math.Exp(-0.10*2) +
		400*math.Exp(-0.10*3)

	if got := cfs.NPV(r, anchor); !almostEq(got, exp, epsilon) {
		t.Errorf("NPV got %.10f, want %.10f", got, exp)
	}
}

// -----------------------------------------------------------------------------
// IRR
// -----------------------------------------------------------------------------

func TestIRRSimpleTwoPeriod(t *testing.T) {
	cfs := CashFlows{
		{-100, anchor},
		{110, anchor.AddDate(1, 0, 0)},
	}
	irr, err := cfs.IRR()
	if err != nil {
		t.Fatalf("IRR error: %v", err)
	}
	want := math.Log(1.1) // ≈ 0.09531
	if !almostEq(irr.RateAnnualContinuous(), want, 1e-6) {
		t.Errorf("IRR got %.6f, want %.6f", irr.RateAnnualContinuous(), want)
	}
}

func TestIRRMultiplePeriods(t *testing.T) {
	cfs := CashFlows{
		{-1000, anchor},
		{400, anchor.AddDate(1, 0, 0)},
		{400, anchor.AddDate(2, 0, 0)},
		{400, anchor.AddDate(3, 0, 0)},
	}
	irr, err := cfs.IRR()
	if err != nil {
		t.Fatalf("IRR error: %v", err)
	}

	// IRR must make NPV ≈ 0
	if npv := cfs.NPV(irr, anchor); !almostEq(npv, 0, 1e-6) {
		t.Errorf("IRR root check failed: NPV at IRR = %f, want 0", npv)
	}
}

func TestIRRErrors(t *testing.T) {
	// empty slice
	if _, err := (CashFlows{}).IRR(); err == nil {
		t.Error("IRR expected error for empty slice, got nil")
	}

	// all inflows cannot bracket a root
	cfs := CashFlows{
		{+10, anchor},
		{+10, anchor.AddDate(1, 0, 0)},
	}
	if _, err := cfs.IRR(); err == nil {
		t.Error("IRR expected error for un-bracketable root, got nil")
	}
}
