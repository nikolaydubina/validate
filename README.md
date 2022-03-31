# 🥬 validate. simply.

> no reflection. no gencode. hierarchical and extendable. fairly fast. 100LOC. generics.

[![codecov](https://codecov.io/gh/nikolaydubina/validate/branch/main/graph/badge.svg?token=76JC6fX7DP)](https://codecov.io/gh/nikolaydubina/validate)
[![Go Reference](https://pkg.go.dev/badge/github.com/nikolaydubina/validate.svg)](https://pkg.go.dev/github.com/nikolaydubina/validate)
[![Go Report Card](https://goreportcard.com/badge/github.com/nikolaydubina/validate)](https://goreportcard.com/report/github.com/nikolaydubina/validate)

This is convenient when you have custom validation and nested structures.  
Your type has to satisfy `Validate() error` interface and you are good to go!

```go
// Employee is example of struct with validatable fields and nested structure
type Employee struct {
	Name      string
	Age       int
	Color     Color     // custom func Validate()
	Education Education // nested with Validate()
	Salary    float64
}

func (s Employee) Validate() error {
	return validate.All(
		validate.OneOf[string]{Value: s.Name, Values: []string{"Zeus", "Hera"}},
		validate.OneOf[int]{Value: s.Age, Values: []int{35, 55}},
		validate.MinMax[int]{Value: s.Age, Min: 10, Max: 100}, // same field validated again
		s.Color,
		s.Education,
		validate.MinMax[float64]{Value: s.Salary, Min: -10, Max: 123.456},
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
BenchmarkEmployee_Error_Use-10                   	 4381866	      2750 ns/op	    1856 B/op	      62 allocs/op
BenchmarkEmployee_Error_Ignore-10                	 9370138	      1286 ns/op	    1120 B/op	      39 allocs/op
BenchmarkEmployee_Success-10                     	31928192	       374 ns/op	     376 B/op	      15 allocs/op
BenchmarkEmployeeSimple_Error_Use-10             	 7597821	      1580 ns/op	    1056 B/op	      39 allocs/op
BenchmarkEmployeeSimple_Error_Ignore-10          	21584463	       548 ns/op	     568 B/op	      22 allocs/op
BenchmarkEmployeeSimple_Success-10               	51159644	       233 ns/op	     240 B/op	      10 allocs/op
BenchmarkEmployeeNoContainers_Error_Use-10       	14266191	       829 ns/op	     520 B/op	      23 allocs/op
BenchmarkEmployeeNoContainers_Error_Ignore-10    	32474282	       369 ns/op	     296 B/op	      15 allocs/op
BenchmarkEmployeeNoContainers_Success-10         	87722424	       137 ns/op	     112 B/op	       6 allocs/op
PASS
ok  	github.com/nikolaydubina/validate	116.387s
```

## Appendix A: Comparison to other validators

#### `github.com/go-playground/validator`

It uses struct tags and reflection.
Binding custom validations require defining validation function with special name and using interface typecast then registering this to validator instance.

It has instance of validator that is reused.

Its speed is mostly few hundred ns and up to 1µs.
Its memory allocation can be 0 and reaches up to few dozen.
It wins in both speed and memory allocation.
