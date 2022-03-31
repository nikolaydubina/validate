package validate

import (
	"fmt"
	"strings"

	"golang.org/x/exp/constraints"
)

type Validatable interface {
	Validate() error
}

type ValidationError struct {
	Errors []error
}

func (e ValidationError) Error() string {
	if len(e.Errors) == 0 {
		return ""
	}
	msg := make([]string, 0, len(e.Errors))
	for _, q := range e.Errors {
		msg = append(msg, q.Error())
	}
	return strings.Join(msg, ";")
}

func All(vs ...Validatable) error {
	var errs []error
	for _, f := range vs {
		if err := f.Validate(); err != nil {
			errs = append(errs, f.Validate())
		}
	}
	if len(errs) > 0 {
		return ValidationError{Errors: errs}
	}
	return nil
}

type Min[T constraints.Ordered] struct {
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
	return fmt.Sprintf("(%v) smaller than min (%v)", s.Value, s.Min)
}

type Max[T constraints.Ordered] struct {
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
	return fmt.Sprintf("(%v) higher than max (%v)", s.Value, s.Max)
}

type MinMax[T constraints.Ordered] struct {
	Value T
	Min   T
	Max   T
}

func (s MinMax[T]) Validate() error {
	return All(
		Max[T]{Value: s.Value, Max: s.Max},
		Min[T]{Value: s.Value, Min: s.Min},
	)
}

type OneOf[T comparable] struct {
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
	return fmt.Sprintf("(%v) is not in %v", s.Value, s.Values)
}
