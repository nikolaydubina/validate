# validate

No reflection. No gencode. Hierarchical and extendable. Fairly fast. 100LOC. Generics.

This is convenient when you have custom validation and nested structures.  
Your type has to satisfy `Validate() error` interface, and you are good to go!

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

Everything to format error, we already have in validators.
Thus, reusing it as Errors.

## Benchmarks

```
$ go test -timeout=1h -bench=. -benchtime=10s -benchmem ./...
goos: darwin
goarch: amd64
pkg: github.com/nikolaydubina/validate
cpu: VirtualApple @ 2.50GHz
BenchmarkEmployee_Error_Ignore-10    	 5050152	      2339 ns/op	    1920 B/op	      71 allocs/op
BenchmarkEmployee_Error_Use-10       	 3125104	      3827 ns/op	    2657 B/op	      94 allocs/op
BenchmarkEmployee_Success-10         	32163120	       373 ns/op	     376 B/op	      15 allocs/op
PASS
ok  	github.com/nikolaydubina/validate	42.913s
```