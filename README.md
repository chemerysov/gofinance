# gofinance
goal: the comprehensive pure Go finance library

## current status
supports:

interest rates: annual percentage rate, effective annual rate, continuously compunded rate,
conversion between interest rates, discount factors

time: user input of time in an approximate format, for example only the year, which takes the midpoint of the year, or a time range in which cash flow was occuring

## getting started
run the following commands:

`go get github.com/chemerysov/gofinance@latest`

`go mod tidy`

in your code, add line:

`import "github.com/chemerysov/gofinance"`
