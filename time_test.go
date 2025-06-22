package gofinance

import (
	"testing"
	"time"
)

// mid is re-implemented in the test so we can compute the *independent* oracle.
func mid(start, end time.Time) time.Time {
	return start.Add(end.Sub(start) / 2)
}

// -----------------------------------------------------------------------------
// Single-period cases
// -----------------------------------------------------------------------------
func TestStringToTime_SinglePeriod(t *testing.T) {
	t.Parallel()

	loc := time.UTC // all results must be forced into UTC

	cases := []struct {
		name   string
		input  string
		expect func() time.Time
	}{
		{
			"year-non-leap",
			"2021",
			func() time.Time {
				s := time.Date(2021, 1, 1, 0, 0, 0, 0, loc)
				e := time.Date(2021, 12, 31, 23, 59, 59, 999_999_999, loc)
				return mid(s, e)
			},
		},
		{
			"year-leap",
			"2020",
			func() time.Time {
				s := time.Date(2020, 1, 1, 0, 0, 0, 0, loc)
				e := time.Date(2020, 12, 31, 23, 59, 59, 999_999_999, loc)
				return mid(s, e)
			},
		},
		{
			"month-iso",
			"2021-05",
			func() time.Time {
				s := time.Date(2021, 5, 1, 0, 0, 0, 0, loc)
				e := time.Date(2021, 6, 0, 23, 59, 59, 999_999_999, loc) // day 0 == last of prev month
				return mid(s, e)
			},
		},
		{
			"month-slash-sep",
			"2021/06",
			func() time.Time {
				s := time.Date(2021, 6, 1, 0, 0, 0, 0, loc)
				e := time.Date(2021, 7, 0, 23, 59, 59, 999_999_999, loc)
				return mid(s, e)
			},
		},
		{
			"day-dot-sep",
			"2021.07.04",
			func() time.Time {
				s := time.Date(2021, 7, 4, 0, 0, 0, 0, loc)
				e := s.Add(24*time.Hour - time.Nanosecond)
				return mid(s, e)
			},
		},
		{
			"hour",
			"2021-07-04 15",
			func() time.Time {
				s := time.Date(2021, 7, 4, 15, 0, 0, 0, loc).Truncate(time.Hour)
				e := s.Add(time.Hour - time.Nanosecond)
				return mid(s, e)
			},
		},
		{
			"minute",
			"2021-07-04 15:30",
			func() time.Time {
				s := time.Date(2021, 7, 4, 15, 30, 0, 0, loc).Truncate(time.Minute)
				e := s.Add(time.Minute - time.Nanosecond)
				return mid(s, e)
			},
		},
		{
			"second",
			"2021-07-04 15:30:45",
			func() time.Time {
				s := time.Date(2021, 7, 4, 15, 30, 45, 0, loc).Truncate(time.Second)
				e := s.Add(time.Second - time.Nanosecond)
				return mid(s, e)
			},
		},
		{
			"millisecond",
			"2021-07-04 15:30:45.123",
			func() time.Time {
				s := time.Date(2021, 7, 4, 15, 30, 45, 123_000_000, loc).Truncate(time.Millisecond)
				e := s.Add(time.Millisecond - time.Nanosecond)
				return mid(s, e)
			},
		},
	}

	for _, caseStruct := range cases {
		caseStruct := caseStruct // capture
		t.Run(caseStruct.name, func(t *testing.T) {
			t.Parallel()
			got, err := StringToTime(caseStruct.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			want := caseStruct.expect()
			if !got.Equal(want) {
				t.Errorf("StringToTime(%q)\n  got : %v\n  want: %v", caseStruct.input, got, want)
			}
			if got.Location() != loc {
				t.Errorf("result not in UTC: loc=%v", got.Location())
			}
		})
	}
}

// -----------------------------------------------------------------------------
// Two-period cases (commutativity & mixed resolutions)
// -----------------------------------------------------------------------------
func TestStringToTime_TwoPeriods(t *testing.T) {
	t.Parallel()

	check := func(input1, input2 string) {
		loc := time.UTC

		gotInput1, _ := StringToTime(input1)
		gotInput2, _ := StringToTime(input2)
		wantInput1Input2 := mid(gotInput1, gotInput2)

		gotInput1Input2, err := StringToTime(input1, input2)
		if err != nil {
			t.Fatalf("StringToTime(%q,%q) unexpected error: %v", input1, input2, err)
		}
		if !gotInput1Input2.Equal(wantInput1Input2) {
			t.Errorf("StringToTime(%q,%q)\n  got : %v\n  want: %v",
				input1, input2, gotInput1Input2, wantInput1Input2)
		}

		// commutativity
		gotInput2Input1, _ := StringToTime(input2, input1)
		if !gotInput2Input1.Equal(gotInput1Input2) {
			t.Errorf("StringToTime should be commutative.\n  forward : %v\n  reverse : %v",
				gotInput1Input2, gotInput2Input1)
		}

		if gotInput1Input2.Location() != loc {
			t.Errorf("result not in UTC: loc=%v", gotInput1Input2.Location())
		}
	}

	check("2021", "2023-12")
	check("2020-02", "2020-03-15 12:00:00")
}

// -----------------------------------------------------------------------------
// Error paths
// -----------------------------------------------------------------------------
func TestStringToTime_Errors(t *testing.T) {
	t.Parallel()

	errCases := [][]string{
		{},                         // 0 argumnets
		{"too", "many", "strings"}, // >2 arguments
		{"21-01-01"},               // unsupported format
	}

	for _, errCase := range errCases {
		errCase := errCase
		t.Run(
			"error_"+func() string {
				if len(errCase) == 0 {
					return "0arguments"
				}
				return errCase[0]
			}(),
			func(t *testing.T) {
				t.Parallel()

				if _, err := StringToTime(errCase...); err == nil {
					t.Fatalf("expected error for input %v, got nil", errCase)
				}
			})
	}
}
