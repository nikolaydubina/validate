package validate

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/constraints"
)

type Validatable interface {
	Validate() error
}

type ValidationError struct {
	Errors []error
}

func (e ValidationError) Error() string {
	msg := make([]string, 0, len(e.Errors))
	for _, q := range e.Errors {
		msg = append(msg, q.Error())
	}
	return "validate: " + strconv.Itoa(len(e.Errors)) + " errors: [" + strings.Join(msg, "; ") + "]"
}

func All(vs ...Validatable) error {
	var errs []error
	for _, f := range vs {
		if err := f.Validate(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return ValidationError{Errors: errs}
	}
	return nil
}

type Min[T constraints.Ordered] struct {
	Name  string
	Value T
	Min   T
}

func (s Min[T]) Validate() error {
	if s.Value < s.Min {
		return s
	}
	return nil
}

func (s Min[T]) Error() string {
	return fmt.Sprintf("%s(%v) smaller than min(%v)", s.Name, s.Value, s.Min)
}

type Max[T constraints.Ordered] struct {
	Name  string
	Value T
	Max   T
}

func (s Max[T]) Validate() error {
	if s.Value > s.Max {
		return s
	}
	return nil
}

func (s Max[T]) Error() string {
	return fmt.Sprintf("%s(%v) higher than max(%v)", s.Name, s.Value, s.Max)
}

type OneOf[T comparable] struct {
	Name   string
	Value  T
	Values []T
}

func (s OneOf[T]) Validate() error {
	for _, q := range s.Values {
		if q == s.Value {
			return nil
		}
	}
	return s
}

func (s OneOf[T]) Error() string {
	return fmt.Sprintf("%s(%v) is not in %v", s.Name, s.Value, s.Values)
}

type Before struct {
	Name  string
	Value time.Time
	Time  time.Time
}

func (s Before) Validate() error {
	if !s.Value.Before(s.Time) {
		return s
	}
	return nil
}

func (s Before) Error() string {
	return fmt.Sprintf("%s(%v) is not before (%v)", s.Name, s.Value, s.Time)
}

type After struct {
	Name  string
	Value time.Time
	Time  time.Time
}

func (s After) Validate() error {
	if !s.Value.After(s.Time) {
		return s
	}
	return nil
}

func (s After) Error() string {
	return fmt.Sprintf("%s(%v) is not after (%v)", s.Name, s.Value, s.Time)
}
