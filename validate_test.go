package validate_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/nikolaydubina/validate"
)

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
		return fmt.Errorf("color wrong value(%s), expected(%v)", s, []Color{
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
				Salary:        256.99,
				Experience:    time.Duration(10) * time.Hour,
				Birthday:      time.Date(1984, 1, 1, 0, 0, 0, 0, time.UTC),
				VacationStart: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			err: errors.New("validate: 8 errors: [name(Bob) not in [Zeus Hera]; age(101) not in [35 55]; color wrong value(orange), expected([red green blue]); validate: 1 errors: [(Berkeley) not in [KAIST Stanford]]; salary(256.99) higher than max (123.456); duration(10h0m0s) higher than max (1h0m0s); birthday(1984-01-01 00:00:00 +0000 UTC) is not after (1984-01-01 00:00:00 +0000 UTC); vacation_start(2025-01-01 00:00:00 +0000 UTC) is not before (2024-01-01 00:00:00 +0000 UTC)]"),
		},
		{
			e: Employee{
				Name:  "Bob",
				Age:   -10,
				Color: "orange",
				Education: Education{
					Duration:   75,
					SchoolName: "Berkeley",
				},
				Salary: 256.99,
			},
			err: errors.New("validate: 7 errors: [name(Bob) not in [Zeus Hera]; age(-10) not in [35 55]; age(-10) smaller than min (10); color wrong value(orange), expected([red green blue]); validate: 1 errors: [(Berkeley) not in [KAIST Stanford]]; salary(256.99) higher than max (123.456); birthday(0001-01-01 00:00:00 +0000 UTC) is not after (1984-01-01 00:00:00 +0000 UTC)]"),
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
			Salary:   79,
			Birthday: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
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

func BenchmarkEmployee_Error_Message(b *testing.B) {
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
func BenchmarkEmployee_Error(b *testing.B) {
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

func BenchmarkEmployee_Success(b *testing.B) {
	e := Employee{
		Name:  "Hera",
		Age:   55,
		Color: "red",
		Education: Education{
			Duration:   75,
			SchoolName: "KAIST",
		},
		Salary:   79,
		Birthday: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
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
		validate.OneOf("", s.Name, "Zeus", "Hera"),
		validate.OneOf("", s.Age, 35, 55),
		validate.Min("", s.Age, 10),
		validate.Max("", s.Salary, 123.456),
	)
}
func BenchmarkEmployeeSimple_Error_Message(b *testing.B) {
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

func BenchmarkEmployeeSimple_Error(b *testing.B) {
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

type EmployeeNoContainers struct {
	Age    int
	Salary float64
}

func (s EmployeeNoContainers) Validate() error {
	return validate.All(
		validate.Min("", s.Age, 10),
		validate.Max("", s.Salary, 123.456),
	)
}
func BenchmarkEmployeeNoContainers_Error_Message(b *testing.B) {
	e := EmployeeNoContainers{
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

func BenchmarkEmployeeNoContainers_Error(b *testing.B) {
	e := EmployeeNoContainers{
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

func BenchmarkEmployeeNoContainers_Success(b *testing.B) {
	e := EmployeeNoContainers{
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
