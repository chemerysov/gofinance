package gofinance

import "math"

type Rate interface {
	// convert any rate to effective annual rate
	RateAnnualEffective() float64
	// convert any rate to annual precentage rate compounded annually
	RateAnnualPercentageCompoundedAnnually() float64
	// convert any rate to continuous rate
	RateContinuous() float64
	// derive (present value) discount factor based on number of years
	DiscountFactor(years float64) float64
}

// implements Rate
type RateAnnualPercentage struct {
	Value          float64
	PeriodsPerYear float64
}

func (r RateAnnualPercentage) RateAnnualEffective() float64 {
	return math.Pow(1+r.Value/r.PeriodsPerYear, r.PeriodsPerYear) - 1
}

func (r RateAnnualPercentage) RateAnnualPercentageCompoundedAnnually() float64 {
	return r.Value
}

func (r RateAnnualPercentage) RateContinuous() float64 {
	return math.Log(1 + r.RateAnnualEffective())
}

func (r RateAnnualPercentage) DiscountFactor(years float64) float64 {
	return math.Pow(1+r.Value/r.PeriodsPerYear, -r.PeriodsPerYear*years)
}

type RateEffective struct {
	Value          float64
	PeriodsPerYear float64
}

func (r RateEffective) RateAnnualEffective() float64 {
	return math.Pow(1+r.Value, r.PeriodsPerYear) - 1
}

func (r RateEffective) RateAnnualPercentageCompoundedAnnually() float64 {
	return r.RateAnnualEffective()
}

func (r RateEffective) RateContinuous() float64 {
	return math.Log(1 + r.RateAnnualEffective())
}

func (r RateEffective) DiscountFactor(years float64) float64 {
	return math.Pow(1+r.Value, -r.PeriodsPerYear*years)
}

type RateContinuous struct {
	R float64
}

func (c RateContinuous) RateAnnualEffective() float64 {
	return math.Exp(c.R) - 1
}

func (c RateContinuous) RateAnnualNominal() float64 {
	return c.RateAnnualEffective()
}

func (c RateContinuous) RateContinuous() float64 {
	return c.R
}

func (c RateContinuous) DiscountFactor(t float64) float64 {
	return math.Exp(-c.R * t)
}

type RateSimple struct {
	R float64
}

func (s RateSimple) RateAnnualEffective() float64 {
	return s.R
}
func (s RateSimple) RateAnnualNominal() float64 {
	return s.R
}
func (s RateSimple) RateContinuous() float64 {
	return math.Log(1 + s.R)
}
func (s RateSimple) DiscountFactor(t float64) float64 {
	return 1 / (1 + s.R*t)
}
