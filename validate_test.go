package validate_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/nikolaydubina/validate"
)

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

func TestEmployee_Error(t *testing.T) {
	tests := []struct {
		e   Employee
		err error
	}{
		{
			e: Employee{
				Name:  "Bob",
				Age:   101,
				Color: "orange",
				Education: Education{
					Duration:   75,
					SchoolName: "Berkeley",
				},
				Salary: 256.99,
			},
			err: errors.New("(Bob) is not in [Zeus Hera];(101) is not in [35 55];(101) higher than max (100);wrong value(orange), expected([red green blue]);(Berkeley) is not in [KAIST Stanford];(256.99) higher than max (123.456)"),
		},
	}
	for i, tc := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			err := tc.e.Validate()
			if tc.err.Error() != err.Error() {
				t.Errorf("\ngot: %s\nexpected: %s", err.Error(), tc.err.Error())
			}
		})
	}
}

func TestEmployee_Success(t *testing.T) {
	tests := []Employee{
		{
			Name:  "Hera",
			Age:   55,
			Color: "red",
			Education: Education{
				Duration:   75,
				SchoolName: "KAIST",
			},
			Salary: 79,
		},
	}
	for i, tc := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			err := tc.Validate()
			if err != nil {
				t.Errorf("got error %v", err)
			}
		})
	}
}

func BenchmarkEmployee_Error_Ignore(b *testing.B) {
	e := Employee{
		Name:  "Bob",
		Age:   101,
		Color: "orange",
		Education: Education{
			Duration:   75,
			SchoolName: "Berkeley",
		},
		Salary: 256.99,
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if e.Validate() == nil {
			b.FailNow()
		}
	}
}

func BenchmarkEmployee_Error_Use(b *testing.B) {
	e := Employee{
		Name:  "Bob",
		Age:   101,
		Color: "orange",
		Education: Education{
			Duration:   75,
			SchoolName: "Berkeley",
		},
		Salary: 256.99,
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		msg := e.Validate().Error()
		if msg == "" {
			b.FailNow()
		}
	}
}

func BenchmarkEmployee_Success(b *testing.B) {
	e := Employee{
		Name:  "Hera",
		Age:   55,
		Color: "red",
		Education: Education{
			Duration:   75,
			SchoolName: "KAIST",
		},
		Salary: 79,
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if e.Validate() != nil {
			b.FailNow()
		}
	}
}

type EmployeeSimple struct {
	Name   string
	Age    int
	Salary float64
}

func (s EmployeeSimple) Validate() error {
	return validate.All(
		validate.OneOf[string]{Value: s.Name, Values: []string{"Zeus", "Hera"}},
		validate.OneOf[int]{Value: s.Age, Values: []int{35, 55}},
		validate.MinMax[int]{Value: s.Age, Min: 10, Max: 100},
		validate.MinMax[float64]{Value: s.Salary, Min: -10, Max: 123.456},
	)
}

func BenchmarkEmployeeSimple_Error_Ignore(b *testing.B) {
	e := EmployeeSimple{
		Name:   "Bob",
		Age:    101,
		Salary: 256.99,
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if e.Validate() == nil {
			b.FailNow()
		}
	}
}

func BenchmarkEmployeeSimple_Error_Use(b *testing.B) {
	e := EmployeeSimple{
		Name:   "Bob",
		Age:    101,
		Salary: 256.99,
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		msg := e.Validate().Error()
		if msg == "" {
			b.FailNow()
		}
	}
}

func BenchmarkEmployeeSimple_Success(b *testing.B) {
	e := EmployeeSimple{
		Name:   "Hera",
		Age:    55,
		Salary: 79,
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if e.Validate() != nil {
			b.FailNow()
		}
	}
}
