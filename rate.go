package gofinance

import "math"

// # Rate represents an interest rate, discount rate, compound rate, etc.
type Rate interface {
	// DiscountFactor returns discount factor based on number of years.
	// Math details:
	//
	// PresentValueOfCashFlow = DiscountFactor * FutureValueOfCashFlow
	//
	// FutureValueOfCashFlow = CompoundFactor * PresentValueOfCashFlow
	//
	// CompoundFactor = 1 / DiscountFactor
	DiscountFactor(years float64) float64

	// RateAnnualEffective converts any rate to effective annual rate.
	// Note that an effective annual rate is the same as
	// annual percentage rate compounded once a year / annually.
	// For more on this rate see [RateAnnualEffective].
	RateAnnualEffective() float64

	// RateContinuous converts any rate to continuous rate.
	// For more on this rate see [RateContinuous].
	RateContinuous() float64
}

// RateAnnualPercentage implements [Rate] for annual percentage rate.
// To specify this rate fully, a compounding frequency is needed, here PeriodsPerYear.
// That is, 5% annual percentage rate compounded monthly
// and 5% annual percentage rate compunded quarterly are different rates.
type RateAnnualPercentage struct {
	Value          float64
	PeriodsPerYear float64
}

// DiscountFactor implements [Rate].
// DiscountFactor returns discount factor based on number of years.
// Math details:
//
// DiscountFactor = (1 + AnnualPercentageRate / Periods)^{-Periods * Years}
func (r RateAnnualPercentage) DiscountFactor(years float64) float64 {
	return math.Pow(1+r.Value/r.PeriodsPerYear, -r.PeriodsPerYear*years)
}

// RateAnnualEffective implements [Rate].
// RateAnnualEffective converts any rate to effective annual rate.
// Math details:
//
// CompoundFactor = (1 + AnnualPercentageRate / Periods)^Periods = 1 + EffectiveAnnualRate
//
// EffectiveAnnualRate = (1 + AnnualPercentageRate / Periods)^Periods - 1
func (r RateAnnualPercentage) RateAnnualEffective() float64 {
	return math.Pow(1+r.Value/r.PeriodsPerYear, r.PeriodsPerYear) - 1
}

// RateContinuous implements [Rate].
// RateContinuous converts any rate to continuous rate.
// Math details:
//
// CompoundFactor = 1 + EffectiveAnnualRate = e^{ContinuousRate}
//
// ContinuousRate = ln(1 + EffectiveAnnualRate)
func (r RateAnnualPercentage) RateContinuous() float64 {
	return math.Log(1 + r.RateAnnualEffective())
}

// RateEffective implements [Rate] for effective rate.
// To specify this rate fully, a compounding frequency is needed, here PeriodsPerYear.
// That is, 5% effective annual rate and 5% effective monthly rate are different rates.
type RateEffective struct {
	Value          float64
	PeriodsPerYear float64
}

// DiscountFactor implements [Rate].
// DiscountFactor returns discount factor based on number of years.
// Math details:
//
// DicountFactor = (1 + EffectiveRate)^{-Periods * Years}
func (r RateEffective) DiscountFactor(years float64) float64 {
	return math.Pow(1+r.Value, -r.PeriodsPerYear*years)
}

// RateAnnualEffective implements [Rate].
// RateAnnualEffective converts any rate to effective annual rate.
// Math details:
//
// CompoundFactor = (1 + EffectiveRate)^Periods = 1 + EffectiveAnnualRate
//
// EffectiveAnnualRate = (1 + EffectiveRate)^Periods - 1
func (r RateEffective) RateAnnualEffective() float64 {
	return math.Pow(1+r.Value, r.PeriodsPerYear) - 1
}

// RateContinuous implements [Rate].
// RateContinuous converts any rate to continuous rate.
// Math details:
//
// CompoundFactor = 1 + EffectiveAnnualRate = e^{ContinuousRate}
//
// ContinuousRate = ln(1 + EffectiveAnnualRate)
func (r RateEffective) RateContinuous() float64 {
	return math.Log(1 + r.RateAnnualEffective())
}

// RateContinuous implements [Rate] for continuous rate.
// No need to specify compounding frequency as it is already specified (continuous).
// For more details, see [RateContinuous.DiscountFactor].
type RateContinuous struct {
	Value float64
}

// DiscountFactor implements [Rate].
// DiscountFactor returns discount factor based on number of years.
// Math details:
//
// DiscountFactor = \lim_{Periods -> \infin} (1 + AnnualPercentageRate / Periods)^{-Periods * Years}
//
// e^x = \lim_{n -> \infin} (1 + x/n)^n   by definition
//
// e^AnnualPercentageRate = \lim_{Periods -> \infin} (1 + AnnualPercentageRate / Periods)^Periods
//
// DiscountFactor = e^{AnnualPercentageRate * -Years}
//
// AnnualPercentageRate compounded continuously = ContinuousRate
//
// DiscountFactor = e^{ContinuousRate * -Years}
func (r RateContinuous) DiscountFactor(years float64) float64 {
	return math.Exp(r.Value * -years)
}

// RateAnnualEffective implements [Rate].
// RateAnnualEffective converts any rate to effective annual rate.
// Math details:
//
// CompoundFactor = e^ContinuousRate = 1 + EffectiveAnnualRate
//
// EffectiveAnnualRate = e^ContinuousRate - 1
func (r RateContinuous) RateAnnualEffective() float64 {
	return math.Exp(r.Value) - 1
}

// RateContinuous implements [Rate].
// RateContinuous converts any rate to continuous rate.
// Math details:
//
// ContinuousRate = ContinuousRate
func (r RateContinuous) RateContinuous() float64 {
	return r.Value
}
