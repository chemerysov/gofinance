package gofinance

import "math"

type Rate interface {
	// derive (present value) discount factor based on number of years
	// discount factor is a factor DF such that
	// present value of cash flow = DF * future value of cash flow
	// compound factor = 1 / discount factor
	DiscountFactor(years float64) float64

	// convert any rate to effective annual rate
	// for more on this rate see struct RateAnnualEffective
	RateAnnualEffective() float64

	// convert any rate to continuous rate
	// continuous rate = \lim_{n->\infin} (1 + APR/n)^n
	// so an APR compounded n approaches infinity times
	RateContinuous() float64
}

// implements Rate for annual percentage rate
// to specify this rate fully you need compounding frequency, here PeriodsPerYear
type RateAnnualPercentage struct {
	Value          float64
	PeriodsPerYear float64
}

// DF = (1 + APR/n)^{-n*t}
func (r RateAnnualPercentage) DiscountFactor(years float64) float64 {
	return math.Pow(1+r.Value/r.PeriodsPerYear, -r.PeriodsPerYear*years)
}

// CF = (1 + APR/n)^n = 1 + EAR
// EAR = (1 + APR/n)^n - 1
func (r RateAnnualPercentage) RateAnnualEffective() float64 {
	return math.Pow(1+r.Value/r.PeriodsPerYear, r.PeriodsPerYear) - 1
}

// CF = 1 + EAR = e^{CR}
// CR = log(1 + EAR)
// converting from EAR as opposed to from APR simplifies math
func (r RateAnnualPercentage) RateContinuous() float64 {
	return math.Log(1 + r.RateAnnualEffective())
}

// implements Rate for effective rate
// to specify this rate fully you need compounding frequency, here PeriodsPerYear
// that is, there can exist effective annual rate, effective monthly rate, etc
type RateEffective struct {
	Value          float64
	PeriodsPerYear float64
}

// DF = (1 + ER)^{-n*t}
func (r RateEffective) DiscountFactor(years float64) float64 {
	return math.Pow(1+r.Value, -r.PeriodsPerYear*years)
}

// CF = (1 + ER)^n = 1 + EAR
// EAR = (1 + ER)^n - 1
func (r RateEffective) RateAnnualEffective() float64 {
	return math.Pow(1+r.Value, r.PeriodsPerYear) - 1
}

// CF = 1 + EAR = e^{CR}
// CR = log(1 + EAR)
// converting from EAR as opposed to from APR simplifies math
func (r RateEffective) RateContinuous() float64 {
	return math.Log(1 + r.RateAnnualEffective())
}

// implements Rate for continuous rate
// no need to specify compounding frequency as it is already specified (continuous)
type RateContinuous struct {
	Value float64
}

// DF = \lim_{n->\infin} (1 + APR/n)^{-nt}
// e^x = \lim_{n->\infin} (1 + x/n)^n   by definition
// e^APR = \lim_{n->\infin} (1 + APR/n)^n
// DF = e^{APR * -t}
// APR compounded continuously = CR
// DF = e^{CR * -t}
func (r RateContinuous) DiscountFactor(years float64) float64 {
	return math.Exp(r.Value * -years)
}

// CF = e^APR = 1 + EAR
// EAR = e^APR - 1
func (r RateContinuous) RateAnnualEffective() float64 {
	return math.Exp(r.Value) - 1
}

// CR = CR
func (r RateContinuous) RateContinuous() float64 {
	return r.Value
}
