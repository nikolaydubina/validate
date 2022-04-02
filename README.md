# 🥬 validate. simply.

> no reflection. no gencode. hierarchical and extendable. fast. ~100LOC. generics.

[![codecov](https://codecov.io/gh/nikolaydubina/validate/branch/main/graph/badge.svg?token=76JC6fX7DP)](https://codecov.io/gh/nikolaydubina/validate)
[![Go Reference](https://pkg.go.dev/badge/github.com/nikolaydubina/validate.svg)](https://pkg.go.dev/github.com/nikolaydubina/validate)
[![Go Report Card](https://goreportcard.com/badge/github.com/nikolaydubina/validate)](https://goreportcard.com/report/github.com/nikolaydubina/validate)

This is convenient when you have custom validation and nested structures.  

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
		validate.OneOf("name", s.Name, "Zeus", "Hera"),
		validate.OneOf("age", s.Age, 35, 55),
		validate.Min("age", s.Age, 10), // same field validated again
		s.Color.Validate(),
		s.Education.Validate(),
		validate.Max("salary", s.Salary, 123.456),
		validate.Max("duration", s.Experience, time.Duration(1)*time.Hour),
		validate.After("birthday", s.Birthday, time.Date(1984, 1, 1, 0, 0, 0, 0, time.UTC)),
		validate.Before("vacation_start", s.VacationStart, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
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
		validate.Min("", e.Duration, 10),
		validate.OneOf("", e.SchoolName, "KAIST", "Stanford"),
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
validate: 8 errors: [name(Bob) not in [Zeus Hera]; age(101) not in [35 55]; color wrong value(orange), expected([red green blue]); validate: 1 errors: [(Berkeley) not in [KAIST Stanford]]; salary(256.99) higher than max (123.456); duration(10h0m0s) higher than max (1h0m0s); birthday(1984-01-01 00:00:00 +0000 UTC) is not after (1984-01-01 00:00:00 +0000 UTC); vacation_start(2025-01-01 00:00:00 +0000 UTC) is not before (2024-01-01 00:00:00 +0000 UTC)]
```

## Implementation Details

Printing error takes a lot of time. 
Thus, it is good to delay constructor of error message as much as possible.
And sometimes user code does not need to print error at all and only `nil` check is performed.
This is done by moving construction of error message in `Error` methods.

It is advisable to avoid memory allocations and creation of structures.
Such in case of success flow, we ideally will not have any memory allocations at all.
This is why we make validators as functions and call them in chain.
We do not delay nor wrap validation function calls.
We use function arguments as storage for validation parameters, they are simple params and likely to be on stack which is fast.
For example, for `OneOf` we are using variadic arguments.
Other alternative is to use arrays since in Go they are on stack as well.

We also hope Go compiler
- can detect that argument to function is constant and inline it in assembly or stack
- does not use expensive memory for variadic parameters
- can inline functions

Defining custom validators with `switch` is expected to be even faster.

## Benchmarks

```
$ go test -bench=. -benchtime=10s -benchmem ./...
goos: darwin
goarch: amd64
pkg: github.com/nikolaydubina/validate
cpu: VirtualApple @ 2.50GHz
BenchmarkEmployee_Error_Message-10                	   3744121	      3229 ns/op	    2376 B/op	      56 allocs/op
BenchmarkEmployee_Error-10                        	  12533948	       958 ns/op	     904 B/op	      23 allocs/op
BenchmarkEmployee_Success-10                      	 100000000	       115 ns/op	      80 B/op	       3 allocs/op
BenchmarkEmployeeSimple_Error_Message-10          	   9488436	      1263 ns/op	     840 B/op	      25 allocs/op
BenchmarkEmployeeSimple_Error-10                  	  44261380	       270 ns/op	     344 B/op	       9 allocs/op
BenchmarkEmployeeSimple_Success-10                	 243491635	        49 ns/op	      48 B/op	       2 allocs/op
BenchmarkEmployeeNoContainers_Error_Message-10    	  28089966	       427 ns/op	     248 B/op	       9 allocs/op
BenchmarkEmployeeNoContainers_Error-10            	 142881793	        85 ns/op	      88 B/op	       3 allocs/op
BenchmarkEmployeeNoContainers_Success-10        	1000000000	         4 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	github.com/nikolaydubina/validate	120.565s
```

## Appendix A: Comparison to other validators

#### `github.com/go-playground/validator`

It uses struct tags and reflection.
Binding custom validations require defining validation function with special name and using interface typecast then registering this to validator instance.

It has instance of validator that is reused.

Its speed is mostly few hundred ns and up to 1µs.
Its memory allocation can be 0 and reaches up to few dozen.

## Appendix B: Wrapping validators into interface

Early version of this library was wrapping each validation operation into a `interface { Validate() error }`.
In this approach, we already had in validators everything needed to format error message, which is why we were reusing them as error containers.
However, there were few drawbacks.

Code looked more verbose:
```go
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
```

Performance was slightly worse for error case, and much worse for success case:
```
$ go test -bench=. -benchtime=10s -benchmem ./...
goos: darwin
goarch: amd64
pkg: github.com/nikolaydubina/validate
cpu: VirtualApple @ 2.50GHz
BenchmarkEmployee_Error_Message-10                	 3579223	      3379 ns/op	    2761 B/op	      62 allocs/op
BenchmarkEmployee_Error-10                        	 9361948	      1277 ns/op	    1344 B/op	      34 allocs/op
BenchmarkEmployee_Success-10                      	25418672	       474 ns/op	     552 B/op	      14 allocs/op
BenchmarkEmployeeSimple_Error_Message-10          	 8757170	      1364 ns/op	     992 B/op	      28 allocs/op
BenchmarkEmployeeSimple_Error-10                  	30418941	       394 ns/op	     504 B/op	      13 allocs/op
BenchmarkEmployeeSimple_Success-10                	65194581	       184 ns/op	     224 B/op	       6 allocs/op
BenchmarkEmployeeNoContainers_Error_Message-10    	24971338	       483 ns/op	     280 B/op	      10 allocs/op
BenchmarkEmployeeNoContainers_Error-10            	72736639	       165 ns/op	     136 B/op	       5 allocs/op
BenchmarkEmployeeNoContainers_Success-10          	143333276	        83 ns/op	      64 B/op	       2 allocs/op
PASS
ok  	github.com/nikolaydubina/validate	124.950s
```

## Appendix C: Binding validator functions in map to field names

It might appear that it is more efficient not to pass `name` of field in validator.
Such it is tempting to run slice or map of functions.
However, performance deteriorates with this approach.
Likely this is due to compiler using stack or not efficiently inlining.

Code sample:
```go
func All(vs map[string]error) error {
	errs := make(map[string]error, len(vs))
	for k, err := range vs {
		if err != nil {
			errs[k] = err
		}
	}
	if len(errs) > 0 {
		return errMultiple(errs)
	}
	return nil
}

...

func (s Employee) Validate() error {
	return validate.All(map[string]error{
		"name":           validate.OneOf(s.Name, "Zeus", "Hera"),
		"age":            validate.OneOf(s.Age, 35, 55),
		"age_2":          validate.Min(s.Age, 10), // same field validated again
		"color":          s.Color.Validate(),
		"education":      s.Education.Validate(),
		"salary":         validate.Max(s.Salary, 123.456),
		"duration":       validate.Max(s.Experience, time.Duration(1)*time.Hour),
		"birthday":       validate.After(s.Birthday, time.Date(1984, 1, 1, 0, 0, 0, 0, time.UTC)),
		"vacation_start": validate.Before(s.VacationStart, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
	})
}
```

Performance is worse across the board:
```
goos: darwin
goarch: amd64
pkg: github.com/nikolaydubina/validate
cpu: VirtualApple @ 2.50GHz
BenchmarkEmployee_Error_Message-10                	 2698065	      4359 ns/op	    3866 B/op	      57 allocs/op
BenchmarkEmployee_Error-10                        	 6993564	      1714 ns/op	    2058 B/op	      21 allocs/op
BenchmarkEmployee_Success-10                      	14948445	       810 ns/op	    1329 B/op	       7 allocs/op
BenchmarkEmployeeSimple_Error_Message-10          	 8243392	      1460 ns/op	    1112 B/op	      26 allocs/op
BenchmarkEmployeeSimple_Error-10                  	33798837	       356 ns/op	     496 B/op	       7 allocs/op
BenchmarkEmployeeSimple_Success-10                	66953932	       182 ns/op	      96 B/op	       3 allocs/op
BenchmarkEmployeeNoContainers_Error_Message-10    	19754359	       600 ns/op	     576 B/op	      10 allocs/op
BenchmarkEmployeeNoContainers_Error-10            	61285670	       194 ns/op	     368 B/op	       3 allocs/op
BenchmarkEmployeeNoContainers_Success-10          	123293473	        98 ns/op	      48 B/op	       1 allocs/op
PASS
ok  	github.com/nikolaydubina/validate	128.263s
```

## Reference

- As of `2022-04-01`, Go does not support generic arrays.
