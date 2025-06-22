package gofinance

import (
	"errors"
	"time"
)

// resolution indicates the granularity of the parsed input: year, month, day, etc.
type resolution int

const (
	yearRes resolution = iota
	monthRes
	dayRes
	hourRes
	minuteRes
	secondRes
	milliRes
)

// resolutionToLayouts maps a resolution to a slice of layouts that fit that resolution.
type resolutionToLayouts struct {
	resolution resolution
	layouts    []string
}

// sliceResolutionToLayouts is a slice of resolutionToLayouts.
// It implements multiple resolutions and provides supported layouts for each resolution.
var sliceResolutionToLayouts []resolutionToLayouts = []resolutionToLayouts{
	{
		milliRes,
		[]string{
			"2006-01-02 15:04:05.000",
			"2006/01/02 15:04:05.000",
			"2006.01.02 15:04:05.000",
		},
	},
	{
		secondRes,
		[]string{
			"2006-01-02 15:04:05",
			"2006/01/02 15:04:05",
			"2006.01.02 15:04:05",
		},
	},
	{
		minuteRes,
		[]string{
			"2006-01-02 15:04",
			"2006/01/02 15:04",
			"2006.01.02 15:04",
		},
	},
	{
		hourRes,
		[]string{
			"2006-01-02 15",
			"2006/01/02 15",
			"2006.01.02 15",
		},
	},
	{
		dayRes,
		[]string{
			"2006-01-02",
			"2006/01/02",
			"2006.01.02",
		},
	},
	{
		monthRes,
		[]string{
			"2006-01",
			"2006/01",
			"2006.01",
		},
	},
	{
		yearRes,
		[]string{
			"2006",
		},
	},
}

// mid returns the midpoint between start and end.
func midOfStartEnd(start, end time.Time) time.Time {
	return start.Add(end.Sub(start) / 2)
}

// parseStringToStartEnd returns the mid of the time period represented by inputString.
// UTC location is forced.
func parseStringToMidTime(input string) (mid time.Time, err error) {
	location := time.UTC
	for _, resolutionToLayouts := range sliceResolutionToLayouts {
		for _, layout := range resolutionToLayouts.layouts {
			timeParsed, error := time.ParseInLocation(layout, input, location)
			if error == nil {
				switch resolutionToLayouts.resolution {
				case yearRes:
					y := timeParsed.Year()
					start := time.Date(y, 1, 1, 0, 0, 0, 0, location)
					end := time.Date(y, 12, 31, 23, 59, 59, 999999999, location)
					mid = midOfStartEnd(start, end)
				case monthRes:
					y, m, _ := timeParsed.Date()
					start := time.Date(y, m, 1, 0, 0, 0, 0, location)
					end := time.Date(y, m+1, 0, 23, 59, 59, 999999999, location)
					mid = midOfStartEnd(start, end)
				case dayRes:
					y, m, d := timeParsed.Date()
					start := time.Date(y, m, d, 0, 0, 0, 0, location)
					end := start.Add(24*time.Hour - time.Nanosecond)
					mid = midOfStartEnd(start, end)
				case hourRes:
					start := timeParsed.Truncate(time.Hour)
					end := start.Add(time.Hour - time.Nanosecond)
					mid = midOfStartEnd(start, end)
				case minuteRes:
					start := timeParsed.Truncate(time.Minute)
					end := start.Add(time.Minute - time.Nanosecond)
					mid = midOfStartEnd(start, end)
				case secondRes:
					start := timeParsed.Truncate(time.Second)
					end := start.Add(time.Second - time.Nanosecond)
					mid = midOfStartEnd(start, end)
				case milliRes:
					start := timeParsed.Truncate(time.Millisecond)
					end := start.Add(time.Millisecond - time.Nanosecond)
					mid = midOfStartEnd(start, end)
				}
				return
			}
		}
	}
	err = errors.New("unsupported time format: " + input)
	return
}

// Midpoint returns the midpoint of the provided period strings.
//
// If one string is supplied, it interprets that string as a period and returns
// the midpoint of that period.
//
// If two strings are supplied, it interprets each as separate periods and
// returns the midpoint between the two period midpoints.
//
// Supported formats:
//
//   - YYYY
//   - YYYY-MM, YYYY/MM, YYYY.MM
//   - YYYY-MM-DD, YYYY/MM/DD, YYYY.MM.DD
//   - YYYY-MM-DD HH, YYYY/MM/DD HH, YYYY.MM.DD HH
//   - YYYY-MM-DD HH:MM, YYYY/MM/DD HH:MM, YYYY.MM.DD HH:MM
//   - YYYY-MM-DD HH:MM:SS, YYYY/MM/DD HH:MM:SS, YYYY.MM.DD HH:MM:SS
//   - YYYY-MM-DD HH:MM:SS.mmm, YYYY/MM/DD HH:MM:SS.mmm, YYYY.MM.DD HH:MM:SS.mmm
//
// The function always returns a time in UTC, so UTC location is forced.
func StringToTime(periods ...string) (time.Time, error) {
	if len(periods) == 0 || len(periods) > 2 {
		return time.Time{}, errors.New("StringToTime requires 1 or 2 period strings")
	} else if len(periods) == 1 {
		mid, err := parseStringToMidTime(periods[0])
		if err != nil {
			return time.Time{}, err
		}
		return mid, nil
	} else if len(periods) == 2 {
		mid1, err1 := parseStringToMidTime(periods[0])
		if err1 != nil {
			return time.Time{}, err1
		}
		mid2, err2 := parseStringToMidTime(periods[1])
		if err2 != nil {
			return time.Time{}, err2
		}
		if mid1.Before(mid2) {
			return midOfStartEnd(mid1, mid2), nil
		} else {
			return midOfStartEnd(mid2, mid1), nil
		}
	}
	return time.Time{}, errors.New("unknown error")
}
