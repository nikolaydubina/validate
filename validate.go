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
			errs = append(errs, err)
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

type MinLen[T any] struct {
	Value  []T
	MinLen int
}

func (s MinLen[T]) Validate() error {
	if len(s.Value) < s.MinLen {
		return s
	}
	return nil
}

func (s MinLen[T]) Error() string {
	return fmt.Sprintf("len(%d) is smaller than min len(%d)", len(s.Value), s.MinLen)
}

type MaxLen[T any] struct {
	Value  []T
	MaxLen int
}

func (s MaxLen[T]) Validate() error {
	if len(s.Value) > s.MaxLen {
		return s
	}
	return nil
}

func (s MaxLen[T]) Error() string {
	return fmt.Sprintf("len(%d) is higher than max len(%d)", len(s.Value), s.MaxLen)
}

type MinMaxLen[T any] struct {
	Value  []T
	MinLen int
	MaxLen int
}

func (s MinMaxLen[T]) Validate() error {
	return All(
		MinLen[T]{Value: s.Value, MinLen: s.MinLen},
		MaxLen[T]{Value: s.Value, MaxLen: s.MaxLen},
	)
}

type MinLenMap[K comparable, V any] struct {
	Value  map[K]V
	MinLen int
}

func (s MinLenMap[T, K]) Validate() error {
	if len(s.Value) < s.MinLen {
		return s
	}
	return nil
}

func (s MinLenMap[T, K]) Error() string {
	return fmt.Sprintf("len(%d) is smaller than min len(%d)", len(s.Value), s.MinLen)
}

type MaxLenMap[K comparable, V any] struct {
	Value  map[K]V
	MaxLen int
}

func (s MaxLenMap[K, V]) Validate() error {
	if len(s.Value) > s.MaxLen {
		return s
	}
	return nil
}

func (s MaxLenMap[K, V]) Error() string {
	return fmt.Sprintf("len(%d) is higher than max len(%d)", len(s.Value), s.MaxLen)
}

type MinMaxLenMap[K comparable, V any] struct {
	Value  map[K]V
	MinLen int
	MaxLen int
}

func (s MinMaxLenMap[K, V]) Validate() error {
	return All(
		MinLenMap[K, V]{Value: s.Value, MinLen: s.MinLen},
		MinMaxLenMap[K, V]{Value: s.Value, MaxLen: s.MaxLen},
	)
}

type MinLenStr struct {
	Value  string
	MinLen int
}

func (s MinLenStr) Validate() error {
	if len(s.Value) < s.MinLen {
		return s
	}
	return nil
}

func (s MinLenStr) Error() string {
	return fmt.Sprintf("len(%d) is smaller than min len(%d)", len(s.Value), s.MinLen)
}

type MaxLenStr struct {
	Value  string
	MaxLen int
}

func (s MaxLenStr) Validate() error {
	if len(s.Value) > s.MaxLen {
		return s
	}
	return nil
}

func (s MaxLenStr) Error() string {
	return fmt.Sprintf("len(%d) is higher than max len(%d)", len(s.Value), s.MaxLen)
}

type MinMaxLenStr struct {
	Value  string
	MinLen int
	MaxLen int
}

func (s MinMaxLenStr) Validate() error {
	return All(
		MinLenStr{Value: s.Value, MinLen: s.MinLen},
		MinMaxLenStr{Value: s.Value, MaxLen: s.MaxLen},
	)
}
