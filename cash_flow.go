package gofinance

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/khezen/rootfinding" // for [IRR]
)

// CashFlow represents a single dated cash‑flow.
// A positive Value denotes an inflow, a negative Value an outflow.
// Date is always stored in UTC (see NewCashFlow for details).
//
// Example:
//
//	rent, _ := NewCashFlow(1000, "2025-07-01")
//
// Units: Value in the currency of the analysis, Date in UTC.
type CashFlow struct {
	Value float64
	Date  time.Time
}

// NewCashFlow builds a [CashFlow] from a numeric value and a human‑friendly time
// specification. The time string is parsed by [StringToTime] which supports
// granularities from years down to milliseconds and always returns the midpoint
// of the specified period in UTC.
//
// Note that you can supply one or two dates.
// For example, if you supply "2020", the [Date] of the [CashFlow]
// will be the middle of the year 2020. If you supply "2020-10-15" and
// "2021-10-14", then the Date will be the midpoint between these two dates.
//
// Note that conceptually a cash flow is discrete, occurs at a point in time,
// not throughout a period of time.
func NewCashFlow(value float64, timeStr ...string) (CashFlow, error) {
	date, err := StringToTime(timeStr...)
	if err != nil {
		return CashFlow{}, err
	}
	return CashFlow{value, date}, nil
}

// daysInYear helper returns number of days in a given year.
// Helper for [yearsBetween]
func daysInYear(y int) int {
	if (y%4 == 0 && y%100 != 0) || y%400 == 0 {
		return 366
	}
	return 365
}

// yearsBetween counts full calendar years first, then
// expresses the leftover part as a fraction of the length of the
// next full calendar year.
//
//	2023-03-15 → 2026-02-10
//	   fullYears  = 2  (to 2025-03-15)
//	   remainder  = 2025-03-15 → 2026-02-10
//	   yearLength = 365 (2025 is not leap)
//	   frac       = remainder / yearLength
//	   result     = sign * (fullYears + frac)
func yearsBetween(a, b time.Time) float64 {
	if a.Equal(b) {
		return 0
	}

	// normalise order, keep sign
	sign := 1.0
	if b.Before(a) {
		a, b = b, a
		sign = -1
	}

	// 1. count whole calendar years
	fullYears := 0.0
	for {
		next := a.AddDate(1, 0, 0) // same month/day/time next year
		if !next.After(b) {
			fullYears++
			a = next
		} else {
			break
		}
	}

	// 2. leftover fraction
	if a.Equal(b) {
		return sign * fullYears
	}
	yearLen := daysInYear(a.Year())
	remDays := b.Sub(a).Hours() / 24.0
	frac := remDays / float64(yearLen)

	return sign * (fullYears + frac)
}

// YearsFrom returns the signed year distance between valuationDate and the
// cash‑flow's occurrence Date. If the cash‑flow happens after the valuation
// date the result is positive, if it happened before it is negative.
// A calendar year is taken to be 365.25 days in accordance
// with the ACT/365.25 day‑count convention.
func (cf CashFlow) YearsFrom(valuationDate time.Time) float64 {
	return yearsBetween(valuationDate, cf.Date)
}

// PresentValue discounts the cash‑flow to valuationDate using the supplied
// Rate. The absolute time distance is used so that both past and future flows
// are handled gracefully: a past inflow is compounded forward, a future inflow
// is discounted back.
//
// If valuationDate is nil, the UTC [time.Now] is used.
func (cf CashFlow) PresentValue(r Rate, valuationDate time.Time) float64 {
	years := cf.YearsFrom(valuationDate)
	return cf.Value * r.DiscountFactor(years)
}

// PresentValueNow is a convenience wrapper for [PresentValue]
// that sets valuationDate to the current UTC time using [time.Now].
func (cf CashFlow) PresentValueNow(r Rate) float64 {
	return cf.PresentValue(r, time.Now().UTC())
}

// CashFlows is a helper alias that adds portfolio‑level analytics to a slice
// of CashFlow.
//
//	cfs := CashFlows{cf1, cf2, cf3}
//	npv := cfs.NPV(discountRate, time.Now())
//	irr, _ := cfs.IRR(0.1)
//
// All collection methods leave the original slice untouched (except [Sort]).
type CashFlows []CashFlow

// Sort orders the cash‑flows in‑place by ascending [Date]. Useful before [IRR]
// calculations that implicitly take the first cash‑flow as the time‑zero
// reference.
func (cfs CashFlows) Sort() {
	sort.Slice(cfs, func(i, j int) bool {
		return cfs[i].Date.Before(cfs[j].Date)
	})
}

// NPV computes the net present value of the collection at valuationDate using
// the provided discount Rate.
func (cfs CashFlows) NPV(r Rate, valuationDate time.Time) float64 {
	npv := 0.0
	for _, cf := range cfs {
		npv += cf.PresentValue(r, valuationDate)
	}
	return npv
}

// IRR estimates the internal [Rate] of return by finding the rate (r)
// that makes the NPV of the cash-flow stream equal to zero.
// It brackets a root automatically and then refines it with
// [github.com/khezen/rootfinding.Brent]. The function returns an error if it cannot
// bracket a root or if Brent fails to converge.
func (cashFlows CashFlows) IRR() (Rate, error) {
	if len(cashFlows) == 0 {
		return RateAnnualContinuous{}, errors.New("IRR requires at least one cash-flow")
	}

	// work on a sorted copy so the caller’s slice remains untouched
	ordered := make(CashFlows, len(cashFlows))
	copy(ordered, cashFlows)
	ordered.Sort()
	anchor := ordered[0].Date

	// helper: wraps [NPV] with valuationDate = anchor
	npv := func(r float64) float64 {
		return ordered.NPV(RateAnnualContinuous{Value: r}, anchor)
	}

	//----------------------------------------------------------------------
	// 1.  Bracket a root
	// ---------------------------------------------------------------------
	// Starts with a very low rate just shy of −100 % (continuous); this gives
	// a huge discount factor and usually pushes NPV positive when the first
	// cash-flow is an outflow.
	lowerBoundRate := -0.999999 // ~-99.9999 %
	upperBoundRate := 0.10      // 10 % p.a. to begin with
	npvLowerBound := npv(lowerBoundRate)
	npvUpperBound := npv(upperBoundRate)

	// If NPV signs do not differ, expand the upper bound exponentially
	// until we hit a sign change or a reasonable ceiling.
	for npvLowerBound*npvUpperBound > 0 && upperBoundRate < 1000 {
		upperBoundRate *= 2
		npvUpperBound = npv(upperBoundRate)
	}
	if npvLowerBound*npvUpperBound > 0 {
		return RateAnnualContinuous{}, errors.New("IRR: could not bracket a root")
	}

	//----------------------------------------------------------------------
	// 2.  Refine with Brent
	// ---------------------------------------------------------------------
	root, err := rootfinding.Brent(npv, lowerBoundRate, upperBoundRate, 12)
	if err != nil {
		return RateAnnualContinuous{}, fmt.Errorf("IRR: %w", err)
	} // this if statement is not covered by tests because difficult to provoke error here
	return RateAnnualContinuous{Value: root}, nil
}
