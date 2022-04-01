# 🥬 validate. simply.

> no reflection. no gencode. hierarchical and extendable. fairly fast. ~100LOC. generics.

[![codecov](https://codecov.io/gh/nikolaydubina/validate/branch/main/graph/badge.svg?token=76JC6fX7DP)](https://codecov.io/gh/nikolaydubina/validate)
[![Go Reference](https://pkg.go.dev/badge/github.com/nikolaydubina/validate.svg)](https://pkg.go.dev/github.com/nikolaydubina/validate)
[![Go Report Card](https://goreportcard.com/badge/github.com/nikolaydubina/validate)](https://goreportcard.com/report/github.com/nikolaydubina/validate)

This is convenient when you have custom validation and nested structures.  
Your type has to satisfy `Validate() error` interface and you are good to go!

```go
// Employee is example of struct with validatable fields and nested structure
type Employee struct {
	Name          string
	Age           int
	Color         Color     // custom func Validate()
	Education     Education // nested with Validate()
	Salary        float64
	Experience    time.Duration
	Birthday      time.Time
	VacationStart time.Time
}

func (s Employee) Validate() error {
	return validate.All(
		validate.OneOf[string]{Name: "name", Value: s.Name, Values: []string{"Zeus", "Hera"}},
		validate.OneOf[int]{Name: "age", Value: s.Age, Values: []int{35, 55}},
		validate.Min[int]{Name: "age", Value: s.Age, Min: 10}, // same field validated again
		s.Color,
		s.Education,
		validate.Max[float64]{Name: "salary", Value: s.Salary, Max: 123.456},
		validate.Max[time.Duration]{Name: "duration", Value: s.Experience, Max: time.Duration(1) * time.Hour},
		validate.After{Name: "birthday", Value: s.Birthday, Time: time.Date(1984, 1, 1, 0, 0, 0, 0, time.UTC)},
		validate.Before{Name: "vacation_start", Value: s.VacationStart, Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
	)
}

// Education is another custom struct
type Education struct {
	Duration   int
	SchoolName string
}

func (e Education) Validate() error {
	if (e.Duration % 17) == 5 {
		return errors.New("my special error")
	}
	return validate.All(
		validate.Min[int]{Value: e.Duration, Min: 10},
		validate.OneOf[string]{Value: e.SchoolName, Values: []string{"KAIST", "Stanford"}},
	)
}

// Color is custom enum
type Color string

const (
	Red   Color = "red"
	Green Color = "green"
	Blue  Color = "blue"
)

func (s Color) Validate() error {
	switch s {
	case Red, Green, Blue:
		return nil
	default:
		return fmt.Errorf("wrong value(%s), expected(%v)", s, []Color{
			"red",
			"green",
			"blue",
		})
	}
}
```

Example error message:
```
name(Bob) is not in [Zeus Hera];age(101) is not in [35 55];color wrong value(orange), expected([red green blue]);(Berkeley) is not in [KAIST Stanford];salary(256.99) higher than max(123.456);duration(10h0m0s) higher than max(1h0m0s);birthday(1984-01-01 00:00:00 +0000 UTC) is not after (1984-01-01 00:00:00 +0000 UTC);vacation_start(2025-01-01 00:00:00 +0000 UTC) is not before (2024-01-01 00:00:00 +0000 UTC)
```

## Implementation Details

It is notable that printing error takes lots of time. 
Thus, it is good to delay constructor of error message as much as possible.
This is done by moving construction of error message in `Error` methods.

We already have in validators everything needed to format error message, which is why reusing them as error containers.

Most of time and memory allocations happen in validators that use containers.
Thus it is advised to avoid `OneOf` and alike.
If possible, define your own validators with arrays and not slices or maps.
Custom validators with switch cases are expected to be even more performant.
As of 2022-04-01, Go does not support generic arrays. Otherwise, we would use arrays.

## Benchmarks

```
$ go test -timeout=1h -bench=. -benchtime=10s -benchmem ./...
goos: darwin
goarch: amd64
pkg: github.com/nikolaydubina/validate
cpu: VirtualApple @ 2.50GHz
BenchmarkEmployee_Error_Message-10                	 3430935	      3500 ns/op	    2376 B/op	      63 allocs/op
BenchmarkEmployee_Error-10                        	 9050138	      1319 ns/op	    1344 B/op	      34 allocs/op
BenchmarkEmployee_Success-10                      	25547962	       482.1 ns/op	     552 B/op	      14 allocs/op
BenchmarkEmployeeSimple_Error_Message-10          	 9065190	      1322 ns/op	     880 B/op	      27 allocs/op
BenchmarkEmployeeSimple_Error-10                  	29836638	       398.7 ns/op	     504 B/op	      13 allocs/op
BenchmarkEmployeeSimple_Success-10                	65272360	       181.7 ns/op	     224 B/op	       6 allocs/op
BenchmarkEmployeeNoContainers_Error_Message-10    	27753966	       432.7 ns/op	     216 B/op	       9 allocs/op
BenchmarkEmployeeNoContainers_Error-10            	73233061	       163.1 ns/op	     136 B/op	       5 allocs/op
BenchmarkEmployeeNoContainers_Success-10          	147265029	        81.30 ns/op	      64 B/op	       2 allocs/op
PASS
ok  	github.com/nikolaydubina/validate	124.499s
```

## Appendix A: Comparison to other validators

#### `github.com/go-playground/validator`

It uses struct tags and reflection.
Binding custom validations require defining validation function with special name and using interface typecast then registering this to validator instance.

It has instance of validator that is reused.

Its speed is mostly few hundred ns and up to 1µs.
Its memory allocation can be 0 and reaches up to few dozen.
It wins in both speed and memory allocation.
